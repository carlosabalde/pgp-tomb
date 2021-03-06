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

const HookExtension = ".hook"
const PublicKeyExtension = ".pub"
const SecretExtension = ".secret"
const TemplateSchemaExtension = ".schema"
const TemplateSkeletonExtension = ".skeleton"

const DefaultEditor = "vim"

type Hook struct {
	Alias string
	Path  string
}

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
	Alias    string
	Schema   *gojsonschema.Schema
	Skeleton []byte
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

func GetRoot() string {
	return viper.GetString("root")
}

func GetHooks() map[string]Hook {
	return viper.Get("hooks").(map[string]Hook)
}

func GetIdentity() *pgp.PublicKey {
	return viper.Get("identity").(*pgp.PublicKey)
}

func GetPrivateKey() *pgp.PrivateKey {
	return viper.Get("key").(*pgp.PrivateKey)
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

func GetTags() *gojsonschema.Schema {
	return viper.Get("tags").(*gojsonschema.Schema)
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

func GetGPGConnectAgent() string {
	return viper.GetString("gpg-connect-agent")
}

func GetEditor() string {
	return viper.GetString("editor")
}
