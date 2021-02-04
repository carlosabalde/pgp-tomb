package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/carlosabalde/pgp-tomb/internal/core"
	"github.com/carlosabalde/pgp-tomb/internal/core/config"
	"github.com/carlosabalde/pgp-tomb/internal/core/secret"
)

const (
	bashCompletionFunction = `
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
		COMPREPLY=( $(cd "$secrets" && compgen -o plusdirs -f -X '!*` + config.SecretExtension + `' -- "$cur"))

		if [ "${#COMPREPLY[@]}" == "1" ]; then
			if [ -d "$secrets/$COMPREPLY" ]; then
				LASTCHAR=/
				COMPREPLY=$(printf %q%s "$COMPREPLY" "$LASTCHAR")
			elif [ -f "$secrets/$COMPREPLY" ]; then
				COMPREPLY=$(printf %q "${COMPREPLY%` + config.SecretExtension + `}")
			fi
		else
			for ((i=0; i < ${#COMPREPLY[@]}; i++)); do
				[ -d "$secrets/${COMPREPLY[$i]}" ] && \
					COMPREPLY[$i]=${COMPREPLY[$i]}/
				[ -f "$secrets/${COMPREPLY[$i]}" ] && \
					COMPREPLY[$i]=$(printf %q "${COMPREPLY[$i]%` + config.SecretExtension + `}")
			done
		fi
	else
		echo "Please set PGP_TOMB_ROOT!"
	fi
}

__pgp-tomb_custom_func() {
	__pgp-tomb_complete_secret_uri
}
`
)

