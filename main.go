package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/kybin/tor/cell"
	"github.com/kybin/tor/syntax"
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

type Tor struct {
	screen     tcell.Screen
	mainArea   *Area
	statusArea *Area

	// current is a mode that will handle terminal events.
	current Mode

	// All modes that could be current mode.
	normal   *NormalMode
	find     *FindMode
	replace  *ReplaceMode
	gotoline *GotoLineMode
	exit     *ExitMode
}

// tor will be initialized in main
var tor *Tor = nil

func (t *Tor) InitAreas() {
	w, h := t.screen.Size()
	left := w/2 - 80/2
	if left < 0 {
		left = 0
	}
	t.mainArea = NewArea(cell.Pt{0, left}, cell.Pt{h - 1, w - left})
	t.statusArea = NewArea(cell.Pt{h - 1, 0}, cell.Pt{1, w})
}

// Refit refits it's areas.
func (t *Tor) RefitAreas() {
	w, h := t.screen.Size()
	left := w/2 - 80/2
	if left < 0 {
		left = 0
	}
	t.mainArea.Set(cell.Pt{0, left}, cell.Pt{h - 1, w - left})
	t.statusArea.Set(cell.Pt{h - 1, 0}, cell.Pt{1, w})
}

// ChangeMode changes current mode.
// It also calls old current's End() and new current's Start().
func (t *Tor) ChangeMode(m Mode) {
	t.current.End()
	t.current = m
	t.current.Start()
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
	text, err := readOrCreate(editFile, newFlag)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	screen, err := tcell.NewScreen()
	if err != nil {
		panic(err)
	}
	if err := screen.Init(); err != nil {
		panic(err)
	}
	defer screen.Fini()
	screen.SetStyle(tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset))
	screen.EnablePaste()
	screen.Clear()

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
	tor = &Tor{}
	tor.screen = screen
	tor.InitAreas()
	tor.normal = &NormalMode{
		text:      text,
		cursor:    cursor,
		selection: selection,
		history:   history,
		f:         editFile,
		parser:    parser,
		copied:    loadConfig("copy"),
		area:      tor.mainArea,
	}
	tor.find = &FindMode{
		str: loadConfig("find"),
	}
	tor.replace = &ReplaceMode{
		str: loadConfig("replace"),
	}
	tor.gotoline = &GotoLineMode{
		cursor: cursor,
	}
	tor.exit = &ExitMode{
		f:      editFile,
		cursor: cursor,
	}
	tor.current = tor.normal // start as normal mode.

	tor.exit.exit = func() {
		saveLastPosition(editFile, cursor.l, cursor.b)
		screen.Fini()
		os.Exit(0)
	}

	// main loop
	for {
		tor.normal.area.Win.Follow(tor.normal.cursor, 3)

		screen.Clear()
		drawScreen(screen, tor.normal)
		drawStatus(screen, tor.current)
		if tor.current == tor.normal {
			winP := cursor.Position().Sub(tor.normal.area.Win.Min())
			screen.ShowCursor(winP.O+tor.normal.area.min.O, winP.L)
		} else {
			_, h := screen.Size()
			screen.ShowCursor(vlen(tor.current.Status(), tor.normal.text.tabWidth), h)
		}
		screen.Show()

		// wait for keyboard input
		ev := screen.PollEvent()

		switch ev := ev.(type) {
		case *tcell.EventKey:
			tor.current.Handle(ev)
		case *tcell.EventResize:
			tor.RefitAreas()
			screen.Sync()
		}
	}
}
