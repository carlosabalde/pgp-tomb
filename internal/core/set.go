package core

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/carlosabalde/pgp-tomb/internal/core/config"
	"github.com/carlosabalde/pgp-tomb/internal/core/secret"
)

func Set(uri, inputPath string, tags []secret.Tag, ignoreSchema bool) {
	if !set(uri, inputPath, tags, ignoreSchema) {
		os.Exit(1)
	}
	fmt.Println("Done!")
}

func set(uri, inputPath string, tags []secret.Tag, ignoreSchema bool) bool {
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
				"file":  inputPath,
			}).Error("Failed to open input file!")
			return false
		}
		defer file.Close()
		input = file
	}

	// Check tags?
	if !ignoreSchema {
		if serializedTags, err := s.GetSerializedTags(); err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("Failed to serialize tags!")
			return false
		} else {
			if valid, errs := validateSchema(serializedTags, config.GetTags()); !valid {
				fmt.Fprintf(os.Stderr, "Tags do not match schema!\n")
				for _, err := range errs {
					fmt.Fprintf(os.Stderr, "  - %s\n", err)
				}
				return false
			}
		}
	}

	// Check template?
	if template := s.GetTemplate(); template != nil && template.Schema != nil && !ignoreSchema {
		buffer := new(bytes.Buffer)
		if _, err := io.Copy(buffer, input); err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("Failed to read input!")
			return false
		}

		if valid, errs := validateSchema(buffer.String(), template.Schema); !valid {
			fmt.Fprintf(os.Stderr, "Secret does not match '%s' schema!\n", template.Alias)
			for _, err := range errs {
				fmt.Fprintf(os.Stderr, "  - %s\n", err)
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
