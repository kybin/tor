package main

import (
	"fmt"
	"unicode/utf8"

	term "github.com/nsf/termbox-go"
)

type ReplaceMode struct {
	str   string
	start bool
	olds  []string

	mode *ModeSelector
}

func (m *ReplaceMode) Start() {
	term.SetInputMode(term.InputEsc)

	nm := m.mode.normal
	if nm.selection.on {
		m.str = nm.text.DataInside(nm.selection.MinMax())
	}
	m.start = true
}

func (m *ReplaceMode) End() {}

func (m *ReplaceMode) Handle(ev term.Event) {
	switch ev.Key {
	case term.KeyEsc, term.KeyCtrlK:
		if len(m.olds) == 0 {
			m.str = ""
		} else {
			m.str = m.olds[len(m.olds)-1]
		}
		m.mode.ChangeTo(m.mode.normal)
	case term.KeyEnter:
		m.mode.ChangeTo(m.mode.normal)
		m.olds = append(m.olds, m.str)
		saveConfig("replace", m.str)
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
			}
			m.str += string(ev.Ch)
		}
		m.start = false
	}
}

func (m *ReplaceMode) Status() string {
	return fmt.Sprintf("replace : %v", m.str)
}

func (m *ReplaceMode) Error() string {
	return ""
}
