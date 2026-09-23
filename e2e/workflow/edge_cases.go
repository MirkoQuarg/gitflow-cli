/*
SPDX-FileCopyrightText: 2026 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package workflow

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mercedes-benz/gitflow-cli/core"
	"github.com/mercedes-benz/gitflow-cli/e2e"
	"github.com/stretchr/testify/assert"
)

// --- Push disabled tests ---

func RunReleaseStartNoPush(t *testing.T) {
	t.Helper()
	env := e2e.SetupTestEnv(t)

	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0", "main")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.1.0-dev", "develop")

	configPath := env.WriteConfig("workflow:\n  push: false\n")
	env.ExecuteGitflow("release", "start", "--config", configPath)

	env.AssertBranchExists("release/1.1.0")
	env.AssertBranchNotOnRemote("release/1.1.0")
	env.AssertCurrentBranchEquals("release/1.1.0")
}

func RunReleaseFinishNoPush(t *testing.T) {
	t.Helper()
	env := e2e.SetupTestEnv(t)

	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0", "main")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.1.0-dev", "develop")
	env.CreateBranch("release/1.1.0", "develop")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.1.0", "release/1.1.0")

	configPath := env.WriteConfig("workflow:\n  push: false\n")
	env.ExecuteGitflow("release", "finish", "--config", configPath)

	env.AssertTagNotOnRemote("1.1.0")
	env.AssertCurrentBranchEquals("develop")
}

func RunHotfixStartNoPush(t *testing.T) {
	t.Helper()
	env := e2e.SetupTestEnv(t)

	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0", "main")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.1.0-dev", "develop")

	configPath := env.WriteConfig("workflow:\n  push: false\n")
	env.ExecuteGitflow("hotfix", "start", "--config", configPath)

	env.AssertBranchExists("hotfix/1.0.1")
	env.AssertBranchNotOnRemote("hotfix/1.0.1")
	env.AssertCurrentBranchEquals("hotfix/1.0.1")
}

func RunHotfixFinishNoPush(t *testing.T) {
	t.Helper()
	env := e2e.SetupTestEnv(t)

	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0", "main")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.1.0-dev", "develop")
	env.CreateBranch("hotfix/1.0.1", "main")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.1", "hotfix/1.0.1")

	configPath := env.WriteConfig("workflow:\n  push: false\n")
	env.ExecuteGitflow("hotfix", "finish", "--config", configPath)

	env.AssertTagNotOnRemote("1.0.1")
	env.AssertCurrentBranchEquals("develop")
}

// --- --no-push flag tests ---

func RunReleaseStartNoPushFlag(t *testing.T) {
	t.Helper()
	env := e2e.SetupTestEnv(t)

	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0", "main")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.1.0-dev", "develop")

	env.ExecuteGitflow("release", "start", "--no-push")

	env.AssertBranchExists("release/1.1.0")
	env.AssertBranchNotOnRemote("release/1.1.0")
	env.AssertCurrentBranchEquals("release/1.1.0")
}

func RunHotfixStartNoPushFlag(t *testing.T) {
	t.Helper()
	env := e2e.SetupTestEnv(t)

	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0", "main")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.1.0-dev", "develop")

	env.ExecuteGitflow("hotfix", "start", "--no-push")

	env.AssertBranchExists("hotfix/1.0.1")
	env.AssertBranchNotOnRemote("hotfix/1.0.1")
	env.AssertCurrentBranchEquals("hotfix/1.0.1")
}

func RunReleaseFinishNoPushFlag(t *testing.T) {
	t.Helper()
	env := e2e.SetupTestEnv(t)

	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0", "main")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.1.0-dev", "develop")
	env.CreateBranch("release/1.1.0", "develop")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.1.0", "release/1.1.0")

	env.ExecuteGitflow("release", "finish", "--no-push")

	env.AssertTagNotOnRemote("1.1.0")
	env.AssertCurrentBranchEquals("develop")
}

func RunHotfixFinishNoPushFlag(t *testing.T) {
	t.Helper()
	env := e2e.SetupTestEnv(t)

	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0", "main")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.1.0-dev", "develop")
	env.CreateBranch("hotfix/1.0.1", "main")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.1", "hotfix/1.0.1")

	env.ExecuteGitflow("hotfix", "finish", "--no-push")

	env.AssertTagNotOnRemote("1.0.1")
	env.AssertCurrentBranchEquals("develop")
}

// --- Rollback tests ---

func RunRollbackPreservesExistingBranches(t *testing.T) {
	t.Helper()
	env := e2e.SetupTestEnv(t)

	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0", "main")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.1.0-dev", "develop")

	// Try release finish without a release branch — triggers an error
	configPath := env.WriteConfig("workflow:\n  rollback: true\n")
	errMsg := env.ExecuteGitflowExpectError("release", "finish", "--config", configPath)

	assert.Contains(t, errMsg, "'release'")

	// develop branch must still exist after rollback
	env.AssertBranchExists("develop")
}

func RunRollbackDisabledLeavesState(t *testing.T) {
	t.Helper()
	env := e2e.SetupTestEnv(t)

	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0", "main")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.1.0-dev", "develop")

	configPath := env.WriteConfig("workflow:\n  rollback: false\n")
	errMsg := env.ExecuteGitflowExpectError("release", "finish", "--config", configPath)

	assert.Contains(t, errMsg, "'release'")

	// develop branch must still exist (rollback disabled = no cleanup at all)
	env.AssertBranchExists("develop")
}

// --- Branch sync tests ---

func RunReleaseStartCreatesDevBranch(t *testing.T) {
	t.Helper()

	env := e2e.SetupTestEnvWithoutDevelop(t)

	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0", "main")

	// Set up sync that creates develop and sets a qualified version
	oldSync := core.BranchSync
	core.BranchSync = func(req core.BranchSyncRequest) (core.BranchSyncResult, error) {
		if req.BranchType == core.Development {
			// After syncBranch creates the branch, we need a qualified version on develop.
			// We handle this by creating the branch ourselves with the right version.
			if err := req.Repository.CheckoutBranch(req.CreateFrom); err != nil {
				return core.BranchSyncResult{}, err
			}
			if err := req.Repository.CreateBranch(req.Configured); err != nil {
				return core.BranchSyncResult{}, err
			}
			if err := req.Repository.WriteFile("version.txt", "1.1.0-dev"); err != nil {
				return core.BranchSyncResult{}, err
			}
			if err := req.Repository.AddFile("version.txt"); err != nil {
				return core.BranchSyncResult{}, err
			}
			if err := req.Repository.CommitChanges("Set development version"); err != nil {
				return core.BranchSyncResult{}, err
			}
			if err := req.Repository.PushChanges(req.Configured); err != nil {
				return core.BranchSyncResult{}, err
			}
			// Return Created: false because we already created it ourselves
			return core.BranchSyncResult{ResolvedName: req.Configured}, nil
		}
		return core.BranchSyncResult{ResolvedName: req.Configured}, nil
	}
	t.Cleanup(func() { core.BranchSync = oldSync })

	env.ExecuteGitflow("release", "start")

	env.AssertBranchExists("develop")
	env.AssertBranchExists("release/1.1.0")
}

func RunReleaseStartDeclinedCreatesDev(t *testing.T) {
	t.Helper()

	env := e2e.SetupTestEnvWithoutDevelop(t)

	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0", "main")

	// Set up declining sync
	oldSync := core.BranchSync
	core.BranchSync = func(req core.BranchSyncRequest) (core.BranchSyncResult, error) {
		if req.BranchType == core.Development {
			return core.BranchSyncResult{}, nil
		}
		return core.BranchSyncResult{ResolvedName: req.Configured}, nil
	}
	t.Cleanup(func() { core.BranchSync = oldSync })

	errMsg := env.ExecuteGitflowExpectError("release", "start")

	assert.Contains(t, errMsg, "required but was not resolved")
}

// --- Production branch sync tests ---

func RunReleaseStartWithMasterBranch(t *testing.T) {
	t.Helper()
	t.Cleanup(func() { core.ResetBranchNames() })

	// Repo has 'master' as production branch, but config says 'main' (default)
	env := e2e.SetupTestEnv(t, e2e.WithProductionBranch("master"))

	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0", "master")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.1.0-dev", "develop")

	// Sync callback: resolve 'main' → 'master' (found as candidate)
	oldSync := core.BranchSync
	core.BranchSync = func(req core.BranchSyncRequest) (core.BranchSyncResult, error) {
		if req.BranchType == core.Production && req.Configured == "main" {
			return core.BranchSyncResult{ResolvedName: "master"}, nil
		}
		return core.BranchSyncResult{ResolvedName: req.Configured}, nil
	}
	t.Cleanup(func() { core.BranchSync = oldSync })

	env.ExecuteGitflow("release", "start")

	env.AssertBranchExists("release/1.1.0")
}

func RunReleaseStartWithMasterDeclined(t *testing.T) {
	t.Helper()
	t.Cleanup(func() { core.ResetBranchNames() })

	env := e2e.SetupTestEnv(t, e2e.WithProductionBranch("master"))

	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0", "master")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.1.0-dev", "develop")

	// Sync callback: decline resolution
	oldSync := core.BranchSync
	core.BranchSync = func(req core.BranchSyncRequest) (core.BranchSyncResult, error) {
		if req.BranchType == core.Production {
			return core.BranchSyncResult{}, nil
		}
		return core.BranchSyncResult{ResolvedName: req.Configured}, nil
	}
	t.Cleanup(func() { core.BranchSync = oldSync })

	errMsg := env.ExecuteGitflowExpectError("release", "start")

	assert.Contains(t, errMsg, "required but was not resolved")
}

func RunReleaseStartWithDevBranch(t *testing.T) {
	t.Helper()
	t.Cleanup(func() { core.ResetBranchNames() })

	// Repo has 'dev' instead of 'develop'
	env := e2e.SetupTestEnv(t, e2e.WithDevelopmentBranch("dev"))

	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0", "main")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.1.0-dev", "dev")

	// Sync callback: resolve 'develop' → 'dev'
	oldSync := core.BranchSync
	core.BranchSync = func(req core.BranchSyncRequest) (core.BranchSyncResult, error) {
		if req.BranchType == core.Development && req.Configured == "develop" {
			return core.BranchSyncResult{ResolvedName: "dev"}, nil
		}
		return core.BranchSyncResult{ResolvedName: req.Configured}, nil
	}
	t.Cleanup(func() { core.BranchSync = oldSync })

	env.ExecuteGitflow("release", "start")

	env.AssertBranchExists("release/1.1.0")
}

// --- Auto-confirm (--yes) tests ---

func RunReleaseStartYesAutoResolvesBranch(t *testing.T) {
	t.Helper()
	t.Cleanup(func() { core.ResetBranchNames() })

	// Repo has 'master' but config defaults to 'main'
	env := e2e.SetupTestEnv(t, e2e.WithProductionBranch("master"))

	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0", "master")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.1.0-dev", "develop")

	env.ExecuteGitflow("release", "start", "--yes")

	env.AssertBranchExists("release/1.1.0")
}

func RunReleaseStartYesCreatesDevBranch(t *testing.T) {
	t.Helper()
	t.Cleanup(func() { core.ResetBranchNames() })

	env := e2e.SetupTestEnvWithoutDevelop(t)

	// No version file on main — the beforeReleaseStart hook will create one on develop
	env.ExecuteGitflow("release", "start", "--yes")

	env.AssertBranchExists("develop")
	env.AssertBranchExists("release/1.0.0")
}

// --- Robustness tests ---

func RunReleaseStartDirtyRepo(t *testing.T) {
	t.Helper()
	env := e2e.SetupTestEnv(t)

	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0", "main")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.1.0-dev", "develop")

	// Make the repo dirty
	env.ExecuteGit("checkout", "develop")
	dirtyFile := env.LocalPath + "/dirty.txt"
	_ = os.WriteFile(dirtyFile, []byte("uncommitted"), 0644)
	env.ExecuteGit("add", dirtyFile)

	errMsg := env.ExecuteGitflowExpectError("release", "start")

	assert.Contains(t, errMsg, "not clean")
}

func RunReleaseStartDuplicateRelease(t *testing.T) {
	t.Helper()
	env := e2e.SetupTestEnv(t)

	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0", "main")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.1.0-dev", "develop")
	env.CreateBranch("release/1.1.0", "develop")

	errMsg := env.ExecuteGitflowExpectError("release", "start")

	assert.Contains(t, errMsg, "already has")
}

func RunHotfixStartDuplicateHotfix(t *testing.T) {
	t.Helper()
	env := e2e.SetupTestEnv(t)

	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0", "main")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.1.0-dev", "develop")
	env.CreateBranch("hotfix/1.0.1", "main")

	errMsg := env.ExecuteGitflowExpectError("hotfix", "start")

	assert.Contains(t, errMsg, "already has")
}

// --- Merge conflict tests ---

// RunReleaseFinishForeignConflict asserts that a conflict in anything but the
// version file stops the workflow. Only an isolated version file conflict can
// be decided without asking anyone; carrying on would build the merge, the tag
// and the back-merge on top of a conflicted tree.
func RunReleaseFinishForeignConflict(t *testing.T) {
	t.Helper()
	env := e2e.SetupTestEnv(t)

	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0", "main")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.1.0-dev", "develop")

	// a file both branches introduce differently, so that merging the release
	// into the production branch conflicts in more than the version file
	env.CommitFile("notes.txt", []byte("from the production branch\n"), "main")
	env.CommitFile("notes.txt", []byte("from the development branch\n"), "develop")

	env.ExecuteGitflow("release", "start")

	errMsg := env.ExecuteGitflowExpectError("release", "finish")

	assert.Contains(t, errMsg, "notes.txt")
	assert.Contains(t, errMsg, "cannot be resolved automatically")

	// the workflow stopped at the merge, so the release was never tagged
	_, err := env.ExecuteGitAllowError("rev-parse", "--verify", "1.1.0")
	assert.Error(t, err, "the release must not have been tagged")
}

// RunHotfixFinishVersionFileInSubdirectory asserts that the workflow also
// works when --path points at a subdirectory of the repository rather than at
// its root. Git names a conflicting file relative to the repository root,
// which is not how the plugin names its version file, and the back-merge of a
// hotfix into the development branch always conflicts in that file.
func RunHotfixFinishVersionFileInSubdirectory(t *testing.T) {
	t.Helper()
	env := e2e.SetupTestEnv(t)

	env.CommitTemplateContent("{{.Version}}", "app/version.txt", "1.0.0", "main")
	env.CommitTemplateContent("{{.Version}}", "app/version.txt", "1.1.0-dev", "develop")

	// the last --path wins, so the project path becomes the subdirectory
	projectPath := filepath.Join(env.LocalPath, "app")

	env.ExecuteGitflow("--path", projectPath, "hotfix", "start")
	env.ExecuteGitflow("--path", projectPath, "hotfix", "finish")

	// the hotfix is released on the production branch and tagged
	env.AssertTagEquals("1.0.1", "main")
	env.AssertTemplateVersionEquals("{{.Version}}", "app/version.txt", "1.0.1", "main")

	// the development version survives the back-merge unchanged
	env.AssertTemplateVersionEquals("{{.Version}}", "app/version.txt", "1.1.0-dev", "develop")
	env.AssertBranchDoesNotExist("hotfix/1.0.1")
}
