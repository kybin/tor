package main

import (
	"fmt"
	term "github.com/nsf/termbox-go"
	"strconv"
	"unicode/utf8"
)

type GotoLineMode struct {
	linestr string

	cursor *Cursor
	mode   *ModeSelector
}

func (m *GotoLineMode) Start() {}

func (m *GotoLineMode) End() {}

func (m *GotoLineMode) Handle(ev term.Event) {
	switch ev.Key {
	case term.KeyCtrlK:
		m.linestr = ""
		m.mode.ChangeTo(m.mode.normal)
	case term.KeyEnter:
		if m.linestr == "" {
			m.mode.ChangeTo(m.mode.normal)
			return
		}
		n, err := strconv.Atoi(m.linestr)
		if err != nil {
			panic("cannot convert gotoline string to int")
		}
		// line number starts with 1.
		// but internally it starts with 0.
		// so we should n - 1, except 0 will treated as 0.
		if n != 0 {
			n--
		}
		m.cursor.GotoLine(n)
		m.linestr = ""
		m.mode.ChangeTo(m.mode.normal)
	case term.KeyBackspace, term.KeyBackspace2:
		if m.linestr == "" {
			return
		}
		_, rlen := utf8.DecodeLastRuneInString(m.linestr)
		m.linestr = m.linestr[:len(m.linestr)-rlen]
	default:
		if ev.Ch != 0 {
			_, err := strconv.Atoi(string(ev.Ch))
			if err == nil {
				m.linestr += string(ev.Ch)
			}
		}
	}
}

func (m *GotoLineMode) Status() string {
	return fmt.Sprintf("goto : %v", m.linestr)
}

func (m *GotoLineMode) Error() string {
	return ""
}
