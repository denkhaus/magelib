package golang

import (
	"fmt"
	"os"

	"github.com/denkhaus/logging"
	"github.com/denkhaus/magelib"

	"github.com/denkhaus/magelib/git"
	"github.com/juju/errors"
	"github.com/magefile/mage/sh"
)

func InGoPackageDir(pkg string, fn func() error) error {
	return magelib.InDirectory(
		PackageDir(os.ExpandEnv(pkg)), fn,
	)
}

func PackageDir(pkg string) string {
	return fmt.Sprintf("%s/src/%s", Env("GOPATH"), pkg)
}

func Env(value string) string {
	out, err := magelib.GoEnvOut(value)
	magelib.HandleError(err)
	if out == "" {
		magelib.HandleError(
			errors.Errorf("%s is undefined", value),
		)
	}

	return out
}

func IsPackageCleanCmd(pkg string) magelib.Cmd {
	return func() error {
		return IsPackageClean(pkg)
	}
}

func IsPackageClean(pkg string) error {
	status, err := git.GitStatus(
		PackageDir(pkg),
	)
	if err != nil {
		return errors.Annotate(err, "GitStatus")
	}

	return git.FormatStatusError(pkg, status)
}

func EnsureBranchInPackageCmd(pkg string, branchName string) magelib.Cmd {
	return func() error {
		return EnsureBranchInPackage(pkg, branchName)
	}
}

func EnsureBranchInPackage(pkg string, branchName string) error {
	return InGoPackageDir(pkg, func() error {
		branch, err := git.Branch()
		if err != nil {
			return errors.Annotate(err, "GitBranch")
		}

		if branch != branchName {
			logging.Infof("checkout [%s] in go pkg [%s]", branchName, pkg)
			return git.Checkout(branchName)
		}

		logging.Infof("branch [%s] is checked out in go pkg [%s]", branchName, pkg)
		return nil
	})
}

func UpdateModuleCmd(path string, vendor bool) magelib.Cmd {
	return func() error {
		return UpdateGoModule(path, vendor)
	}
}

func UpdateGoModule(path string, vendor bool) error {
	env := magelib.ArgsMap{
		"GO111MODULE": "on",
	}

	return magelib.InDirectory(path, magelib.ChainCmdsCmd(
		func() error {
			logging.Info("run -> go get -d")
			return sh.RunWithV(env, "go", "get", "-d")
		},
		func() error {
			logging.Info("run -> go mod tidy")
			return sh.RunWithV(env, "go", "mod", "tidy")
		},
		func() error {
			if !vendor {
				return nil
			}

			logging.Info("run -> go mod vendor")
			return sh.RunWithV(env, "go", "mod", "vendor")
		}))
}
