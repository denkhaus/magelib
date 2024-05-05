package git

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/denkhaus/logging"
	"github.com/denkhaus/magelib"

	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
)

var (
	Checkout = sh.RunCmd("git", "checkout")
	Branch   = sh.OutCmd("git", "rev-parse", "--abbrev-ref", "HEAD")
)

var (
	ErrCommitNotDefined = errors.New("commit not defined")
)

func GitStatus(path string) (*StatusInfo, error) {
	gitOut, err := GitStatusOutput(path)
	if err != nil {
		return nil, errors.Wrap(err, "GitStatusOutput")
	}

	info := NewStatusInfo(path)
	if err := info.parseStatusOutput(gitOut); err != nil {
		return nil, errors.Wrap(err, "ParseStatusOutput")
	}

	return info, nil
}

func FormatStatusError(path string, status *StatusInfo) error {
	if status.IsDirty() {
		return errors.Errorf("unstaged files have been changed in repo %q", path)
	}

	if status.IsModified() {
		return errors.Errorf("staged files have been modified in repo %q", path)
	}

	if !status.IsSynced() {
		return errors.Errorf("repo %q is not in sync with remote repo", path)
	}

	return nil
}

// EnsureBranchInRepositoryCmd returns a function that ensures that a branch is checked out in a repository.
//
// The function takes two parameters:
// - path (string): the path to the repository.
// - branchName (string): the name of the branch to ensure.
//
// It returns a function that, when called, ensures the specified branch exists in the repository.
// The returned function returns an error if there was a problem ensuring the branch.
func EnsureBranchInRepositoryCmd(path string, branchName string) magelib.Cmd {
	return func() error {
		return EnsureBranchInRepository(path, branchName)
	}
}

// EnsureBranchInRepository ensures that a branch is checked out in a repository.
//
// It takes two parameters:
// - path (string): the path to the repository.
// - branchName (string): the name of the branch to ensure.
//
// It returns an error if there was a problem ensuring the branch.
func EnsureBranchInRepository(path string, branchName string) error {
	return magelib.InDirectory(path, func() error {
		branch, err := Branch()
		if err != nil {
			return errors.Wrap(err, "GitBranch")
		}

		if branch != branchName {
			logging.Infof("checkout [%s] in repository [%s]", branchName, path)
			return Checkout(branchName)
		}

		logging.Infof("branch [%s] is checked out in repository [%s]", branchName, path)
		return nil
	})
}

func IsRepoCleanCmd(path string) magelib.Cmd {
	return func() error {
		return IsRepoClean(path)
	}
}

func IsRepoClean(path string) error {
	path, err := filepath.Abs(os.ExpandEnv(path))
	if err != nil {
		return errors.Wrap(err, "Abs")
	}

	status, err := GitStatus(path)
	if err != nil {
		return errors.Wrap(err, "GitStatus")
	}

	return FormatStatusError(path, status)
}

// CurrentCommit retrieves the current commit hash of the Git repository.
//
// It returns the commit hash as a string and an error if the command fails.
func CurrentCommit() (commit string, err error) {
	commit, err = sh.Output("git", "rev-parse", "HEAD")
	if err != nil {
		err = errors.Wrap(err, "git [rev-parse]")
	}
	return
}

// TagsByCommit retrieves the tags associated with a given commit in a Git repository.
//
// It takes a commit string as a parameter and returns a slice of strings representing the tags and an error.
// The commit string should not be empty. If it is, the function returns an error with the value ErrCommitNotDefined.
// The function executes the "git tag --contains <commit> --sort=-creatordate" command to retrieve the tags.
// If the command fails, the function returns an error.
// The function splits the output of the command into tags using the newline character as a delimiter.
// If the output contains only an empty string, the function returns an empty slice.
// Otherwise, it returns the slice of tags.
func TagsByCommit(commit string) ([]string, error) {
	if commit == "" {
		return nil, ErrCommitNotDefined
	}

	output, err := sh.Output("git", "tag", "--contains", commit, "--sort=-creatordate")
	if err != nil {
		return nil, errors.Wrap(err, "git [tag --contains]")
	}

	tags := strings.Split(output, "\n")
	if len(tags) == 1 && tags[0] == "" {
		return nil, nil
	}

	return tags, nil
}

// IsCommitTagged checks if a commit is tagged in a Git repository.
//
// It takes a commit hash string as a parameter and returns a boolean indicating whether the commit is tagged and an error.
func IsCommitTagged(commit string) (bool, error) {
	currentCommit, err := CurrentCommit()
	if err != nil {
		return false, errors.Wrap(err, "CurrentCommit")
	}

	tags, err := TagsByCommit(currentCommit)
	if err != nil {
		return false, errors.Wrap(err, "TagsByCommit")
	}

	if len(tags) > 0 {
		return true, nil
	}

	return false, nil
}

// MostRecentTag retrieves the current tag of a Git repository at the specified path.
//
// It takes a path string as a parameter and returns a string representing the current tag and an error.
func MostRecentTag() (tag string, err error) {
	tag, err = sh.Output("git", "describe", "--tags", "--abbrev=0")
	if err != nil {
		err = errors.Wrap(err, "git [describe --tags]")
	}

	return
}
