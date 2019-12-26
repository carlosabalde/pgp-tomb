package config

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/xeipuuv/gojsonschema"

	"github.com/carlosabalde/pgp-tomb/internal/core/query"
	"github.com/carlosabalde/pgp-tomb/internal/helpers/maps"
	"github.com/carlosabalde/pgp-tomb/internal/helpers/pgp"
)

const (
	configSchema = `
		{
		  "required": [
		    "keepers"
		  ],
		  "type": "object",
		  "properties": {
		    "root": {
		      "type": ["string", "null"]
		    },
		    "identity": {
		      "type": ["string", "null"]
		    },
		    "key": {
		      "type": ["string", "null"]
		    },
		    "keepers": {
		      "type": "array",
		      "items": {
		        "type": "string"
		      },
		      "minItems": 1
		    },
		    "teams": {
		      "type": ["object", "null"],
		      "patternProperties": {
		        ".*": {
		          "type": ["array", "null"],
		          "items": {
		            "type": "string"
		          }
		        }
		      }
		    },
		    "tags": {
		      "type": ["string", "null"]
		    },
		    "permissions": {
		      "type": ["array", "null"],
		      "items": {
		        "type": "object",
		        "minProperties": 1,
		        "maxProperties": 1,
		        "patternProperties": {
		          ".*": {
		            "type": ["array", "null"],
		            "items": {
		              "type": "string",
		              "pattern": "^(?:\\+|\\-).+"
		            }
		          }
		        }
		      }
		    },
		    "templates": {
		      "type": ["array", "null"],
		      "items": {
		        "type": "object",
		        "minProperties": 1,
		        "maxProperties": 1,
		        "patternProperties": {
		          ".*": {
		            "type": "string"
		          }
		        }
		      }
		    }
		  },
		  "additionalProperties": false
		}
		`

	defaultTagsSchema = `
		{
		  "type": "object",
		  "required": [],
		  "patternProperties": {
		    ".*": {
		      "type": "string"
		    }
		  },
		  "additionalProperties": true
		}
		`
)

func Init(file string) {
	checkSchema(file)
	initRootConfig()
	initGPGConfig()
	initEditorConfig()
	initHooksConfig()
	initKeyConfig()
	initPublicKeysConfig()
	initIdentity()
	initSecretsConfig()
	initKeepersConfig()
	initTeamsConfig()
	initTagsConfig()
	initPermissionRulesConfig()
	initTemplatesConfig()
	initTemplateRulesConfig()
}

func checkSchema(file string) {
	configYaml, err := ioutil.ReadFile(file)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"file":  file,
			"error": err,
		}).Fatal("Failed to load configuration file!")
	}

	configJson, err := yaml.YAMLToJSON([]byte(configYaml))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"file":  file,
			"error": err,
		}).Fatal("Failed to covert configuration to JSON!")
	}

	schemaLoader := gojsonschema.NewStringLoader(configSchema)
	documentLoader := gojsonschema.NewStringLoader(string(configJson))
	validation, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"file":  file,
			"error": err,
		}).Fatal("Failed to validate configuration file!")
	}
	if !validation.Valid() {
		errors := make([]string, 0)
		for _, err := range validation.Errors() {
			errors = append(errors, err.String())
		}
		logrus.WithFields(logrus.Fields{
			"file":   file,
			"errors": strings.Join(errors, ", "),
		}).Fatal("Invalid configuration file!")
	}

	logrus.WithFields(logrus.Fields{
		"file": file,
	}).Info("Configuration schema checked")
}

func initRootConfig() {
	if !viper.IsSet("root") || viper.GetString("root") == "" {
		root, err := os.Getwd()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("Failed to determine root folder!")
		}
		viper.Set("root", root)
	}

	logrus.WithFields(logrus.Fields{
		"folder": viper.Get("root"),
	}).Info("Root folder initialized")
}

func initGPGConfig() {
	var command string

	viper.Set("gpg", "")
	viper.Set("gpg-connect-agent", "")

	if !viper.IsSet("key") || viper.GetString("key") == "" {
		command = "gpg"
	} else {
		command = "gpg-connect-agent"
	}

	executable, err := exec.LookPath(command)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":   err,
			"command": command,
		}).Fatal("Failed to locate GPG executable!")
	}

	logrus.WithFields(logrus.Fields{
		"executable": executable,
	}).Info("GPG initialized")

	viper.Set(command, executable)
}

