package main

import (
	"fmt"
	"os"
	term "github.com/nsf/termbox-go"
)

func newTermCursor(c *cursor, l *layout) {
	viewbound := l.mainViewerBound()
	viewx, viewy := viewbound.Min.X, viewbound.Min.Y
	term.SetCursor(viewx, viewy)
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

func clipDrawBuffer(drawbuf [][]rune, v *viewer) [][]rune {
	clipbuf := make([][]rune, 0)
	xstart, ystart := v.min.X, v.min.Y
	xend, yend := v.max.X, v.max.Y
	//xstart, ystart := 0,0
	//xend, yend := 20, 10
	yend = min(yend, len(drawbuf))
	if yend < ystart {
		// if then, we don't have a place for draw
		return clipbuf
	}
	for _, origbuf := range drawbuf[ystart:yend] {
		minoff := xstart
		maxoff := xend
		maxoff = min(maxoff, len(origbuf))
		if maxoff > minoff {
			clipbuf = append(clipbuf, origbuf[minoff:maxoff])
		} else {
			clipbuf = append(clipbuf, make([]rune, 0))
		}
	}
	return clipbuf
}

func draw(clipbuf [][]rune, l *layout) {
	drawrect := l.mainViewerBound()
	minx, maxx := drawrect.Min.X, drawrect.Max.X
	miny, maxy := drawrect.Min.Y, drawrect.Max.Y
	for x := minx ; x < maxx ; x++ {
		for y := miny ; y < maxy ; y++ {
			term.SetCell(x, y, ' ', term.ColorDefault, term.ColorDefault)
		}
	}
	for linenum, line := range clipbuf {
		for off, r := range line {
			term.SetCell(minx+off, miny+linenum, r, term.ColorWhite, term.ColorDefault)
		}
	}
	term.Flush()
}

func setState(c *cursor, v *viewer) {
	termw, termh := term.Size()
	stateline := termh - 1
	linenum := c.line
	byteoff := c.boff
	visoff := c.voff
	cursoroff := c.offset()
	cy, cx := c.positionInViewer(v)
	vminy, vminx, vmaxy, vmaxx := v.min.Y, v.min.X, v.max.Y, v.max.X
	state := fmt.Sprintf("linenum:%v, byteoff:%v, visoff:%v, cursoroff:%v, cpos:(%v,%v), vpos:(%v,%v, %v,%v)", linenum, byteoff, visoff, cursoroff, cy, cx, vminy, vminx, vmaxy, vmaxx)
	for off:=0 ; off<termw ; off++ {
		term.SetCell(off, stateline, ' ', term.ColorBlack, term.ColorWhite)
	}
	for off, ch := range state {
		term.SetCell(off, stateline, rune(ch), term.ColorBlack, term.ColorWhite)
	}
}

func main() {
	// check there is an destination file. ex)tor some.file
	args := os.Args[1:]
	if len(args)==0 {
		fmt.Println("please, set text file")
		return
	}
	f := args[0]

	err := term.Init()
	if err != nil {
		panic(err)
	}
	defer term.Close()
	term.SetInputMode(term.InputAlt)
	term.Clear(term.ColorDefault, term.ColorDefault)
	term.Flush()

	text := open(f)

	layout := newLayout()
	view := newViewer(layout)
	cursor := newCursor(text)
	newTermCursor(cursor, layout)

	db := textToDrawBuffer(text)
	draw(db, layout)

	setState(cursor, view)
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
				if (ev.Mod&term.ModAlt) != 0 {
					switch ev.Ch {
					case 'j': cursor.moveLeft()
					case 'l': cursor.moveRight()
					case 'i': cursor.moveUp()
					case 'k': cursor.moveDown()
					case 'm': cursor.moveBow()
					case '.': cursor.moveEow()
					case 'u': cursor.pageUp()
					case 'o': cursor.pageDown()
					}
				}
			}
		// case term.EventResize:
		//	view.resize()
		//	view.clear()
		//	view.draw()
		}
		view.moveToCursor(cursor)
		cb := clipDrawBuffer(db, view)
		draw(cb, layout)
		setState(cursor, view)
		setTermboxCursor(cursor, view, layout)
		term.Flush()

	}
}
