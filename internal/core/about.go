package core

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/carlosabalde/pgp-tomb/internal/core/secret"
	"github.com/carlosabalde/pgp-tomb/internal/helpers/slices"
)

func About(uri string) {
	// Initializations.
	s := secret.New(uri)

	// Check secret exists.
	if !s.Exists() {
		fmt.Fprintln(os.Stderr, "Secret does not exist!")
		os.Exit(1)
	}

	// Extract current recipients.
	currentRecipientKeyIds, err := s.GetCurrentRecipientsKeyIds()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"uri":   uri,
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
		}).Fatal("Failed to get expected public keys!")
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
	tmpCurrentAliases, errCurrentAliases := slices.Difference(
		currentAliases, expectedAliases)
	if errCurrentAliases != nil {
		logrus.Fatal(errCurrentAliases)
	}
	rubbishRecipients := tmpCurrentAliases.Interface().([]string)
	if len(rubbishRecipients) > 0 {
		sort.Strings(rubbishRecipients)
		fmt.Printf(
			"! Rubbish recipients: %s\n",
			strings.Join(rubbishRecipients, ", "))
	}

	// Render missing recipients?
	tmpExpectedAliases, errExpectedAliases := slices.Difference(
		expectedAliases, currentAliases)
	if errExpectedAliases != nil {
		logrus.Fatal(errExpectedAliases)
	}
	missingRecipients := tmpExpectedAliases.Interface().([]string)
	if len(missingRecipients) > 0 {
		sort.Strings(missingRecipients)
		fmt.Printf(
			"! Missing recipients: %s\n",
			strings.Join(missingRecipients, ", "))
	}

	// Render tags.
	fmt.Println("- Tags:");
	for _, tag := range s.GetTags() {
		fmt.Printf("  + %s: %s\n", tag.Name, tag.Value);
	}
}
