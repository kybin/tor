package main

import (
	"fmt"
	"os"
	term "github.com/nsf/termbox-go"
)

func clear_term() {
	term.Clear(term.ColorDefault, term.ColorDefault)
	term.Flush()
}

func draw(txt text) {
	//termw, termh := term.Size()
	for linenum, line := range txt {
		visoff := 0
		for off, ch := range line {
			if ch == '\t' {
				off++
				visoff += taboffset
			} else {
				term.SetCell(visoff, linenum, rune(ch), term.ColorWhite, term.ColorDefault)
				off++
				visoff++
			}
		}
	}
	term.Flush()
}



func setState(c *cursor) {
	termw, termh := term.Size()
	stateline := termh - 1
	linenum := c.linenum
	byteoff := c.off
	visoff := c.visoff
	cursoroff := c.cursorOffset()

	state := fmt.Sprintf("linenum:%v, byteoff:%v, visoff:%v, cursoroff:%v", linenum, byteoff, visoff, cursoroff)
	for off:=0 ; off<termw ; off++ {
		term.SetCell(off, stateline, ' ', term.ColorBlack, term.ColorWhite)
	}
	for off, ch := range state {
		term.SetCell(off, stateline, rune(ch), term.ColorBlack, term.ColorWhite)
	}
}

func main() { // main loop
	err := term.Init()
	if err != nil {
		panic(err)
	}
	defer term.Close()

	clear_term()

	args := os.Args[1:]
	if len(args)==0 {
		print("please, set text file")
		os.Exit(1)
	}
	f := args[0]
	text := open(f)
	draw(text)
	cursor := initializeCursor(text)
	setState(cursor)
	term.Flush()

loop:
	for {
		ev := term.PollEvent()
		switch ev.Type {
		case term.EventKey:
			switch ev.Key {
			case term.KeyCtrlW:
				break loop
			case term.KeyArrowLeft:
				cursor.moveLeft()
			case term.KeyArrowRight:
				cursor.moveRight()
			case term.KeyArrowUp:
				cursor.moveUp()
			case term.KeyArrowDown:
				cursor.moveDown()
			}

			// for unknown reason ev.Mod is not a term.ModAlt so I enforce it.
			ev.Mod = term.ModAlt
			if ev.Mod == term.ModAlt {
				switch ev.Ch {
				case 'j': cursor.moveLeft()
				case 'l': cursor.moveRight()
				case 'i': cursor.moveUp()
				case 'k': cursor.moveDown()
				case 'm': cursor.moveBow()
				case '.': cursor.moveEow()
				case 'u': cursor.moveBol()
				case 'o': cursor.moveEol()
				}
			}
		// case term.EventResize:
		//	something()
		}
		setVisualCursor(cursor)
		setState(cursor)
		term.Flush()
	}
}
