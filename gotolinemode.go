package main

import (
	term "github.com/nsf/termbox-go"
	"strconv"
	"unicode/utf8"
)

type GotoLineMode struct {
	linestr string
}

func (g *GotoLineMode) Handle(ev term.Event, cursor *Cursor, mode *string) {
	switch ev.Key {
	case term.KeyCtrlK:
		g.linestr = ""
		*mode = "normal"
	case term.KeyEnter:
		if g.linestr == "" {
			*mode = "normal"
			return
		}
		n, err := strconv.Atoi(g.linestr)
		if err != nil {
			panic("cannot convert gotoline string to int")
		}
		// line number starts with 1.
		// but internally it starts with 0.
		// so we should n - 1, except 0 will treated as 0.
		if n != 0 {
			n--
		}
		cursor.GotoLine(n)
		g.linestr = ""
		*mode = "normal"
	case term.KeyBackspace, term.KeyBackspace2:
		if g.linestr == "" {
			return
		}
		_, rlen := utf8.DecodeLastRuneInString(g.linestr)
		g.linestr = g.linestr[:len(g.linestr)-rlen]
	default:
		if ev.Ch != 0 {
			_, err := strconv.Atoi(string(ev.Ch))
			if err == nil {
				g.linestr += string(ev.Ch)
			}
		}
	}
}
