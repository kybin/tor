package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	term "github.com/nsf/termbox-go"
)

// NormalMode is a mode for text editing.
type NormalMode struct {
	text      *Text
	cursor    *Cursor
	selection *Selection
	history   *History
	f         string

	copied string
	status string
	err    string

	mode *ModeSelector
}

// Start prepare things to start a normal mode.
func (m *NormalMode) Start() {
	term.SetInputMode(term.InputAlt)
}

// End prepare things to end a normal mode.
func (m *NormalMode) End() {}

// Handle handles a terminal event.
// It will run appropriate actions, and save it in history.
func (m *NormalMode) Handle(ev term.Event) {
	m.status = ""
	m.err = ""

	rememberActions := make([]*Action, 0)
	cut := false
	actions := m.parseEvent(ev)
	for _, a := range actions {
		m.do(a)
		a.text = m.text
		// delete selection usally don't delete anything.
		if a.kind == "delete" && a.value == "" {
			continue
		}
		// skip action types that are not specified below.
		switch a.kind {
		case "insert", "delete", "backspace", "insertTab", "removeTab", "move":
			if a.kind != "move" {
				m.text.edited = true
			}
			nc := m.history.Cut(m.history.head)
			if nc != 0 {
				cut = true
			}
		default:
			continue
		}
		// joining repeative same kind of actions.
		if a.kind == "insert" || a.kind == "delete" || a.kind == "backspace" || a.kind == "move" {
			var last *Action
			if len(rememberActions) != 0 {
				last = rememberActions[len(rememberActions)-1]
			} else if !cut && m.history.Len() != 0 {
				lastGroup := m.history.Last()
				last = lastGroup[len(lastGroup)-1]
			}
			if last != nil && a.kind == last.kind {
				if last.kind == "insert" || a.kind == "delete" {
					last.value = last.value + a.value
				} else if a.kind == "backspace" {
					last.value = a.value + last.value
				}
				last.afterCursor = a.afterCursor
				continue
			}
		}
		rememberActions = append(rememberActions, a)
	}
	if len(rememberActions) != 0 {
		m.history.Add(rememberActions)
	}
}

