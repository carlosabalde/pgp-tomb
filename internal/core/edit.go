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

func Edit(uri string, dropTags bool, tags []secret.Tag, ignoreSchema bool) {
	// Initialize output writer.
	output, err := ioutil.TempFile("", "pgp-tomb-")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to open temporary file!")
	}
	defer os.Remove(output.Name())

	// Try to load secret / dump initial skeleton.
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
		s = secret.New(uri)
		s.SetTags(tags)
		if template := s.GetTemplate(); template != nil {
			if err := ioutil.WriteFile(output.Name(), template.Skeleton, 0644); err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
					"uri":   uri,
				}).Fatal("Failed to dump skeleton!")
			}
		}
	default:
		logrus.WithFields(logrus.Fields{
			"error": err,
			"uri":   uri,
		}).Fatal("Failed to load secret!")
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

	// Compute initial digest.
	digest := md5File(output.Name())

	// Avoid loosing edited changes.
	for {
		// Open decrypted secret / initial skeleton in an external editor.
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

		// Check if secret needs to be updated.
		if updateTags || digest != md5File(output.Name()) {
			if set(uri, output.Name(), newTags, ignoreSchema) {
				fmt.Println("Done!")
				break
			} else {
				fmt.Print("\nReopen editor? (Y/n) ")
				var response string
				_, err := fmt.Scanln(&response)
				if err == nil && response == "n" {
					fmt.Println("Aborted!")
					break
				}
			}
		} else {
			fmt.Println("No changes!")
			break
		}
	}
}

func md5File(path string) string {
	file, err := os.Open(path)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"file":  path,
		}).Fatal("Failed to open file!")
	}
	defer file.Close()

	digest := md5.New()
	if _, err := io.Copy(digest, file); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"file":  path,
		}).Fatal("Failed to MD5 file!")
	}

	return fmt.Sprintf("%x", digest.Sum(nil))
}
