package shx

import (
	"github.com/denkhaus/magelib"
	"github.com/magefile/mage/sh"
)

// RunWithVCmd -> sh.RunWithV as chainable Cmd
// RunWithV is like RunWith, but always sends the command's stdout to os.Stdout.
func RunWithVCmd(env magelib.Args, cmd string, args ...string) magelib.CmdWithArgs {
	return func(args2 ...string) error {
		return sh.RunWithV(env, cmd, append(args, args2...)...)
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
