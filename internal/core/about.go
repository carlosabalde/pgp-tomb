package core

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/carlosabalde/pgp-tomb/internal/helpers/pgp"
	"github.com/carlosabalde/pgp-tomb/internal/helpers/slices"
)

func About(uri string) {
	// Check secret exists.
	secretPath := findSecret(uri)
	if secretPath == "" {
		fmt.Fprintln(os.Stderr, "Secret does not exist!")
		os.Exit(1)
	}

	// Initialize input reader.
	input, err := os.Open(secretPath)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"path":  secretPath,
		}).Fatal("Failed to open input file!")
	}
	defer input.Close()

	// Extract current recipients.
	currentRecipientKeyIds, err := pgp.GetRecipientKeyIdsForEncryptedMessage(input)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"uri":   uri,
		}).Error("Failed to determine recipients!")
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
	for _, key := range getPublicKeysForSecret(uri) {
		expectedAliases = append(expectedAliases, key.Alias)
	}

	// Render expected recipients.
	sort.Strings(expectedAliases)
	fmt.Printf(
		"- Expected recipients: %s\n",
		strings.Join(expectedAliases, ", "))

	// Render unknown recipients?
	if len(unknownRecipients) > 0 {
		sort.Strings(unknownRecipients)
		fmt.Printf(
			"! Unknown rubbish recipients: %s\n",
			strings.Join(unknownRecipients, ", "))
	}

	// Render rubbish recipients?
	rubbishRecipients := slices.Difference(
		currentAliases,
		expectedAliases).Interface().([]string)
	if len(rubbishRecipients) > 0 {
		sort.Strings(rubbishRecipients)
		fmt.Printf(
			"! Rubbish recipients: %s\n",
			strings.Join(rubbishRecipients, ", "))
	}

	// Render missing recipients?
	missingRecipients := slices.Difference(
		expectedAliases,
		currentAliases).Interface().([]string)
	if len(missingRecipients) > 0 {
		sort.Strings(missingRecipients)
		fmt.Printf(
			"! Missing recipients: %s\n",
			strings.Join(missingRecipients, ", "))
	}
}
