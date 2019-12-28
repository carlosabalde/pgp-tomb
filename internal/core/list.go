package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/carlosabalde/pgp-tomb/internal/core/config"
	"github.com/carlosabalde/pgp-tomb/internal/core/query"
	"github.com/carlosabalde/pgp-tomb/internal/core/secret"
	"github.com/carlosabalde/pgp-tomb/internal/helpers/maps"
	"github.com/carlosabalde/pgp-tomb/internal/helpers/pgp"
)

func List(
	folderOrUri string, long bool, queryString, keyAlias string,
	ignoreSchema bool, enableJson bool) {
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
		key = config.GetIdentity()
	}

	// Define walk function.
	listed := 0
	walk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == config.SecretExtension {
			if listSecret(path, long, queryParsed, key, ignoreSchema, enableJson, listed == 0) {
				listed++
			}
		}
		return nil
	}

	// Check folder vs. URI & walk file system.
	root := ""
	if folderOrUri != "" {
		item := path.Join(config.GetSecretsRoot(), folderOrUri+config.SecretExtension)
		if info, err := os.Stat(item); err == nil && !info.IsDir() {
			if enableJson {
				fmt.Print("{")
			}
			walk(item, info, err)
			if enableJson {
				fmt.Print("}")
			}
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
		if enableJson {
			fmt.Print("{")
		}
		if err := filepath.Walk(root, walk); err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("Failed to list secrets!")
		}
		if enableJson {
			fmt.Print("}")
		}
	}
}

func listSecret(
	path string, long bool, q query.Query, key *pgp.PublicKey,
	ignoreSchema bool, enableJson bool, isFirst bool) bool {
	uri := strings.TrimPrefix(path, config.GetSecretsRoot())
	uri = strings.TrimPrefix(uri, string(os.PathSeparator))
	uri = strings.TrimSuffix(uri, config.SecretExtension)

	s, err := secret.Load(uri)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"uri":   uri,
		}).Error("Failed to load secret!")
		return false
	}

	if !q.Eval(s) {
		return false
	}

	if key != nil {
		if readable, err := s.IsReadableBy(key); err == nil {
			if !readable {
				return false
			}
		} else {
			logrus.WithFields(logrus.Fields{
				"error": err,
				"uri":   s.GetUri(),
			}).Error("Failed to check if secret is readable!")
			return false
		}
	}

	if !enableJson {
		fmt.Printf("- %s\n", s.GetUri())
	}

	if enableJson || long {
		if es, err := exportSecret(s, ignoreSchema); err == nil {
			if !enableJson {
				renderSecretDetails(es)
			} else {
				if serializedUri, err := json.Marshal(s.GetUri()); err == nil {
					if serializedSecret, err := json.Marshal(es); err == nil {
						if !isFirst {
							fmt.Print(",")
						}
						fmt.Printf("%s:%s", string(serializedUri), string(serializedSecret))
					} else {
						logrus.WithFields(logrus.Fields{
							"error": err,
							"uri":   s.GetUri(),
						}).Error("Failed to serialize secret details!")
						return false
					}
				} else {
					logrus.WithFields(logrus.Fields{
						"error": err,
						"uri":   s.GetUri(),
					}).Error("Failed to serialize secret URI!")
					return false
				}
			}
		} else {
			logrus.WithFields(logrus.Fields{
				"error": err,
				"uri":   s.GetUri(),
			}).Error("Failed to export secret!")
			return false
		}
	}

	return true
}

type exportedSecret struct {
	Recipients exportedRecipients `json:"recipients"`
	Template   *exportedTemplate  `json:"template"`
	Tags       exportedTags       `json:"tags"`
}

type exportedRecipients struct {
	Expected []string `json:"expected"`
	Unknown  []string `json:"unknown"`
	Rubbish  []string `json:"rubbish"`
	Missing  []string `json:"missing"`
}

type exportedTemplate struct {
	Alias string `json:"alias"`
	State string `json:"state"`
}

type exportedTags struct {
	State string            `json:"state"`
	Tags  map[string]string `json:"tags"`
}

