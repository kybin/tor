package main

import (
	"os"

	term "github.com/nsf/termbox-go"
)

type ExitMode struct {
	f      string // name of this file.
	cursor *Cursor
	mode   *ModeSelector
}

func (m *ExitMode) Start() {
	term.SetInputMode(term.InputEsc)
}

func (m *ExitMode) End() {}

func (m *ExitMode) Handle(ev term.Event) {
	if ev.Ch == 'y' {
		saveLastPosition(m.f, m.cursor.l, m.cursor.b)
		term.Close()
		os.Exit(0)
	} else if ev.Ch == 'n' || ev.Key == term.KeyEsc || ev.Key == term.KeyCtrlK {
		m.mode.ChangeTo(m.mode.normal)
	}
}

func (m *ExitMode) Status() string {
	return "Buffer modified. Do you really want to quit? (y/n)"
}

func (m *ExitMode) Error() string {
	return ""
}
