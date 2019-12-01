package core

import (
	"github.com/carlosabalde/pgp-tomb/internal/core/config"
	"github.com/carlosabalde/pgp-tomb/internal/helpers/pgp"
)

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
