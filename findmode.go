package main

import (
	term "github.com/nsf/termbox-go"
	"unicode/utf8"
	"fmt"
)

// TODO: handle aborted situation

type FindMode struct {
	// TODO: olds []string
	str   string
	start bool
	set   bool

	text *Text
	selection *Selection
	mode *ModeSelector
}

func (m *FindMode) Start() {
	if m.selection.on {
		m.set = true
		min, max := m.selection.MinMax()
		m.str = m.text.DataInside(min, max)
		m.selection.on = false
		return
	}
	m.start = true
}

func (m *FindMode) End() {}

func (m *FindMode) Handle(ev term.Event) {
	switch ev.Key {
	case term.KeyCtrlK:
		// TODO: revert to old
		m.mode.ChangeTo(m.mode.normal)
	case term.KeyEnter:
		m.set = true
		m.mode.ChangeTo(m.mode.normal)
	case term.KeySpace:
		if m.start {
			m.str = ""
			m.start = false
		}
		m.str += " "
	case term.KeyBackspace, term.KeyBackspace2:
		if m.start {
			m.str = ""
			m.start = false
			return
		}
		_, rlen := utf8.DecodeLastRuneInString(m.str)
		m.str = m.str[:len(m.str)-rlen]
	default:
		if ev.Mod&term.ModAlt != 0 {
			return
		}
		if ev.Ch != 0 {
			if m.start {
				m.str = ""
				m.start = false
			}
			m.str += string(ev.Ch)
		}
	}
}

func (m *FindMode) Status() string {
	return fmt.Sprintf("find : %v", m.str)
}
