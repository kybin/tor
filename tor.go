package main

import (
	"fmt"
	term "github.com/nsf/termbox-go"
)

var (
	taboffset = 4
	termWhite   = term.ColorWhite
	termDefault = term.ColorDefault
)

type cursor struct {
	line,
	offset int
}

func init_cursor() cursor {
	term.SetCursor(0, 0)
	c := cursor{0, 0}
	return c
}

func (c *cursor) move(l, o int) {
	// shift cursor but prevent negative fields
	c.line = c.line + l
	if c.line < 0 {
		c.line = 0
	}
	c.offset = c.offset + o
	if c.offset < 0 {
		c.offset = 0
	}

	term.SetCursor(c.line, c.offset)
}

func clear_term() {
	term.Clear(termDefault, termDefault)
	term.Flush()
}

func print_text(t text) {
	width, height := term.Size()
	height -= 1 // for print state
	print_state(fmt.Sprintf("width : %v, height : %v ", width, height))
	fmt.Println("num line :", len(t), "height of termbox :", height)

	for l := 0; l < height && l < len(t); l++ { // should not exceeded both line num and term size
		line := t[l]
		choff, visoff := 0, 0
		for visoff < width && choff < len(line) { // same here
			ch := line[choff]
			if ch == '\t' {
				choff++
				visoff += taboffset
			} else {
				term.SetCell(visoff, l, rune(ch), termWhite, termDefault)
				choff++
				visoff++
			}
		}
		fmt.Println("")
	}
	term.Flush()
}

func print_state(s string) {
	_, termh := term.Size()
	for off, ch := range s {
		term.SetCell(off, termh, rune(ch), termWhite, termDefault)
	}
}

func main() { // main loop
	err := term.Init()
	if err != nil {
		panic(err)
	}
	defer term.Close()

	clear_term()

	f := "/home/kybin/go/src/github.com/coldmine/tor/text"
	txt := open(f)
	print_text(txt)

	cursor := init_cursor()

loop:
	for {
		ev := term.PollEvent()
		switch ev.Type {
		case term.EventKey:
			switch ev.Key {
			case term.KeyCtrlW:
				break loop
			case term.KeyArrowLeft:
				cursor.move(-1, 0)
			case term.KeyArrowRight:
				cursor.move(1, 0)
			case term.KeyArrowUp:
				cursor.move(0, -1)
			case term.KeyArrowDown:
				cursor.move(0, 1)
			}

			// for unknown reason ev.Mod is not a term.ModAlt so I enforce it.
			ev.Mod = term.ModAlt
			if ev.Mod == term.ModAlt {
				switch ev.Ch {
				case 'j':
					cursor.move(-1, 0)
				case 'l':
					cursor.move(1, 0)
				case 'i':
					cursor.move(0, -1)
				case 'k':
					cursor.move(0, 1)
				}
			}
		// case term.EventResize:
		//	something()
		}
		term.Flush()
	}
}
