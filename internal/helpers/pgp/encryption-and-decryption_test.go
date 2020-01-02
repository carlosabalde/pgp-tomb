package pgp

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptAndDecrypt(t *testing.T) {
	secret := "Hello world!"

	decrypt := func(input io.Reader) {
		privateKeyFile, err := os.Open("testdata/private-key.unencrypted.asc")
		if assert.NoError(t, err) {
			privateKey, err := LoadASCIIArmoredPrivateKey(privateKeyFile)
			if assert.NoError(t, err) {
				assert.Equal(t, privateKey.Entity.PrivateKey.Encrypted, false)
				output := new(bytes.Buffer)
				err := Decrypt("dummy-agent", input, output, &privateKey)
				if assert.NoError(t, err) {
					assert.Equal(t, output.String(), secret)
				}
			}
		}
	}

	encrypt := func() {
		publicKeyFile, err := os.Open("testdata/public-key.asc")
		if assert.NoError(t, err) {
			publicKey, err := LoadASCIIArmoredPublicKey("key", publicKeyFile)
			if assert.NoError(t, err) {
				output := new(bytes.Buffer)
				err := Encrypt(
					bytes.NewBufferString(secret),
					output,
					[]*PublicKey{ &publicKey })
				if assert.NoError(t, err) {
					decrypt(bytes.NewBufferString(output.String()))
				}
			}
		}
	}

	encrypt()
}