func initEditorConfig() {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = DefaultEditor
	}

	executable, err := exec.LookPath(editor)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"editor": editor,
			"error":  err,
		}).Fatal("Failed to locate editor executable!")
	}

	logrus.WithFields(logrus.Fields{
		"executable": executable,
	}).Info("Editor initialized")

	viper.Set("editor", executable)
}

func initHooksConfig() {
	hooks := make(map[string]Hook)
	hooksRoot := path.Join(GetRoot(), "hooks")

	if info, err := os.Stat(hooksRoot); os.IsNotExist(err) || !info.IsDir() {
		logrus.WithFields(logrus.Fields{
			"folder": hooksRoot,
			"error":  err,
		}).Fatal("Failed to access to hooks folder!")
	}

	for _, alias := range [...]string{"pre", "post"} {
		path := path.Join(hooksRoot, alias+HookExtension)
		if info, err := os.Stat(path); err == nil && !info.IsDir() && info.Mode()&0111 != 0 {
			hooks[alias] = Hook{
				Alias: alias,
				Path:  path,
			}
		}
	}

	res, err := maps.KeysSlice(hooks)
	if err != nil {
		logrus.Fatal(err)
	}
	hooksAliases := res.Interface().([]string)
	sort.Strings(hooksAliases)
	logrus.WithFields(logrus.Fields{
		"folder": hooksRoot,
		"hooks":  strings.Join(hooksAliases, ", "),
	}).Info("Hooks initialized")

	viper.Set("hooks", hooks)
}

func initKeyConfig() {
	path := viper.GetString("key")
	if path != "" {
		input, err := os.Open(path)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"file":  path,
				"error": err,
			}).Fatal("Failed to read private key!")
		}
		defer input.Close()

		key, err := pgp.LoadASCIIArmoredPrivateKey(input)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"file":  path,
				"error": err,
			}).Fatal("Failed to load private key!")
		}

		logrus.WithFields(logrus.Fields{
			"file": path,
		}).Info("Private key initialized")

		viper.Set("key", &key)
	} else {
		viper.Set("key", (*pgp.PrivateKey)(nil))
	}
}

func initPublicKeysConfig() {
	keys := make(map[string]*pgp.PublicKey)
	keysRoot := path.Join(GetRoot(), "keys")

	if info, err := os.Stat(keysRoot); os.IsNotExist(err) || !info.IsDir() {
		logrus.WithFields(logrus.Fields{
			"folder": keysRoot,
			"error":  err,
		}).Fatal("Failed to access to public keys folder!")
	}

	if err := filepath.Walk(
		keysRoot,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if filepath.Ext(path) == PublicKeyExtension {
				input, err := os.Open(path)
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"file":  path,
						"error": err,
					}).Fatal("Failed to read public key!")
				}
				defer input.Close()

				key, err := pgp.LoadASCIIArmoredPublicKey(
					strings.TrimSuffix(filepath.Base(path), PublicKeyExtension),
					input)
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"file":  path,
						"error": err,
					}).Fatal("Failed to load public key!")
				}
				keys[key.Alias] = &key
			}

			return nil
		}); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to load public keys!")
	}

	res, err := maps.KeysSlice(keys)
	if err != nil {
		logrus.Fatal(err)
	}
	keysAliases := res.Interface().([]string)
	sort.Strings(keysAliases)
	logrus.WithFields(logrus.Fields{
		"folder": keysRoot,
		"keys":   strings.Join(keysAliases, ", "),
	}).Info("Public keys initialized")

	viper.Set("keys", keys)
}

func initIdentity() {
	keyAlias := viper.GetString("identity")
	if keyAlias != "" {
		keys := GetPublicKeys()
		key, found := keys[keyAlias]
		if !found {
			logrus.WithFields(logrus.Fields{
				"key": keyAlias,
			}).Fatal("Unknown key used as identity!")
		}

		logrus.WithFields(logrus.Fields{
			"key": keyAlias,
		}).Info("Identity initialized")

		viper.Set("identity", key)
	} else {
		viper.Set("identity", (*pgp.PublicKey)(nil))
	}
}

