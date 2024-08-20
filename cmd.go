package magelib

import (
	"github.com/pkg/errors"
	"gopkg.in/pipe.v2"
)

type Cmd func() error
type CmdWithArgs func(args ...string) error
type OutCmdFunc func(args ...string) (string, error)
type ArgsMap map[string]string

// PipeOutCmd creates a pipe task function that executes the given OutCmdFunc function with the provided arguments and writes the output to stdout.
//
// Parameters:
// - fn: The OutCmdFunc function to be executed.
// - args: The arguments to be passed to the OutCmdFunc function.
//
// Returns:
// - pipe.Pipe: The pipe task function.
func PipeOutCmd(fn OutCmdFunc, args ...string) pipe.Pipe {
	return pipe.TaskFunc(func(s *pipe.State) error {
		output, err := fn(args...)
		if len(output) > 0 {
			if _, err := s.Stdout.Write([]byte(output)); err != nil {
				return errors.Wrap(err, "write [stdout]")
			}
		}

		return err
	})
}

// PipeCmd creates a pipe task function that executes the given CmdWithArgs function with the provided arguments.
//
// Parameters:
// - fn: The CmdWithArgs function to be executed.
// - args: The arguments to be passed to the CmdWithArgs function.
//
// Returns:
// - pipe.Pipe: The pipe task function.
func PipeCmd(fn CmdWithArgs, args ...string) pipe.Pipe {
	return pipe.TaskFunc(func(s *pipe.State) error {
		return fn(args...)
	})
}

// Chain executes a chain of commands in sequence.
//
// Parameters:
// - fns: A variable number of commands to be executed.
//
// Returns:
// - error: The error returned by the first failing command, or nil if all commands succeed.
func Chain(fns ...Cmd) error {
	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

// ChainCmds creates a new command that chains multiple commands together.
//
// Parameters:
// - fns: A variable number of commands to be chained.
//
// Returns:
// - Cmd: A new command that executes the chained commands.
func ChainCmds(fns ...Cmd) Cmd {
	return func() error {
		return Chain(fns...)
	}
}
