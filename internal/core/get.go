package core

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/atotto/clipboard"
	"github.com/sirupsen/logrus"

	"github.com/carlosabalde/pgp-tomb/internal/core/config"
	"github.com/carlosabalde/pgp-tomb/internal/helpers/pgp"
)

func Get(uri, outputPath string, copyToClipboard bool) {
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
		}).Fatal("Failed to open secret file!")
	}
	defer input.Close()

	// Initialize output writer.
	var output io.Writer
	if copyToClipboard {
		output = new(bytes.Buffer)
	} else {
		if outputPath == "" {
			output = os.Stdout
		} else {
			file, err := os.Create(outputPath)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
					"path":  outputPath,
				}).Fatal("Failed to open output file!")
			}
			defer file.Close()
			output = file
		}
	}

	// Decrypt secret.
	if err := pgp.DecryptWithGPG(config.GetGPG(), input, output); err != nil {
		fmt.Fprintln(
			os.Stderr,
			"Unable to decrypt secret! Are you allowed to access it?")
		os.Exit(1)
	}

	// Copy decrypted secret to system clipboard?
	if copyToClipboard {
		secret := output.(*bytes.Buffer).String()
		if err := clipboard.WriteAll(secret); err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("Failed to copy to system clipboard!")
		}
	}
}
