package git

import (
	"github.com/denkhaus/logging"
	"github.com/denkhaus/magelib/common"
	"github.com/juju/errors"
	"github.com/magefile/mage/sh"
)

var (
	GitCheckout = sh.RunCmd("git", "checkout")
	GitBranch   = sh.OutCmd("git", "rev-parse", "--abbrev-ref", "HEAD")
)

func EnsureBranchInRepositoryCmd(path string, branchName string) func() error {
	return func() error {
		return EnsureBranchInRepository(path, branchName)
	}
}

func EnsureBranchInRepository(path string, branchName string) error {
	return common.InDirectory(path, func() error {
		branch, err := GitBranch()
		if err != nil {
			return errors.Annotate(err, "GitBranch")
		}

		if branch != branchName {
			logging.Infof("checkout [%s] in repository [%s]", branchName, path)
			return GitCheckout(branchName)
		}

		logging.Infof("branch [%s] is checked out in repository [%s]", branchName, path)
		return nil
	})
}

func EnsureBranchInGoPackageCmd(pkg string, branchName string) func() error {
	return func() error {
		return EnsureBranchInGoPackage(pkg, branchName)
	}
}

func EnsureBranchInGoPackage(pkg string, branchName string) error {
	return common.InGoPackageDir(pkg, func() error {
		branch, err := GitBranch()
		if err != nil {
			return errors.Annotate(err, "GitBranch")
		}

		if branch != branchName {
			logging.Infof("checkout [%s] in go pkg [%s]", branchName, pkg)
			return GitCheckout(branchName)
		}

		logging.Infof("branch [%s] is checked out in go pkg [%s]", branchName, pkg)
		return nil
	})
}
