package secret

import (
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"

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

type DoesNotExist struct{}

func (self *DoesNotExist) Error() string {
	return "secret does not exist"
}

func New(uri string) *Secret {
	return &Secret{
		uri:  uri,
		tags: make([]Tag, 0),
		path: path.Join(config.GetSecretsRoot(), uri+config.SecretExtension),
	}
}

func Load(uri string) (*Secret, error) {
	result := New(uri)

	if info, err := os.Stat(result.path); os.IsNotExist(err) || info.IsDir() {
		return nil, &DoesNotExist{}
	}

	// This is required in order to populate tags.
	input, err := result.NewReader()
	if err != nil {
		return nil, errors.Wrap(err, "failed to open secret")
	}
	defer input.Close()

	return result, nil
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

	key := config.GetPrivateKey()
	if key == nil {
		if err := pgp.DecryptWithGPG(config.GetGPG(), input, output); err != nil {
			return errors.Wrap(err, "failed to decrypt secret")
		}
	} else {
		if err := pgp.Decrypt(config.GetGPGConnectAgent(), input, output, key); err != nil {
			return errors.Wrap(err, "failed to decrypt secret")
		}
	}

	return nil
}

func (self *Secret) GetExpectedPublicKeys() ([]*pgp.PublicKey, error) {
	result := make([]*pgp.PublicKey, 0)

	for _, rule := range config.GetPermissionRules() {
		if rule.Query.Eval(self) {
			for _, expression := range rule.Expressions {
				var tmp reflect.Value
				var err error
				if expression.Deny {
					tmp, err = slices.Difference(result, expression.Keys)
				} else {
					tmp, err = slices.Union(result, expression.Keys)
				}
				if err != nil {
					return nil, errors.Wrap(err, "unexpected error")
				}
				result = tmp.Interface().([]*pgp.PublicKey)
			}
		}
	}

	tmp, err := slices.Union(result, config.GetKeepers())
	if err != nil {
		return nil, errors.Wrap(err, "unexpected error")
	}
	result = tmp.Interface().([]*pgp.PublicKey)

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

func (self *Secret) GetTemplate() *config.Template {
	for _, rule := range config.GetTemplateRules() {
		if rule.Query.Eval(self) {
			return rule.Template
		}
	}

	return nil
}

func (self *Secret) GetRecipients() (expected, unknown, rubbish, missing []string, e error) {
	// Initializations.
	expected = make([]string, 0)
	current := make([]string, 0)
	unknown = make([]string, 0)
	rubbish = make([]string, 0)
	missing = make([]string, 0)
	e = nil

	// Extract current recipients.
	currentRecipientKeyIds, err := self.GetCurrentRecipientsKeyIds()
	if err != nil {
		e = errors.Wrap(err, "failed to determine current recipients")
		return
	}

	// Determine current & unknown recipients.
	for _, keyId := range currentRecipientKeyIds {
		key := findPublicKeyByKeyId(keyId)
		if key != nil {
			current = append(current, key.Alias)
		} else {
			unknown = append(
				unknown,
				fmt.Sprintf("0x%x", keyId))
		}
	}
	sort.Strings(unknown)

	// Determine expected recipients.
	if keys, err := self.GetExpectedPublicKeys(); err == nil {
		for _, key := range keys {
			expected = append(expected, key.Alias)
		}
		sort.Strings(expected)
	} else {
		e = errors.Wrap(err, "failed to determine expected recipients")
		return
	}

	// Determine rubbish recipients.
	if tmp, err := slices.Difference(current, expected); err == nil {
		rubbish = tmp.Interface().([]string)
		sort.Strings(rubbish)
	} else {
		e = errors.Wrap(err, "failed to determine rubbish recipients")
		return
	}

	// Determine missing recipients.
	if tmp, err := slices.Difference(expected, current); err == nil {
		missing = tmp.Interface().([]string)
		sort.Strings(missing)
	} else {
		e = errors.Wrap(err, "failed to determine missing recipients")
		return
	}

	// Done!
	return
}

func (self *Secret) IsReadableBy(key *pgp.PublicKey) (bool, error) {
	// Try expected recipients.
	if keys, err := self.GetExpectedPublicKeys(); err == nil {
		for _, aKey := range keys {
			if aKey == key {
				return true, nil
			}
		}
	} else {
		return false, errors.Wrap(err, "failed to determine expected recipients")
	}

	// Try current recipients.
	if currentRecipientKeyIds, err := self.GetCurrentRecipientsKeyIds(); err == nil {
		for _, keyId := range currentRecipientKeyIds {
			aKey := findPublicKeyByKeyId(keyId)
			if aKey != nil && aKey == key {
				return true, nil
			}
		}
	} else {
		return false, errors.Wrap(err, "failed to determine current recipients")
	}

	// Not readable.
	return false, nil
}

// Implementation of 'query.Context' interface.
func (self *Secret) GetIdentifier(key string) string {
	if key == "uri" {
		return self.uri
	} else if strings.HasPrefix(key, "tags.") {
		name := strings.ToLower(key[5:])
		for _, tag := range self.tags {
			if strings.ToLower(tag.Name) == name {
				return tag.Value
			}
		}
	}
	return ""
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
