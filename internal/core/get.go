package core

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/atotto/clipboard"
	"github.com/sirupsen/logrus"

	"github.com/carlosabalde/pgp-tomb/internal/core/secret"
)

func Get(uri, outputPath string, copyToClipboard bool) {
	// Load secret.
	s, err := secret.Load(uri)
	if err != nil {
		switch err := err.(type) {
		case *secret.DoesNotExist:
			fmt.Fprintln(os.Stderr, "Secret does not exist!")
			os.Exit(1)
		default:
			logrus.WithFields(logrus.Fields{
				"error": err,
				"uri":   uri,
			}).Error("Failed to load secret!")
			return
		}
	}

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
	if err := s.Decrypt(output); err != nil {
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
