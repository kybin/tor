package main

import (
	"unicode/utf8"
	term "github.com/nsf/termbox-go"
)

type FindMode struct {
	findstr string
	juststart bool
}

func (f *FindMode) Handle(ev term.Event, cursor *Cursor) {
	switch ev.Key {
	case term.KeyEnter:
		if f.findstr == "" {
			return
		}
		cursor.GotoNext(f.findstr)
		return
	case term.KeySpace:
		f.findstr += " "
		return
	case term.KeyBackspace, term.KeyBackspace2:
		if f.juststart {
			f.findstr = ""
			return
		}
		_, rlen := utf8.DecodeLastRuneInString(f.findstr)
		f.findstr = f.findstr[:len(f.findstr)-rlen]
		return
	}
	if ev.Mod & term.ModAlt != 0 && ev.Ch != 0 {
		switch ev.Ch {
		case 'j':
			if f.findstr == "" {
				return
			}
			cursor.GotoPrev(f.findstr)
		case 'l':
			if f.findstr == "" {
				return
			}
			cursor.GotoNext(f.findstr)
		case 'i':
			if f.findstr == "" {
				return
			}
			cursor.GotoFirst(f.findstr)
		case 'k':
			if f.findstr == "" {
				return
			}
			cursor.GotoLast(f.findstr)
		case 'm':
			if f.findstr == "" {
				return
			}
			cursor.GotoPrevWord(f.findstr)
		case '.':
			if f.findstr == "" {
				return
			}
			cursor.GotoNextWord(f.findstr)
		}
	} else if ev.Ch != 0 {
		if f.juststart {
			f.juststart = false
			f.findstr = ""
		}
		f.findstr += string(ev.Ch)
	}
}
