package core

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/carlosabalde/pgp-tomb/internal/core/config"
	"github.com/carlosabalde/pgp-tomb/internal/core/query"
	"github.com/carlosabalde/pgp-tomb/internal/core/secret"
	"github.com/carlosabalde/pgp-tomb/internal/helpers/pgp"
	"github.com/carlosabalde/pgp-tomb/internal/helpers/slices"
)

func List(folder string, long bool, queryString, keyAlias string) {
	// Initialize query.
	var queryParsed query.Query
	if queryString != "" {
		var err error
		queryParsed, err = query.Parse(queryString)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"query": queryString,
				"error": err,
			}).Fatal("Failed to parse query!")
		}
	} else {
		queryParsed = query.True
	}

	// Initialize key.
	var key *pgp.PublicKey
	if keyAlias != "" {
		key = findPublicKey(keyAlias)
		if key == nil {
			fmt.Fprintln(os.Stderr, "Key does not exist!")
			os.Exit(1)
		}
	} else {
		key = nil
	}

	// Check folder.
	root := path.Join(config.GetSecretsRoot(), folder)
	if folder != "" {
		if info, err := os.Stat(root); os.IsNotExist(err) || !info.IsDir() {
			fmt.Fprintln(os.Stderr, "Folder does not exist!")
			os.Exit(1)
		}
	}

	// Walk file system.
	if err := filepath.Walk(
		root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && filepath.Ext(path) == config.SecretExtension {
				listSecret(path, long, queryParsed, key)
			}
			return nil
		}); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to list secrets!")
	}
}

func listSecret(path string, long bool, q query.Query, key *pgp.PublicKey) {
	uri := strings.TrimPrefix(path, config.GetSecretsRoot())
	uri = strings.TrimPrefix(uri, string(os.PathSeparator))
	uri = strings.TrimSuffix(uri, config.SecretExtension)

	s, err := secret.Load(uri)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"uri":   s.GetUri(),
		}).Error("Failed to load secret!")
		return
	}

	if !q.Eval(s) {
		return
	}

	if key != nil {
		found := false
		if keys, err := s.GetExpectedPublicKeys(); err == nil {
			for _, aKey := range keys {
				if aKey == key {
					found = true
					break
				}
			}
		} else {
			logrus.WithFields(logrus.Fields{
				"error": err,
				"uri":   s.GetUri(),
			}).Error("Failed to get expected public keys!")
			return
		}
		if !found {
			return
		}
	}

	fmt.Printf("- %s\n", s.GetUri())
	if long {
		renderSecretDetails(s)
	}
}

func renderSecretDetails(s *secret.Secret) {
	// Extract current recipients.
	currentRecipientKeyIds, err := s.GetCurrentRecipientsKeyIds()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"uri":   s.GetUri(),
		}).Error("Failed to determine current recipients!")
		return
	}

	// Determine current & unknown recipients.
	currentAliases := make([]string, 0)
	unknownRecipients := make([]string, 0)
	for _, keyId := range currentRecipientKeyIds {
		key := findPublicKeyByKeyId(keyId)
		if key != nil {
			currentAliases = append(currentAliases, key.Alias)
		} else {
			unknownRecipients = append(
				unknownRecipients,
				fmt.Sprintf("0x%x", keyId))
		}
	}

	// Determine expected recipients.
	expectedAliases := make([]string, 0)
	if keys, err := s.GetExpectedPublicKeys(); err == nil {
		for _, key := range keys {
			expectedAliases = append(expectedAliases, key.Alias)
		}
	} else {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"uri":   s.GetUri(),
		}).Error("Failed to determine expected recipients!")
		return
	}

	// Render expected recipients.
	sort.Strings(expectedAliases)
	fmt.Printf(
		"  + recipients: %s\n",
		strings.Join(expectedAliases, ", "))

	// Render unknown recipients?
	if len(unknownRecipients) > 0 {
		sort.Strings(unknownRecipients)
		fmt.Printf(
			"    * unknown : %s\n",
			strings.Join(unknownRecipients, ", "))
	}

	// Render rubbish recipients?
	tmpCurrentAliases, errCurrentAliases := slices.Difference(
		currentAliases, expectedAliases)
	if errCurrentAliases != nil {
		logrus.WithFields(logrus.Fields{
			"error": errCurrentAliases,
			"uri":   s.GetUri(),
		}).Error("Failed to determine rubbish recipients!")
		return
	}
	rubbishRecipients := tmpCurrentAliases.Interface().([]string)
	if len(rubbishRecipients) > 0 {
		sort.Strings(rubbishRecipients)
		fmt.Printf(
			"    * rubbish: %s\n",
			strings.Join(rubbishRecipients, ", "))
	}

	// Render missing recipients?
	tmpExpectedAliases, errExpectedAliases := slices.Difference(
		expectedAliases, currentAliases)
	if errExpectedAliases != nil {
		logrus.WithFields(logrus.Fields{
			"error": errExpectedAliases,
			"uri":   s.GetUri(),
		}).Error("Failed to determine missing recipients!")
		return
	}
	missingRecipients := tmpExpectedAliases.Interface().([]string)
	if len(missingRecipients) > 0 {
		sort.Strings(missingRecipients)
		fmt.Printf(
			"    * missing: %s\n",
			strings.Join(missingRecipients, ", "))
	}

	// Render tags.
	for _, tag := range s.GetTags() {
		fmt.Printf("  + tags.%s: %s\n", tag.Name, tag.Value)
	}

	// Done!
	fmt.Println()
}
