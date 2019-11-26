package config

import (
	"regexp"

	"github.com/spf13/viper"

	"github.com/carlosabalde/pgp-tomb/internal/helpers/pgp"
)

const PublicKeyExtension = ".pub"
const SecretExtension = ".pgp"
const DefaultEditor = "vim"

type Permission struct {
	Regexp      *regexp.Regexp
	Expressions []PermissionExpression
}

type PermissionExpression struct {
	Deny bool
	Keys []string
}

func GetPublicKeys() map[string]*pgp.PublicKey {
	return viper.Get("keys").(map[string]*pgp.PublicKey)
}

func GetSecretsRoot() string {
	return viper.GetString("secrets")
}

func GetKeepers() []string {
	return viper.GetStringSlice("keepers")
}

func GetTeams() map[string][]string {
	return viper.Get("teams").(map[string][]string)
}

func GetPermissions() []Permission {
	return viper.Get("permissions").([]Permission)
}

func GetGPG() string {
	return viper.GetString("gpg")
}

func GetEditor() string {
	return viper.GetString("editor")
}
