package main

import (
	term "github.com/nsf/termbox-go"
)

// Mode interface takes an event from terminal and handle it.
type Mode interface {
	Start()            // Start set up mode variables.
	End()              // End clear mode variables.
	Handle(term.Event) // Handle handles a terminal event.
	Status() string    // Status returns a current status of the mode.
	Error() string     // Error indicates an error from the last event. It should be empty when there was no error.
}

type ModeSelector struct {
	// current is a mode that will handle terminal events.
	current Mode

	// All modes that could be current mode.
	normal   *NormalMode
	find     *FindMode
	replace  *ReplaceMode
	gotoline *GotoLineMode
	exit     *ExitMode
}

// ChangeTo changes current mode.
// It also calls old current's End() and new current's Start().
func (ms *ModeSelector) ChangeTo(m Mode) {
	ms.current.End()
	ms.current = m
	ms.current.Start()
}
