package main

import (
	term "github.com/nsf/termbox-go"
)

// Mode interface takes an event from terminal and handle it.
type Mode interface {
	Start()            // Start setup mode variables.
	End()              // End clear mode variables.
	Handle(term.Event) // Handle handles a terminal event.
	Status() string    // Status return current status of the mode.
	Error() string
}

type ModeSelector struct {
	current Mode // it stores one of follow modes.

	normal   *NormalMode
	find     *FindMode
	replace  *ReplaceMode
	gotoline *GotoLineMode
	exit     *ExitMode
}

// ChangeTo chage current mode.
// It also calls old current's End() and new current's Start().
func (ms *ModeSelector) ChangeTo(m Mode) {
	ms.current.End()
	ms.current = m
	ms.current.Start()
}
