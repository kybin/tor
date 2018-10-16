package main

import (
	"os/exec"
)

// cmdGroupType defines how commands in group will run.
type cmdGroupType int

const (
	// andCmdGroup will run all commands.
	// It will not run a command if one of
	// the prev commands NOT exists OR failed to run.
	andCmdGroup = cmdGroupType(iota)
	// orCmdGroup will run first command which exists.
	// It will not run a command if one of
	// the prev commands exists BUT failed to run.
	orCmdGroup
)

// cmdGroup is a group of commands to run.
type cmdGroup struct {
	kind cmdGroupType
	cmds []*exec.Cmd
}

// CombinedOutput runs registered commands and returns 'first' error
// as exec.Cmd.CombinedOutput style.
// If an error occurred, it will not run the rest of commands.
func (r cmdGroup) CombinedOutput() ([]byte, error) {
	for _, c := range r.cmds {
		if c.Path == "" {
			// Didn't find the command.
			if r.kind == orCmdGroup {
				continue
			}
			// If a command missing in andCmdGroup, it should fail.
			// The best way to do that is letting it run.
		}
		out, err := c.CombinedOutput()
		if err != nil {
			return out, err
		}
		if r.kind == orCmdGroup {
			break
		}
	}
	return nil, nil
}
