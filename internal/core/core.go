package core

import (
	"os"
	"path"
	"reflect"

	"github.com/sirupsen/logrus"

	"github.com/carlosabalde/pgp-tomb/internal/core/config"
	"github.com/carlosabalde/pgp-tomb/internal/helpers/pgp"
	"github.com/carlosabalde/pgp-tomb/internal/helpers/slices"
)

func getPathForSecret(uri string) string {
	return path.Join(config.GetSecretsRoot(), uri+config.SecretExtension)
}

func findSecret(uri string) string {
	path := getPathForSecret(uri)
	if info, err := os.Stat(path); os.IsNotExist(err) || info.IsDir() {
		return ""
	}
	return path
}

func getPublicKeysForSecret(uri string) []*pgp.PublicKey {
	// Build list of key aliases according to the configured permissions &
	// keepers.
	permissions := config.GetPermissions()
	aliases := make([]string, 0)
	for _, permission := range permissions {
		if permission.Regexp.Match([]byte(uri)) {
			for _, expression := range permission.Expressions {
				var tmp reflect.Value
				var err error
				if expression.Deny {
					tmp, err = slices.Difference(
						aliases,
						expression.Keys)
				} else {
					tmp, err = slices.Union(
						aliases,
						expression.Keys)
				}
				if err != nil {
					logrus.Fatal(err)
				}
				aliases = tmp.Interface().([]string)
			}
		}
	}
	tmp, err := slices.Union(aliases, config.GetKeepers())
	if err != nil {
		logrus.Fatal(err)
	}
	aliases = tmp.Interface().([]string)

	// Expand key aliases to full 'pgp.PublicKey' instances.
	keys := config.GetPublicKeys()
	result := make([]*pgp.PublicKey, len(aliases))
	for i, alias := range aliases {
		result[i] = keys[alias]
	}
	return result
}

func findPublicKey(alias string) *pgp.PublicKey {
	keys := config.GetPublicKeys()
	if key, found := keys[alias]; found {
		return key
	}
	return nil
}

func findPublicKeyByKeyId(id uint64) *pgp.PublicKey {
	for _, key := range config.GetPublicKeys() {
		if key.Entity.PrimaryKey.KeyId == id {
			return key
		}
		for _, subkey := range key.Entity.Subkeys {
			if subkey.PublicKey.KeyId == id {
				return key
			}
		}
	}
	return nil
}
