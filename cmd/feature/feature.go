/*
SPDX-FileCopyrightText: 2026 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package feature

import (
	"github.com/mercedes-benz/gitflow-cli/core"

	"github.com/spf13/cobra"
)

// FeatureCmd represents the feature subcommand of RootCmd.
var FeatureCmd = &cobra.Command{
	Args:  cobra.NoArgs,
	Use:   "feature",
	Short: "Develop a new feature for an upcoming release",

	Long: `Develop a new feature for an upcoming release.

Feature branches are where the actual work happens. They branch off the
development branch and merge back into it, so a feature that is not finished in
time simply stays out of the next release.

A feature branch never carries a version: the version is only moved by the
release and hotfix workflows. Nothing on a feature branch touches the version
file, which is why no build tool is required to run these commands.`,
}

// startCmd represents the start subcommand of FeatureCmd.
var startCmd = &cobra.Command{
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	Use:          "start <name>",
	Short:        "Create a new feature branch from the development branch",

	Long: `Create a new feature branch from the development branch.

The name is the part after the feature prefix, so

    gitflow-cli feature start 142-book-a-service-appointment

creates and checks out 'feature/142-book-a-service-appointment'. Passing the
prefix as well ('feature/142-...') is accepted and not doubled.

The development branch is pulled first, so the feature does not start from a
stale state. Only the new branch is pushed.`,

	RunE: func(c *cobra.Command, args []string) error {
		return core.StartFeature(args[0], core.ProjectPath)
	},
}

// finishCmd represents the finish subcommand of FeatureCmd.
var finishCmd = &cobra.Command{
	Args:         cobra.MaximumNArgs(1),
	SilenceUsage: true,
	Use:          "finish [name]",
	Short:        "Merge the feature branch back into the development branch",

	Long: `Merge the feature branch back into the development branch.

Without an argument the currently checked out feature branch is finished.
The branch is merged with a merge commit (--no-ff), so the individual commits
of the feature stay visible in the history, and is then deleted locally and on
the remote.

Note that this merges locally. Where changes reach the development branch
through a reviewed merge request, use that instead of this command.`,

	RunE: func(c *cobra.Command, args []string) error {
		name := ""

		if len(args) > 0 {
			name = args[0]
		}

		return core.FinishFeature(name, core.ProjectPath)
	},
}

// Initialize Cobra flags for the feature subcommand.
func init() {
	// add subcommands to the feature command
	FeatureCmd.AddCommand(startCmd, finishCmd)
}
