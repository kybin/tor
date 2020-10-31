package main

import (
	"github.com/gdamore/tcell/v2"
)

type ExitMode struct {
	f      string // name of this file.
	cursor *Cursor
	tor    *Tor
	exit   func()
}

func (m *ExitMode) Start() {
	if !tor.normal.text.edited {
		m.exit()
	}
}

func (m *ExitMode) End() {}

func (m *ExitMode) Handle(ev *tcell.EventKey) {
	if ev.Rune() == 'y' {
		m.exit()
	} else if ev.Rune() == 'n' || ev.Key() == tcell.KeyEsc || ev.Key() == tcell.KeyCtrlK {
		tor.ChangeMode(tor.normal)
	}
}

func (m *ExitMode) Status() string {
	return "Buffer modified. Do you really want to quit? (y/n)"
}

func (m *ExitMode) Error() string {
	return ""
}