var (
	cfgFile  string
	verbose  bool
	root     string
	key      string
	identity string
	rootCmd  = &cobra.Command{
		Use:                    "pgp-tomb",
		Version:                config.GetVersion(),
		SilenceErrors:          true,
		BashCompletionFunction: bashCompletionFunction,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initConfig()
			executeHook("pre", cmd.Name())
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			executeHook("post", cmd.Name())
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
	config.Init(viper.ConfigFileUsed())
}

func executeHook(alias string, command string) {
	hooks := config.GetHooks()
	if hook, found := hooks[alias]; found {
		args := []string{
			config.GetRoot(),
			command,
		}
		cmd := exec.Command(hook.Path, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			switch err := err.(type) {
			case *exec.ExitError:
				fmt.Fprintf(os.Stderr, "Aborted by '%s' hook!\n", alias)
				os.Exit(1)
			default:
				logrus.WithFields(logrus.Fields{
					"error": err,
					"hook":  alias,
				}).Fatal("Failed to execute hook!")
			}
		}
	}
}

func parseTags(tags []string) []secret.Tag {
	result := make([]secret.Tag, 0)
	for _, tag := range tags {
		if index := strings.Index(tag, ":"); index > 0 {
			result = append(result, secret.Tag{
				Name:  strings.TrimSpace(tag[:index]),
				Value: strings.TrimSpace(tag[index+1:]),
			})
		}
	}
	return result
}

func main() {
	// Initializations.
	setUmask()

	// Customize version template.
	rootCmd.SetVersionTemplate(fmt.Sprintf(
		"PGP Tomb version {{.Version}} (%s)\n"+
			"Copyright (c) 2019-2021 Carlos Abalde\n", config.GetRevision()))

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
	rootCmd.PersistentFlags().StringVar(
		&identity, "identity", "",
		"override 'identity' option in config file")
	viper.BindPFlag("identity", rootCmd.PersistentFlags().Lookup("identity"))
	rootCmd.PersistentFlags().StringVar(
		&key, "key", "",
		"override 'key' option in config file")
	viper.BindPFlag("key", rootCmd.PersistentFlags().Lookup("key"))

	// 'get' command.
	var cmdGetFile string
	var cmdGetCopy bool
	cmdGet := &cobra.Command{
		Use:     "get <secret URI>",
		Aliases: []string{"cat", "show"},
		Short:   "Show secret (defaults to stdout)",
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
	var cmdSetTags []string
	var cmdSetIgnoreSchema bool
	cmdSet := &cobra.Command{
		Use:     "set <secret URI>",
		Aliases: []string{"add", "insert"},
		Short:   "Create / update secret (defaults to stdin)",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("requires a secret URI argument")
			}
			for _, tag := range cmdSetTags {
				if !strings.Contains(tag, ":") {
					return errors.New("expected tag format is 'name: value'")
				}
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			core.Set(args[0], cmdSetFile, parseTags(cmdSetTags), cmdSetIgnoreSchema)
		},
	}
	cmdSet.PersistentFlags().StringVarP(
		&cmdSetFile, "file", "f", "",
		"read secret from file")
	cmdSet.PersistentFlags().StringArrayVar(
		&cmdSetTags, "tag", nil,
		"tag secret using 'name: value' pair")
	cmdSet.PersistentFlags().BoolVar(
		&cmdSetIgnoreSchema, "ignore-schema", false,
		"skip schema validations, both for tags and secrets")

	// 'edit' command.
	var cmdEditDropTags bool
	var cmdEditTags []string
	var cmdEditIgnoreSchema bool
	cmdEdit := &cobra.Command{
		Use:   "edit <secret URI>",
		Short: "Edit secret using your preferred editor (defaults to $EDITOR)",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("requires a secret URI argument")
			}
			for _, tag := range cmdEditTags {
				if !strings.Contains(tag, ":") {
					return errors.New("expected tag format is 'name: value'")
				}
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			tags := parseTags(cmdEditTags)
			core.Edit(args[0], cmdEditDropTags, tags, cmdEditIgnoreSchema)
		},
	}
	cmdEdit.PersistentFlags().BoolVar(
		&cmdEditDropTags, "drop-tags", false,
		"drop existing tags")
	cmdEdit.PersistentFlags().StringArrayVar(
		&cmdEditTags, "tag", nil,
		"tag secret using 'name: value' pair")
	cmdEdit.PersistentFlags().BoolVar(
		&cmdEditIgnoreSchema, "ignore-schema", false,
		"skip schema validations, both for tags and secrets")

	// 'rebuild' command.
	var cmdRebuildQuery string
	var cmdRebuildRecipient string
	var cmdRebuildWorkers int
	var cmdRebuildForce bool
	var cmdRebuildDryRun bool
	cmdRebuild := &cobra.Command{
		Use:   "rebuild [<folder>|<secret URI<]",
		Short: "Rebuild / check secrets",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return errors.New("rebuilding multiple folders / URIs is not supported")
			}
			if cmdRebuildWorkers < 1 {
				return errors.New("at least one worker is needed")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			var folderOrUri string = ""
			if len(args) > 0 {
				folderOrUri = args[0]
			}
			core.Rebuild(
				folderOrUri, cmdRebuildQuery, cmdRebuildRecipient, cmdRebuildWorkers,
				cmdRebuildForce, cmdRebuildDryRun)
		},
	}
	cmdRebuild.PersistentFlags().StringVarP(
		&cmdRebuildQuery, "query", "q", "",
		"limit rebuild to secrets matching this query")
	cmdRebuild.PersistentFlags().StringVar(
		&cmdRebuildRecipient, "recipient", "",
		"limit rebuild to secrets readable by this key alias (defaults to --identity)")
	cmdRebuild.PersistentFlags().IntVar(
		&cmdRebuildWorkers, "workers", runtime.NumCPU(),
		"set preferred number of workers")
	cmdRebuild.PersistentFlags().BoolVar(
		&cmdRebuildForce, "force", false,
		"force rebuild even when not needed")
	cmdRebuild.PersistentFlags().BoolVar(
		&cmdRebuildDryRun, "dry-run", false,
		"run without actually executing any side effect")

	// 'list' command.
	var cmdListLong bool
	var cmdListQuery string
	var cmdListRecipient string
	var cmdListIgnoreSchema bool
	var cmdListJson bool
	cmdList := &cobra.Command{
		Use:     "list [<folder>|<secret URI>]",
		Short:   "List secrets",
		Aliases: []string{"ls", "dir"},
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return errors.New("listing multiple folders / URIs is not supported")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			var folderOrUri string = ""
			if len(args) > 0 {
				folderOrUri = args[0]
			}
			core.List(
				folderOrUri, cmdListLong, cmdListQuery, cmdListRecipient,
				cmdListIgnoreSchema, cmdListJson)
		},
	}
	cmdList.PersistentFlags().BoolVarP(
		&cmdListLong, "long", "l", false,
		"list using the long format (requires decryption for schema validation)")
	cmdList.PersistentFlags().StringVarP(
		&cmdListQuery, "query", "q", "",
		"limit listing to secrets matching this query")
	cmdList.PersistentFlags().StringVar(
		&cmdListRecipient, "recipient", "",
		"limit listing to secrets readable by this key alias (defaults to --identity)")
	cmdList.PersistentFlags().BoolVar(
		&cmdListIgnoreSchema, "ignore-schema", false,
		"skip schema validations, both for tags and secrets")
	cmdList.PersistentFlags().BoolVarP(
		&cmdListJson, "json", "j", false,
		"enable JSON output")

	// 'init' command.
	cmdInit := &cobra.Command{
		Use:   "init <path>",
		Short: "Initialize a new tomb",
		Args:  cobra.ExactArgs(1),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Skip initialization of configuration & execution of hooks.
		},
		Run: func(cmd *cobra.Command, args []string) {
			core.Initialize(args[0])
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			// Skip execution of hooks.
		},
	}

	// 'bash' command.
	cmdBash := &cobra.Command{
		Use:   "bash",
		Short: "Generate Bash completion script",
		Args:  cobra.NoArgs,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Skip initialization of configuration & execution of hooks.
		},
		Run: func(cmd *cobra.Command, args []string) {
			rootCmd.GenBashCompletion(os.Stdout)
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			// Skip execution of hooks.
		},
	}

	// 'zsh' command.
	cmdZsh := &cobra.Command{
		Use:   "zsh",
		Short: "Generate Zsh completion script",
		Args:  cobra.NoArgs,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Skip initialization of configuration & execution of hooks.
		},
		Run: func(cmd *cobra.Command, args []string) {
			// XXX: custom completion function for secret URIs not yet available
			// for Zsh. See:
			//   - https://github.com/spf13/cobra/pull/884
			rootCmd.GenZshCompletion(os.Stdout)
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			// Skip execution of hooks.
		},
	}

	// Register commands & execute.
	rootCmd.AddCommand(
		cmdGet, cmdSet, cmdEdit, cmdRebuild, cmdList, cmdInit, cmdBash, cmdZsh)
	if err := rootCmd.Execute(); err != nil {
		args := append([]string{"get"}, os.Args[1:]...)
		rootCmd.SetArgs(args)
		rootCmd.SilenceErrors = false
		rootCmd.Execute()
	}
}
