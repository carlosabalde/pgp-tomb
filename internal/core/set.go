package core

import (
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/carlosabalde/pgp-tomb/internal/core/secret"
)

func Set(uri, inputPath string) {
	// Initializations.
	s := secret.New(uri)

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

	// Encrypt secret.
	if err := s.Encrypt(input); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to encrypt file!")
	}

	// Done!
	fmt.Println("Done!")
}
