package git

import (
	"os"
	"path/filepath"

	"github.com/denkhaus/logging"
	"github.com/denkhaus/magelib"

	"github.com/juju/errors"
	"github.com/magefile/mage/sh"
)

var (
	Checkout = sh.RunCmd("git", "checkout")
	Branch   = sh.OutCmd("git", "rev-parse", "--abbrev-ref", "HEAD")
)

func GitStatus(path string) (*StatusInfo, error) {
	gitOut, err := GitStatusOutput(path)
	if err != nil {
		return nil, errors.Annotate(err, "GitStatusOutput")
	}

	info := NewStatusInfo(path)
	if err := info.parseStatusOutput(gitOut); err != nil {
		return nil, errors.Annotate(err, "ParseStatusOutput")
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

func EnsureBranchInRepositoryCmd(path string, branchName string) func() error {
	return func() error {
		return EnsureBranchInRepository(path, branchName)
	}
}

func EnsureBranchInRepository(path string, branchName string) error {
	return magelib.InDirectory(path, func() error {
		branch, err := Branch()
		if err != nil {
			return errors.Annotate(err, "GitBranch")
		}

		if branch != branchName {
			logging.Infof("checkout [%s] in repository [%s]", branchName, path)
			return Checkout(branchName)
		}

		logging.Infof("branch [%s] is checked out in repository [%s]", branchName, path)
		return nil
	})
}

func IsRepoCleanCmd(path string) func() magelib.Cmd {
	return func() error {
		return IsRepoClean(path)
	}
}

func IsRepoClean(path string) error {
	path, err := filepath.Abs(os.ExpandEnv(path))
	if err != nil {
		return errors.Annotate(err, "Abs")
	}

	status, err := GitStatus(path)
	if err != nil {
		return errors.Annotate(err, "GitStatus")
	}

	return FormatStatusError(path, status)
}
