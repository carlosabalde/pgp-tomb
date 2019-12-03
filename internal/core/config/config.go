package config

import (
	"github.com/spf13/viper"
	"github.com/xeipuuv/gojsonschema"

	"github.com/carlosabalde/pgp-tomb/internal/core/query"
	"github.com/carlosabalde/pgp-tomb/internal/helpers/pgp"
)

var (
	version  string
	revision string
)

const PublicKeyExtension = ".pub"
const SecretExtension = ".secret"
const TemplateExtension = ".template"
const DefaultEditor = "vim"

type PermissionRule struct {
	Query       query.Query
	Expressions []PermissionExpression
}

type PermissionExpression struct {
	Deny bool
	Keys []string
}

type TemplateRule struct {
	Query    query.Query
	Template string
}

func GetVersion() string {
	return version
}

func GetRevision() string {
	return revision
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

func GetPermissionRules() []PermissionRule {
	return viper.Get("permission-rules").([]PermissionRule)
}

func GetTemplates() map[string]*gojsonschema.Schema {
	return viper.Get("templates").(map[string]*gojsonschema.Schema)
}

func GetTemplateRules() []TemplateRule {
	return viper.Get("template-rules").([]TemplateRule)
}

func GetGPG() string {
	return viper.GetString("gpg")
}

func GetEditor() string {
	return viper.GetString("editor")
}
