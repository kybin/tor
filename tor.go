package main

import (
	"fmt"
	"os"
	"time"
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

func clearScreen(l *layout) {
	drawrect := l.mainViewerBound()
	minx, maxx := drawrect.Min.X, drawrect.Max.X
	miny, maxy := drawrect.Min.Y, drawrect.Max.Y
	for x := minx ; x < maxx ; x++ {
		for y := miny ; y < maxy ; y++ {
			term.SetCell(x, y, ' ', term.ColorDefault, term.ColorDefault)
		}
	}
}

func draw(clipbuf [][]rune, l *layout) {
	min := l.mainViewerBound().Min
	for linenum, line := range clipbuf {
		for off, r := range line {
			term.SetCell(min.X+off, min.Y+linenum, r, term.ColorWhite, term.ColorDefault)
		}
	}
}

func printStatus(status string) {
	termw, termh := term.Size()
	statusLine := termh - 1
	for off:=0 ; off<termw ; off++ {
		term.SetCell(off, statusLine, ' ', term.ColorBlack, term.ColorWhite)
	}
	for off, ch := range status {
		term.SetCell(off, statusLine, rune(ch), term.ColorBlack, term.ColorWhite)
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
	drawbuf := textToDrawBuffer(text)
	cursor := newCursor(text)
	newTermCursor(cursor, layout)

	edit := false
	events := make(chan term.Event, 20)
	go func() {
		for {
			events <- term.PollEvent()
		}
	}()
	for {
		// draw buffer
		view.moveToCursor(cursor)
		cb := clipDrawBuffer(drawbuf, view)
		clearScreen(layout)
		draw(cb, layout)
		cy, cx := cursor.positionInViewer(view)
		status := fmt.Sprintf("linenum:%v, byteoff:%v, visoff:%v, cursoroff:%v, cpos:(%v,%v), vpos:(%v,%v, %v,%v)", cursor.line, cursor.boff, cursor.voff, cursor.offset(), cy, cx, view.min.Y, view.min.X, view.max.Y, view.max.X)
		if edit == true {
			status += " editing..."
		} else {
			status += " idle"
		}
		printStatus(status)
		setTermboxCursor(cursor, view, layout)
		term.Flush()

		// wait for keyboard input
		select {
		case ev := <-events:
			edit = true
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
					case 'j':
						cursor.moveLeft()
					case 'l':
						cursor.moveRight()
					case 'i':
						cursor.moveUp()
					case 'k':
						cursor.moveDown()
					case 'm':
						cursor.moveBow()
					case '.':
						cursor.moveEow()
					case 'u':
						cursor.moveBol()
					case 'o':
						cursor.moveEol()
					case 'h':
						cursor.pageUp()
					case 'n':
						cursor.pageDown()
					case 'a':
						cursor.moveBof()
					case 'z':
						cursor.moveEof()
					}
				}
			}
		case <-time.After(time.Second):
			// OK. It's idle time. We should check if any edit applied on contents.
			if edit == true {
				// remember the action. (or difference)
			}
			edit = false
		// case term.EventResize:
		//	view.resize()
		//	view.clear()
		//	view.draw()
		}
	}
}
