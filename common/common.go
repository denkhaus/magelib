package common

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/denkhaus/logging"
	"github.com/juju/errors"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	pipe "gopkg.in/pipe.v2"
)

type OutCmdFunc func(args ...string) (string, error)

var (
	GitCheckout    = sh.RunCmd("git", "checkout")
	MkTempDir      = sh.OutCmd("mktemp", "-d")
	GoInstall      = sh.RunCmd("go", "install")
	GoUpdate       = sh.RunCmd("go", "get", "-u")
	GoModuleVendor = sh.RunCmd("go", "mod", "vendor")
	GoModuleTidy   = sh.RunCmd("go", "mod", "tidy")
	GoGet          = sh.RunCmd("go", "get")
	GoEnvOut       = sh.OutCmd("go", "env")
	GoModOut       = sh.RunCmd("go", "mod")
)

func PipeOutCmd(fn OutCmdFunc, args ...string) pipe.Pipe {
	return pipe.TaskFunc(func(s *pipe.State) error {
		output, err := fn(args...)
		if len(output) > 0 {
			if _, err := s.Stdout.Write([]byte(output)); err != nil {
				return errors.Annotate(err, "write [stdout]")
			}
		}

		return err
	})
}

func HandleError(err error) {
	if err != nil {
		mg.Fatal(1, err)
	}
}

func InDirectory(path string, fn func() error) (err error) {
	path = os.ExpandEnv(path)
	if !filepath.IsAbs(path) {
		path, err = filepath.Abs(path)
		if err != nil {
			return errors.Annotate(err, "Abs")
		}
	}

	oldPath, err := os.Getwd()
	if err != nil {
		return errors.Annotate(err, "Getwd")
	}

	if err := os.Chdir(path); err != nil {
		return errors.Annotate(err, "Chdir")
	}
	defer func(p string) {
		err = os.Chdir(p)
	}(oldPath)

	return fn()
}

// RunVWith is like RunV, but with env variables.
func RunVWith(env map[string]string, cmd string, args ...string) error {
	_, err := sh.Exec(env, os.Stdout, os.Stderr, cmd, args...)
	return err
}

func InGoPackageDir(pkg string, fn func() error) error {
	path := GoPackageDir(os.ExpandEnv(pkg))
	return InDirectory(path, fn)
}

func EnsureBranchInRepositoryFunc(path string, branchName string) func() error {
	return func() error {
		return EnsureBranchInRepository(path, branchName)
	}
}

func EnsureBranchInRepository(path string, branchName string) error {
	return InDirectory(path, func() error {
		branch, err := git.GitBranch(path)
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

func EnsureBranchInGoPackageFunc(pkg string, branchName string) func() error {
	return func() error {
		return EnsureBranchInGoPackage(pkg, branchName)
	}
}

func EnsureBranchInGoPackage(pkg string, branchName string) error {
	return InGoPackageDir(pkg, func() error {
		logging.Infof("checkout %q in go pkg %s", branchName, pkg)
		return GitCheckout(branchName)
	})
}

func GoPackageDir(pkg string) string {
	return fmt.Sprintf("%s/src/%s", GoEnvValue("GOPATH"), pkg)
}

func GoEnvValue(value string) string {
	out, err := GoEnvOut(value)
	HandleError(err)
	if out == "" {
		HandleError(errors.Errorf("%s is undefined", value))
	}

	return out
}
