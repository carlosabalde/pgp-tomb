package core

import (
	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
	"github.com/xeipuuv/gojsonschema"

	"github.com/carlosabalde/pgp-tomb/internal/core/config"
	"github.com/carlosabalde/pgp-tomb/internal/core/query"
	"github.com/carlosabalde/pgp-tomb/internal/helpers/pgp"
)

func findPublicKey(alias string) *pgp.PublicKey {
	keys := config.GetPublicKeys()
	if key, found := keys[alias]; found {
		return key
	}
	return nil
}

func parseQuery(queryString string) (result query.Query) {
	if queryString != "" {
		var err error
		result, err = query.Parse(queryString)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"query": queryString,
				"error": err,
			}).Fatal("Failed to parse query!")
		}
	} else {
		result = query.True
	}
	return
}

func validateSchema(value string, schema *gojsonschema.Schema) (bool, []string) {
	errors := make([]string, 0)

	// This will succeed if 'value' is valid JSON since JSON is a subset of
	// YAML 1.2!
	jsonValue, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return false, errors
	}

	loader := gojsonschema.NewStringLoader(string(jsonValue))
	validation, err := schema.Validate(loader)
	if err != nil || !validation.Valid() {
		if err == nil {
			for _, err := range validation.Errors() {
				errors = append(errors, err.String())
			}
		}
		return false, errors
	}

	return true, errors
}
