/*
SPDX-FileCopyrightText: 2026 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package cmd_test

import (
	"strings"
	"testing"

	"github.com/mercedes-benz/gitflow-cli/e2e"
	"github.com/stretchr/testify/assert"
)

// The feature workflow does not read or write a version, so these tests run
// without any plugin version file and never need a build tool.
const featureName = "142-book-a-service-appointment"

// TestFeatureStart tests that a feature branch is created off develop and pushed
func TestFeatureStart(t *testing.T) {
	env := e2e.SetupTestEnv(t)

	env.ExecuteGitflow("feature", "start", featureName)

	branch := "feature/" + featureName
	env.AssertBranchExists(branch)
	env.AssertBranchExists("origin/" + branch)
	env.AssertCurrentBranchEquals(branch)

	// a feature branch carries no commit of its own and no version change
	env.AssertCommitMessageEquals("Initial empty commit", branch)
}

// TestFeatureStartAcceptsThePrefix tests that passing the prefix does not double it
func TestFeatureStartAcceptsThePrefix(t *testing.T) {
	env := e2e.SetupTestEnv(t)

	env.ExecuteGitflow("feature", "start", "feature/"+featureName)

	env.AssertBranchExists("feature/" + featureName)
	env.AssertBranchDoesNotExist("feature/feature/" + featureName)
}

// TestFeatureStartRejectsReservedNames tests that a feature cannot shadow a long-lived branch
func TestFeatureStartRejectsReservedNames(t *testing.T) {
	for _, name := range []string{"develop", "main", "release", "hotfix", "release/1.0.0"} {
		t.Run(name, func(t *testing.T) {
			env := e2e.SetupTestEnv(t)

			message := env.ExecuteGitflowExpectError("feature", "start", name)

			assert.Contains(t, message, "reserved")
			env.AssertBranchDoesNotExist("feature/" + name)
		})
	}
}

// TestFeatureStartRefusesAnExistingBranch tests that an existing feature branch is never reused
func TestFeatureStartRefusesAnExistingBranch(t *testing.T) {
	env := e2e.SetupTestEnv(t)

	env.ExecuteGitflow("feature", "start", featureName)
	env.ExecuteGit("checkout", "develop")

	message := env.ExecuteGitflowExpectError("feature", "start", featureName)

	assert.Contains(t, message, "already has a 'feature/"+featureName+"' branch")
}

// TestFeatureStartWithoutPush tests that --no-push keeps the branch local
func TestFeatureStartWithoutPush(t *testing.T) {
	env := e2e.SetupTestEnv(t)

	env.ExecuteGitflow("feature", "start", featureName, "--no-push")

	env.AssertBranchNotOnRemote("feature/" + featureName)
}

// TestFeatureFinish tests that the feature is merged into develop and the branch removed
func TestFeatureFinish(t *testing.T) {
	env := e2e.SetupTestEnv(t)
	branch := "feature/" + featureName

	env.ExecuteGitflow("feature", "start", featureName)
	env.CommitFile("feature.txt", []byte("the work\n"), branch)

	env.ExecuteGitflow("feature", "finish", featureName)

	// the merge commit keeps the individual commits of the feature visible
	env.AssertCommitMessageEquals("Merge branch '"+branch+"' into develop", "develop")
	env.AssertCurrentBranchEquals("develop")

	// the file arrived on develop
	content := env.ExecuteGit("show", "develop:feature.txt")
	assert.Equal(t, "the work", strings.TrimSpace(content))

	// the branch is gone locally and on the remote
	env.AssertBranchDoesNotExist(branch)
	env.AssertBranchDoesNotExist("origin/" + branch)
}

// TestFeatureFinishUsesTheCurrentBranch tests that finish without an argument
// resolves to the checked out feature branch
func TestFeatureFinishUsesTheCurrentBranch(t *testing.T) {
	env := e2e.SetupTestEnv(t)
	branch := "feature/" + featureName

	env.ExecuteGitflow("feature", "start", featureName)
	env.CommitFile("feature.txt", []byte("the work\n"), branch)
	env.ExecuteGit("checkout", branch)

	env.ExecuteGitflow("feature", "finish")

	env.AssertCommitMessageEquals("Merge branch '"+branch+"' into develop", "develop")
	env.AssertBranchDoesNotExist(branch)
}

// TestFeatureFinishOutsideAFeatureBranch tests that finish refuses to guess
func TestFeatureFinishOutsideAFeatureBranch(t *testing.T) {
	env := e2e.SetupTestEnv(t)

	env.ExecuteGitflow("feature", "start", featureName)
	env.ExecuteGit("checkout", "develop")

	message := env.ExecuteGitflowExpectError("feature", "finish")

	assert.Contains(t, message, "is not a 'feature' branch")
	assert.Contains(t, message, "feature/"+featureName)
	env.AssertBranchExists("feature/" + featureName)
}

// TestFeatureFinishOfAnUnknownBranch tests the error for a feature that does not exist
func TestFeatureFinishOfAnUnknownBranch(t *testing.T) {
	env := e2e.SetupTestEnv(t)

	message := env.ExecuteGitflowExpectError("feature", "finish", "999-never-started")

	assert.Contains(t, message, "does not have a 'feature/999-never-started' branch to finish")
}

// TestFeatureFinishOfALocalOnlyBranch tests that a feature that was never pushed can be finished
func TestFeatureFinishOfALocalOnlyBranch(t *testing.T) {
	env := e2e.SetupTestEnv(t)
	branch := "feature/" + featureName

	env.ExecuteGitflow("feature", "start", featureName, "--no-push")

	// commit without pushing, so the branch stays local
	env.WriteWorkingFile("feature.txt", []byte("the work\n"))
	env.ExecuteGit("add", "feature.txt")
	env.ExecuteGit("commit", "-m", "feat: the work")
	env.AssertBranchNotOnRemote(branch)

	env.ExecuteGitflow("feature", "finish", featureName)

	env.AssertCommitMessageEquals("Merge branch '"+branch+"' into develop", "develop")
	env.AssertBranchDoesNotExist(branch)
}

// TestFeatureWithConfigFile tests the feature workflow with a custom branch prefix
func TestFeatureWithConfigFile(t *testing.T) {
	env, configPath := setupCustomBranchTest(t)

	const customFeatureBranch = "custom-feature"
	branch := customFeatureBranch + "/" + featureName

	env.ExecuteGitflow("feature", "start", featureName, "--config", configPath)

	env.AssertBranchExists(branch)
	env.AssertBranchExists("origin/" + branch)
	env.AssertCurrentBranchEquals(branch)

	env.CommitFile("feature.txt", []byte("the work\n"), branch)

	env.ExecuteGitflow("feature", "finish", featureName, "--config", configPath)

	env.AssertCommitMessageEquals(
		"Merge branch '"+branch+"' into "+developmentBranch, developmentBranch)
	env.AssertBranchDoesNotExist(branch)
	env.AssertCurrentBranchEquals(developmentBranch)
}
