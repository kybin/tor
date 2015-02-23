package main

import (
	"fmt"
	"os"
	"time"
	term "github.com/nsf/termbox-go"
	"io/ioutil"
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

// Move keys are...
// KeyArrowLeft, KeyArrowRight, KeyArrowUp, KeyArrowDown, alt+(any character)
func isMoveEvent(ev term.Event) bool {
	switch ev.Type {
	case term.EventKey:
		switch ev.Key {
		case term.KeyArrowLeft, term.KeyArrowRight, term.KeyArrowUp, term.KeyArrowDown:
			return true
		default:
			if ev.Mod & term.ModAlt != 0 {
				if ev.Ch == 0 {
					return false
				}
				return true
			}
			return false
		}
	default:
		return false
	}
}

// Add keys are...
// KeyEnter, KeySpace, KeyTab, (any character)
func isAddEvent(ev term.Event) bool {
	switch ev.Type {
	case term.EventKey:
		switch ev.Key {
		case term.KeyEnter, term.KeySpace, term.KeyTab:
			return true
		default:
			if ev.Mod & term.ModAlt != 0 {
				return false
			}
			if ev.Ch == 0 {
				return false
			}
			return true
		}
	default:
		return false
	}
}

// Delete keys are...
// KeyDelete, KeyBackspace, KeyBackspace2
func isDeleteEvent(ev term.Event) bool {
	switch ev.Type {
	case term.EventKey:
		switch ev.Key {
		case term.KeyDelete, term.KeyBackspace, term.KeyBackspace2:
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func handleMoveEvent(ev term.Event, cursor *Cursor) {
	switch ev.Type {
	case term.EventKey:
		switch ev.Key {
		case term.KeyArrowLeft:
			cursor.MoveLeft()
		case term.KeyArrowRight:
			cursor.MoveRight()
		case term.KeyArrowUp:
			cursor.MoveUp()
		case term.KeyArrowDown:
			cursor.MoveDown()
		default:
			if ev.Mod & term.ModAlt != 0 {
				switch ev.Ch {
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
			}
		}
	}
}

func handleAddEvent(ev term.Event, cursor *Cursor) string {
	switch ev.Type {
	case term.EventKey:
		switch ev.Key {
		case term.KeyEnter:
			cursor.SplitLine()
			return "\n"
		case term.KeySpace:
			cursor.Insert(' ')
			return " "
		case term.KeyTab:
			cursor.Insert('\t')
			return "\t"
		default:
			if ev.Mod & term.ModAlt != 0 {
				return ""
			}
			if ev.Ch == 0 {
				return ""
			}
			cursor.Insert(ev.Ch)
			return string(ev.Ch)
		}
	}
	return ""
}

func handleDeleteEvent(ev term.Event, cursor *Cursor, selection *Selection) (string, string) {
	switch ev.Type {
	case term.EventKey:
		switch ev.Key {
		case term.KeyDelete:
			if selection.on {
				return cursor.DeleteSelection(selection), "deleteSelection"
			} else {
				return cursor.Delete(), "delete"
			}
		case term.KeyBackspace, term.KeyBackspace2:
			if selection.on {
				return cursor.DeleteSelection(selection), "deleteSelection"
			} else {
				return cursor.Backspace(), "backspace"
			}
		}
	}
	return "", "deleteSelection"
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
	history := newHistory()
	SetCursor(mainview.min.l, mainview.min.o)

	status := ""
	holdStatus := false
	lastActStr := ""
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

		if !holdStatus {
			if selection.on {
				status = "selection on - " + fmt.Sprintf("(%v, %v) - (%v, %v)", selection.start.l, selection.start.o, selection.end.l, selection.end.o)
			} else {
				status = fmt.Sprintf("linenum:%v, byteoff:%v, visoff:%v, cursoroff:%v, vpos:(%v,%v, %v,%v)", cursor.l, cursor.b, cursor.v, cursor.o, win.min.l, win.min.o, win.max.l, win.max.o)
			}
		}
		printStatus(status)
		holdStatus = false

		SetTermboxCursor(cursor, win, layout)
		term.Flush()

		// wait for keyboard input
		select {
		case ev := <-events:
			switch ev.Type {
			case term.EventKey:
				// on every key input, we should determine we need to keep selection.
				// all key with SHIFT will keep selection.
				keepSelection := false

				if isMoveEvent(ev) {
					if withShift(ev.Ch) {
						if !selection.on {
							selection.on = true
							selection.SetStart(cursor)
						}
						keepSelection = true
					}
					handleMoveEvent(ev, cursor)
					lastActStr = "move"
				} else if isAddEvent(ev) {
					beforeCursor := *cursor
					addedBefore := ""
					addedNow := handleAddEvent(ev, cursor)
					actContinue := false
					history.Cut(history.head)
					if lastActStr == "add" {
						actContinue = true
						lastAct := history.Last()
						addedBefore = lastAct.value
						beforeCursor = lastAct.beforeCursor
						history.RemoveLast()
					}
					history.Add(&Action{kind:"add", value:addedBefore+addedNow, beforeCursor:beforeCursor, afterCursor:*cursor})
					if !actContinue {
						history.head++
					}
					lastActStr = "add"
				} else if isDeleteEvent(ev) {
					deletedBefore := ""
					beforeCursor := *cursor
					// afterCursor := *cursor
					actContinue := false
					history.Cut(history.head)

					deletedNow, kind := handleDeleteEvent(ev, cursor, selection)

					var deleted string
					if kind == "backspace" && lastActStr == "backspace"  {
						actContinue = true
						deletedBefore = history.Last().value
						beforeCursor = history.Last().beforeCursor
						history.RemoveLast()
						deleted = deletedNow + deletedBefore
						lastActStr = "backspace"
					} else if kind == "delete" && lastActStr == "delete" {
						actContinue = true
						deletedBefore = history.Last().value
						beforeCursor = history.Last().beforeCursor
						history.RemoveLast()
						deleted = deletedBefore + deletedNow
						lastActStr = "delete"
					} else if kind == "deleteSelection" {
						// "deleteSelection" will always treat as discontinued action.
						kind = "delete"
						deleted = deletedNow
						min, _ := selection.MinMax()
						beforeCursor = min
						lastActStr = "deleteSelection"
					} else {
						// all other situation will fall off here.
						deleted = deletedNow
						lastActStr = kind
					}
					history.Add(&Action{kind:kind, value:deleted, beforeCursor:beforeCursor, afterCursor:*cursor})
					if !actContinue {
						history.head++
					}
				} else {
					switch ev.Key {
					case term.KeyCtrlW:
						return
					// case term.KeyCtrlC:
						// copySelection()
					case term.KeyCtrlS:
						err := save(f, text)
						if err != nil {
							status = fmt.Sprintf("%v", err)
							holdStatus = true
							continue
						}
						status = fmt.Sprintf("successfully saved : %v", f)
						holdStatus = true
					case term.KeyCtrlZ:
						// undo
						if history.head == 0 {
							continue
						}
						history.head--
						action := history.At(history.head)
						status = fmt.Sprintf("undo : %v", action)
						holdStatus = true
						switch action.kind {
						case "add":
							cursor.Copy(action.afterCursor)
							for range action.value {
								cursor.Backspace()
							}
						case "backspace":
							cursor.Copy(action.afterCursor)
							for _, r := range action.value {
								cursor.Insert(r)
							}
						case "delete":
							cursor.Copy(action.afterCursor)
							for _, r := range action.value {
								cursor.Insert(r)
							}
						}
						lastActStr = "undo"
					case term.KeyCtrlY:
						// redo
						if history.head == history.Len() {
							continue
						}
						action := history.At(history.head)
						status = fmt.Sprintf("redo : %v", action)
						holdStatus = true
						history.head++
						switch action.kind {
						case "add":
							cursor.Copy(action.beforeCursor)
							for _, r := range action.value {
								cursor.Insert(r)
							}
						case "backspace":
							cursor.Copy(action.beforeCursor)
							for range action.value {
								cursor.Backspace()
							}
						case "delete":
							cursor.Copy(action.beforeCursor)
							for range action.value {
								cursor.Delete()
							}
						}
						lastActStr = "redo"
					default:
						panic("what the")
					}
				}
				if !keepSelection {
					selection.on = false
				}
				if selection.on {
					selection.SetEnd(cursor)
				}
				lastAct := history.Last()
				if lastAct != nil {
					historyFileString := ""
					for i, a := range history.actions {
						if i != 0 {
							historyFileString += "\n"
						}
						historyFileString += fmt.Sprintf("%v", a)
					}
					ioutil.WriteFile(extendFileName(f, ".history"), []byte(historyFileString), 0755)
				}
			}
		case <-time.After(time.Second):
			holdStatus = true
		// case term.EventResize:
		//	win.resize()
		//	win.clear()
		//	win.draw()
		}
	}
}
