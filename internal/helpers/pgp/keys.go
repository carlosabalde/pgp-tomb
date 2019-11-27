package pgp

import (
	"io"

	"github.com/pkg/errors"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

type PublicKey struct {
	Alias  string
	Entity *openpgp.Entity
}

func LoadASCIIArmoredPublicKey(alias string, input io.Reader) (*PublicKey, error) {
	block, err := armor.Decode(input)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode ASCII armor of key '%s'", alias)
	}

	if block.Type != openpgp.PublicKeyType {
		return nil, errors.Errorf("invalid public key '%s'", alias)
	}

	reader := packet.NewReader(block.Body)

	entity, err := openpgp.ReadEntity(reader)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create entity from public key '%s'", alias)
	}

	return &PublicKey{alias, entity}, nil
}

func GetRecipientKeyIdsForEncryptedMessage(input io.Reader) ([]uint64, error) {
	result := make([]uint64, 0)

	// Strongly based on openpgp.ReadMessage().
	packets := packet.NewReader(input)
	for {
		item, err := packets.Next()
		if err != nil {
			return nil, err
		}
		switch packet := item.(type) {
		case *packet.EncryptedKey:
			result = append(result, packet.KeyId)
		default:
			return result, nil
			break
		}
	}

	return result, nil
}
