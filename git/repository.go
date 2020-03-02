package git

import (
	"bytes"
	"os"
	"os/exec"
	"strings"

	"github.com/denkhaus/logging"
	"github.com/denkhaus/magelib/common"
	"github.com/juju/errors"
	"github.com/magefile/mage/sh"
)

var (
	GitCheckout = sh.RunCmd("git", "checkout")
)

func GitBranch(cwd string) (string, error) {
	var stderr = new(bytes.Buffer)
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Stderr = stderr
	cmd.Dir = cwd

	out, err := cmd.Output()
	if err != nil {
		return "", errors.Annotatef(err,
			"git rev-parse --abbrev-ref HEAD err: [%s]", stderr.String())
	}

	return strings.TrimSpace(string(out)), nil
}

func EnsureBranchInRepositoryCmd(path string, branchName string) func() error {
	return func() error {
		return EnsureBranchInRepository(path, branchName)
	}
}

func EnsureBranchInRepository(path string, branchName string) error {
	return common.InDirectory(path, func() error {
		branch, err := GitBranch(path)
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
		path := common.GoPackageDir(os.ExpandEnv(pkg))
		branch, err := GitBranch(path)
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