// parseEvent parses a terminal event and return actions.
func (m *NormalMode) parseEvent(ev term.Event) []*Action {
	if ev.Type != term.EventKey {
		panic(fmt.Sprintln("what the..", ev.Type, "event?"))
	}

	switch ev.Key {
	case term.KeyCtrlW:
		return []*Action{{kind: "selection", value: "off"}, {kind: "exit"}}
	case term.KeyCtrlS:
		return []*Action{{kind: "selection", value: "off"}, {kind: "save"}}
	case term.KeyCtrlK:
		return []*Action{{kind: "selection", value: "off"}}
	// move
	case term.KeyArrowLeft:
		return []*Action{{kind: "selection", value: "off"}, {kind: "move", value: "left"}}
	case term.KeyArrowRight:
		return []*Action{{kind: "selection", value: "off"}, {kind: "move", value: "right"}}
	case term.KeyArrowUp:
		return []*Action{{kind: "selection", value: "off"}, {kind: "move", value: "up"}}
	case term.KeyArrowDown:
		return []*Action{{kind: "selection", value: "off"}, {kind: "move", value: "down"}}
	case term.KeyPgup:
		return []*Action{{kind: "selection", value: "off"}, {kind: "move", value: "pageup"}}
	case term.KeyPgdn:
		return []*Action{{kind: "selection", value: "off"}, {kind: "move", value: "pagedown"}}
	case term.KeyHome:
		return []*Action{{kind: "selection", value: "off"}, {kind: "move", value: "bol"}}
	case term.KeyEnd:
		return []*Action{{kind: "selection", value: "off"}, {kind: "move", value: "eol"}}
	// insert
	case term.KeyEnter:
		return []*Action{{kind: "delete", value: "selection"}, {kind: "insert", value: "\n"}}
	case term.KeyCtrlN:
		if ev.Mod&term.ModAlt != 0 {
			return []*Action{{kind: "selection", value: "off"}, {kind: "move", value: "eol"}, {kind: "insert", value: "\n"}, {kind: "insert", value: "autoIndent"}}
		}
		return []*Action{{kind: "delete", value: "selection"}, {kind: "insert", value: "\n"}, {kind: "insert", value: "autoIndent"}}
	case term.KeySpace:
		return []*Action{{kind: "delete", value: "selection"}, {kind: "insert", value: " "}}
	case term.KeyTab:
		tab := "\t"
		if m.text.tabToSpace {
			tab = strings.Repeat(" ", m.text.tabWidth)
		}
		return []*Action{{kind: "delete", value: "selection"}, {kind: "insert", value: tab}}
	case term.KeyCtrlU:
		return []*Action{{kind: "removeTab"}}
	case term.KeyCtrlO:
		return []*Action{{kind: "insertTab"}}
	// delete : value will added after actual deletion.
	case term.KeyDelete:
		if m.selection.on {
			return []*Action{{kind: "delete", value: "selection"}}
		} else {
			if ev.Mod&term.ModAlt != 0 {
				return []*Action{{kind: "selection", value: "on"}, {kind: "move", value: "nextBowEow"}, {kind: "delete", value: "selection"}}
			}
			return []*Action{{kind: "delete"}}
		}
	case term.KeyBackspace, term.KeyBackspace2:
		if m.selection.on {
			return []*Action{{kind: "delete", value: "selection"}}
		} else {
			if ev.Mod&term.ModAlt != 0 {
				return []*Action{{kind: "selection", value: "on"}, {kind: "move", value: "prevBowEow"}, {kind: "delete", value: "selection"}}
			}
			return []*Action{{kind: "backspace"}}
		}
	// undo, redo
	case term.KeyCtrlZ:
		return []*Action{{kind: "undo"}}
	case term.KeyCtrlY:
		return []*Action{{kind: "redo"}}
	// copy, paste, cut
	case term.KeyCtrlC:
		if m.selection.on {
			return []*Action{{kind: "copy"}, {kind: "selection", value: "off"}}
		} else {
			return []*Action{}
		}
	case term.KeyCtrlV:
		if m.selection.on {
			return []*Action{{kind: "delete", value: "selection"}, {kind: "insert", value: m.copied}}
		}
		return []*Action{{kind: "insert", value: m.copied}}
	case term.KeyCtrlJ:
		if m.selection.on {
			return []*Action{{kind: "delete", value: "selection"}, {kind: "insert", value: m.mode.replace.str}}
		}
		return []*Action{}
	case term.KeyCtrlX:
		if m.selection.on {
			return []*Action{{kind: "copy"}, {kind: "delete", value: "selection"}}
		} else {
			return []*Action{{kind: "copy"}, {kind: "delete"}}
		}
	// find
	case term.KeyCtrlD, term.KeyF3:
		return []*Action{{kind: "selection", value: "off"}, {kind: "move", value: "findNextSelect"}}
	case term.KeyCtrlB, term.KeyF2:
		return []*Action{{kind: "selection", value: "off"}, {kind: "move", value: "findPrevSelect"}}
	case term.KeyCtrlF:
		return []*Action{{kind: "modeChange", value: "find"}}
	case term.KeyCtrlR:
		return []*Action{{kind: "modeChange", value: "replace"}}
	case term.KeyCtrlG:
		return []*Action{{kind: "modeChange", value: "gotoline"}}
	case term.KeyCtrlA:
		return []*Action{{kind: "selectAll"}}
	case term.KeyCtrlL:
		return []*Action{{kind: "selectLine"}}
	default:
		if ev.Ch == 0 {
			return []*Action{}
		}
		if ev.Mod&term.ModAlt != 0 {
			switch ev.Ch {
			case 'j':
				return []*Action{{kind: "move", value: "selLeft"}, {kind: "selection", value: "off"}}
			case 'J':
				return []*Action{{kind: "selection", value: "on"}, {kind: "move", value: "left"}}
			case 'l':
				return []*Action{{kind: "move", value: "selRight"}, {kind: "selection", value: "off"}}
			case 'L':
				return []*Action{{kind: "selection", value: "on"}, {kind: "move", value: "right"}}
			case 'i':
				return []*Action{{kind: "selection", value: "off"}, {kind: "move", value: "up"}}
			case 'I':
				return []*Action{{kind: "selection", value: "on"}, {kind: "move", value: "up"}}
			case 'k':
				return []*Action{{kind: "selection", value: "off"}, {kind: "move", value: "down"}}
			case 'K':
				return []*Action{{kind: "selection", value: "on"}, {kind: "move", value: "down"}}
			case 'm':
				return []*Action{{kind: "selection", value: "off"}, {kind: "move", value: "prevBowEow"}}
			case 'M':
				return []*Action{{kind: "selection", value: "on"}, {kind: "move", value: "prevBowEow"}}
			case '.':
				return []*Action{{kind: "selection", value: "off"}, {kind: "move", value: "nextBowEow"}}
			case '>':
				return []*Action{{kind: "selection", value: "on"}, {kind: "move", value: "nextBowEow"}}
			case 'u':
				return []*Action{{kind: "selection", value: "off"}, {kind: "move", value: "bocBolRepeat"}}
			case 'U':
				return []*Action{{kind: "selection", value: "on"}, {kind: "move", value: "bocBolRepeat"}}
			case 'y':
				return []*Action{{kind: "selection", value: "off"}, {kind: "move", value: "bol"}}
			case 'Y':
				return []*Action{{kind: "selection", value: "on"}, {kind: "move", value: "bol"}}
			case 'o':
				return []*Action{{kind: "selection", value: "off"}, {kind: "move", value: "eol"}}
			case 'O':
				return []*Action{{kind: "selection", value: "on"}, {kind: "move", value: "eol"}}
			case 'w':
				return []*Action{{kind: "selection", value: "off"}, {kind: "move", value: "pageup"}}
			case 'W':
				return []*Action{{kind: "selection", value: "on"}, {kind: "move", value: "pageup"}}
			case 's':
				return []*Action{{kind: "selection", value: "off"}, {kind: "move", value: "pagedown"}}
			case 'S':
				return []*Action{{kind: "selection", value: "on"}, {kind: "move", value: "pagedown"}}
			case 'e':
				return []*Action{{kind: "selection", value: "off"}, {kind: "move", value: "bof"}}
			case 'E':
				return []*Action{{kind: "selection", value: "on"}, {kind: "move", value: "bof"}}
			case 'd':
				return []*Action{{kind: "selection", value: "off"}, {kind: "move", value: "eof"}}
			case 'D':
				return []*Action{{kind: "selection", value: "on"}, {kind: "move", value: "eof"}}
			case '1':
				return []*Action{{kind: "selection", value: "off"}, {kind: "move", value: "prevGlobal"}}
			case '!':
				return []*Action{{kind: "selection", value: "on"}, {kind: "move", value: "prevGlobal"}}
			case '2':
				return []*Action{{kind: "selection", value: "off"}, {kind: "move", value: "nextGlobal"}}
			case '@':
				return []*Action{{kind: "selection", value: "on"}, {kind: "move", value: "nextGlobal"}}
			case '9':
				return []*Action{{kind: "selection", value: "off"}, {kind: "move", value: "prevGlobal"}}
			case '(':
				return []*Action{{kind: "selection", value: "on"}, {kind: "move", value: "prevGlobal"}}
			case '0':
				return []*Action{{kind: "selection", value: "off"}, {kind: "move", value: "nextGlobal"}}
			case ')':
				return []*Action{{kind: "selection", value: "on"}, {kind: "move", value: "nextGlobal"}}
			case 'q':
				return []*Action{{kind: "selection", value: "off"}, {kind: "move", value: "prevIndentMatch"}}
			case 'Q':
				return []*Action{{kind: "selection", value: "on"}, {kind: "move", value: "prevIndentMatch"}}
			case 'a':
				return []*Action{{kind: "selection", value: "off"}, {kind: "move", value: "nextIndentMatch"}}
			case 'A':
				return []*Action{{kind: "selection", value: "on"}, {kind: "move", value: "nextIndentMatch"}}
			case ']', 'x':
				return []*Action{{kind: "selection", value: "off"}, {kind: "move", value: "nextArg"}}
			case '}', 'X':
				return []*Action{{kind: "selection", value: "on"}, {kind: "move", value: "nextArg"}}
			case '[', 'z':
				return []*Action{{kind: "selection", value: "off"}, {kind: "move", value: "prevArg"}}
			case '{', 'Z':
				return []*Action{{kind: "selection", value: "on"}, {kind: "move", value: "prevArg"}}
			case 'c':
				return []*Action{{kind: "selection", value: "off"}, {kind: "move", value: "matchingBracket"}}
			case 'C':
				return []*Action{{kind: "selection", value: "on"}, {kind: "move", value: "matchingBracket"}}
			default:
				return []*Action{}
			}
		}

		// key pressed without modifier
		if m.selection.on {
			return []*Action{{kind: "delete", value: "selection"}, {kind: "insert", value: string(ev.Ch)}}
		} else {
			return []*Action{{kind: "insert", value: string(ev.Ch)}}
		}
	}
}

