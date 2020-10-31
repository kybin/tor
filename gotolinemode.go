package main

import (
	"fmt"
	"strconv"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

type GotoLineMode struct {
	linestr string

	cursor *Cursor
}

func (m *GotoLineMode) Start() {}

func (m *GotoLineMode) End() {}

func (m *GotoLineMode) Handle(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyEsc, tcell.KeyCtrlK:
		m.linestr = ""
		tor.ChangeMode(tor.normal)
	case tcell.KeyEnter:
		if m.linestr == "" {
			tor.ChangeMode(tor.normal)
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
		tor.ChangeMode(tor.normal)
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if m.linestr == "" {
			return
		}
		_, rlen := utf8.DecodeLastRuneInString(m.linestr)
		m.linestr = m.linestr[:len(m.linestr)-rlen]
	default:
		if ev.Rune() != 0 {
			_, err := strconv.Atoi(string(ev.Rune()))
			if err == nil {
				m.linestr += string(ev.Rune())
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
