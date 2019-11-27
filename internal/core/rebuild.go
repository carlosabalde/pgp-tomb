package core

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/carlosabalde/pgp-tomb/internal/core/config"
	"github.com/carlosabalde/pgp-tomb/internal/helpers/maps"
	"github.com/carlosabalde/pgp-tomb/internal/helpers/pgp"
)

func Rebuild(grep string, dryRun bool) {
	// Initializations.
	checked := 0
	var grepRegexp *regexp.Regexp
	if grep != "" {
		var err error
		grepRegexp, err = regexp.Compile(grep)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"regexp": grep,
				"error":  err,
			}).Fatal("Failed to compile limit regexp!")
		}
	} else {
		grepRegexp = nil
	}

	// Walk file system.
	if err := filepath.Walk(
		config.GetSecretsRoot(),
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				if checkFile(path, info, grepRegexp, dryRun) {
					checked++
				}
			}
			return nil
		}); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to rebuild secrets!")
	}

	// Done!
	fmt.Printf("Done! %d files checked.\n", checked)
}

func checkFile(path string, info os.FileInfo, grep *regexp.Regexp, dryRun bool) bool {
	var checker func()
	uri := strings.TrimPrefix(path, config.GetSecretsRoot())
	uri = strings.TrimPrefix(uri, string(os.PathSeparator))

	if filepath.Ext(path) == config.SecretExtension {
		uri = strings.TrimSuffix(uri, config.SecretExtension)
		checker = func() {
			checkSecret(uri, path, dryRun)
		}
	} else {
		checker = func() {
			checkUnexpectFile(path, dryRun)
		}
	}

	if grep == nil || grep.Match([]byte(uri)) {
		checker()
		return true
	}

	return false
}

func checkSecret(uri, path string, dryRun bool) {
	// Initialize input reader.
	input, err := os.Open(path)
	if err != nil {
		fmt.Printf("! Failed to open file '%s'", path)
		logrus.WithFields(logrus.Fields{
			"error": err,
			"path":  path,
		}).Error("Failed to open file!")
		return
	}
	defer input.Close()

	// Parse secret.
	currentRecipientKeyIds, err := pgp.GetRecipientKeyIdsForEncryptedMessage(input)
	if err != nil {
		fmt.Printf("! Failed to determine recipients for '%s'", uri)
		logrus.WithFields(logrus.Fields{
			"error": err,
			"uri":   uri,
		}).Error("Failed to determine recipients!")
		return
	}

	// Determine expected recipients.
	expectedKeys := getPublicKeysForSecret(uri)

	// Check current recipients.
	reEncrypt := false
	expectedKeysByAlias := make(map[string]*pgp.PublicKey)
	for _, key := range expectedKeys {
		expectedKeysByAlias[key.Alias] = key
	}
	for _, keyId := range currentRecipientKeyIds {
		key := findPublicKeyByKeyId(keyId)
		if key != nil {
			if _, found := expectedKeysByAlias[key.Alias]; !found {
				fmt.Printf(
					"- Re-encrypting '%s': rubbish recipients (%s, etc.)...",
					uri, key.Alias)
				reEncrypt = true
				break
			} else {
				delete(expectedKeysByAlias, key.Alias)
			}
		} else {
			fmt.Printf(
				"- Re-encrypting '%s': unknown rubbish recipients (0x%x, etc.)...",
				uri, keyId)
			reEncrypt = true
			break
		}
	}
	if !reEncrypt && len(expectedKeysByAlias) > 0 {
		keysAliases := maps.StringKeysSlice(expectedKeysByAlias)
		sort.Strings(keysAliases)
		fmt.Printf(
			"- Re-encrypting '%s': missing recipients (%s)...",
			uri, strings.Join(keysAliases, ", "))
		reEncrypt = true
	}

	// Re-encrypt?
	if reEncrypt {
		if !dryRun {
			reEncryptSecret(path, expectedKeys)
		} else {
			success()
		}
	}
}

func reEncryptSecret(path string, keys []*pgp.PublicKey) {
	// Initialize input reader.
	input, err := os.Open(path)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"path":  path,
		}).Fatal("Failed to open secret file!")
	}
	defer input.Close()

	// Initialize output buffer.
	buffer := new(bytes.Buffer)

	// Decrypt secret.
	if err := pgp.DecryptWithGPG(config.GetGPG(), input, buffer); err != nil {
		failure()
		logrus.WithFields(logrus.Fields{
			"error": err,
			"path":  path,
		}).Error("Failed to decrypt file for re-encryption! Are you allowed to access it?")
		return
	}

	// Initialize output writer.
	output, err := os.Create(path)
	if err != nil {
		failure()
		logrus.WithFields(logrus.Fields{
			"error": err,
			"path":  path,
		}).Error("Failed to open file for re-encryption!")
		return
	}
	defer output.Close()

	// Encrypt secret.
	if err := pgp.Encrypt(buffer, output, keys); err != nil {
		failure()
		logrus.WithFields(logrus.Fields{
			"error": err,
			"path":  path,
		}).Error("Failed to re-encrypt file!")
		return
	}

	// Done!
	success()
}

func checkUnexpectFile(path string, dryRun bool) {
	fmt.Printf("- Removing unexpected file '%s'...", path)

	if !dryRun {
		if err := os.Remove(path); err != nil {
			failure()
			logrus.WithFields(logrus.Fields{
				"error": err,
				"path":  path,
			}).Error("Failed to remove unexpected file!")
		} else {
			success()
		}
	} else {
		success()
	}
}

func success() {
	fmt.Println(" ✓")
}

func failure() {
	fmt.Println(" ✗")
}
