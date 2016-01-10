package main

import (
	"flag"
	"fmt"
	term "github.com/nsf/termbox-go"
	"os"
)

func main() {
	var newFlag bool
	flag.BoolVar(&newFlag, "new", false, "let tor to edit a new file.")
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
		text = &Text{lines: lines, tabToSpace: false, tabWidth: 4, edited: false}
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
	selection := NewSelection()
	history := newHistory()

	mode := &ModeSelector{}
	mode.normal = &NormalMode{
		text:      text,
		cursor:    cursor,
		selection: selection,
		history:   history,
		f:         f,
		mode:      mode,
	}
	mode.find = &FindMode{
		text:      text,
		selection: selection,
		mode:      mode,
	}
	mode.replace = &ReplaceMode{
		text:      text,
		selection: selection,
		mode:      mode,
	}
	mode.gotoline = &GotoLineMode{
		cursor: cursor,
		mode: mode,
	}
	mode.exit = &ExitMode{
		f:      f,
		cursor: cursor,
		mode:   mode,
	}
	mode.current = mode.normal // will start tor as normal mode.

	events := make(chan term.Event, 20)
	go func() {
		for {
			events <- term.PollEvent()
		}
	}()
	for {
		win.Follow(cursor, 3)
		clearScreen(mainarea)
		drawScreen(mainarea, win, text, selection, cursor)
		printStatus(mode.current.Status())
		winP := cursor.Position().Sub(win.min)
		SetCursor(mainarea.min.l+winP.l, mainarea.min.o+winP.o)

		term.Flush()

		// wait for keyboard input
		select {
		case ev := <-events:
			switch ev.Type {
			case term.EventKey:
				mode.current.Handle(ev)
			case term.EventResize:
				term.Clear(term.ColorDefault, term.ColorDefault)
				resizeScreen(mainarea, win)
			}
		}
	}
}
