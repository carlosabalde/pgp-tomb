package core

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

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
	case *secret.DoesNotExist:
		s = secret.New(uri)
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
	output.Close()

	// Compute initial digests.
	tagsDigest := md5Tags(s)
	secretDigest := md5File(output.Name())

	// Adjust tags?
	if dropTags {
		s.SetTags(make([]secret.Tag, 0))
	} else if len(tags) > 0 {
		s.SetTags(tags)
	}

	// Avoid loosing edited changes.
loop:
	for {
		openEditor(output.Name())
		if tagsDigest != md5Tags(s) || secretDigest != md5File(output.Name()) {
			if set(s, output.Name(), ignoreSchema) {
				fmt.Println("Done!")
				break
			} else {
				for {
					fmt.Print("\nWhat now? edit [t]ags; edit [s]ecret; [a]bort ")
					var response string
					_, err := fmt.Scanln(&response)
					if err == nil {
						switch response {
						case "t":
							editTags(s)
							continue loop
						case "s":
							continue loop
						case "a":
							fmt.Println("Aborted!")
							break loop
						}
					}
				}
			}
		} else {
			fmt.Println("No changes!")
			break
		}
	}
}

func editTags(s *secret.Secret) {
	// Initialize output writer.
	output, err := ioutil.TempFile("", "pgp-tomb-")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to open temporary file!")
	}
	defer os.Remove(output.Name())

	// Dump & edit tags.
	for _, tag := range s.GetTags() {
		if _, err := fmt.Fprintf(output, "%s: %s\n", tag.Name, tag.Value); err != nil {
			logrus.WithFields(logrus.Fields{
				"file": output.Name(),
				"error":  err,
			}).Fatal("Failed to dump tags!")
		}
	}
	output.Close()
	openEditor(output.Name())

	// Process new tags.
	items, err := ioutil.ReadFile(output.Name())
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"file": output.Name(),
			"error":  err,
		}).Fatal("Failed to load tags!")
	}
	tags := make([]secret.Tag, 0)
	for _, item := range strings.Split(string(items), `\n`) {
		if index := strings.Index(item, ":"); index > 0 {
			tags = append(tags, secret.Tag{
				Name:  strings.TrimSpace(item[:index]),
				Value: strings.TrimSpace(item[index+1:]),
			})
		}
	}
	s.SetTags(tags)
}

func openEditor(path string) {
	cmd := exec.Command(config.GetEditor(), path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.WithFields(logrus.Fields{
			"editor": config.GetEditor(),
			"error":  err,
		}).Fatal("Failed to open external editor!")
	}
}

func md5Tags(s *secret.Secret) string {
	digest := md5.New()
	for _, tag := range s.GetTags() {
		digest.Write([]byte(tag.Name))
		digest.Write([]byte(tag.Value))
	}
	return fmt.Sprintf("%x", digest.Sum(nil))
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
