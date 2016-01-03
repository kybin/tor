package main

import (
	"errors"
	"flag"
	"fmt"
	term "github.com/nsf/termbox-go"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

// parseFileArg returns (filepath, linenum, offset, error).
// if the linenum is given, but 0 or negative, it will be 1.
// if the offset is given, but negative, it will be 0.
// when only filepath is given, linenum and offset will be set to -1.
func parseFileArg(farg string) (string, int, int, error) {
	finfo := strings.Split(farg, ":")
	f := finfo[0]
	l, o := -1, -1
	err := error(nil)

	if len(finfo) >= 4 {
		return "", -1, -1, errors.New("too many colons in file argument")
	}

	if len(finfo) == 1 {
		return f, -1, -1, nil
	}
	if len(finfo) >= 2 {
		l, err = strconv.Atoi(finfo[1])
		if err != nil {
			return "", -1, -1, err
		}
		if len(finfo) == 3 {
			o, err = strconv.Atoi(finfo[2])
			if err != nil {
				return "", -1, -1, err
			}
		}
	}

	if l < 0 {
		l = 0
	}
	if o < 0 {
		o = 0
	}
	return f, l, o, nil
}

func main() {
	var newFlag, debugFlag bool
	flag.BoolVar(&newFlag, "new", false, "let tor to edit a new file.")
	flag.BoolVar(&debugFlag, "debug", false, "tor will create .history file for debugging.")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("please, set text file")
		os.Exit(1)
	}
	farg := args[len(args)-1]

	f, initL, initO, err := parseFileArg(farg)
	if err != nil {
		fmt.Println("file arg is invalid: ", err)
		os.Exit(1)
	}

	exist := true
	if _, err := os.Stat(f); os.IsNotExist(err) {
		exist = false
	}
	if !exist && !newFlag {
		fmt.Println("file not exist. please retry with -new flag.")
		os.Exit(1)
	} else if exist && newFlag {
		fmt.Println("file already exist.")
		os.Exit(1)
	}

	var text *Text
	if exist {
		text, err = open(f)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else {
		lines := make([]Line, 0)
		lines = append(lines, Line{""})
		text = &Text{lines:lines, tabToSpace:false, tabWidth:4, edited:false}
	}

	err = term.Init()
	if err != nil {
		panic(err)
	}
	defer term.Close()
	term.SetInputMode(term.InputAlt)
	// term.SetOutputMode(term.Output256)
	term.Clear(term.ColorDefault, term.ColorDefault)
	term.Flush()

	termw, termh := term.Size()
	mainarea := NewArea(Point{0, 0}, Point{termh - 1, termw})
	win := NewWindow(mainarea.Size())

	cursor := NewCursor(text)
	if initL != -1 {
		l := initL
		// to internal line number
		if l != 0 {
			l--
		}
		cursor.GotoLine(l)
		if initO != -1 {
			cursor.SetO(initO)
		}
	} else {
		l, b := loadLastPosition(f)
		cursor.GotoLine(l)
		cursor.SetCloseToB(b)
	}

	findmode := &LineInputMode{}
	replacemode := &LineInputMode{}
	gotolinemode := &GotoLineMode{}
	selection := NewSelection()
	history := newHistory()

	mode := "normal"

	status := ""
	holdStatus := false
	lastActStr := ""
	copied := ""

	events := make(chan term.Event, 20)
	go func() {
		for {
			events <- term.PollEvent()
		}
	}()
	for {
		win.Follow(cursor, 3)
		clearScreen(mainarea)
		drawScreen(mainarea, win, text, selection, cursor, mode)

		if mode == "exit" {
			status = fmt.Sprintf("Buffer modified. Do you really want to quit? (y/n)")
		} else if mode == "gotoline" {
			status = fmt.Sprintf("goto : %v", gotolinemode.linestr)
		} else if mode == "find" {
			status = fmt.Sprintf("find : %v", findmode.str)
		} else if mode == "replace" {
			status = fmt.Sprintf("replace : %v", replacemode.str)
		} else if !holdStatus {
			status = fmt.Sprintf("%v:%v:%v", f, cursor.l+1, cursor.O())
			if findmode.set {
				status = fmt.Sprintf("find \"%v\": ctrl+d, ctrl+b", findmode.str)
			} else if replacemode.set {
				status = fmt.Sprintf("replace \"%v\": ctrl+j", replacemode.str)
			}
		}
		printStatus(status)
		holdStatus = false

		// reset variables
		findmode.set = false
		replacemode.set = false

		winP := cursor.Position().Sub(win.min)
		SetCursor(mainarea.min.l+winP.l, mainarea.min.o+winP.o)

		term.Flush()

		// wait for keyboard input
		select {
		case ev := <-events:
			switch ev.Type {
			case term.EventKey:
				if mode == "exit" {
					if ev.Ch == 'y' {
						saveLastPosition(f, cursor.l, cursor.b)
						return
					} else if ev.Ch == 'n' || ev.Key == term.KeyCtrlK {
						mode = "normal"
					}
					continue
				} else if mode == "gotoline" {
					gotolinemode.Handle(ev, cursor, &mode)
					continue
				} else if mode == "find" {
					findmode.Handle(ev, cursor, &mode)
					continue
				} else if mode == "replace" {
					replacemode.Handle(ev, cursor, &mode)
					continue
				}

				actions := parseEvent(ev, text, selection, &mode)
				for _, a := range actions {
					if a.kind == "modeChange" {
						if a.value == "find" {
							if selection.on {
								min, max := selection.MinMax()
								findmode.set = true
								findmode.str = text.DataInside(min, max)
								selection.on = false
								continue
							}
							findmode.start = true
						} else if a.value == "replace" {
							if selection.on {
								min, max := selection.MinMax()
								replacemode.set = true
								replacemode.str = text.DataInside(min, max)
								selection.on = false
								continue
							}
							replacemode.start = true
						}
						mode = a.value
						continue
					}

					beforeCursor := *cursor

					if a.kind == "exit" {
						if !text.edited {
							saveLastPosition(f, cursor.l, cursor.b)
							return
						}
						mode = "exit"
						continue
					} else if a.kind == "save" {
						err := save(f, text)
						if err != nil {
							panic(err)
						}
						text.edited = false
						status = fmt.Sprintf("successfully saved : %v", f)
						holdStatus = true
					} else if a.kind == "copy" {
						if selection.on {
							minc, maxc := selection.MinMax()
							copied = text.DataInside(minc, maxc)
						} else {
							r, _ := cursor.RuneAfter()
							copied = string(r)
						}
						saveCopyString(copied)
					} else if a.kind == "paste" {
						if copied == "" {
							copied = loadCopyString()
						}
						cursor.Insert(copied)
						a.value = copied
					} else if a.kind == "replace" {
						if replacemode.str != "" {
							cursor.Insert(replacemode.str)
							a.value = replacemode.str
						}
					} else {
						do(a, text, cursor, selection, history, &status, &holdStatus, findmode.str)
					}
					switch a.kind {
					case "insert", "delete", "backspace", "deleteSelection", "paste", "replace", "insertTab", "removeTab":
						// remember the action.
						text.edited = true
						nc := history.Cut(history.head)
						if nc != 0 {
							lastActStr = ""
						}
						if a.kind == "insert" || a.kind == "delete" || a.kind == "backspace" {
							if a.kind == lastActStr {
								lastAct, err := history.Pop()
								if err != nil {
									panic(err)
								}
								history.head--
								beforeCursor = lastAct.beforeCursor
								if a.kind == "insert" || a.kind == "delete" {
									a.value = lastAct.value + a.value
								} else if a.kind == "backspace" {
									a.value = a.value + lastAct.value
								}
							}
						}
						a.beforeCursor = beforeCursor
						if a.kind == "deleteSelection" {
							a.beforeCursor, _ = selection.MinMax()
						}
						a.afterCursor = *cursor
						history.Add(a)
						history.head++
					}
					lastActStr = a.kind
					lastAct := history.Last()
					if debugFlag && lastAct != nil {
						historyFileString := ""
						for i, a := range history.actions {
							if i != 0 {
								historyFileString += "\n"
							}
							historyFileString += fmt.Sprintf("%v, %v", a, history.head)
						}
						ioutil.WriteFile(extendFileName(f, ".history"), []byte(historyFileString), 0755)
					}
				}
			case term.EventResize:
				term.Clear(term.ColorDefault, term.ColorDefault)
				resizeScreen(mainarea, win)
			}
		case <-time.After(time.Second):
			holdStatus = true
		}
	}
}
