package config

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/carlosabalde/pgp-tomb/internal/helpers/maps"
	"github.com/carlosabalde/pgp-tomb/internal/helpers/pgp"
)

func Init(root string) {
	initRootConfig(root)
	initGPGConfig()
	initEditorConfig()
	initPublicKeysConfig()
	initSecretsConfig()
	initKeepersConfig()
	initTeamsConfig()
	initPermissionsConfig()
}

func initRootConfig(root string) {
	if !viper.IsSet("root") || viper.GetString("root") == "" {
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
				keys[key.Alias] = key
			}

			return nil
		}); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to load public keys!")
	}

	keysAliases := maps.StringKeysSlice(keys)
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
	keepers := viper.GetStringSlice("keepers")

	if len(keepers) < 1 {
		logrus.Fatal("At least one keeper is required!")
	}

	keys := GetPublicKeys()
	for _, keyAlias := range keepers {
		if _, found := keys[keyAlias]; !found {
			logrus.WithFields(logrus.Fields{
				"key": keyAlias,
			}).Fatal("Found an unknown keeper!")
		}
	}

	sort.Strings(keepers)
	logrus.WithFields(logrus.Fields{
		"keys": strings.Join(keepers, ", "),
	}).Info("Keepers initialized")
}

func initTeamsConfig() {
	teams := make(map[string][]string)

	keys := GetPublicKeys()
	for teamId, teamMapValue := range viper.GetStringMap("teams") {
		teams[teamId] = make([]string, 0)
		if teamMapValue != nil {
			teamMembers := teamMapValue.([]interface{})
			for _, keySliceValue := range teamMembers {
				if keySliceValue != nil {
					keyAlias := keySliceValue.(string)
					if _, found := keys[keyAlias]; !found {
						logrus.WithFields(logrus.Fields{
							"key":  keyAlias,
							"team": teamId,
						}).Fatal("Found unknown key in team!")
					}
					teams[teamId] = append(teams[teamId], keyAlias)
				}
			}
		}
	}

	teamsIds := maps.StringKeysSlice(teams)
	sort.Strings(teamsIds)
	logrus.WithFields(logrus.Fields{
		"teams": strings.Join(teamsIds, ", "),
	}).Info("Teams initialized")

	viper.Set("teams", teams)
}

func initPermissionsConfig() {
	permissions := make([]Permission, 0)

	if _, ok := viper.Get("permissions").([]interface{}); ok {
		keys := GetPublicKeys()
		teams := GetTeams()
		for _, itemSliceValue := range viper.Get("permissions").([]interface{}) {
			item := itemSliceValue.(map[interface{}]interface{})
			for regexpStringMapKey, expressionsMapValue := range item {
				if expressionsMapValue != nil {
					regexpString := regexpStringMapKey.(string)
					expressions := expressionsMapValue.([]interface{})

					var permission Permission

					regexp, err := regexp.Compile(regexpString)
					if err != nil {
						logrus.WithFields(logrus.Fields{
							"regexp": regexpString,
							"error":  err,
						}).Fatal("Failed to compile permissions regexp!")
					}
					permission.Regexp = regexp

					permission.Expressions = make([]PermissionExpression, 0)
					for _, expressionStringSliceValue := range expressions {
						var expression PermissionExpression

						expressionString := expressionStringSliceValue.(string)

						if expressionString[0] != '+' && expressionString[0] != '-' {
							logrus.WithFields(logrus.Fields{
								"regexp":     regexpString,
								"expression": expressionString,
							}).Fatal("Found invalid permissions expression!")
						}

						expression.Deny = expressionString[0] == '-'

						expression.Keys = make([]string, 0)
						subject := expressionString[1:]
						if _, found := keys[subject]; !found {
							if keys, found := teams[subject]; !found {
								logrus.WithFields(logrus.Fields{
									"regexp":     regexpString,
									"expression": expressionString,
								}).Fatal("Found unknown key or team in permissions expression!")
							} else {
								expression.Keys = keys
							}
						} else {
							expression.Keys = append(expression.Keys, subject)
						}

						permission.Expressions = append(permission.Expressions, expression)
					}

					permissions = append(permissions, permission)
					break
				}
			}
		}
	}

	logrus.WithFields(logrus.Fields{
		"len": len(permissions),
	}).Info("Permissions initialized")

	viper.Set("permissions", permissions)
}
