package magelib

import (
	"github.com/juju/errors"
	"gopkg.in/pipe.v2"
)

type Cmd func() error
type CmdWithArgs func(args ...string) error
type OutCmdFunc func(args ...string) (string, error)
type ArgsMap map[string]string

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
