package golang

import (
	"fmt"
	"os"

	"github.com/denkhaus/logging"
	"github.com/denkhaus/magelib"

	"github.com/denkhaus/magelib/git"
	"github.com/denkhaus/magelib/shx"
	"github.com/magefile/mage/sh"
)

var (
	UpdatePackage  = sh.RunCmd("go", "get", "-u")
	InstallPackage = sh.RunCmd("go", "install")
	Mod            = sh.RunCmd("go", "mod")
)

// InGoPackageDir executes a function within the directory of a Go package in GOPATH.
//
// pkg: the name of the Go package.
// fn: the function to be executed within the package directory.
// error: any error that occurred during execution.
func InGoPackageDir(pkg string, fn func() error) error {
	dir, err := PackageDir(os.ExpandEnv(pkg))
	if err != nil {
		return magelib.Fatal(err, "PackageDir")
	}

	return magelib.InDirectory(dir, fn)
}

// PackageDir returns the directory path of a Go package in GOPATH.
//
// pkg: the name of the Go package.
// string: the directory path of the Go package.
// error: any error that occurred while retrieving the directory path.
func PackageDir(pkg string) (string, error) {
	gopath, err := Env("GOPATH")
	if err != nil {
		return "", magelib.Fatal(err, "Env [GOPATH]")
	}

	return fmt.Sprintf("%s/src/%s", gopath, pkg), nil
}

// Env returns the value of the specified golang environment variable.
//
// value: the name of the environment variable to retrieve.
// string: the value of the environment variable, or an empty string if it is undefined.
// error: any error that occurred while retrieving the environment variable value.
func Env(value string) (string, error) {
	out, err := magelib.GoEnvOut(value)
	if err != nil {
		return "", magelib.Fatal(err, "GoEnvOut")
	}

	if out == "" {
		return "", magelib.Fatalf("%s is undefined", value)
	}

	return out, nil
}

// IsPackageCleanCmd returns a function that checks if a Go package is clean.
//
// pkg: the path to the Go package.
// magelib.Cmd: a function that returns an error if the package is not clean.
func IsPackageCleanCmd(pkg string) magelib.Cmd {
	return func() error {
		return IsPackageClean(pkg)
	}
}

// IsPackageClean checks if a Go package has
// unstaged files that have been changed in repo
// or staged files have been modified in repo
// or repo is not in sync with the remote repo
//
// path is the path to the repository.
// status is the status information of the repository.
// Returns an error if the repository status is not valid, otherwise nil.
func IsPackageClean(pkg string) error {
	dir, err := PackageDir(os.ExpandEnv(pkg))
	if err != nil {
		return magelib.Fatal(err, "PackageDir")
	}
	status, err := git.GitStatus(dir)
	if err != nil {
		return magelib.Fatal(err, "GitStatus")
	}

	return git.FormatStatusError(pkg, status)
}

// EnsureBranchInPackageCmd returns a function that ensures a specific branch is checked out in a Go package.
//
// pkg: the path to the Go package.
// branchName: the name of the branch to ensure.
//
// magelib.Cmd: a function that returns an error if there was a problem ensuring the branch.
func EnsureBranchInPackageCmd(pkg string, branchName string) magelib.Cmd {
	return func() error {
		return EnsureBranchInPackage(pkg, branchName)
	}
}

// EnsureBranchInPackage ensures that a specific branch is checked out in a Go package.
//
// pkg: the path to the Go package.
// branchName: the name of the branch to ensure.
//
// error: an error if there was a problem ensuring the branch.
func EnsureBranchInPackage(pkg string, branchName string) error {
	return InGoPackageDir(pkg, func() error {
		branch, err := git.Branch()
		if err != nil {
			return magelib.Fatal(err, "GitBranch")
		}

		if branch != branchName {
			logging.Infof("checkout [%s] in go pkg [%s]", branchName, pkg)
			return git.Checkout(branchName)
		}

		logging.Infof("branch [%s] is checked out in go pkg [%s]", branchName, pkg)
		return nil
	})
}

// UpdateModuleCmd returns a command that updates the Go module at the specified path.
//
// path: the directory path where the Go module is located.
// vendor: a boolean indicating whether to vendor the dependencies.
//
// magelib.Cmd: a command that updates the Go module.
func UpdateModuleCmd(path string, vendor bool) magelib.Cmd {
	return func() error {
		return UpdateGoModule(path, vendor)
	}
}

// UpdateGoModule updates the Go module at the specified path.
//
// path: the directory path where the Go module is located.
// vendor: a boolean indicating whether to vendor the dependencies.
//
// error: an error if the update operation fails.
func UpdateGoModule(path string, vendor bool) error {
	env := magelib.ArgsMap{
		"GO111MODULE": "on",
	}

	return magelib.InDirectory(path, magelib.ChainCmds(
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

func InstallPackageCmd(packageName string) magelib.Cmd {
	return func() error {
		return InstallPackage(packageName)
	}
}

func EnsurePackageInstalledCmd(appName string, packageName string) magelib.Cmd {
	return func() error {
		ok, err := shx.IsAppInstalled(appName)
		if err != nil {
			return magelib.Fatal(err, "IsAppInstalled")
		}
		if !ok {
			return InstallPackage(packageName)
		}

		return nil
	}
}
