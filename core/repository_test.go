/*
SPDX-FileCopyrightText: 2026 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package core

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Git reports conflicting files relative to the repository root, while a
// plugin names its version file relative to the project path. These are the
// cases that translation has to get right for a project path that is not the
// repository root itself.
func TestRelativeToPrefix(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		fileName string
		expected string
	}{
		{
			name:     "project path is the repository root",
			prefix:   "",
			fileName: "pom.xml",
			expected: "pom.xml",
		},
		{
			name:     "file inside the project path",
			prefix:   "agent/",
			fileName: "agent/pyproject.toml",
			expected: "pyproject.toml",
		},
		{
			name:     "file deeper inside the project path",
			prefix:   "agent/",
			fileName: "agent/sub/version.txt",
			expected: "sub/version.txt",
		},
		{
			name:     "nested project path",
			prefix:   "services/agent/",
			fileName: "services/agent/pyproject.toml",
			expected: "pyproject.toml",
		},
		{
			name:     "file outside the project path climbs out of it",
			prefix:   "agent/",
			fileName: "README.md",
			expected: "../README.md",
		},
		{
			name:     "file outside a nested project path climbs out twice",
			prefix:   "services/agent/",
			fileName: "version.txt",
			expected: "../../version.txt",
		},
		{
			// the trailing slash is what keeps this from being mistaken for a
			// file inside the project path
			name:     "a sibling whose name starts the same is not inside",
			prefix:   "agent/",
			fileName: "agentic/pyproject.toml",
			expected: "../agentic/pyproject.toml",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, relativeToPrefix(test.prefix, test.fileName))
		})
	}
}

// gitIn runs a git command in a directory and fails the test when git does.
func gitIn(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %s failed: %s", strings.Join(args, " "), output)

	return strings.TrimSpace(string(output))
}

// cloneWithCommit sets up a bare remote with one commit on main and returns
// the path of a clone of it.
func cloneWithCommit(t *testing.T) (remotePath, localPath string) {
	t.Helper()

	dir := t.TempDir()
	remotePath = filepath.Join(dir, "remote")
	localPath = filepath.Join(dir, "local")

	gitIn(t, dir, "init", "--bare", "--initial-branch=main", remotePath)
	gitIn(t, dir, "clone", remotePath, localPath)
	gitIn(t, localPath, "config", "user.name", "Test User")
	gitIn(t, localPath, "config", "user.email", "noreply@mercedes-benz.com")
	gitIn(t, localPath, "commit", "--allow-empty", "-m", "Initial empty commit")
	gitIn(t, localPath, "push", "origin", "main")

	return remotePath, localPath
}

func TestHasLocalBranch(t *testing.T) {
	_, localPath := cloneWithCommit(t)
	repository := NewRepository(localPath, "origin")

	gitIn(t, localPath, "branch", "feature/local")
	// a tag of the same name must not count as a branch
	gitIn(t, localPath, "tag", "feature/tagged")

	for name, expected := range map[string]bool{
		"main":           true,
		"feature/local":  true,
		"feature/tagged": false,
		"feature/absent": false,
	} {
		t.Run(name, func(t *testing.T) {
			exists, err := repository.HasLocalBranch(name)
			require.NoError(t, err)
			assert.Equal(t, expected, exists)
		})
	}
}

func TestIsAncestor(t *testing.T) {
	_, localPath := cloneWithCommit(t)
	repository := NewRepository(localPath, "origin")

	gitIn(t, localPath, "switch", "-c", "feature/ahead")
	gitIn(t, localPath, "commit", "--allow-empty", "-m", "ahead")

	contained, err := repository.IsAncestor("main", "feature/ahead")
	require.NoError(t, err)
	assert.True(t, contained)

	contained, err = repository.IsAncestor("feature/ahead", "main")
	require.NoError(t, err)
	assert.False(t, contained)

	// an unknown revision is an error, not a "no"
	_, err = repository.IsAncestor("does-not-exist", "main")
	assert.Error(t, err)
}

func TestPushDeletionIfUnchanged(t *testing.T) {
	remotePath, localPath := cloneWithCommit(t)
	repository := NewRepository(localPath, "origin")

	gitIn(t, localPath, "switch", "-c", "feature/shared")
	gitIn(t, localPath, "push", "origin", "feature/shared")

	// a colleague pushes to the branch after the local clone last fetched
	colleaguePath := filepath.Join(t.TempDir(), "colleague")
	gitIn(t, filepath.Dir(colleaguePath), "clone", "--branch", "feature/shared", remotePath, colleaguePath)
	gitIn(t, colleaguePath, "config", "user.name", "Colleague")
	gitIn(t, colleaguePath, "config", "user.email", "noreply@mercedes-benz.com")
	gitIn(t, colleaguePath, "commit", "--allow-empty", "-m", "colleague")
	gitIn(t, colleaguePath, "push", "origin", "feature/shared")

	// the stale remote-tracking branch makes the deletion refuse
	assert.Error(t, repository.PushDeletionIfUnchanged("feature/shared"))
	gitIn(t, remotePath, "rev-parse", "--verify", "refs/heads/feature/shared")

	// once the clone knows the current state, the deletion goes through
	gitIn(t, localPath, "fetch", "origin")
	require.NoError(t, repository.PushDeletionIfUnchanged("feature/shared"))

	cmd := exec.Command("git", "rev-parse", "--verify", "refs/heads/feature/shared")
	cmd.Dir = remotePath
	assert.Error(t, cmd.Run(), "branch should be gone from the remote")
}
