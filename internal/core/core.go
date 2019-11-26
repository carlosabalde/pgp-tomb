package core

import (
	"os"
	"path"

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
				if expression.Deny {
					aliases = slices.Difference(
						aliases,
						expression.Keys).Interface().([]string)
				} else {
					aliases = slices.Union(
						aliases,
						expression.Keys).Interface().([]string)
				}
			}
		}
	}
	aliases = slices.Union(aliases, config.GetKeepers()).Interface().([]string)

	// Expand key aliases to full 'pgp.PublicKey' instances.
	keys := config.GetPublicKeys()
	result := make([]*pgp.PublicKey, 0)
	for _, alias := range aliases {
		result = append(result, keys[alias])
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
