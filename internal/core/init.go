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
	configSkeleton = `#
# This file has been auto-generated using 'pgp-tomb init'. Please, check
# https://github.com/carlosabalde/pgp-tomb for documentation and additional
# examples.
#

root: %s

identity:

key:

keepers:
  - alice

teams:

tags:

permissions:

templates:
`
	preHookSkeleton = `#!/usr/bin/env bash

# Configuration.
MAX_AGE=86400

# Check input arguments.
if [ "$#" -ne 2 ]; then
    exit 1
fi
ROOT=$1
COMMAND=$2

# Initializations.
set -e
cd "$ROOT"
cd "$(git rev-parse --show-toplevel)"

# Do not continue if Git is not installed of if the current folder does not look
# like a Git repository.
if [ "$(git rev-parse --is-inside-work-tree 2> /dev/null)" != "true" ]; then
    exit 0
fi

# Check if the repository has not been updated in a while.
NOW=$(date +%s)
STAT_OPTS='-c %Y'
if uname | grep -q 'Darwin'; then
    STAT_OPTS='-f %m'
fi
LAST_UPDATE=$(stat $STAT_OPTS .git/FETCH_HEAD)
if [ "$[$NOW-$LAST_UPDATE]" -ge "$MAX_AGE" ]; then
    # About ASCII art:
    #   - http://patorjk.com/software/taag/#p=display&h=1&f=Bloody&t=Pull%20please!
    echo
    echo ' ██▓███   █    ██  ██▓     ██▓        ██▓███   ██▓    ▓█████ ▄▄▄        ██████ ▓█████  ▐██▌ '
    echo '▓██░  ██▒ ██  ▓██▒▓██▒    ▓██▒       ▓██░  ██▒▓██▒    ▓█   ▀▒████▄    ▒██    ▒ ▓█   ▀  ▐██▌ '
    echo '▓██░ ██▓▒▓██  ▒██░▒██░    ▒██░       ▓██░ ██▓▒▒██░    ▒███  ▒██  ▀█▄  ░ ▓██▄   ▒███    ▐██▌ '
    echo '▒██▄█▓▒ ▒▓▓█  ░██░▒██░    ▒██░       ▒██▄█▓▒ ▒▒██░    ▒▓█  ▄░██▄▄▄▄██   ▒   ██▒▒▓█  ▄  ▓██▒ '
    echo '▒██▒ ░  ░▒▒█████▓ ░██████▒░██████▒   ▒██▒ ░  ░░██████▒░▒████▒▓█   ▓██▒▒██████▒▒░▒████▒ ▒▄▄  '
    echo '▒▓▒░ ░  ░░▒▓▒ ▒ ▒ ░ ▒░▓  ░░ ▒░▓  ░   ▒▓▒░ ░  ░░ ▒░▓  ░░░ ▒░ ░▒▒   ▓▒█░▒ ▒▓▒ ▒ ░░░ ▒░ ░ ░▀▀▒ '
    echo '░▒ ░     ░░▒░ ░ ░ ░ ░ ▒  ░░ ░ ▒  ░   ░▒ ░     ░ ░ ▒  ░ ░ ░  ░ ▒   ▒▒ ░░ ░▒  ░ ░ ░ ░  ░ ░  ░ '
    echo '░░        ░░░ ░ ░   ░ ░     ░ ░      ░░         ░ ░      ░    ░   ▒   ░  ░  ░     ░       ░ '
    echo '            ░         ░  ░    ░  ░                ░  ░   ░  ░     ░  ░      ░     ░  ░ ░    '
    echo
    echo

    # Run git pull only if explicitly allowed by the user.
    read -p "> Run 'git pull --rebase' now? (Y/n) " -n 1 -r REPLY
    REPLY=${REPLY:-Y}
    if [ "$REPLY" = "y" -o "$REPLY" = "Y" ]; then
        git pull --rebase
    fi
fi
`
	postHookSkeleton = `#!/usr/bin/env bash

# Check input arguments.
if [ "$#" -ne 2 ]; then
    exit 1
fi
ROOT=$1
COMMAND=$2

# Initializations.
set -e
cd "$ROOT"

# Do not continue if Git is not installed of if the current folder does not look
# like a Git repository.
if [ "$(git rev-parse --is-inside-work-tree 2> /dev/null)" != "true" ]; then
    exit 0
fi

# Look for uncommitted or untracked files in the secrets folder.
if [ -n "$(git status --porcelain -- secrets 2> /dev/null)" ]; then
    # About ASCII art:
    #   - http://patorjk.com/software/taag/#p=display&h=1&f=Bloody&t=Commit%20please!
    echo
    echo
    echo ' ▄████▄   ▒█████   ███▄ ▄███▓ ███▄ ▄███▓ ██▓▄▄▄█████▓    ██▓███   ██▓    ▓█████ ▄▄▄        ██████ ▓█████  ▐██▌'
    echo '▒██▀ ▀█  ▒██▒  ██▒▓██▒▀█▀ ██▒▓██▒▀█▀ ██▒▓██▒▓  ██▒ ▓▒   ▓██░  ██▒▓██▒    ▓█   ▀▒████▄    ▒██    ▒ ▓█   ▀  ▐██▌'
    echo '▒▓█    ▄ ▒██░  ██▒▓██    ▓██░▓██    ▓██░▒██▒▒ ▓██░ ▒░   ▓██░ ██▓▒▒██░    ▒███  ▒██  ▀█▄  ░ ▓██▄   ▒███    ▐██▌'
    echo '▒▓▓▄ ▄██▒▒██   ██░▒██    ▒██ ▒██    ▒██ ░██░░ ▓██▓ ░    ▒██▄█▓▒ ▒▒██░    ▒▓█  ▄░██▄▄▄▄██   ▒   ██▒▒▓█  ▄  ▓██▒'
    echo '▒ ▓███▀ ░░ ████▓▒░▒██▒   ░██▒▒██▒   ░██▒░██░  ▒██▒ ░    ▒██▒ ░  ░░██████▒░▒████▒▓█   ▓██▒▒██████▒▒░▒████▒ ▒▄▄'
    echo '░ ░▒ ▒  ░░ ▒░▒░▒░ ░ ▒░   ░  ░░ ▒░   ░  ░░▓    ▒ ░░      ▒▓▒░ ░  ░░ ▒░▓  ░░░ ▒░ ░▒▒   ▓▒█░▒ ▒▓▒ ▒ ░░░ ▒░ ░ ░▀▀▒'
    echo '  ░  ▒     ░ ▒ ▒░ ░  ░      ░░  ░      ░ ▒ ░    ░       ░▒ ░     ░ ░ ▒  ░ ░ ░  ░ ▒   ▒▒ ░░ ░▒  ░ ░ ░ ░  ░ ░  ░'
    echo '░        ░ ░ ░ ▒  ░      ░   ░      ░    ▒ ░  ░         ░░         ░ ░      ░    ░   ▒   ░  ░  ░     ░       ░'
    echo '░ ░          ░ ░         ░          ░    ░                           ░  ░   ░  ░     ░  ░      ░     ░  ░ ░'
    echo '░'
    echo
    exit 0
fi

# Look for pending pushes in any branch.
if [ -n "$(git log --branches --not --remotes 2> /dev/null)" ]; then
    # About ASCII art:
    #   - http://patorjk.com/software/taag/#p=display&h=1&f=Bloody&t=Push%20please!
    echo
    echo
    echo ' ██▓███   █    ██   ██████  ██░ ██     ██▓███   ██▓    ▓█████ ▄▄▄        ██████ ▓█████  ▐██▌ '
    echo '▓██░  ██▒ ██  ▓██▒▒██    ▒ ▓██░ ██▒   ▓██░  ██▒▓██▒    ▓█   ▀▒████▄    ▒██    ▒ ▓█   ▀  ▐██▌ '
    echo '▓██░ ██▓▒▓██  ▒██░░ ▓██▄   ▒██▀▀██░   ▓██░ ██▓▒▒██░    ▒███  ▒██  ▀█▄  ░ ▓██▄   ▒███    ▐██▌ '
    echo '▒██▄█▓▒ ▒▓▓█  ░██░  ▒   ██▒░▓█ ░██    ▒██▄█▓▒ ▒▒██░    ▒▓█  ▄░██▄▄▄▄██   ▒   ██▒▒▓█  ▄  ▓██▒ '
    echo '▒██▒ ░  ░▒▒█████▓ ▒██████▒▒░▓█▒░██▓   ▒██▒ ░  ░░██████▒░▒████▒▓█   ▓██▒▒██████▒▒░▒████▒ ▒▄▄  '
    echo '▒▓▒░ ░  ░░▒▓▒ ▒ ▒ ▒ ▒▓▒ ▒ ░ ▒ ░░▒░▒   ▒▓▒░ ░  ░░ ▒░▓  ░░░ ▒░ ░▒▒   ▓▒█░▒ ▒▓▒ ▒ ░░░ ▒░ ░ ░▀▀▒ '
    echo '░▒ ░     ░░▒░ ░ ░ ░ ░▒  ░ ░ ▒ ░▒░ ░   ░▒ ░     ░ ░ ▒  ░ ░ ░  ░ ▒   ▒▒ ░░ ░▒  ░ ░ ░ ░  ░ ░  ░ '
    echo '░░        ░░░ ░ ░ ░  ░  ░   ░  ░░ ░   ░░         ░ ░      ░    ░   ▒   ░  ░  ░     ░       ░ '
    echo '            ░           ░   ░  ░  ░                ░  ░   ░  ░     ░  ░      ░     ░  ░ ░    '
    echo
    exit 0
fi
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

	// Dump 'pre' hook.
	if err := ioutil.WriteFile(path.Join(root, "hooks", "pre.hook"), []byte(preHookSkeleton), 0755); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to dump pre hook skeleton!")
	}

	// Dump 'post' hook.
	if err := ioutil.WriteFile(path.Join(root, "hooks", "post.hook"), []byte(postHookSkeleton), 0755); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to dump post hook skeleton!")
	}

	// Dump next steps.
	fmt.Println("Done! Next steps:")
	fmt.Println("  1. Add at least one ASCII armored public PGP key (.pub files) to 'keys/'.")
	fmt.Println("  2. Optionally add templates (.schema & .skeleton files) to 'templates/'.")
	fmt.Println("  3. Optionally customize hooks (executable .hook files) in 'hooks/'.")
	fmt.Println("  4. Customize 'pgp-tomb.conf': keepers, teams, permissions, etc.")
	fmt.Println("  5. Optionally configure Bash / Zsh completions.")
	fmt.Printf("  6. Optionally set 'PGP_TOMB_ROOT' environment variable to '%s'.\n", root)
	fmt.Println("  7. Optionally set 'pgp-tomb' alias as 'pgp-tomb --root \"$PGP_TOMB_ROOT\"'.")
	fmt.Println("  8. Start enjoying your brand new tomb! :)")
	fmt.Println()
}
