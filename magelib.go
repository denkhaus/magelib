package magelib

import (
	"os"
	"path/filepath"

	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
)

var (
	MkTempDir   = sh.OutCmd("mktemp", "-d")
	GoInstall   = sh.RunCmd("go", "install")
	GoUpdate    = sh.RunCmd("go", "get", "-u")
	GoGet       = sh.RunCmd("go", "get")
	GoEnvOut    = sh.OutCmd("go", "env")
	GoModOut    = sh.RunCmd("go", "mod")
	GoModVendor = sh.RunCmd("go", "mod", "vendor")
	GoModTidy   = sh.RunCmd("go", "mod", "tidy")
)

func InDirectory(path string, cmd Cmd) (err error) {
	path = os.ExpandEnv(path)
	if !filepath.IsAbs(path) {
		path, err = filepath.Abs(path)
		if err != nil {
			return errors.Wrap(err, "Abs")
		}
	}

	oldPath, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "Getwd")
	}

	if err := os.Chdir(path); err != nil {
		return errors.Wrap(err, "Chdir")
	}

	defer func(p string) {
		err = os.Chdir(p)
	}(oldPath)

	return cmd()
}
