package main

import (
	"fmt"
	"os"
	"time"
	term "github.com/nsf/termbox-go"
)

// we use line, offset style. termbox use o, l style.
func SetCursor(l, o int) {
	term.SetCursor(o, l)
}

func SetCell(l, o int, ch rune, fg, bg term.Attribute) {
	term.SetCell(o, l, ch, fg, bg)
}

func newTermCursor(c *cursor, l *layout) {
	viewbound := l.mainViewerBound()
	viewl, viewo := viewbound.min.l, viewbound.min.o
	SetCursor(viewl, viewo)
}


func clearScreen(l *layout) {
	viewer := l.mainViewerBound()
	for l := viewer.min.l ; l < viewer.max.l ; l++ {
		for o := viewer.min.o ; o < viewer.max.o ; o++ {
			SetCell(l, o, ' ', term.ColorDefault, term.ColorDefault)
		}
	}
}

// draw text inside of window at mainviewer
func drawScreen(l *layout, w *viewer, t text, sel *selection) {
	viewer := l.mainViewerBound()
	for l , lbyte := range t {
		if l < w.min.l || l >= w.max.l {
			continue
		}
		o := 0 // we cannot use index of line([]rune) because some rune have multiple-visible length. ex) tab, korean
		for _, ch := range lbyte {
			if o >= w.max.o {
				break
			}
			bgColor := term.ColorDefault
			if sel.on && sel.Contains(Point{l,o}) {
				bgColor = term.ColorGreen
			}
			// append cell to buffer
			if ch == '\t' {
				for i:=0 ; i<taboffset ; i++ {
					if o >= w.min.o {
						SetCell(l-w.min.l+viewer.min.l, o-w.min.o+viewer.min.o, rune(' '), term.ColorWhite, bgColor)
					}
					o += 1
				}
			} else {
				if o >= w.min.o {
					SetCell(l-w.min.l+viewer.min.l, o-w.min.o+viewer.min.o, rune(ch), term.ColorWhite, bgColor)
				}
				o += 1
			}
		}
	}
}

func printStatus(status string) {
	termw, termh := term.Size()
	statusLine := termh - 1
	for off:=0 ; off<termw ; off++ {
		SetCell(statusLine, off, ' ', term.ColorBlack, term.ColorWhite)
	}
	for off, ch := range status {
		SetCell(statusLine, off, rune(ch), term.ColorBlack, term.ColorWhite)
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
		view.moveToCursor(cursor)
		clearScreen(layout)
		drawScreen(layout, view, text, selection)
		// cl, co := cursor.positionInViewer(view)
		// status := fmt.Sprintf("linenum:%v, byteoff:%v, visoff:%v, cursoroff:%v, cpos:(%v,%v), vpos:(%v,%v, %v,%v)", cursor.line, cursor.boff, cursor.voff, cursor.offset(), cl, co, view.min.l, view.min.o, view.max.l, view.max.o)
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
					printStatus("selection on - " + fmt.Sprintf("(%v, %v) - (%v, %v)", selection.start.o, selection.start.l, selection.end.o, selection.end.l))
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
