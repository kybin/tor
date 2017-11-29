package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/kybin/tor/syntax"
	term "github.com/nsf/termbox-go"
)

var usage = `

  tor [flag...] file

file
  filename[:line[:offset]]

flag
`

func printUsage(f *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, "Usage of %s:", os.Args[0])
	fmt.Fprintf(os.Stderr, usage)
	f.PrintDefaults()
}

func main() {
	flagset := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	var newFlag bool
	flagset.BoolVar(&newFlag, "new", false, "let tor to edit a new file.")

	// sort args, so let flags always placed ahead of file arg.
	args := os.Args[1:]
	sort.Strings(args)
	flagset.Parse(args)

	fileArgs := flagset.Args()
	if len(fileArgs) != 1 {
		printUsage(flagset)
		os.Exit(1)
	}
	farg := fileArgs[0]

	f, initL, initO, err := parseFileArg(farg)
	if err != nil {
		printUsage(flagset)
		os.Exit(1)
	}

	exist := true
	if _, err := os.Stat(f); os.IsNotExist(err) {
		exist = false
	}
	if !exist && !newFlag {
		fmt.Println("file not exist. please retry with -new flag.")
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

	ext := filepath.Ext(farg)
	var lang syntax.Language
	if ext != "" {
		lang = syntax.Languages[ext[1:]]
	}

	err = term.Init()
	if err != nil {
		panic(err)
	}
	defer term.Close()
	term.SetInputMode(term.InputAlt)

	termw, termh := term.Size()
	mainarea := NewArea(Point{0, 0}, Point{termh - 1, termw})
	win := NewWindow(mainarea.Size())
	cursor := NewCursor(text)
	selection := NewSelection(text)
	history := NewHistory()

	mode := &ModeSelector{}
	mode.normal = &NormalMode{
		text:      text,
		cursor:    cursor,
		selection: selection,
		history:   history,
		f:         f,
		mode:      mode,
		copied:    loadConfig("copy"),
	}
	mode.find = &FindMode{
		mode: mode,
		str:  loadConfig("find"),
	}
	mode.replace = &ReplaceMode{
		mode: mode,
		str:  loadConfig("replace"),
	}
	mode.gotoline = &GotoLineMode{
		cursor: cursor,
		mode:   mode,
	}
	mode.exit = &ExitMode{
		f:      f,
		cursor: cursor,
		mode:   mode,
	}
	mode.current = mode.normal // start tor as normal mode.

	// initial cursor position
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

	events := make(chan term.Event, 20)
	go func() {
		for {
			events <- term.PollEvent()
		}
	}()

	// main loop
	matches := lang.Parse(mode.normal.text.Bytes())
	for {
		moved := win.Follow(cursor, 3)
		if lang != nil {
			if moved || (mode.current == mode.normal && mode.normal.dirty) {
				// recalculate syntax matches from window's top.
				// it will better to recalculate from edited position,
				// but seems little harder to implement.
				matches = lang.ParseRange(matches, mode.normal.text.Bytes(), syntax.Pos{win.min.l, 0}, syntax.Pos{win.Max().l + 1, 0})
				mode.normal.dirty = false
			}
		}

		term.Clear(term.ColorDefault, term.ColorDefault)
		drawScreen(mainarea, win, mode.normal.text, selection, matches)
		if mode.current.Error() != "" {
			printErrorStatus(mode.current.Error())
		} else {
			printStatus(mode.current.Status())
		}
		if mode.current == mode.normal {
			winP := cursor.Position().Sub(win.min)
			term.SetCursor(mainarea.min.o+winP.o, mainarea.min.l+winP.l)
		} else {
			term.SetCursor(vlen(mode.current.Status(), mode.normal.text.tabWidth), termh)
		}
		term.Flush()
		term.Sync()

		// wait for keyboard input
		select {
		case ev := <-events:
			switch ev.Type {
			case term.EventKey:
				mode.current.Handle(ev)
			case term.EventResize:
				term.Clear(term.ColorDefault, term.ColorDefault)
				termw, termh = term.Size()
				resizeScreen(mainarea, win, termw, termh)
			}
		}
	}
}
