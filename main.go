package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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

// parseFileArg parses farg. Which looks like "filepath:linenum(1):offset(1)"
// The correspond return values are (filepath, linenum(0), offset(0)).
//
// Offset from the arg treated as 1 based offsets.
// They will changed to 0 based offsets, means they are subtracted by 1.
//
// If linenum or offset is given but invalid, it's return value will be 0.
// When both of them are ungiven, their return value will be -1, -1.
func parseFileArg(farg string) (string, int, int) {
	// final ":" is invalid but ignore it is sufficient.
	if strings.HasSuffix(farg, ":") {
		farg = farg[:len(farg)-1]
	}

	finfo := strings.Split(farg, ":")
	f := finfo[0]
	if len(finfo) == 1 {
		return f, -1, -1
	}

	l, err := strconv.Atoi(finfo[1])
	if err != nil || l < 1 {
		l = 1
	}
	l -= 1 // to base 0.
	if len(finfo) == 2 {
		return f, l, 0
	}

	o, err := strconv.Atoi(finfo[2])
	if err != nil || o < 1 {
		o = 1
	}
	o -= 1 // to base 0.
	return f, l, o
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

	editFile, initL, initB := parseFileArg(fileArgs[0])
	if initL == -1 {
		initL, initB = loadLastPosition(editFile)
	}

	// get text from file or make new.
	text, err := openOrCreate(editFile, newFlag)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = term.Init()
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
	ext := filepath.Ext(editFile)
	if ext != "" {
		ext = ext[1:]
	}
	parser := syntax.NewParser(text, ext)

	// create modes for handling events.
	mode := &ModeSelector{}
	mode.normal = &NormalMode{
		text:      text,
		cursor:    cursor,
		selection: selection,
		history:   history,
		f:         editFile,
		parser:    parser,
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
		f:      editFile,
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

	// main loop
	for {
		win.Follow(cursor, 3)

		if mode.normal.dirty {
			mode.normal.parser.ClearFrom(cell.Pt{L: win.min.L, O: 0})
			mode.normal.dirty = false
		}
		mode.normal.parser.ParseTo(cell.Pt{L: win.Max().L + 1, O: 0})

		mu.Lock()
		term.Clear(term.ColorDefault, term.ColorDefault)
		drawScreen(win, mode.normal.text, selection, parser)
		drawStatus(mode.current)
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
