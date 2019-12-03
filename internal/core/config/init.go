package config

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/xeipuuv/gojsonschema"

	"github.com/carlosabalde/pgp-tomb/internal/core/query"
	"github.com/carlosabalde/pgp-tomb/internal/helpers/maps"
	"github.com/carlosabalde/pgp-tomb/internal/helpers/pgp"
)

func Init() {
	initRootConfig()
	initGPGConfig()
	initEditorConfig()
	initPublicKeysConfig()
	initSecretsConfig()
	initKeepersConfig()
	initTeamsConfig()
	initPermissionRulesConfig()
	initTemplatesConfig()
	initTemplateRulesConfig()
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
	executable, err := exec.LookPath("gpg")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to locate GPG executable!")
	}

	logrus.WithFields(logrus.Fields{
		"executable": executable,
	}).Info("GPG initialized")

	viper.Set("gpg", executable)
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

func initPublicKeysConfig() {
	keys := make(map[string]*pgp.PublicKey)
	keysRoot := path.Join(viper.GetString("root"), "keys")

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
						"path":  path,
						"error": err,
					}).Fatal("Failed to load public key!")
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

func initSecretsConfig() {
	secretsRoot := path.Join(viper.GetString("root"), "secrets")

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
	rawKeepers := viper.GetStringSlice("keepers")

	if len(rawKeepers) < 1 {
		logrus.Fatal("At least one keeper is required!")
	}

	keepers := make([]*pgp.PublicKey, 0)
	aliases := make([]string, 0)

	keys := GetPublicKeys()
	for _, keyAlias := range rawKeepers {
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
				if keySliceValue != nil {
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
			}
		}
		teams[teamAlias] = team
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

						if expressionString[0] != '+' && expressionString[0] != '-' {
							logrus.WithFields(logrus.Fields{
								"query":      queryString,
								"expression": expressionString,
							}).Fatal("Found invalid permissions expression!")
						}

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
	templatesRoot := path.Join(viper.GetString("root"), "templates")

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

			if filepath.Ext(path) == TemplateExtension {
				alias := strings.TrimSuffix(filepath.Base(path), TemplateExtension)
				schema, err := gojsonschema.NewSchema(gojsonschema.NewReferenceLoader("file://" + path))
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"file":  path,
						"error": err,
					}).Fatal("Failed to load template!")
				}
				templates[alias] = &Template{
					Alias:  alias,
					Schema: schema,
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
				if templateAliasMapValue != nil {
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
	}

	logrus.WithFields(logrus.Fields{
		"len": len(rules),
	}).Info("Template rules initialized")

	viper.Set("template-rules", rules)
}
