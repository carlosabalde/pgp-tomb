package core

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/carlosabalde/pgp-tomb/internal/core/config"
	"github.com/carlosabalde/pgp-tomb/internal/core/query"
	"github.com/carlosabalde/pgp-tomb/internal/core/secret"
	"github.com/carlosabalde/pgp-tomb/internal/helpers/maps"
	"github.com/carlosabalde/pgp-tomb/internal/helpers/pgp"
)

func Rebuild(folder, queryString string, workers int, force, dryRun bool) {
	// Initialize counter.
	checked := 0

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

	// Check folder.
	root := path.Join(config.GetSecretsRoot(), folder)
	if folder != "" {
		if info, err := os.Stat(root); os.IsNotExist(err) || !info.IsDir() {
			fmt.Fprintln(os.Stderr, "Folder does not exist!")
			os.Exit(1)
		}
	}

	// Launch workers.
	var waitGroup sync.WaitGroup
	tasksChannel := make(chan func() string, 32)
	for i := 0; i < workers; i++ {
		waitGroup.Add(1)
		go taskDispatcher(tasksChannel, &waitGroup)
	}

	// Walk file system.
	if err := filepath.Walk(
		root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				if checkFile(path, info, tasksChannel, queryParsed, force, dryRun) {
					checked++
				}
			}
			return nil
		}); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to rebuild secrets!")
	}

	// Wait for completion.
	close(tasksChannel)
	waitGroup.Wait()

	// Done!
	fmt.Printf("Done! %d files checked.\n", checked)
}

func checkFile(
	path string, info os.FileInfo, tasksChannel chan func() string,
	q query.Query, force, dryRun bool) bool {
	if filepath.Ext(path) == config.SecretExtension {
		uri := strings.TrimPrefix(path, config.GetSecretsRoot())
		uri = strings.TrimPrefix(uri, string(os.PathSeparator))
		uri = strings.TrimSuffix(uri, config.SecretExtension)

		s, err := secret.Load(uri)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
				"uri":   s.GetUri(),
			}).Fatal("Failed to load secret!")
			return false
		}

		if q.Eval(s) {
			task := func() string {
				return checkSecret(s, force, dryRun)
			}
			tasksChannel <- task
			return true
		}
	} else {
		task := func() string {
			return checkUnexpectFile(path, dryRun)
		}
		tasksChannel <- task
		return true
	}

	return false
}

func taskDispatcher(tasksChannel <-chan func() string, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	for task := range tasksChannel {
		message := task()
		if message != "" {
			fmt.Println(message)
		}
	}
}

func checkSecret(s *secret.Secret, force, dryRun bool) string {
	// Extract current recipients.
	currentRecipientKeyIds, err := s.GetCurrentRecipientsKeyIds()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"uri":   s.GetUri(),
		}).Error("Failed to determine recipients!")
		return fmt.Sprintf("! Failed to determine recipients for '%s'", s.GetUri())
	}

	// Determine expected recipients.
	expectedKeys, err := s.GetExpectedPublicKeys()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"uri":   s.GetUri(),
		}).Error("Failed to determine expected public keys!")
		return fmt.Sprintf("! Failed to determine expected public keys for '%s'", s.GetUri())
	}

	// Check current recipients.
	result := ""
	expectedKeysByAlias := make(map[string]*pgp.PublicKey)
	for _, key := range expectedKeys {
		expectedKeysByAlias[key.Alias] = key
	}
	for _, keyId := range currentRecipientKeyIds {
		key := findPublicKeyByKeyId(keyId)
		if key != nil {
			if _, found := expectedKeysByAlias[key.Alias]; !found {
				result = fmt.Sprintf(
					"- Re-encrypting '%s': rubbish recipients (%s, etc.)...",
					s.GetUri(), key.Alias)
				break
			} else {
				delete(expectedKeysByAlias, key.Alias)
			}
		} else {
			result = fmt.Sprintf(
				"- Re-encrypting '%s': unknown rubbish recipients (0x%x, etc.)...",
				s.GetUri(), keyId)
			break
		}
	}
	if result == "" && len(expectedKeysByAlias) > 0 {
		res, err := maps.KeysSlice(expectedKeysByAlias)
		if err != nil {
			logrus.Fatal(err)
		}
		keysAliases := res.Interface().([]string)
		sort.Strings(keysAliases)
		result = fmt.Sprintf(
			"- Re-encrypting '%s': missing recipients (%s)...",
			s.GetUri(), strings.Join(keysAliases, ", "))
	}
	if result == "" && force {
		result = fmt.Sprintf("- Re-encrypting '%s': forced...", s.GetUri())
	}

	// Re-encrypt?
	if result != "" {
		if !dryRun {
			if reEncryptSecret(s) {
				result += " ✓"
			} else {
				result += " ✗"
			}
		} else {
			result += " ✓"
		}
	}

	// Done!
	return result
}

func reEncryptSecret(s *secret.Secret) bool {
	// Initialize output buffer.
	buffer := new(bytes.Buffer)

	// Decrypt secret.
	if err := s.Decrypt(buffer); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"uri":   s.GetUri(),
		}).Error("Failed to decrypt file for re-encryption! Are you allowed to access it?")
		return false
	}

	// Encrypt secret.
	if err := s.Encrypt(buffer); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"uri":   s.GetUri(),
		}).Error("Failed to re-encrypt file!")
		return false
	}

	// Done!
	return true
}

func checkUnexpectFile(path string, dryRun bool) string {
	result := fmt.Sprintf("- Removing unexpected file '%s'...", path)

	if !dryRun {
		if err := os.Remove(path); err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
				"path":  path,
			}).Error("Failed to remove unexpected file!")
			result += " ✗"
		} else {
			result += " ✓"
		}
	} else {
		result += " ✓"
	}

	return result
}
