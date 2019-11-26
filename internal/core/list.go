package core

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/carlosabalde/pgp-tomb/internal/core/config"
	"github.com/carlosabalde/pgp-tomb/internal/helpers/pgp"
)

func List(limit, keyAlias string) {
	// Initialize limit.
	var limitRegexp *regexp.Regexp
	if limit != "" {
		var err error
		limitRegexp, err = regexp.Compile(limit)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"regexp": limit,
				"error":  err,
			}).Fatal("Failed to compile limit regexp!")
		}
	} else {
		limitRegexp = nil
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

	// Walk file system.
	if err := filepath.Walk(
		config.GetSecretsRoot(),
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && filepath.Ext(path) == config.SecretExtension {
				listSecret(path, limitRegexp, key)
			}
			return nil
		}); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to list secrets!")
	}
}

func listSecret(path string, limit *regexp.Regexp, key *pgp.PublicKey) {
	uri := strings.TrimPrefix(path, config.GetSecretsRoot())
	uri = strings.TrimPrefix(uri, string(os.PathSeparator))
	uri = strings.TrimSuffix(uri, config.SecretExtension)

	if limit != nil && !limit.Match([]byte(uri)) {
		return
	}

	if key != nil {
		found := false
		for _, aKey := range getPublicKeysForSecret(uri) {
			if aKey == key {
				found = true
				break
			}
		}
		if !found {
			return
		}
	}

	fmt.Printf("- %s\n", uri)
}
