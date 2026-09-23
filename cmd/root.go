/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mercedes-benz/gitflow-cli/cmd/feature"
	"github.com/mercedes-benz/gitflow-cli/cmd/hotfix"
	"github.com/mercedes-benz/gitflow-cli/cmd/release"
	"github.com/mercedes-benz/gitflow-cli/core"
	"github.com/mercedes-benz/gitflow-cli/core/plugin"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Configuration file of the workflow automation command line tool.
var cfgFile string

// RootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Args: cobra.NoArgs,
	Use:  "gitflow-cli",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	// Cobra keeps flag values on the command, and the command is a package level
	// singleton. When Execute is called more than once in one process, a flag of
	// an earlier call would still be set, so restore the defaults before the
	// arguments are parsed again.
	rootCmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		_ = f.Value.Set(f.DefValue)
		f.Changed = false
	})

	return rootCmd.Execute()
}

// Initialize Cobra flags and configuration settings.
func init() {
	rootCmd.Version = buildVersion()
	rootCmd.SetVersionTemplate("{{.Version}}\n")

	// sets the passed functions to be run when each command's ExecuteHook method is called
	cobra.OnInitialize(initConfiguration)

	// set up interactive prompts (branch resolver, docker fallback)
	initPrompts()

	// add subcommands to the root command
	rootCmd.AddCommand(release.ReleaseCmd, hotfix.HotfixCmd, feature.FeatureCmd)

	// persistent flags, which, if defined here, will be global for the application
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.gitflow-cli.yaml)")
	rootCmd.PersistentFlags().StringVarP(&core.ProjectPath, "path", "p", ".", "path to git repository (default is current directory)")
	rootCmd.PersistentFlags().Bool("docker-mode", false, "run plugin commands inside a Docker container")
	rootCmd.PersistentFlags().Bool("native-mode", false, "run plugin commands natively on the host (default)")
	rootCmd.PersistentFlags().Bool("no-push", false, "do not push changes to remote repository")
	rootCmd.PersistentFlags().BoolP("yes", "y", false, "automatically confirm all interactive prompts")
	rootCmd.MarkFlagsMutuallyExclusive("docker-mode", "native-mode")
}

// Read in Viper config file and environment variables if set.
func initConfiguration() {
	if docker, _ := rootCmd.Flags().GetBool("docker-mode"); docker {
		plugin.ExecutorModeOverride = plugin.ModeDocker
	} else if native, _ := rootCmd.Flags().GetBool("native-mode"); native {
		plugin.ExecutorModeOverride = plugin.ModeNative
	} else {
		// neither flag given: drop an override of an earlier run in this process
		plugin.ExecutorModeOverride = ""
	}

	if noPush, _ := rootCmd.Flags().GetBool("no-push"); noPush {
		push := false
		core.PushOverride = &push
	} else {
		core.PushOverride = nil
	}

	if cfgFile != "" {
		// use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// find home directory
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// search config in home directory with name ".gitflow-cli" (without extension)
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".gitflow-cli")
	}

	// read in environment variables that match
	viper.AutomaticEnv()

	// if a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	} else if cfgFile == "" {
		if err := initDefaultConfig(); err != nil {
			fmt.Fprintln(os.Stderr, "Warning: could not create default config:", err)
		} else {
			_ = viper.ReadInConfig()
		}
	}
}

const defaultConfig = `branches:
  production: main
  development: develop
  release: release
  hotfix: hotfix
  feature: feature

workflow:
  push: true
  rollback: false
  docker-fallback: true

logging: "off"
`

func initDefaultConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(home, ".gitflow-cli.yaml")

	if _, err := os.Stat(configPath); err == nil {
		return nil
	}

	if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "Created default config file:", configPath)
	return nil
}
