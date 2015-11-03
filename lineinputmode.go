package main

import (
	term "github.com/nsf/termbox-go"
	"unicode/utf8"
)

// TODO: handle aborted situation

type LineInputMode struct {
	// TODO: olds []string
	str   string
	start bool
	set   bool
}

func (f *LineInputMode) Handle(ev term.Event, cursor *Cursor, mode *string) {
	switch ev.Key {
	case term.KeyCtrlK:
		// TODO: revert to old
		*mode = "normal"
	case term.KeyEnter:
		f.set = true
		*mode = "normal"
	case term.KeySpace:
		if f.start {
			f.str = ""
			f.start = false
		}
		f.str += " "
	case term.KeyBackspace, term.KeyBackspace2:
		if f.start {
			f.str = ""
			f.start = false
			return
		}
		_, rlen := utf8.DecodeLastRuneInString(f.str)
		f.str = f.str[:len(f.str)-rlen]
	default:
		if ev.Mod&term.ModAlt != 0 {
			return
		}
		if ev.Ch != 0 {
			if f.start {
				f.str = ""
				f.start = false
			}
			f.str += string(ev.Ch)
		}
	}
}
