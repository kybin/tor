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

func textToDrawBuffer(txt text) [][]rune {
	drawbuf := make([][]rune, 0)
	for _ , line := range txt {
		linebuf := make([]rune, 0)
		for _, ch := range line {
			if ch == '\t' {
				for i:=0 ; i<taboffset ; i++ { linebuf = append(linebuf, rune(' ')) }
			} else {
				linebuf = append(linebuf, rune(ch))
			}
		}
		drawbuf = append(drawbuf, linebuf)
	}
	return drawbuf
}

func clipDrawBuffer(drawbuf [][]rune, window *viewer) [][]rune {
	clipbuf := make([][]rune, 0)
	xstart, ystart := window.min.X, window.min.Y
	xend, yend := window.max.X, window.max.Y
	//xstart, ystart := 0,0
	//xend, yend := 20, 10
	yend = min(yend+1, len(drawbuf))
	if yend < ystart {
		// if then, we don't have a place for draw
		return clipbuf
	}
	for _, origbuf := range drawbuf[ystart:yend] {
		minoff := xstart
		maxoff := xend+1
		maxoff = min(maxoff, len(origbuf))
		if maxoff > minoff {
			clipbuf = append(clipbuf, origbuf[minoff:maxoff])
		} else {
			clipbuf = append(clipbuf, make([]rune, 0))
		}
	}
	return clipbuf
}

func draw(clipbuf [][]rune) {
	sizex, sizey := term.Size()
	for x := 0 ; x < sizex ; x++ {
		for y := 0 ; y < sizey ; y++ {
			term.SetCell(x, y, ' ', term.ColorDefault, term.ColorDefault)
		}
	}
	for linenum, line := range clipbuf {
		for off, r := range line {
			term.SetCell(off, linenum, r, term.ColorWhite, term.ColorDefault)
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
	//term.SetInputMode(term.InputEsc)
	clear_term()

	args := os.Args[1:]
	if len(args)==0 {
		fmt.Println("please, set text file")
		return
	}
	f := args[0]
	view := newViewer()
	text := open(f)
	db := textToDrawBuffer(text)
	draw(db)
	cursor := initializeCursor(text)
	setState(cursor)
	term.Flush()

	events := make(chan term.Event, 20)
	go func() {
		for {
			events <- term.PollEvent()
		}
	}()
	for {
		select {
		case ev := <-events:
			switch ev.Type {
			case term.EventKey:
				switch ev.Key {
				case term.KeyCtrlW:
					return
				case term.KeyArrowLeft:
					cursor.moveLeft()
				case term.KeyArrowRight:
					cursor.moveRight()
				case term.KeyArrowUp:
					cursor.moveUp()
				case term.KeyArrowDown:
					cursor.moveDown()
				}
				ev.Mod=term.ModAlt
				if (ev.Mod&term.ModAlt) != 0 {
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
			}
		// case term.EventResize:
		//	view.resize()
		//	view.clear()
		//	view.draw()
		}
		setVisualCursor(cursor)
		setState(cursor)
		view.moveToCursor(cursor)
		cb := clipDrawBuffer(db, view)
		draw(cb)
		term.Flush()

	}
}
