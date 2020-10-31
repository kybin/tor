package main

import "github.com/gdamore/tcell/v2"

// Mode interface takes an event from terminal and handle it.
type Mode interface {
	Start()                 // Start set up mode variables.
	End()                   // End clear mode variables.
	Handle(*tcell.EventKey) // Handle handles a terminal event.
	Status() string         // Status returns a current status of the mode.
	Error() string          // Error indicates an error from the last event. It should be empty when there was no error.
}