func initSecretsConfig() {
	secretsRoot := path.Join(GetRoot(), "secrets")

	if info, err := os.Stat(secretsRoot); os.IsNotExist(err) || !info.IsDir() {
		logrus.WithFields(logrus.Fields{
			"folder": secretsRoot,
			"error":  err,
		}).Fatal("Failed to access to secrets folder!")
	}

	logrus.WithFields(logrus.Fields{
		"folder": secretsRoot,
	}).Info("Secrets folder initialized")

	viper.Set("secrets", secretsRoot)
}

func initKeepersConfig() {
	keepers := make([]*pgp.PublicKey, 0)
	aliases := make([]string, 0)

	keys := GetPublicKeys()
	for _, keyAlias := range viper.GetStringSlice("keepers") {
		key, found := keys[keyAlias]
		if !found {
			logrus.WithFields(logrus.Fields{
				"key": keyAlias,
			}).Fatal("Found an unknown keeper!")
		}
		keepers = append(keepers, key)
		aliases = append(aliases, key.Alias)
	}

	sort.Strings(aliases)
	logrus.WithFields(logrus.Fields{
		"keys": strings.Join(aliases, ", "),
	}).Info("Keepers initialized")
	viper.Set("keepers", keepers)
}

func initTeamsConfig() {
	teams := make(map[string]Team)

	keys := GetPublicKeys()
	for teamAlias, teamMapValue := range viper.GetStringMap("teams") {
		team := Team{
			Alias: teamAlias,
			Keys:  make([]*pgp.PublicKey, 0),
		}
		if teamMapValue != nil {
			teamMembers := teamMapValue.([]interface{})
			for _, keySliceValue := range teamMembers {
				keyAlias := keySliceValue.(string)
				key, found := keys[keyAlias]
				if !found {
					logrus.WithFields(logrus.Fields{
						"key":  keyAlias,
						"team": teamAlias,
					}).Fatal("Found unknown key in team!")
				}
				team.Keys = append(team.Keys, key)
			}
			teams[teamAlias] = team
		}
	}

	if _, found := teams["all"]; !found {
		team := Team{
			Alias: "all",
			Keys:  make([]*pgp.PublicKey, 0),
		}
		for _, key := range keys {
			team.Keys = append(team.Keys, key)
		}
		teams["all"] = team
	}

	res, err := maps.KeysSlice(teams)
	if err != nil {
		logrus.Fatal(err)
	}
	aliases := res.Interface().([]string)
	sort.Strings(aliases)
	logrus.WithFields(logrus.Fields{
		"teams": strings.Join(aliases, ", "),
	}).Info("Teams initialized")

	viper.Set("teams", teams)
}

func initTagsConfig() {
	schemaString := viper.GetString("tags")
	if schemaString == "" {
		schemaString = defaultTagsSchema
	}

	schemaLoader := gojsonschema.NewStringLoader(schemaString)
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to load tags schema!")
	}

	logrus.Info("Tags initialized")

	viper.Set("tags", schema)
}

func initPermissionRulesConfig() {
	rules := make([]PermissionRule, 0)

	if _, ok := viper.Get("permissions").([]interface{}); ok {
		keys := GetPublicKeys()
		teams := GetTeams()
		for _, itemSliceValue := range viper.Get("permissions").([]interface{}) {
			item := itemSliceValue.(map[interface{}]interface{})
			for queryStringMapKey, expressionsMapValue := range item {
				if expressionsMapValue != nil {
					queryString := queryStringMapKey.(string)
					expressions := expressionsMapValue.([]interface{})

					var rule PermissionRule

					queryParsed, err := query.Parse(queryString)
					if err != nil {
						logrus.WithFields(logrus.Fields{
							"query": queryString,
							"error": err,
						}).Fatal("Failed to parse permissions query!")
					}
					rule.Query = queryParsed

					rule.Expressions = make([]PermissionExpression, 0)
					for _, expressionStringSliceValue := range expressions {
						var expression PermissionExpression

						expressionString := expressionStringSliceValue.(string)

						expression.Deny = expressionString[0] == '-'

						expression.Keys = make([]*pgp.PublicKey, 0)
						subject := expressionString[1:]
						if key, found := keys[subject]; !found {
							if team, found := teams[subject]; !found {
								logrus.WithFields(logrus.Fields{
									"query":      queryString,
									"expression": expressionString,
								}).Fatal("Found unknown key or team in permissions expression!")
							} else {
								expression.Keys = team.Keys
							}
						} else {
							expression.Keys = append(expression.Keys, key)
						}

						rule.Expressions = append(rule.Expressions, expression)
					}

					rules = append(rules, rule)
					break
				}
			}
		}
	}

	logrus.WithFields(logrus.Fields{
		"len": len(rules),
	}).Info("Permission rules initialized")

	viper.Set("permissions", nil)
	viper.Set("permission-rules", rules)
}