func exportSecret(s *secret.Secret, ignoreSchema bool) (exportedSecret, error) {
	// Initializations.
	result := exportedSecret{}

	// Export recipients.
	expected, unknown, rubbish, missing, err := s.GetRecipients()
	if err != nil {
		return exportedSecret{}, errors.Wrap(err, "Failed to determine recipients!")
	}
	result.Recipients.Expected = expected
	result.Recipients.Unknown = unknown
	result.Recipients.Rubbish = rubbish
	result.Recipients.Missing = missing

	// Export template.
	if template := s.GetTemplate(); template != nil && template.Schema != nil {
		result.Template = &exportedTemplate{
			Alias: template.Alias,
		}
		if !ignoreSchema {
			buffer := new(bytes.Buffer)
			if err := s.Decrypt(buffer); err != nil {
				return exportedSecret{}, errors.Wrap(err, "Failed to decrypt secret!")
			} else {
				if valid, _ := validateSchema(buffer.String(), template.Schema); valid {
					result.Template.State = "valid"
				} else {
					result.Template.State = "invalid"
				}
			}
		} else {
			result.Template.State = "unknown"
		}
	} else {
		result.Template = nil
	}

	// Export tags.
	result.Tags = exportedTags{
		Tags: make(map[string]string),
	}
	if !ignoreSchema {
		if serializedTags, err := s.GetSerializedTags(); err != nil {
			return exportedSecret{}, errors.Wrap(err, "Failed to serialize secret tags!")
		} else {
			if valid, _ := validateSchema(serializedTags, config.GetTags()); valid {
				result.Tags.State = "valid"
			} else {
				result.Tags.State = "invalid"
			}
		}
	} else {
		result.Tags.State = "unknown"
	}
	for _, tag := range s.GetTags() {
		result.Tags.Tags[tag.Name] = tag.Value
	}

	// Done!
	return result, nil
}

func renderSecretDetails(es exportedSecret) {
	// Render expected recipients.
	fmt.Print("  |-- recipients: ")
	if len(es.Recipients.Expected) > 0 {
		fmt.Println(strings.Join(es.Recipients.Expected, ", "))
	} else {
		fmt.Println("-")
	}

	// Render unknown recipients.
	fmt.Print("  |   |-- unknown: ")
	if len(es.Recipients.Unknown) > 0 {
		fmt.Println(strings.Join(es.Recipients.Unknown, ", "))
	} else {
		fmt.Println("-")
	}

	// Render rubbish recipients.
	fmt.Print("  |   |-- rubbish: ")
	if len(es.Recipients.Rubbish) > 0 {
		fmt.Println(strings.Join(es.Recipients.Rubbish, ", "))
	} else {
		fmt.Println("-")
	}

	// Render missing recipients.
	fmt.Print("  |   `-- missing: ")
	if len(es.Recipients.Missing) > 0 {
		fmt.Println(strings.Join(es.Recipients.Missing, ", "))
	} else {
		fmt.Println("-")
	}

	// Render template.
	fmt.Print("  |-- template: ")
	if es.Template != nil {
		var decoration string
		switch es.Template.State {
		case "valid":
			decoration = "✓"
		case "invalid":
			decoration = "✗"
		case "unknown":
			decoration = "?"
		}
		fmt.Printf("%s %s\n", es.Template.Alias, decoration)
	} else {
		fmt.Println("-")
	}

	// Render tags.
	var decoration string
	switch es.Tags.State {
	case "valid":
		decoration = "✓"
	case "invalid":
		decoration = "✗"
	case "unknown":
		decoration = "?"
	}
	fmt.Printf("  `-- tags %s\n", decoration)
	res, _ := maps.KeysSlice(es.Tags.Tags)
	tagsNames := res.Interface().([]string)
	sort.Strings(tagsNames)
	for i, tagsName := range tagsNames {
		decoration := "|"
		if i == len(tagsNames)-1 {
			decoration = "`"
		}
		fmt.Printf("      %s-- %s: %s\n", decoration, tagsName, es.Tags.Tags[tagsName])
	}

	// Done!
	fmt.Println()
}
