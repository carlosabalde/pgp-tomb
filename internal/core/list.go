package core

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/carlosabalde/pgp-tomb/internal/core/config"
	"github.com/carlosabalde/pgp-tomb/internal/core/secret"
	"github.com/carlosabalde/pgp-tomb/internal/helpers/pgp"
)

func List(folder, grep, keyAlias string) {
	// Initialize grep.
	var grepRegexp *regexp.Regexp
	if grep != "" {
		var err error
		grepRegexp, err = regexp.Compile(grep)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"regexp": grep,
				"error":  err,
			}).Fatal("Failed to compile grep regexp!")
		}
	} else {
		grepRegexp = nil
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
				listSecret(path, grepRegexp, key)
			}
			return nil
		}); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to list secrets!")
	}
}

func listSecret(path string, grep *regexp.Regexp, key *pgp.PublicKey) {
	uri := strings.TrimPrefix(path, config.GetSecretsRoot())
	uri = strings.TrimPrefix(uri, string(os.PathSeparator))
	uri = strings.TrimSuffix(uri, config.SecretExtension)

	s := secret.New(uri)

	if grep != nil && !grep.Match([]byte(s.GetUri())) {
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
