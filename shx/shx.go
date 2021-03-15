package shx

import (
	"github.com/magefile/mage/sh"
)

type Cmd func() error
type CmdWithArgs func(args ...string) error

// RunWithVCmd -> sh.RunWithV as chainable Cmd
// RunWithV is like RunWith, but always sends the command's stdout to os.Stdout.
func RunWithVCmd(env map[string]string, cmd string, args ...string) CmdWithArgs {
	return func(args2 ...string) error {
		return sh.RunWithV(env, cmd, append(args, args2...)...)
	}
}

// CopyCmd -> sh.Copy as chainable Cmd
// Copy robustly copies the source file to the destination, overwriting the destination if necessary.
func CopyCmd(dst string, src string) Cmd {
	return func() error {
		return sh.Copy(dst, src)
	}
}

// RmCmd -> sh.Rm as chainable Cmd
// Rm removes the given file or directory even if non-empty. It will not return
// an error if the target doesn't exist, only if the target cannot be removed.
func RmCmd(path string) Cmd {
	return func() error {
		return sh.Rm(path)
	}
}

func ChainCmds(fns ...Cmd) error {
	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

func ChainCmdsCmd(fns ...Cmd) Cmd {
	return func() error {
		return ChainCmds(fns...)
	}
}
