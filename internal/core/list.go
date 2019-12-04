package core

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/carlosabalde/pgp-tomb/internal/core/config"
	"github.com/carlosabalde/pgp-tomb/internal/core/query"
	"github.com/carlosabalde/pgp-tomb/internal/core/secret"
	"github.com/carlosabalde/pgp-tomb/internal/helpers/pgp"
)

func List(folderOrUri string, long bool, queryString, keyAlias string, ignoreSchema bool) {
	// Initializations.
	queryParsed := parseQuery(queryString)

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

	// Define walk function.
	walk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == config.SecretExtension {
			listSecret(path, long, queryParsed, key, ignoreSchema)
		}
		return nil
	}

	// Check folder vs. URI & walk file system.
	root := ""
	if folderOrUri != "" {
		item := path.Join(config.GetSecretsRoot(), folderOrUri+config.SecretExtension)
		if info, err := os.Stat(item); err == nil && !info.IsDir() {
			walk(item, info, err)
		} else {
			root = path.Join(config.GetSecretsRoot(), folderOrUri)
			if info, err := os.Stat(root); os.IsNotExist(err) || !info.IsDir() {
				fmt.Fprintln(os.Stderr, "Folder does not exist!")
				os.Exit(1)
			}
		}
	} else {
		root = config.GetSecretsRoot()
	}
	if root != "" {
		if err := filepath.Walk(root, walk); err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("Failed to list secrets!")
		}
	}
}

func listSecret(path string, long bool, q query.Query, key *pgp.PublicKey, ignoreSchema bool) {
	uri := strings.TrimPrefix(path, config.GetSecretsRoot())
	uri = strings.TrimPrefix(uri, string(os.PathSeparator))
	uri = strings.TrimSuffix(uri, config.SecretExtension)

	s, err := secret.Load(uri)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"uri":   s.GetUri(),
		}).Error("Failed to load secret!")
		return
	}

	if !q.Eval(s) {
		return
	}

	if key != nil {
		found := false
		if keys, err := s.GetExpectedPublicKeys(); err == nil {
			for _, aKey := range keys {
				if aKey == key {
					found = true
					break
				}
			}
		} else {
			logrus.WithFields(logrus.Fields{
				"error": err,
				"uri":   s.GetUri(),
			}).Error("Failed to get expected public keys!")
			return
		}
		if !found {
			return
		}
	}

	fmt.Printf("- %s\n", s.GetUri())
	if long {
		renderSecretDetails(s, ignoreSchema)
	}
}

func renderSecretDetails(s *secret.Secret, ignoreSchema bool) {
	// Determine recipients.
	expected, unknown, rubbish, missing, err := s.GetRecipients()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"uri":   s.GetUri(),
		}).Error("Failed to determine recipients!")
		return
	}

	// Render expected recipients.
	fmt.Print("  |-- recipients: ")
	if len(expected) > 0 {
		fmt.Println(strings.Join(expected, ", "))
	} else {
		fmt.Println("-")
	}

	// Render unknown recipients.
	fmt.Print("  |   |-- unknown: ")
	if len(unknown) > 0 {
		fmt.Println(strings.Join(unknown, ", "))
	} else {
		fmt.Println("-")
	}

	// Render rubbish recipients.
	fmt.Print("  |   |-- rubbish: ")
	if len(rubbish) > 0 {
		fmt.Println(strings.Join(rubbish, ", "))
	} else {
		fmt.Println("-")
	}

	// Render missing recipients.
	fmt.Print("  |   `-- missing: ")
	if len(missing) > 0 {
		fmt.Println(strings.Join(missing, ", "))
	} else {
		fmt.Println("-")
	}

	// Render template.
	fmt.Print("  |-- template: ")
	if template := s.GetTemplate(); template != nil {
		var decoration string
		if !ignoreSchema {
			buffer := new(bytes.Buffer)
			if err := s.Decrypt(buffer); err != nil {
				decoration = "?"
			} else {
				if valid, _ := validateSchema(buffer.String(), template.Schema); !valid {
					decoration = "✗"
				} else {
					decoration = "✓"
				}
			}
		} else {
			decoration = "?"
		}
		fmt.Printf("%s %s\n", template.Alias, decoration)
	} else {
		fmt.Println("-")
	}

	// Render tags.
	tags := s.GetTags()
	fmt.Println("  `-- tags")
	for i, tag := range tags {
		decoration := "|"
		if i == len(tags)-1 {
			decoration = "`"
		}
		fmt.Printf("      %s-- %s: %s\n", decoration, tag.Name, tag.Value)
	}

	// Done!
	fmt.Println()
}
