package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/kybin/tor/cell"
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
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprint(os.Stderr, usage)
	f.PrintDefaults()
}

// sortArgs sorts args to make flags always placed ahead of file arg.
func sortArgs(args []string) {
	sort.Slice(args, func(i, j int) bool {
		iIsFlag := strings.HasPrefix(args[i], "-")
		jIsFlag := strings.HasPrefix(args[j], "-")
		if iIsFlag && !jIsFlag {
			return true
		}
		return false
	})
}

func main() {
	flagset := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	var newFlag bool
	flagset.BoolVar(&newFlag, "new", false, "let tor to edit a new file.")

	args := os.Args[1:]
	sortArgs(args)
	flagset.Parse(args)

	fileArgs := flagset.Args()
	if len(fileArgs) != 1 {
		printUsage(flagset)
		os.Exit(1)
	}
	farg := fileArgs[0]

	f, initL, initB := parseFileArg(farg)
	if initL == -1 {
		initL, initB = loadLastPosition(f)
	}

	exist := true
	if _, err := os.Stat(f); os.IsNotExist(err) {
		exist = false
	}
	if !exist && !newFlag {
		fmt.Fprintln(os.Stderr, "file not exist. please retry with -new flag.")
		os.Exit(1)
	}

	var text *Text
	if exist {
		var err error
		text, err = open(f)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else {
		lines := make([]Line, 0)
		lines = append(lines, Line{""})
		text = &Text{lines: lines, tabToSpace: false, tabWidth: 4, edited: false}
	}

	err := term.Init()
	if err != nil {
		panic(err)
	}
	defer term.Close()
	term.SetInputMode(term.InputAlt)

	termw, termh := term.Size()
	win := NewWindow(cell.Pt{termh - 1, termw})
	cursor := NewCursor(text)
	cursor.GotoLine(initL)
	cursor.SetCloseToB(initB)
	selection := NewSelection(text)
	history := NewHistory()

	// create modes for handling events.
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
	mode.current = mode.normal // start as normal mode.

	// get events from termbox.
	events := make(chan term.Event, 20)
	go func() {
		for {
			events <- term.PollEvent()
		}
	}()

	// mutex for commucation with termbox.
	mu := &sync.Mutex{}

	// Sync redraws terminal from buffer.
	// sometimes Flush is insufficient,
	// it is better to Sync frequently.
	go func() {
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-ticker.C:
				mu.Lock()
				term.Sync()
				mu.Unlock()
			}
		}
	}()

	// parse syntax
	ext := filepath.Ext(f)
	var lang *syntax.Language
	if ext != "" {
		lang = syntax.Languages[ext[1:]]
	}
	var matches []syntax.Match
	if lang != nil {
		matches = lang.Parse(mode.normal.text.Bytes())
	}

	// main loop
	for {
		moved := win.Follow(cursor, 3)
		if lang != nil {
			if moved || (mode.current == mode.normal && mode.normal.dirty) {
				// recalculate syntax matches from window's top.
				// it will better to recalculate from edited position,
				// but seems little harder to implement.
				matches = lang.ParseRange(matches, mode.normal.text.Bytes(), cell.Pt{L: win.min.L, O: 0}, cell.Pt{L: win.Max().L + 1, O: 0})
				mode.normal.dirty = false
			}
		}

		mu.Lock()
		term.Clear(term.ColorDefault, term.ColorDefault)
		drawScreen(win, mode.normal.text, selection, lang, matches)
		if mode.current.Error() != "" {
			drawErrorStatus(mode.current.Error())
		} else {
			drawStatus(mode.current.Status())
		}
		if mode.current == mode.normal {
			winP := cursor.Position().Sub(win.min)
			term.SetCursor(winP.O, winP.L)
		} else {
			term.SetCursor(vlen(mode.current.Status(), mode.normal.text.tabWidth), termh)
		}
		term.Flush()
		mu.Unlock()

		// wait for keyboard input
		select {
		case ev := <-events:
			switch ev.Type {
			case term.EventKey:
				mode.current.Handle(ev)
			case term.EventResize:
				mu.Lock()
				term.Clear(term.ColorDefault, term.ColorDefault)
				termw, termh = term.Size()
				win.size = cell.Pt{termh - 1, termw}
				mu.Unlock()
			}
		}
	}
}