func initTemplatesConfig() {
	templates := make(map[string]*Template)
	templatesRoot := path.Join(GetRoot(), "templates")

	if info, err := os.Stat(templatesRoot); os.IsNotExist(err) || !info.IsDir() {
		logrus.WithFields(logrus.Fields{
			"folder": templatesRoot,
			"error":  err,
		}).Fatal("Failed to access to templates folder!")
	}

	if err := filepath.Walk(
		templatesRoot,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			ext := filepath.Ext(path)

			if ext == TemplateSchemaExtension || ext == TemplateSkeletonExtension {
				var alias string
				if ext == TemplateSchemaExtension {
					alias = strings.TrimSuffix(filepath.Base(path), TemplateSchemaExtension)
				} else {
					alias = strings.TrimSuffix(filepath.Base(path), TemplateSkeletonExtension)
				}

				if _, found := templates[alias]; !found {
					templates[alias] = &Template{
						Alias:    alias,
						Schema:   nil,
						Skeleton: nil,
					}
				}

				if ext == TemplateSchemaExtension {
					schema, err := gojsonschema.NewSchema(gojsonschema.NewReferenceLoader("file://" + path))
					if err != nil {
						logrus.WithFields(logrus.Fields{
							"file":  path,
							"error": err,
						}).Fatal("Failed to load template schema!")
					}
					templates[alias].Schema = schema
				} else {
					skeleton, err := ioutil.ReadFile(path)
					if err != nil {
						logrus.WithFields(logrus.Fields{
							"file":  path,
							"error": err,
						}).Fatal("Failed to load template skeleton!")
					}
					templates[alias].Skeleton = skeleton
				}
			}

			return nil
		}); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to load templates!")
	}

	res, err := maps.KeysSlice(templates)
	if err != nil {
		logrus.Fatal(err)
	}
	templatesAliases := res.Interface().([]string)
	sort.Strings(templatesAliases)
	logrus.WithFields(logrus.Fields{
		"folder":    templatesRoot,
		"templates": strings.Join(templatesAliases, ", "),
	}).Info("Templates initialized")

	viper.Set("templates-rules", viper.Get("templates"))
	viper.Set("templates", templates)
}

func initTemplateRulesConfig() {
	rules := make([]TemplateRule, 0)

	if _, ok := viper.Get("templates-rules").([]interface{}); ok {
		templates := GetTemplates()
		for _, itemSliceValue := range viper.Get("templates-rules").([]interface{}) {
			item := itemSliceValue.(map[interface{}]interface{})
			for queryStringMapKey, templateAliasMapValue := range item {
				queryString := queryStringMapKey.(string)
				templateAlias := templateAliasMapValue.(string)

				var rule TemplateRule

				queryParsed, err := query.Parse(queryString)
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"query": queryString,
						"error": err,
					}).Fatal("Failed to parse permissions query!")
				}
				rule.Query = queryParsed

				template, found := templates[templateAlias]
				if !found {
					logrus.WithFields(logrus.Fields{
						"query":    queryString,
						"template": templateAlias,
					}).Fatal("Found unknown template in template rule!")
				}
				rule.Template = template

				rules = append(rules, rule)
				break
			}
		}
	}

	logrus.WithFields(logrus.Fields{
		"len": len(rules),
	}).Info("Template rules initialized")

	viper.Set("template-rules", rules)
}
