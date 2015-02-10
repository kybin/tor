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

func SetTermboxCursor(c *Cursor, w *Window, l *Layout) {
	view := l.MainViewerBound()
	p := c.PositionInWindow(w)
	SetCursor(view.min.l+p.l, view.min.o+p.o)
}

func clearScreen(l *Layout) {
	viewer := l.MainViewerBound()
	for l := viewer.min.l ; l < viewer.max.l ; l++ {
		for o := viewer.min.o ; o < viewer.max.o ; o++ {
			SetCell(l, o, ' ', term.ColorDefault, term.ColorDefault)
		}
	}
}

// draw text inside of window at mainviewer
func drawScreen(l *Layout, w *Window, t *Text, sel *Selection) {
	viewer := l.MainViewerBound()
	for l , ln := range t.lines {
		if l < w.min.l || l >= w.max.l {
			continue
		}
		o := 0 // we cannot use index of line([]rune) because some rune have multiple-visible length. ex) tab, korean
		for _, ch := range ln.data {
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

	text, err := open(f)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = term.Init()
	if err != nil {
		panic(err)
	}
	defer term.Close()
	term.SetInputMode(term.InputAlt)
	term.Clear(term.ColorDefault, term.ColorDefault)
	term.Flush()


	layout := NewLayout()
	mainview := layout.MainViewerBound()
	win := NewWindow(layout)
	// drawbuf := textToDrawBuffer(text, selection)
	cursor := NewCursor(text)
	selection := NewSelection()
	SetCursor(mainview.min.l, mainview.min.o)

	edit := false
	events := make(chan term.Event, 20)
	go func() {
		for {
			events <- term.PollEvent()
		}
	}()
	for {
		win.Follow(cursor)
		clearScreen(layout)
		drawScreen(layout, win, text, selection)
		// c := cursor.PositionInWindow(win)
		// status := fmt.Sprintf("linenum:%v, byteoff:%v, visoff:%v, cursoroff:%v, cpos:(%v,%v), vpos:(%v,%v, %v,%v)", cursor.l, cursor.b, cursor.v, cursor.o, c.l, c.o, win.min.l, win.min.o, win.max.l, win.max.o)
		// if edit == true {
		// 	status += " editing..."
		// } else {
		// 	status += " idle"
		// }
		// printStatus(status)
		SetTermboxCursor(cursor, win, layout)
		term.Flush()

		// wait for keyboard input
		select {
		case ev := <-events:
			edit = true
			switch ev.Type {
			case term.EventKey:
				// on every key input, we should determine we need to keep selection.
				// all key with SHIFT will keep selection.
				keepSelection := false

				switch ev.Key {
				case term.KeyCtrlW:
					return
				// case term.KeyCtrlC:
					// copySelection()
				case term.KeyCtrlS:
					err := save(f, text)
					if err != nil {
						printStatus(fmt.Sprintf("%v", err))
						continue
					}
					printStatus(fmt.Sprintf("successfully saved : %v", f))
				case term.KeyArrowLeft:
					cursor.MoveLeft()
				case term.KeyArrowRight:
					cursor.MoveRight()
				case term.KeyArrowUp:
					cursor.MoveUp()
				case term.KeyArrowDown:
					cursor.MoveDown()
				case term.KeyEnter:
					cursor.SplitLine()
				case term.KeyTab:
					cursor.Insert('\t')
				case term.KeyDelete:
					if selection.on {
						cursor.DeleteSelection(selection)
					} else {
						cursor.Delete()
					}
				case term.KeyBackspace2:
					if selection.on {
						cursor.DeleteSelection(selection)
					} else {
						cursor.Backspace()
					}
				default:
					if (ev.Mod&term.ModAlt) != 0 {
						if withShift(ev.Ch) {
							if !selection.on {
								selection.on = true
								selection.SetStart(cursor)
							}
							keepSelection = true
						}
						switch ev.Ch {
						// if character pressed with shift
						// we will enable cursor selection.
						case 'j', 'J':
							cursor.MoveLeft()
						case 'l', 'L':
							cursor.MoveRight()
						case 'i', 'I':
							cursor.MoveUp()
						case 'k', 'K':
							cursor.MoveDown()
						case 'm', 'M':
							cursor.MoveBow()
						case '.', '>':
							cursor.MoveEow()
						case 'u', 'U':
							cursor.MoveBol()
						case 'o', 'O':
							cursor.MoveEol()
						case 'h', 'H':
							cursor.PageUp()
						case 'n', 'N':
							cursor.PageDown()
						case 'a', 'A':
							cursor.MoveBof()
						case 'z', 'Z':
							cursor.MoveEof()
						}
					} else {
						cursor.Insert(ev.Ch)
					}
				}
				if !keepSelection {
					selection.on = false
				}
				if selection.on {
					selection.SetEnd(cursor)
					printStatus("selection on - " + fmt.Sprintf("(%v, %v) - (%v, %v)", selection.start.l, selection.start.o, selection.end.l, selection.end.o))
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
		//	win.resize()
		//	win.clear()
		//	win.draw()
		}
	}
}
