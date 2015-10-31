package main

import (
	"unicode/utf8"
	term "github.com/nsf/termbox-go"
)

// TODO: print status when FindMode.Set() is executed.
// ex) "find \"some find string\". alt+f to next, alt+b to prev"

type FindMode struct {
	// TODO: olds []string
	findstr string
	juststart bool
}

func (f *FindMode) Handle(ev term.Event, cursor *Cursor, mode *string) {
	switch ev.Key {
	case term.KeyCtrlK:
		// TODO: revert to old
		*mode = "normal"
	case term.KeyEnter:
		*mode = "normal"
	case term.KeySpace:
		if f.juststart {
			f.findstr = ""
			f.juststart = false
		}
		f.findstr += " "
	case term.KeyBackspace, term.KeyBackspace2:
		if f.juststart {
			f.findstr = ""
			f.juststart = false
			return
		}
		_, rlen := utf8.DecodeLastRuneInString(f.findstr)
		f.findstr = f.findstr[:len(f.findstr)-rlen]
	default:
		if ev.Mod & term.ModAlt != 0 {
			return
		}
		if ev.Ch != 0 {
			if f.juststart {
				f.findstr = ""
				f.juststart = false
			}
			f.findstr += string(ev.Ch)
		}
	}
}

