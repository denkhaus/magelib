package common

import (
	"github.com/juju/errors"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"gopkg.in/pipe.v2"
	"path/filepath"
	"os"
)

type OutCmdFunc func(args ...string) (string, error)

var (
	MkTempDir = sh.OutCmd("mktemp", "-d")
	GoInstall = sh.RunCmd("go", "install")
	GoUpdate  = sh.RunCmd("go", "get", "-u")
	GoGet     = sh.RunCmd("go", "get")
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

func InDirectory(path string, fn func() error)(err error){
	if !filepath.IsAbs(path) {
		path, err = filepath.Abs(path)
		if err != nil{
			return errors.Annotate(err, "Abs")
		}		 
	}

	oldPath, err := os.Getwd()
	if err != nil {
		return errors.Annotate(err, "Getwd")
	}
	
	if err := os.Chdir(path);err != nil {
		return errors.Annotate(err, "Chdir")
	}
	defer func(p string){
		err = os.Chdir(p)
	}(oldPath)

	return fn()
}
