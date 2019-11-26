package main

import (
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/carlosabalde/pgp-tomb/internal/core"
	"github.com/carlosabalde/pgp-tomb/internal/core/config"
)

const (
	bashCompletionFunc = `
__pgp-tomb_complete_secret_uri() {
	# See:
	#   - https://superuser.com/questions/564716/bash-completion-for-filename-patterns-or-directories
	#   - https://debian-administration.org/article/317/An_introduction_to_bash_completion_part_2

	compopt -o nospace

	local IFS=$'\n'
	local LASTCHAR=' '
	local secrets

	if [ -z "$PGP_TOMB_ROOT" ]; then
		if [ -d "secrets" ]; then
			secrets=$(pwd)/secrets
		fi
	else
		secrets=$PGP_TOMB_ROOT/secrets
	fi

	if [ ! -z "$secrets" ]; then
		COMPREPLY=( $(cd "$secrets" && compgen -o plusdirs -f -X '!*.pgp' -- "$cur"))

		if [ "${#COMPREPLY[@]}" == "1" ]; then
			if [ -d "$secrets/$COMPREPLY" ]; then
				LASTCHAR=/
				COMPREPLY=$(printf %q%s "$COMPREPLY" "$LASTCHAR")
			elif [ -f "$secrets/$COMPREPLY" ]; then
				COMPREPLY=$(printf %q "${COMPREPLY%.pgp}")
			fi
		else
			for ((i=0; i < ${#COMPREPLY[@]}; i++)); do
				[ -d "$secrets/${COMPREPLY[$i]}" ] && \
					COMPREPLY[$i]=${COMPREPLY[$i]}/
				[ -f "$secrets/${COMPREPLY[$i]}" ] && \
					COMPREPLY[$i]=$(printf %q "${COMPREPLY[$i]%.pgp}")
			done
		fi
	else
		echo "Please set PGP_TOMB_ROOT!"
	fi
}

__pgp-tomb_custom_func() {
	case ${last_command} in
		pgp-tomb_about | pgp-tomb_edit | pgp-tomb_get | pgp-tomb_set)
			__pgp-tomb_complete_secret_uri
			return
			;;
		*)
			;;
	esac
}
`
)

var (
	version  string
	revision string
	cfgFile  string
	verbose  bool
	root     string
	rootCmd  = &cobra.Command{
		Use:                    "pgp-tomb",
		Version:                version,
		BashCompletionFunction: bashCompletionFunc,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initConfig()
		},
	}
)

func initConfig() {
	// Configure logging.
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetOutput(os.Stderr)
	if verbose {
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		logrus.SetLevel(logrus.WarnLevel)
	}

	// Get ready to load configuration.
	viper.SetConfigType("yaml")
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("pgp-tomb")
		viper.AddConfigPath("/etc/pgp-tomb")
		viper.AddConfigPath("$HOME/.pgp-tomb")
		viper.AddConfigPath(".")
		if root != "" {
			viper.AddConfigPath(root)
		}
	}

	// Load configuration.
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logrus.Fatal("Configuration not found!")
		} else {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("Failed to read configuration file!")
		}
	} else {
		logrus.WithFields(logrus.Fields{
			"file": viper.ConfigFileUsed(),
		}).Info("Configuration successfully loaded")
	}

	// Validate & initialize configuration.
	config.Init()
}

func main() {
	// Initializations.
	syscall.Umask(0077)

	// Customize version template.
	rootCmd.SetVersionTemplate(fmt.Sprintf(
		"PGP Tomb version {{.Version}} (%s)\n"+
			"Copyright (c) 2019 Carlos Abalde\n", revision))

	// Global flags.
	rootCmd.PersistentFlags().StringVarP(
		&cfgFile, "config", "c", "",
		"set config file")
	rootCmd.PersistentFlags().BoolVarP(
		&verbose, "verbose", "v", false,
		"enable verbose output")
	rootCmd.PersistentFlags().StringVar(
		&root, "root", "",
		"override 'root' option in config file")
	viper.BindPFlag("root", rootCmd.PersistentFlags().Lookup("root"))

	// 'get' command.
	var cmdGetFile string
	var cmdGetCopy bool
	cmdGet := &cobra.Command{
		Use:     "get",
		Aliases: []string{"cat"},
		Short:   "Read secret",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("requires a secret URI argument")
			}
			if cmdGetCopy && cmdGetFile != "" {
				return errors.New("--file & --copy flags cannot be combined")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			core.Get(args[0], cmdGetFile, cmdGetCopy)
		},
	}
	cmdGet.PersistentFlags().StringVarP(
		&cmdGetFile, "file", "f", "",
		"write secret to file")
	cmdGet.PersistentFlags().BoolVar(
		&cmdGetCopy, "copy", false,
		"copy secret into system clipboard")

	// 'set' command.
	var cmdSetFile string
	cmdSet := &cobra.Command{
		Use:     "set",
		Aliases: []string{"add"},
		Short:   "Create / update secret",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("requires a secret URI argument")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			core.Set(args[0], cmdSetFile)
		},
	}
	cmdSet.PersistentFlags().StringVarP(
		&cmdSetFile, "file", "f", "",
		"read secret from file")

	// 'edit' command.
	cmdEdit := &cobra.Command{
		Use:   "edit",
		Short: "Edit secret",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("requires a secret URI argument")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			core.Edit(args[0])
		},
	}

	// 'about' command.
	cmdAbout := &cobra.Command{
		Use:   "about",
		Short: "Show details about secret",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("requires a secret URI argument")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			core.About(args[0])
		},
	}

	// 'rebuild' command.
	var cmdRebuildLimit string
	var cmdRebuildDryRun bool
	cmdRebuild := &cobra.Command{
		Use:   "rebuild",
		Short: "Rebuild / check secrets",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			core.Rebuild(cmdRebuildLimit, cmdRebuildDryRun)
		},
	}
	cmdRebuild.PersistentFlags().StringVar(
		&cmdRebuildLimit, "limit", "",
		"limit rebuild to secrets with URIs matching this regexp")
	cmdRebuild.PersistentFlags().BoolVar(
		&cmdRebuildDryRun, "dry-run", false,
		"run the rebuild without actually executing any side effect")

	// 'list' command.
	var cmdListLimit string
	var cmdListKey string
	cmdList := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "dir"},
		Short:   "List secrets",
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			core.List(cmdListLimit, cmdListKey)
		},
	}
	cmdList.PersistentFlags().StringVar(
		&cmdListLimit, "limit", "",
		"limit listing to secrets with URIs matching this regexp")
	cmdList.PersistentFlags().StringVar(
		&cmdListKey, "key", "",
		"list only secrets readable by this key alias")

	// 'bash' command.
	cmdBash := &cobra.Command{
		Use:   "bash",
		Short: "Generate Bash completion script",
		Args:  cobra.NoArgs,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Skip initConfig().
		},
		Run: func(cmd *cobra.Command, args []string) {
			rootCmd.GenBashCompletion(os.Stdout)
		},
	}

	// 'zsh' command.
	cmdZsh := &cobra.Command{
		Use:   "zsh",
		Short: "Generate Zsh completion script",
		Args:  cobra.NoArgs,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Skip initConfig().
		},
		Run: func(cmd *cobra.Command, args []string) {
			rootCmd.GenZshCompletion(os.Stdout)
		},
	}

	// Register commands & execute.
	rootCmd.AddCommand(
		cmdGet, cmdSet, cmdEdit, cmdAbout, cmdRebuild, cmdList, cmdBash,
		cmdZsh)
	rootCmd.Execute()
}
