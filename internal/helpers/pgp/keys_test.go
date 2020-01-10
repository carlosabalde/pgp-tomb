package pgp

/*
	$ mkdir /tmp/gnupg/
	$ gpg --homedir /tmp/gnupg --import /mnt/files/keys/*.pub
	$ gpg --homedir /tmp/gnupg --yes --encrypt --compress-algo 1 \
		--output testdata/message.pgp \
		--recipient alice@example.com \
		--recipient bob@example.com \
		--recipient chuck@example.com \
		/mnt/files/plain/answers.md
	$ gpg --homedir /tmp/gnupg --yes --encrypt --armor --compress-algo 1 \
		--output testdata/message.pgp \
		--recipient alice@example.com \
		--recipient bob@example.com \
		--recipient chuck@example.com \
		/mnt/files/plain/answers.md
	$ rm -rf /tmp/gnupg/

	$ cp /mnt/files/keys/alice.pri testdata/private-key.asc
	$ gpg --dearmor testdata/private-key.asc

	$ cp /mnt/files/keys/alice.pub testdata/public-key.asc
	$ gpg --dearmor testdata/public-key.asc
*/

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadASCIIArmoredPrivateKey(t *testing.T) {
	file1, err1 := os.Open("testdata/private-key.asc")
	if assert.NoError(t, err1) {
		key, err := LoadASCIIArmoredPrivateKey(file1)
		if assert.NoError(t, err) {
			assert.Equal(t, key.Entity.PrivateKey.Encrypted, true)
			assert.Equal(t, key.Entity.PrimaryKey.KeyId,
				uint64(16811708897216135528))
			assert.Equal(t, key.Entity.PrivateKey.PublicKey.KeyId,
				uint64(16811708897216135528))
			assert.Equal(t, key.Entity.Subkeys[0].PublicKey.KeyId,
				uint64(14417399214119891650))
			assert.Equal(t, key.Entity.Subkeys[1].PublicKey.KeyId,
				uint64(9978118131145701538))
		}
	}

	file2, err2 := os.Open("testdata/private-key.pgp")
	if assert.NoError(t, err2) {
		_, err := LoadASCIIArmoredPrivateKey(file2)
		assert.Error(t, err)
	}

	file3, err3 := os.Open("testdata/public-key.asc")
	if assert.NoError(t, err3) {
		_, err := LoadASCIIArmoredPrivateKey(file3)
		assert.Error(t, err)
	}
}

func TestLoadASCIIArmoredPublicKey(t *testing.T) {
	file1, err1 := os.Open("testdata/public-key.asc")
	if assert.NoError(t, err1) {
		key, err := LoadASCIIArmoredPublicKey("key1", file1)
		if assert.NoError(t, err) {
			assert.Equal(t, key.Entity.PrimaryKey.KeyId,
				uint64(16811708897216135528))
			assert.Nil(t, key.Entity.PrivateKey)
			assert.Equal(t, key.Entity.Subkeys[0].PublicKey.KeyId,
				uint64(14417399214119891650))
			assert.Equal(t, key.Entity.Subkeys[1].PublicKey.KeyId,
				uint64(9978118131145701538))
		}
	}

	file2, err2 := os.Open("testdata/public-key.pgp")
	if assert.NoError(t, err2) {
		_, err := LoadASCIIArmoredPublicKey("key2", file2)
		assert.Error(t, err)
	}

	file3, err3 := os.Open("testdata/private-key.asc")
	if assert.NoError(t, err3) {
		_, err := LoadASCIIArmoredPublicKey("key3", file3)
		assert.Error(t, err)
	}
}

func TestGetRecipientKeyIdsForEncryptedMessage(t *testing.T) {
	file1, err1 := os.Open("testdata/message.pgp")
	if assert.NoError(t, err1) {
		keyIds, err := GetRecipientKeyIdsForEncryptedMessage(file1)
		if assert.NoError(t, err) {
			assert.ElementsMatch(
				t,
				keyIds,
				[]uint64{
					uint64(14417399214119891650),
					uint64(1763358573985919020),
					uint64(4547273708913783244),
				})
		}
	}

	file2, err2 := os.Open("testdata/message.asc")
	if assert.NoError(t, err2) {
		_, err := GetRecipientKeyIdsForEncryptedMessage(file2)
		assert.Error(t, err)
	}
}
