package secret

import (
	"io"
	"os"
	"path"
	"reflect"

	"github.com/pkg/errors"

	"github.com/carlosabalde/pgp-tomb/internal/core/config"
	"github.com/carlosabalde/pgp-tomb/internal/helpers/pgp"
	"github.com/carlosabalde/pgp-tomb/internal/helpers/slices"
)

type Secret struct {
	uri     string
	headers []Header
	path    string
}

type Header struct {
	Name  string
	Value string
}

func New(uri string) *Secret {
	return &Secret{
		uri:     uri,
		headers: make([]Header, 0),
		path:    path.Join(config.GetSecretsRoot(), uri+config.SecretExtension),
	}
}

func (secret *Secret) Exists() bool {
	if info, err := os.Stat(secret.path); os.IsNotExist(err) || info.IsDir() {
		return false
	}
	return true
}

func (secret *Secret) GetUri() string {
	return secret.uri
}

func (secret *Secret) GetPath() string {
	return secret.path
}

func (secret *Secret) Encrypt(input io.Reader) error {
	output, err := secret.NewWriter()
	if err != nil {
		return errors.Wrap(err, "failed to open secret")
	}
	defer output.Close()

	keys, err := secret.GetExpectedPublicKeys()
	if err != nil {
		return errors.Wrap(err, "failed to get expected public keys")
	}

	if err := pgp.Encrypt(input, output, keys); err != nil {
		return errors.Wrap(err, "failed to encrypt secret")
	}

	return nil
}

func (secret *Secret) Decrypt(output io.Writer) error {
	input, err := secret.NewReader()
	if err != nil {
		return errors.Wrap(err, "failed to open secret")
	}
	defer input.Close()

	if err := pgp.DecryptWithGPG(config.GetGPG(), input, output); err != nil {
		return errors.Wrap(err, "failed to decrypt secret")
	}

	return nil
}

func (secret *Secret) GetExpectedPublicKeys() ([]*pgp.PublicKey, error) {
	// Build list of key aliases according to the configured permissions &
	// keepers.
	permissions := config.GetPermissions()
	aliases := make([]string, 0)
	for _, permission := range permissions {
		if permission.Regexp.Match([]byte(secret.uri)) {
			for _, expression := range permission.Expressions {
				var tmp reflect.Value
				var err error
				if expression.Deny {
					tmp, err = slices.Difference(aliases, expression.Keys)
				} else {
					tmp, err = slices.Union(aliases, expression.Keys)
				}
				if err != nil {
					return nil, errors.Wrap(err, "unexpected error")
				}
				aliases = tmp.Interface().([]string)
			}
		}
	}
	tmp, err := slices.Union(aliases, config.GetKeepers())
	if err != nil {
		return nil, errors.Wrap(err, "unexpected error")
	}
	aliases = tmp.Interface().([]string)

	// Expand key aliases to full 'pgp.PublicKey' instances.
	keys := config.GetPublicKeys()
	result := make([]*pgp.PublicKey, len(aliases))
	for i, alias := range aliases {
		result[i] = keys[alias]
	}
	return result, nil
}

func (secret *Secret) GetCurrentRecipientsKeyIds() ([]uint64, error) {
	input, err := secret.NewReader()
	if err != nil {
		return nil, errors.Wrap(err, "failed to open secret")
	}
	defer input.Close()

	ids, err := pgp.GetRecipientKeyIdsForEncryptedMessage(input)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get current recipients")
	}

	return ids, nil
}
