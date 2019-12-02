package core

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/carlosabalde/pgp-tomb/internal/core/config"
	"github.com/carlosabalde/pgp-tomb/internal/core/query"
	"github.com/carlosabalde/pgp-tomb/internal/core/secret"
	"github.com/carlosabalde/pgp-tomb/internal/helpers/pgp"
)

func List(folder, queryString, keyAlias string) {
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
				listSecret(path, queryParsed, key)
			}
			return nil
		}); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to list secrets!")
	}
}

func listSecret(path string, q query.Query, key *pgp.PublicKey) {
	uri := strings.TrimPrefix(path, config.GetSecretsRoot())
	uri = strings.TrimPrefix(uri, string(os.PathSeparator))
	uri = strings.TrimSuffix(uri, config.SecretExtension)

	s, err := secret.Load(uri)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"uri":   s.GetUri(),
		}).Fatal("Failed to load secret!")
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
			}).Fatal("Failed to get expected public keys!")
		}
		if !found {
			return
		}
	}

	fmt.Printf("- %s\n", s.GetUri())
}
