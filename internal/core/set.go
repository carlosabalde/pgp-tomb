package core

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/xeipuuv/gojsonschema"

	"github.com/carlosabalde/pgp-tomb/internal/core/secret"
)

func Set(uri, inputPath string, tags []secret.Tag) {
	if !set(uri, inputPath, tags) {
		os.Exit(1)
	}
	fmt.Println("Done!")
}

func set(uri, inputPath string, tags []secret.Tag) bool {
	// Initializations.
	s := secret.New(uri)
	s.SetTags(tags)

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
			}).Error("Failed to open input file!")
			return false
		}
		defer file.Close()
		input = file
	}

	// Check template?
	if template := s.GetTemplate(); template != nil {
		buffer := new(bytes.Buffer)
		if _, err := io.Copy(buffer, input); err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("Failed to read input!")
			return false
		}

		loader := gojsonschema.NewStringLoader(buffer.String())
		validation, err := template.Schema.Validate(loader)
		if err != nil || !validation.Valid() {
			fmt.Fprintf(os.Stderr, "Secret does not match '%s' template!\n", template.Alias)
			if err == nil {
				for _, err := range validation.Errors() {
					fmt.Fprintf(os.Stderr, "  - %s\n", err)
				}
			}
			return false
		}

		input = buffer
	}

	// Encrypt secret.
	if err := s.Encrypt(input); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("Failed to encrypt file!")
		return false
	}

	// Done!
	return true
}
