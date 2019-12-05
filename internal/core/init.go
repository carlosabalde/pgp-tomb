package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

const (
	configSkeleton = `
# This file has been auto-generated using 'pgp-tomb init'. Please, check
# https://github.com/carlosabalde/pgp-tomb for documentation and additional
# examples.

root: %s

keepers:
  - alice

teams:

permissions:

templates:
`
)

var (
	folders = [...]string{
		"hooks",
		"keys",
		"secrets",
		"templates",
	}
)

func Initialize(folder string) {
	// Initializations.
	root, _ := filepath.Abs(folder)
	configPath := path.Join(root, "pgp-tomb.conf")

	// Check root folder exists.
	if info, err := os.Stat(root); os.IsNotExist(err) || !info.IsDir() {
		fmt.Fprintln(os.Stderr, "Path does not exist!")
		os.Exit(1)
	}

	// Do not continue if a configuration file already exists.
	if info, err := os.Stat(configPath); err == nil && !info.IsDir() {
		fmt.Fprintln(os.Stderr, "A configuration file already exists!")
		os.Exit(1)
	}

	// Dump configuration.
	config := fmt.Sprintf(configSkeleton, root)
	if err := ioutil.WriteFile(configPath, []byte(config), 0644); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to dump configuration skeleton!")
	}

	// Create folders.
	for _, folder := range folders {
		if err := os.Mkdir(path.Join(root, folder), 0755); err != nil && !os.IsExist(err) {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("Failed to create folder!")
		}
	}

	// Dump next steps.
	fmt.Println("Done! Next steps:")
	fmt.Println("  1. Add at least one ASCII armored public PGP key (.pub files) to 'keys/'.")
	fmt.Println("  2. Optionally add templates (.schema & .skeleton files) to 'templates/'.")
	fmt.Println("  3. Optionally add hooks (executable .hook files) to 'hooks/'.")
	fmt.Println("  4. Customize 'pgp-tomb.conf': keepers, teams, permissions, etc.")
	fmt.Println("  5. Optionally configure Bash / Zsh completions.")
	fmt.Printf("  6. Optionally set 'PGP_TOMB_ROOT' environment variable to '%s'.\n", root)
	fmt.Println("  7. Optionally set 'pgp-tomb' alias as 'pgp-tomb --root \"$PGP_TOMB_ROOT\"'.")
	fmt.Println("  8. Start enjoying your brand new tomb! :)")
	fmt.Println()
}
