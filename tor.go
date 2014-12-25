package main

import (
	"fmt"
	"os"
	"time"
	"image"
	term "github.com/nsf/termbox-go"
)

func newTermCursor(c *cursor, l *layout) {
	viewbound := l.mainViewerBound()
	viewx, viewy := viewbound.Min.X, viewbound.Min.Y
	term.SetCursor(viewx, viewy)
}

type cell struct {
	rune rune
	highlight bool // maybe changed to attributes?
}

func textToDrawBuffer(txt text, sel *selection) [][]cell {
	// highlight start, end should recalculate from selection, so we can ensure start < end.
	var highlight bool
	var hstart, hend image.Point
	if sel.on {
		if sel.end.Y < sel.start.Y {
			hstart, hend = sel.end, sel.start
		} else if sel.end.Y == sel.start.Y && sel.end.X < sel.start.X {
			hstart, hend = sel.end, sel.start
		} else {
			hstart, hend = sel.start, sel.end
		}
	}
	drawbuf := make([][]cell, 0)
	for y , line := range txt {
		linebuf := make([]cell, 0)
		x := 0 // we cannot use index of line([]rune) because some rune have multiple-visible length. ex) tab, korean
		for _, ch := range line {
			// if selection is on, caculate this cell is in highlight range
			if sel.on {
				if y < hstart.Y || y > hend.Y {
					highlight = false
				} else if y == hstart.Y && x < hstart.X {
					highlight = false
				} else if y == hend.Y && x >= hend.X {
					highlight = false
				} else {
					highlight = true
				}
			} else {
				highlight = false
			}
			// append cell to buffer
			if ch == '\t' {
				for i:=0 ; i<taboffset ; i++ {
					linebuf = append(linebuf, cell{rune(' '), highlight})
					x += 1
				}
			} else {
				linebuf = append(linebuf, cell{rune(ch), highlight})
				x += 1
			}
		}
		drawbuf = append(drawbuf, linebuf)
	}
	return drawbuf
}

func clipDrawBuffer(drawbuf [][]cell, v *viewer) [][]cell {
	clipbuf := make([][]cell, 0)
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
			clipbuf = append(clipbuf, make([]cell, 0))
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

func draw(clipbuf [][]cell, l *layout) {
	min := l.mainViewerBound().Min
	for linenum, line := range clipbuf {
		for off, c := range line {
			fgColor := term.ColorWhite
			bgColor := term.ColorDefault
			if c.highlight {
				bgColor = term.ColorGreen
			}
			term.SetCell(min.X+off, min.Y+linenum, c.rune, fgColor, bgColor)
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
	// drawbuf := textToDrawBuffer(text, selection)
	cursor := newCursor(text)
	selection := NewSelection()
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
		drawbuf := textToDrawBuffer(text, selection)
		cb := clipDrawBuffer(drawbuf, view)
		clearScreen(layout)
		draw(cb, layout)
		// cy, cx := cursor.positionInViewer(view)
		// status := fmt.Sprintf("linenum:%v, byteoff:%v, visoff:%v, cursoroff:%v, cpos:(%v,%v), vpos:(%v,%v, %v,%v)", cursor.line, cursor.boff, cursor.voff, cursor.offset(), cy, cx, view.min.Y, view.min.X, view.max.Y, view.max.X)
		// if edit == true {
		// 	status += " editing..."
		// } else {
		// 	status += " idle"
		// }
		// printStatus(status)
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
				// case term.KeyCtrlC:
					// copySelection()
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
					if withShift(ev.Ch) {
						if !selection.on {
							selection.on = true
							selection.SetStart(cursor)
						} else {
							// already on selection. do nothing.
						}
					} else {
						selection.on = false
					}
					switch ev.Ch {
					// if character pressed with shift
					// we will enable cursor selection.
					case 'j', 'J':
						cursor.moveLeft()
					case 'l', 'L':
						cursor.moveRight()
					case 'i', 'I':
						cursor.moveUp()
					case 'k', 'K':
						cursor.moveDown()
					case 'm', 'M':
						cursor.moveBow()
					case '.', '>':
						cursor.moveEow()
					case 'u', 'U':
						cursor.moveBol()
					case 'o', 'O':
						cursor.moveEol()
					case 'h', 'H':
						cursor.pageUp()
					case 'n', 'N':
						cursor.pageDown()
					case 'a', 'A':
						cursor.moveBof()
					case 'z', 'Z':
						cursor.moveEof()
					}
				}
				if selection.on {
					selection.SetEnd(cursor)
					printStatus("selection on - " + fmt.Sprintf("(%v, %v) - (%v, %v)", selection.start.X, selection.start.Y, selection.end.X, selection.end.Y))
				} else {
					printStatus("selection off")
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
