package main

import (
	"unicode/utf8"
	term "github.com/nsf/termbox-go"
)

type FindMode struct {
	// TODO: olds []string
	findstr string
	start bool
	set bool
}

func (f *FindMode) Handle(ev term.Event, cursor *Cursor, mode *string) {
	switch ev.Key {
	case term.KeyCtrlK:
		// TODO: revert to old
		*mode = "normal"
	case term.KeyEnter:
		f.set = true
		*mode = "normal"
	case term.KeySpace:
		if f.start {
			f.findstr = ""
			f.start = false
		}
		f.findstr += " "
	case term.KeyBackspace, term.KeyBackspace2:
		if f.start {
			f.findstr = ""
			f.start = false
			return
		}
		_, rlen := utf8.DecodeLastRuneInString(f.findstr)
		f.findstr = f.findstr[:len(f.findstr)-rlen]
	default:
		if ev.Mod & term.ModAlt != 0 {
			return
		}
		if ev.Ch != 0 {
			if f.start {
				f.findstr = ""
				f.start = false
			}
			f.findstr += string(ev.Ch)
		}
	}
}

