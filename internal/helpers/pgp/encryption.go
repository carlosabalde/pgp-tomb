package pgp

import (
	"io"

	"github.com/pkg/errors"
	"golang.org/x/crypto/openpgp"
)

func Encrypt(input io.Reader, output io.Writer, keys []*PublicKey) error {
	entities := make([]*openpgp.Entity, 0)
	for _, key := range keys {
		entities = append(entities, key.Entity)
	}

	hints := openpgp.FileHints{IsBinary: true}
	plain, err := openpgp.Encrypt(output, entities, nil, &hints, nil)
	if err != nil {
		return errors.Wrap(err, "PGP encryption failed")
	}
	defer plain.Close()

	if _, err := io.Copy(plain, input); err != nil {
		return errors.Wrap(err, "PGP encryption failed")
	}

	return nil
}
