/*
SPDX-FileCopyrightText: 2026 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package core

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// StartFeature creates a new feature branch off the development branch.
// The name is the part after the configured feature prefix, so "142-booking"
// becomes "feature/142-booking".
func StartFeature(name, projectPath string) error {
	pluginRegistryLock.Lock()
	defer pluginRegistryLock.Unlock()

	repository, err := prepareFeatureWorkflow(projectPath)
	if err != nil {
		return err
	}

	branchName, err := featureBranchName(repository, name)
	if err != nil {
		return err
	}

	// format start command messages
	prefix := "Feature Start on branch"
	fmt.Printf("%v %v called: %v\n", prefix, branchName, repository.Local())

	if err := featureStart(repository, branchName); err != nil {
		fmt.Printf("%v %v failed: %v\n", prefix, branchName, repository.Local())
		return err
	}

	fmt.Printf("%v %v completed: %v\n", prefix, branchName, repository.Local())
	return nil
}

// FinishFeature merges a feature branch back into the development branch and
// deletes it. An empty name resolves to the currently checked out branch.
func FinishFeature(name, projectPath string) error {
	pluginRegistryLock.Lock()
	defer pluginRegistryLock.Unlock()

	repository, err := prepareFeatureWorkflow(projectPath)
	if err != nil {
		return err
	}

	branchName, err := resolveFeatureToFinish(repository, name)
	if err != nil {
		return err
	}

	// format finish command messages
	prefix := "Feature Finish on branch"
	fmt.Printf("%v %v called: %v\n", prefix, branchName, repository.Local())

	if err := featureFinish(repository, branchName); err != nil {
		fmt.Printf("%v %v failed: %v\n", prefix, branchName, repository.Local())
		return err
	}

	fmt.Printf("%v %v completed: %v\n", prefix, branchName, repository.Local())
	return nil
}

// prepareFeatureWorkflow applies the configuration and checks the preconditions
// that both feature commands share.
func prepareFeatureWorkflow(projectPath string) (Repository, error) {
	// apply suitable settings from the global configuration to the core package
	applySettings()

	// set path to execute the workflow commands
	ProjectPath = projectPath

	// check if project path exists
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("project path '%v' does not exist", projectPath)
	}

	// get access to the local version control system
	repository := NewRepository(projectPath, Remote)

	// a feature branch never touches the version file, so no plugin and none of
	// its build tools are needed here - only git
	if err := ValidateToolsAvailability(); err != nil {
		return nil, err
	}

	// check if the repository prerequisites are met
	if err := repository.IsClean(); err != nil {
		return nil, err
	}

	// ensure production branch exists (must resolve before development)
	if err := syncBranch(repository, Production); err != nil {
		return nil, err
	}

	// feature branches start from and merge back into the development branch
	if err := syncBranch(repository, Development); err != nil {
		return nil, err
	}

	return repository, nil
}

// featureBranchName turns a user supplied name into a full branch name and
// rejects the names that the Gitflow model reserves.
func featureBranchName(repository Repository, name string) (string, error) {
	prefix := branchNames[Feature]

	// accept both "142-booking" and "feature/142-booking"
	name = strings.TrimSpace(name)
	name = strings.TrimPrefix(name, prefix+"/")
	name = strings.Trim(name, "/")

	if name == "" {
		return "", fmt.Errorf("feature name must not be empty")
	}

	// a feature must not shadow a long-lived branch or another workflow's prefix
	for _, reserved := range []Branch{Production, Development, Release, Hotfix} {
		reservedName := branchNames[reserved]

		if name == reservedName || strings.HasPrefix(name, reservedName+"/") {
			return "", fmt.Errorf(
				"feature name '%v' is reserved for the '%v' branch", name, reservedName)
		}
	}

	branchName := prefix + "/" + name

	// let git decide whether the result is a usable branch name
	if err := repository.CheckRefFormat(branchName); err != nil {
		return "", err
	}

	return branchName, nil
}

// resolveFeatureToFinish determines which feature branch to finish: the one
// named on the command line, or the one currently checked out.
func resolveFeatureToFinish(repository Repository, name string) (string, error) {
	if strings.TrimSpace(name) != "" {
		return featureBranchName(repository, name)
	}

	current, err := repository.CurrentBranch()
	if err != nil {
		return "", err
	}

	prefix := branchNames[Feature] + "/"

	if !strings.HasPrefix(current, prefix) {
		message := fmt.Sprintf(
			"current branch '%v' is not a '%v' branch, so there is nothing to finish"+
				" - name the feature explicitly", current, branchNames[Feature])

		// help the caller along with the feature branches that do exist
		if _, remotes, err := repository.HasBranch(Feature); err == nil && len(remotes) > 0 {
			message = fmt.Sprintf("%v. Candidates: %v", message, strings.Join(remotes, ", "))
		}

		return "", errors.New(message)
	}

	return current, nil
}

func featureStart(repository Repository, branchName string) error {
	// only one branch per feature, and an existing one is never reused - checked
	// before anything is touched, so a refusal leaves the repository as it was
	if exists, err := repository.HasRemoteBranch(branchName); err != nil {
		return err
	} else if exists {
		return fmt.Errorf("repository already has a '%v' branch", branchName)
	}

	if exists, err := repository.HasLocalBranch(branchName); err != nil {
		return err
	} else if exists {
		return fmt.Errorf("repository already has a local '%v' branch", branchName)
	}

	// checkout development branch
	if err := repository.CheckoutBranch(Development.String()); err != nil {
		return err
	}

	// bring it up to date, so the feature does not start from a stale develop
	if err := pullIfPublished(repository, Development.String()); err != nil {
		return err
	}

	// create branch feature/<name> based on the current development branch
	// checkout feature/<name> branch
	if err := repository.CreateBranch(branchName); err != nil {
		return err
	}

	// push only this branch - never all of them, which would publish unrelated work
	if err := pushIfEnabled(func() error { return repository.PushChanges(branchName) }); err != nil {
		return err
	}

	return nil
}

func featureFinish(repository Repository, branchName string) error {
	// remember whether the branch was published, so the deletion below only
	// touches a remote branch that actually exists
	onRemote, err := repository.HasRemoteBranch(branchName)
	if err != nil {
		return err
	}

	// checkout the feature branch - this also creates it locally when it only
	// exists on the remote, and fails when it does not exist at all
	if err := repository.CheckoutBranch(branchName); err != nil {
		return fmt.Errorf("repository does not have a '%v' branch to finish: %v", branchName, err)
	}

	// take in what others pushed to the feature branch, so the merge below
	// carries their commits into the development branch as well
	if onRemote {
		if err := pullOrAbort(repository, branchName); err != nil {
			return err
		}
	}

	// checkout development branch
	if err := repository.CheckoutBranch(Development.String()); err != nil {
		return err
	}

	// bring it up to date, so the merge happens against the current development branch
	if err := pullIfPublished(repository, Development.String()); err != nil {
		return err
	}

	// merge the feature branch into the current development branch (with merge commit --no-ff git flag)
	if err := repository.MergeBranch(branchName, NoFastForward); err != nil {
		// undo the half-finished merge, but leave the feature branch alone:
		// it may carry work that was never pushed
		if abortErr := repository.AbortMerge(); abortErr != nil {
			return errors.Join(err, abortErr)
		}

		return err
	}

	// delete the feature branch locally. The merge above put its commits into
	// the development branch, so nothing is lost - and a plain delete would
	// refuse here whenever the feature carries commits that were never pushed.
	if err := repository.ForceDeleteBranch(branchName); err != nil {
		return err
	}

	// push the development branch
	if err := pushIfEnabled(func() error { return repository.PushChanges(Development.String()) }); err != nil {
		return err
	}

	// delete the feature branch remotely
	if onRemote {
		if err := pushIfEnabled(func() error { return deleteMergedRemoteBranch(repository, branchName) }); err != nil {
			return err
		}
	}

	return nil
}

// deleteMergedRemoteBranch deletes a branch on the remote, but only when every
// commit on it is part of the development branch. Otherwise the deletion
// would drop commits that exist nowhere else.
func deleteMergedRemoteBranch(repository Repository, branchName string) error {
	remoteBranch := Remote + "/" + branchName

	merged, err := repository.IsAncestor(remoteBranch, Development.String())
	if err != nil {
		return err
	}

	if !merged {
		return fmt.Errorf("'%v' has commits that are not in '%v', so it is not deleted"+
			" - merge them and delete the branch by hand", remoteBranch, Development)
	}

	// the lease refuses when someone pushed after the check above
	return repository.PushDeletionIfUnchanged(branchName)
}

// pullIfPublished brings the checked out branch up to date with its remote
// counterpart. A branch that was never pushed has nothing to pull from - for
// example a development branch created locally with pushing disabled.
func pullIfPublished(repository Repository, branchName string) error {
	onRemote, err := repository.HasRemoteBranch(branchName)
	if err != nil {
		return err
	}

	if !onRemote {
		return nil
	}

	return pullOrAbort(repository, branchName)
}

// pullOrAbort pulls the checked out branch and, when the pull fails halfway
// through a merge, undoes that merge so the branch is left as it was.
func pullOrAbort(repository Repository, branchName string) error {
	err := repository.PullBranch(branchName)
	if err == nil {
		return nil
	}

	// the pull may have failed before it started merging, in which case there is
	// nothing to abort - git's own output in err says which of the two it was
	_ = repository.AbortMerge()

	return err
}
