package core

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/carlosabalde/pgp-tomb/internal/helpers/pgp"
)

func Set(uri, inputPath string) {
	// Initialize input reader.
	var input io.Reader
	if inputPath == "" {
		input = os.Stdin
	} else {
		file, err := os.Open(inputPath)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
				"path":  inputPath,
			}).Fatal("Failed to open input file!")
		}
		defer file.Close()
		input = file
	}

	// Create missing folders in path to secret.
	secretPath := getPathForSecret(uri)
	if err := os.MkdirAll(filepath.Dir(secretPath), os.ModePerm); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"path":  secretPath,
		}).Fatal("Failed to create path to secret!")
	}

	// Initialize output writer.
	output, err := os.Create(secretPath)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"path":  secretPath,
		}).Fatal("Failed to open output file!")
	}
	defer output.Close()

	// Encrypt secret.
	if err := pgp.Encrypt(input, output, getPublicKeysForSecret(uri)); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to encrypt file!")
	}

	// Done!
	fmt.Println("Done!")
}
