package core

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/carlosabalde/pgp-tomb/internal/core/config"
	"github.com/carlosabalde/pgp-tomb/internal/core/query"
	"github.com/carlosabalde/pgp-tomb/internal/core/secret"
	"github.com/carlosabalde/pgp-tomb/internal/helpers/pgp"
)

func Rebuild(folderOrUri, queryString, keyAlias string, workers int, force, dryRun bool) {
	// Initializations.
	checked := 0
	queryParsed := parseQuery(queryString)

	// Initialize key.
	var key *pgp.PublicKey
	if keyAlias != "" {
		key = findPublicKey(keyAlias)
		if key == nil {
			fmt.Fprintln(os.Stderr, "Key does not exist!")
			os.Exit(1)
		}
	} else {
		key = config.GetIdentity()
	}

	// Launch workers.
	var waitGroup sync.WaitGroup
	tasksChannel := make(chan func() string, 32)
	for i := 0; i < workers; i++ {
		waitGroup.Add(1)
		go taskDispatcher(tasksChannel, &waitGroup)
	}

	// Define walk function.
	walk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if checkFile(path, info, tasksChannel, queryParsed, key, force, dryRun) {
				checked++
			}
		}
		return nil
	}

	// Check folder vs. URI & walk file system.
	root := ""
	if folderOrUri != "" {
		item := path.Join(config.GetSecretsRoot(), folderOrUri+config.SecretExtension)
		if info, err := os.Stat(item); err == nil && !info.IsDir() {
			walk(item, info, err)
		} else {
			root = path.Join(config.GetSecretsRoot(), folderOrUri)
			if info, err := os.Stat(root); os.IsNotExist(err) || !info.IsDir() {
				fmt.Fprintln(os.Stderr, "Folder does not exist!")
				os.Exit(1)
			}
		}
	} else {
		root = config.GetSecretsRoot()
	}
	if root != "" {
		if err := filepath.Walk(root, walk); err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("Failed to rebuild secrets!")
		}
	}

	// Wait for completion.
	close(tasksChannel)
	waitGroup.Wait()

	// Done!
	if dryRun {
		fmt.Printf("Done! %d files checked (dry run).\n", checked)
	} else {
		fmt.Printf("Done! %d files checked.\n", checked)
	}
}

func checkFile(
	path string, info os.FileInfo, tasksChannel chan func() string,
	q query.Query, key *pgp.PublicKey, force, dryRun bool) bool {
	if filepath.Ext(path) == config.SecretExtension {
		uri := strings.TrimPrefix(path, config.GetSecretsRoot())
		uri = strings.TrimPrefix(uri, string(os.PathSeparator))
		uri = strings.TrimSuffix(uri, config.SecretExtension)

		s, err := secret.Load(uri)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
				"uri":   s.GetUri(),
			}).Error("Failed to load secret!")
			return false
		}

		if !q.Eval(s) {
			return false
		}

		if key != nil {
			if readable, err := s.IsReadableBy(key); err == nil {
				if !readable {
					return false
				}
			} else {
				logrus.WithFields(logrus.Fields{
					"error": err,
					"uri":   s.GetUri(),
				}).Error("Failed to check if secret is readable!")
				return false
			}
		}

		tasksChannel <- func() string {
			return checkSecret(s, force, dryRun)
		}
		return true
	} else {
		tasksChannel <- func() string {
			return checkUnexpectFile(path, dryRun)
		}
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
	// Determine recipients.
	_, unknown, rubbish, missing, err := s.GetRecipients()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"uri":   s.GetUri(),
		}).Error("Failed to determine recipients!")
		return fmt.Sprintf("! Failed to determine recipients for '%s'", s.GetUri())
	}

	// Check recipients.
	result := ""
	if len(unknown) > 0 {
		result = fmt.Sprintf(
			"- Re-encrypting '%s': unknown recipients (%s)...",
			s.GetUri(), strings.Join(unknown, ", "))
	} else if len(rubbish) > 0 {
		result = fmt.Sprintf(
			"- Re-encrypting '%s': rubbish recipients (%s)...",
			s.GetUri(), strings.Join(rubbish, ", "))
	} else if len(missing) > 0 {
		result = fmt.Sprintf(
			"- Re-encrypting '%s': missing recipients (%s)...",
			s.GetUri(), strings.Join(missing, ", "))
	} else if force {
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
				"file":  path,
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
