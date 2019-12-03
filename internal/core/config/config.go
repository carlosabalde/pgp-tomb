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

type Team struct {
	Alias string
	Keys  []*pgp.PublicKey
}

type PermissionRule struct {
	Query       query.Query
	Expressions []PermissionExpression
}

type PermissionExpression struct {
	Deny bool
	Keys []*pgp.PublicKey
}

type Template struct {
	Alias  string
	Schema *gojsonschema.Schema
}

type TemplateRule struct {
	Query    query.Query
	Template *Template
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

func GetKeepers() []*pgp.PublicKey {
	return viper.Get("keepers").([]*pgp.PublicKey)
}

func GetTeams() map[string]Team {
	return viper.Get("teams").(map[string]Team)
}

func GetPermissionRules() []PermissionRule {
	return viper.Get("permission-rules").([]PermissionRule)
}

func GetTemplates() map[string]*Template {
	return viper.Get("templates").(map[string]*Template)
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
