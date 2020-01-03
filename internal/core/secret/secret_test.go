package secret

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/carlosabalde/pgp-tomb/internal/core/config"
)

func bootstrap(t *testing.T, configString string, keys []string) {
	var err error
	var root, wd string

	// Initializations.
	logrus.SetOutput(ioutil.Discard)
	root, err = ioutil.TempDir("", "pgp-tomb-test-")
	require.NoError(t, err)
	configFile := path.Join(root, "pgp-tomb.yaml")
	wd, err = os.Getwd()
	require.NoError(t, err)

	// Create required folders.
	for _, folder := range []string{"hooks", "keys", "secrets", "templates"} {
		err = os.Mkdir(path.Join(root, folder), os.ModePerm)
		require.NoError(t, err)
	}

	// Dump configuration.
	err = ioutil.WriteFile(
		configFile,
		[]byte(fmt.Sprintf(
			"key: %s\n%s",
			path.Join(wd, "testdata", "alice.unencrypted.pri"),
			configString)),
		0644)
	require.NoError(t, err)

	// Link public keys in <working directory>/testdata/.
	for _, key := range keys {
		err = os.Symlink(
			path.Join(wd, "testdata", key),
			path.Join(root, "keys", key))
		require.NoError(t, err)
	}

	// Load & initialize configuration.
	viper.Reset()
	viper.SetConfigType("yaml")
	var file *os.File
	file, err = os.Open(configFile)
	require.NoError(t, err)
	err = viper.ReadConfig(file)
	require.NoError(t, err)
	viper.Set("root", root)
	config.Init(configFile)
}

func TestEncryptAndDecrypt(t *testing.T) {
	bootstrap(t, `
keepers:
  - alice
`, []string{"alice.pub", "bob.pub"})
	defer os.RemoveAll(config.GetRoot())

	message := "Hello world! 🤠"

	decrypt := func(uri string) {
		secret, err := Load(uri)
		if assert.NoError(t, err) {
			output := new(bytes.Buffer)
			err := secret.Decrypt(output)
			if assert.NoError(t, err) {
				assert.Equal(t, output.String(), message)
			}
		}
	}

	encrypt := func(uri string) {
		secret := New(uri)
		err := secret.Encrypt(bytes.NewBufferString(message))
		if assert.NoError(t, err) {
			decrypt(uri)
		}
	}

	encrypt("foo/bar/baz.txt")
}
