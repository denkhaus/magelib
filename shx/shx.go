package shx

import (
	"os"
	"os/exec"
	"strings"

	"github.com/denkhaus/logging"
	"github.com/denkhaus/magelib"
	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
	"gopkg.in/pipe.v2"
)

// RunWithVCmd -> sh.RunWithV as chainable Cmd
// RunWithV is like RunWith, but always sends the command's stdout to os.Stdout.
func RunWithVCmd(env magelib.ArgsMap, cmd string, args ...string) magelib.CmdWithArgs {
	return func(args2 ...string) error {
		return sh.RunWithV(env, cmd, append(args, args2...)...)
	}
}

// RunVCmd returns a function that runs a command with variable arguments.
//
// The cmd parameter is the command to run, and args are the initial arguments.
// The returned function takes additional arguments (args2), which are appended to the initial arguments.
// It returns an error if the command fails.
func RunVCmd(cmd string, args ...string) magelib.CmdWithArgs {
	return func(args2 ...string) error {
		return sh.RunV(cmd, append(args, args2...)...)
	}
}

// CopyCmd -> sh.Copy as chainable Cmd
// Copy robustly copies the source file to the destination, overwriting the destination if necessary.
func CopyCmd(dst string, src string) magelib.Cmd {
	return func() error {
		return sh.Copy(dst, src)
	}
}

// RmCmd -> sh.Rm as chainable Cmd
// Rm removes the given file or directory even if non-empty. It will not return
// an error if the target doesn't exist, only if the target cannot be removed.
func RmCmd(path string) magelib.Cmd {
	return func() error {
		return sh.Rm(path)
	}
}

// RunPipeVerbose runs a pipe.Pipe with verbose output.
//
// The function takes a pipe.Pipe as a parameter and returns an error. It creates a new pipe.State
// with os.Stdout and os.Stderr as the output and error streams. It then calls the pipe function
// with the state and checks if there is an error. If there is no error, it calls the RunTasks
// method of the state. Finally, it returns the error.
func RunPipeVerbose(p pipe.Pipe) error {
	s := pipe.NewState(os.Stdout, os.Stderr)
	err := p(s)
	if err == nil {
		err = s.RunTasks()
	}

	return err
}

// IsAppInstalled returns the install state of an specific app on linux systems
func IsAppInstalled(appName string) (bool, error) {
	appPath, err := exec.LookPath(appName)
	if err != nil {
		return false, nil
	}

	if appPath != "" {
		return true, nil
	}

	// try second way
	out, err := sh.Output("which", appName)
	if err != nil {
		logging.Errorf("error while running 'which' -> %s", out)
		return false, errors.Wrap(err, "which")
	}

	return !strings.Contains(out, "not found"), nil
}
