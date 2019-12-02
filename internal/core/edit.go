package core

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"

	"github.com/carlosabalde/pgp-tomb/internal/core/config"
	"github.com/carlosabalde/pgp-tomb/internal/core/secret"
)

func Edit(uri string, dropTags bool, tags []secret.Tag) {
	// Initialize output writer.
	output, err := ioutil.TempFile("", "pgp-tomb-")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to open temporary file!")
	}
	defer os.Remove(output.Name())

	// Try to load secret.
	s, err := secret.Load(uri)
	switch err := err.(type) {
	case nil:
		if err := s.Decrypt(output); err != nil {
			fmt.Fprintln(
				os.Stderr,
				"Unable to decrypt secret! Are you allowed to access it?")
			os.Exit(1)
		}
		output.Close()
	case *secret.DoesNotExist:
	default:
		logrus.WithFields(logrus.Fields{
			"error": err,
			"uri":   uri,
		}).Error("Failed to load secret!")
		return
	}

	// Compute initial digest.
	digest := md5File(output.Name())

	// Open decrypted secret in an external editor.
	cmd := exec.Command(config.GetEditor(), output.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.WithFields(logrus.Fields{
			"editor": config.GetEditor(),
			"error":  err,
		}).Fatal("Failed to open external editor!")
	}

	// Decide new tags.
	var updateTags bool
	var newTags []secret.Tag
	if dropTags {
		updateTags = true
		newTags = tags
	} else {
		updateTags = false
		newTags = s.GetTags()
	}

	// Check if secret needs to be updated.
	if updateTags || digest != md5File(output.Name()) {
		Set(uri, output.Name(), newTags)
	} else {
		fmt.Println("No changes!")
	}
}

func md5File(path string) string {
	file, err := os.Open(path)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"path":  path,
		}).Fatal("Failed to open file!")
	}
	defer file.Close()

	digest := md5.New()
	if _, err := io.Copy(digest, file); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"path":  path,
		}).Fatal("Failed to MD5 file!")
	}

	return fmt.Sprintf("%x", digest.Sum(nil))
}
