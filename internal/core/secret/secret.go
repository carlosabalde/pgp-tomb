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
	uri  string
	tags []Tag
	path string
}

type Tag struct {
	Name  string
	Value string
}

func New(uri string) *Secret {
	return &Secret{
		uri:  uri,
		tags: make([]Tag, 0),
		path: path.Join(config.GetSecretsRoot(), uri+config.SecretExtension),
	}
}

func (self *Secret) Exists() bool {
	if info, err := os.Stat(self.path); os.IsNotExist(err) || info.IsDir() {
		return false
	}
	return true
}

func (self *Secret) GetUri() string {
	return self.uri
}

func (self *Secret) GetTags() []Tag {
	return self.tags
}

func (self *Secret) SetTags(tags []Tag) {
	self.tags = tags
}

func (self *Secret) GetPath() string {
	return self.path
}

func (self *Secret) Encrypt(input io.Reader) error {
	output, err := self.NewWriter()
	if err != nil {
		return errors.Wrap(err, "failed to open secret")
	}
	defer output.Close()

	keys, err := self.GetExpectedPublicKeys()
	if err != nil {
		return errors.Wrap(err, "failed to get expected public keys")
	}

	if err := pgp.Encrypt(input, output, keys); err != nil {
		return errors.Wrap(err, "failed to encrypt secret")
	}

	return nil
}

func (self *Secret) Decrypt(output io.Writer) error {
	input, err := self.NewReader()
	if err != nil {
		return errors.Wrap(err, "failed to open secret")
	}
	defer input.Close()

	if err := pgp.DecryptWithGPG(config.GetGPG(), input, output); err != nil {
		return errors.Wrap(err, "failed to decrypt secret")
	}

	return nil
}

func (self *Secret) GetExpectedPublicKeys() ([]*pgp.PublicKey, error) {
	// Build list of key aliases according to the configured permissions &
	// keepers.
	permissions := config.GetPermissions()
	aliases := make([]string, 0)
	for _, permission := range permissions {
		if permission.Regexp.Match([]byte(self.uri)) {
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

func (self *Secret) GetCurrentRecipientsKeyIds() ([]uint64, error) {
	input, err := self.NewReader()
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
