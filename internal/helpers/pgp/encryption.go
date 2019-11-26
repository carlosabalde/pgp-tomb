package pgp

import (
	"io"

	"golang.org/x/crypto/openpgp"
)

func Encrypt(input io.Reader, output io.Writer, keys []*PublicKey) error {
	entities := make([]*openpgp.Entity, 0)
	for _, key := range keys {
		entities = append(entities, key.Entity)
	}

	plain, err := openpgp.Encrypt(output, entities, nil, nil, nil)
	if err != nil {
		return err
	}
	defer plain.Close()
	io.Copy(plain, input)

	return nil
}
