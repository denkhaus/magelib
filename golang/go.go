package golang

import (
	"fmt"
	"os"

	"github.com/denkhaus/logging"
	"github.com/denkhaus/magelib"

	"github.com/denkhaus/magelib/git"
	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
)

var (
	UpdatePackage = sh.RunCmd("go", "get", "-u")
	Mod           = sh.RunCmd("go", "mod")
)

func InGoPackageDir(pkg string, fn func() error) error {
	dir, err := PackageDir(os.ExpandEnv(pkg))
	if err != nil {
		return errors.Wrap(err, "PackageDir")
	}

	return magelib.InDirectory(dir, fn)
}

func PackageDir(pkg string) (string, error) {
	gopath, err := Env("GOPATH")
	if err != nil {
		return "", errors.Wrap(err, "Env [GOPATH]")
	}

	return fmt.Sprintf("%s/src/%s", gopath, pkg), nil
}

func Env(value string) (string, error) {
	out, err := magelib.GoEnvOut(value)
	if err != nil {
		return "", errors.Wrap(err, "GoEnvOut")
	}

	if out == "" {
		return "", errors.Errorf("%s is undefined", value)
	}

	return out, nil
}

func IsPackageCleanCmd(pkg string) magelib.Cmd {
	return func() error {
		return IsPackageClean(pkg)
	}
}

func IsPackageClean(pkg string) error {
	dir, err := PackageDir(os.ExpandEnv(pkg))
	if err != nil {
		return errors.Wrap(err, "PackageDir")
	}
	status, err := git.GitStatus(dir)
	if err != nil {
		return errors.Wrap(err, "GitStatus")
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
			return errors.Wrap(err, "GitBranch")
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

func UpdatePackageCmd(packageName string) magelib.Cmd {
	return func() error {
		return UpdatePackage(packageName)
	}
}

func ModTidyCmd() magelib.Cmd {
	return func() error {
		return Mod("tidy")
	}
}
