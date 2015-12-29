package main

import (
	term "github.com/nsf/termbox-go"
)

// Mode interface takes an event from terminal,
// then creates actions, let main loop to do something.
// TODO: also take tor information.
type Mode interface {
	Handle(term.Event) ([]*Action, error)
}