// do takes an action and do it.
// After done the action, it will save result on the action.
func (m *NormalMode) do(a *Action) {
	a.beforeCursor = *m.cursor

	defer func() {
		a.afterCursor = *m.cursor
		if m.selection.on {
			m.selection.SetEnd(m.cursor)
		}
	}()

	switch a.kind {
	case "exit":
		if !m.text.edited {
			saveLastPosition(m.f, m.cursor.l, m.cursor.b)
			term.Close()
			os.Exit(0)
		} else {
			m.mode.ChangeTo(m.mode.exit)
		}
	case "save":
		err := save(m.f, m.text)
		if err != nil {
			m.err = fmt.Sprintf("FAIL TO SAVE: %v", err)
			return
		}
		m.text.edited = false
		m.status = fmt.Sprintf("successfully saved: %v", m.f)

		// post save
		if strings.HasSuffix(m.f, ".go") {
			cmds := []*exec.Cmd{
				exec.Command("goimports", "-w", m.f),
				exec.Command("go", "fmt", m.f),
			}
			for _, cmd := range cmds {
				if cmd.Path == "" {
					// Could not find the command.
					continue
				}
				out, err := cmd.CombinedOutput()
				if err != nil {
					outs := strings.Split(string(out), "\n")
					if len(outs) == 0 {
						m.err = fmt.Sprint(err)
					} else {
						m.err = outs[0]
					}
					return
				}
				// reload the file.
				text, err := open(m.f)
				if err != nil {
					m.err = fmt.Sprint(err)
					return
				}
				m.text = text
				m.cursor.text = text
				m.selection.text = text
				oldl := m.cursor.l
				oldb := m.cursor.b
				m.cursor.GotoLine(oldl)
				m.cursor.SetCloseToB(oldb)
			}
		}
	case "copy":
		if m.selection.on {
			minc, maxc := m.selection.MinMax()
			m.copied = m.text.DataInside(minc, maxc)
		} else {
			r, _ := m.cursor.RuneAfter()
			m.copied = string(r)
		}
		saveConfig("copy", m.copied)
	case "modeChange":
		if a.value == "find" {
			m.mode.ChangeTo(m.mode.find)
		} else if a.value == "replace" {
			m.mode.ChangeTo(m.mode.replace)
		} else if a.value == "gotoline" {
			m.mode.ChangeTo(m.mode.gotoline)
		}
	case "selection":
		if a.value == "on" && !m.selection.on {
			m.selection.on = true
			m.selection.SetStart(m.cursor)
		} else if a.value == "off" {
			m.selection.on = false
		}
	case "move":
		switch a.value {
		case "left":
			m.cursor.MoveLeft()
		case "right":
			m.cursor.MoveRight()
		case "selLeft":
			if m.selection.on {
				minc := m.selection.Min()
				m.cursor.GotoLine(minc.l)
				m.cursor.SetB(minc.b)
			} else {
				m.cursor.MoveLeft()
			}
		case "selRight":
			if m.selection.on {
				maxc := m.selection.Max()
				m.cursor.GotoLine(maxc.l)
				m.cursor.SetB(maxc.b)
			} else {
				m.cursor.MoveRight()
			}
		case "up":
			m.cursor.MoveUp()
		case "down":
			m.cursor.MoveDown()
		case "prevBowEow":
			m.cursor.MovePrevBowEow()
		case "nextBowEow":
			m.cursor.MoveNextBowEow()
		case "bol":
			m.cursor.MoveBol()
		case "eol":
			m.cursor.MoveEol()
		case "bocBolRepeat":
			m.cursor.MoveBocBolRepeat()
		case "pageup":
			m.cursor.PageUp()
		case "pagedown":
			m.cursor.PageDown()
		case "bof":
			m.cursor.MoveBof()
		case "eof":
			m.cursor.MoveEof()
		case "nextGlobal":
			m.cursor.GotoNextGlobalLine()
		case "prevGlobal":
			m.cursor.GotoPrevGlobalLine()
		case "prevIndentMatch":
			m.cursor.GotoPrevIndentMatch()
		case "nextIndentMatch":
			m.cursor.GotoNextIndentMatch()
		case "nextArg":
			m.cursor.GotoNextAny("{(,)}")
			r, _ := m.cursor.RuneAfter()
			if r == '(' || r == '{' {
				m.cursor.MoveRight()
			}
		case "prevArg":
			r, _ := m.cursor.RuneBefore()
			if r == '(' || r == '{' {
				m.cursor.MoveLeft()
			}
			m.cursor.GotoPrevAny("{(,)}")
			r, _ = m.cursor.RuneAfter()
			if r == '(' || r == '{' {
				m.cursor.MoveRight()
			}
		case "matchingBracket":
			m.cursor.GotoMatchingBracket()
		case "findPrev":
			ok := m.cursor.GotoPrev(m.mode.find.str)
			if !ok {
				m.cursor.GotoLast(m.mode.find.str)
			}
		case "findNext":
			ok := m.cursor.GotoNext(m.mode.find.str)
			if !ok {
				m.cursor.GotoFirst(m.mode.find.str)
			}
		case "findPrevWord":
			m.cursor.GotoPrevWord(m.mode.find.str)
		case "findNextWord":
			m.cursor.GotoNextWord(m.mode.find.str)
		// TODO: "findPrevSelect" and "findNextSelect" are hack. make separate action.
		case "findPrevSelect":
			ok := m.cursor.GotoPrev(m.mode.find.str)
			if !ok {
				ok = m.cursor.GotoLast(m.mode.find.str)
			}
			if ok {
				m.selection.on = true
				for range m.mode.find.str {
					m.cursor.MoveRight()
				}
				m.selection.SetStart(m.cursor)
				for range m.mode.find.str {
					m.cursor.MoveLeft()
				}
				m.selection.SetEnd(m.cursor)
			}
		case "findNextSelect":
			ok := m.cursor.GotoNext(m.mode.find.str)
			if !ok {
				ok = m.cursor.GotoFirst(m.mode.find.str)
			}
			if ok {
				m.selection.on = true
				for range m.mode.find.str {
					m.cursor.MoveRight()
				}
				m.selection.SetStart(m.cursor)
				for range m.mode.find.str {
					m.cursor.MoveLeft()
				}
				m.selection.SetEnd(m.cursor)
			}
		default:
			panic(fmt.Sprintln("what the..", a.value, "move?"))
		}
	case "insert":
		if a.value == "autoIndent" {
			prevline := m.text.lines[m.cursor.l-1].data
			trimed := strings.TrimLeft(prevline, " \t")
			indent := prevline[:len(prevline)-len(trimed)]
			m.cursor.Insert(indent)
			a.value = indent
			return
		}
		m.cursor.Insert(a.value)
	case "delete":
		if a.value == "selection" {
			if m.selection.on {
				*m.cursor, _ = m.selection.MinMax()
				a.beforeCursor = *m.cursor // rewrite before cursor.
			}
			a.value = m.selection.Data()
			for range a.value {
				m.cursor.Delete()
			}
			m.selection.on = false
		} else {
			a.value = m.cursor.Delete()
		}
	case "insertTab":
		tab := "\t"
		if m.text.tabToSpace {
			tab = strings.Repeat(" ", m.text.tabWidth)
		}
		lines := make([]int, 0)
		if m.selection.on {
			for _, l := range m.selection.Lines() {
				if m.text.Line(l).data != "" {
					lines = append(lines, l)
				}
			}
		} else {
			lines = append(lines, m.cursor.l)
		}
		tabedLine := ""
		for _, l := range lines {
			m.text.Line(l).Insert(tab, 0)
			if tabedLine != "" {
				tabedLine += ","
			}
			tabedLine += strconv.Itoa(l) + ":" + tab
			if l == m.cursor.l {
				m.cursor.SetB(m.cursor.b + len(tab))
			}
		}
		a.value = tabedLine
	case "removeTab":
		// removeTab is slightly differ from insertTab.
		// removeTab should remember what is removed, not tab string it self.
		lines := make([]int, 0)
		if m.selection.on {
			lines = m.selection.Lines()
		} else {
			lines = append(lines, m.cursor.l)
		}
		untabedLine := ""
		for _, l := range lines {
			removed := ""
			if strings.HasPrefix(m.text.Line(l).data, "\t") {
				removed += m.text.Line(l).Remove(0, 1)
			} else {
				for i := 0; i < m.text.tabWidth; i++ {
					if len(m.text.Line(l).data) == 0 {
						break
					}
					if !strings.HasPrefix(m.text.Line(l).data, " ") {
						break
					}
					removed += m.text.Line(l).Remove(0, 1)
				}
			}
			if untabedLine != "" {
				untabedLine += ","
			}
			untabedLine += strconv.Itoa(l) + ":" + removed
			if l == m.cursor.l && !m.cursor.AtBol() {
				b := m.cursor.b - len(removed)
				if b < 0 {
					b = 0
				}
				m.cursor.SetB(b)
			}
		}
		a.value = untabedLine
	case "backspace":
		a.value = m.cursor.Backspace()
	case "selectAll":
		m.cursor.MoveBof()
		m.selection.on = true
		m.selection.SetStart(m.cursor)
		m.cursor.MoveEof()
		m.selection.SetEnd(m.cursor)
	case "selectLine":
		m.cursor.MoveBol()
		if !m.selection.on {
			m.selection.on = true
			m.selection.SetStart(m.cursor)
		}
		if m.cursor.OnLastLine() {
			m.cursor.MoveEol()
		} else {
			m.cursor.MoveDown()
		}
		m.selection.SetEnd(m.cursor)
	case "selectWord":
		if !m.cursor.AtBow() {
			m.cursor.MovePrevBowEow()
		}
		if !m.selection.on {
			m.selection.on = true
			m.selection.SetStart(m.cursor)
		}
		m.cursor.MoveNextBowEow()
		m.selection.SetEnd(m.cursor)
	case "undo":
		// TODO: Move to history.Undo()
		if m.history.head == 0 {
			return
		}
		m.selection.on = false
		m.history.head--
		undoActions := m.history.At(m.history.head)
		for i := len(undoActions) - 1; i >= 0; i-- {
			u := undoActions[i]
			m.text = u.text
			m.cursor.text = u.text
			m.selection.text = u.text
			switch u.kind {
			case "insert":
				m.cursor.Copy(u.afterCursor)
				for range u.value {
					m.cursor.Backspace()
				}
			case "insertTab":
				lineInfos := strings.Split(u.value, ",")
				for _, li := range lineInfos {
					if li == "" {
						continue
					}
					lis := strings.Split(li, ":")
					lstr := lis[0]
					tab := lis[1]
					l, err := strconv.Atoi(lstr)
					if err != nil {
						panic(err)
					}
					for _, r := range tab {
						rr := m.text.Line(l).Remove(0, 1)
						if rr != string(r) {
							panic("removed and current is not matched")
						}
					}
				}
				m.cursor.Copy(u.beforeCursor)
			case "backspace":
				m.cursor.Copy(u.afterCursor)
				m.cursor.Insert(u.value)
			case "delete":
				m.cursor.Copy(u.afterCursor)
				m.cursor.Insert(u.value)
			case "removeTab":
				lineInfos := strings.Split(u.value, ",")
				for _, li := range lineInfos {
					if li == "" {
						continue
					}
					lis := strings.Split(li, ":")
					lstr := lis[0]
					removed := lis[1]
					l, err := strconv.Atoi(lstr)
					if err != nil {
						panic(err)
					}
					m.text.Line(l).Insert(removed, 0)
				}
				m.cursor.Copy(u.beforeCursor)
			case "move":
				m.cursor.Copy(u.beforeCursor)
			default:
				panic(fmt.Sprintln("what the..", u.kind, "history?"))
			}
		}
	case "redo":
		// TODO: Move to history.Redo()
		if m.history.head == m.history.Len() {
			return
		}
		m.selection.on = false
		redoActions := m.history.At(m.history.head)
		m.history.head++
		for _, r := range redoActions {
			m.text = r.text
			m.cursor.text = r.text
			m.selection.text = r.text
			switch r.kind {
			case "insert":
				m.cursor.Copy(r.beforeCursor)
				m.cursor.Insert(r.value)
			case "insertTab":
				lineInfos := strings.Split(r.value, ",")
				for _, li := range lineInfos {
					if li == "" {
						continue
					}
					lis := strings.Split(li, ":")
					lstr := lis[0]
					tab := lis[1]
					l, err := strconv.Atoi(lstr)
					if err != nil {
						panic(err)
					}
					m.text.Line(l).Insert(tab, 0)
				}
				m.cursor.Copy(r.afterCursor)
			case "backspace":
				m.cursor.Copy(r.beforeCursor)
				for range r.value {
					m.cursor.Backspace()
				}
			case "delete":
				m.cursor.Copy(r.beforeCursor)
				for range r.value {
					m.cursor.Delete()
				}
			case "removeTab":
				lineInfos := strings.Split(r.value, ",")
				for _, li := range lineInfos {
					if li == "" {
						continue
					}
					lis := strings.Split(li, ":")
					lstr := lis[0]
					removed := lis[1]
					l, err := strconv.Atoi(lstr)
					if err != nil {
						panic(err)
					}
					for _, r := range removed {
						rr := m.text.Line(l).Remove(0, 1)
						if rr != string(r) {
							panic("removed and current is not matched")
						}
					}
				}
				m.cursor.Copy(r.afterCursor)
			case "move":
				m.cursor.Copy(r.afterCursor)
			default:
				panic(fmt.Sprintln("what the..", r.kind, "history?"))
			}
		}
	default:
		panic(fmt.Sprintln("what the..", a.kind, "action?"))
	}
}

// Status returns a status as string.
// The status will cleared when normal mode takes another event.
func (m *NormalMode) Status() string {
	if m.status != "" {
		return m.status
	}
	return fmt.Sprintf("%v:%v:%v", m.f, m.cursor.l+1, m.cursor.O())
}

// Error returns an error of the last done action.
// If there was no error, it will return an empty string.
// The error will cleared when normal mode takes another event.
func (m *NormalMode) Error() string {
	return m.err
}
