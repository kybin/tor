package main

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Line
type Line struct {
	data string
}

func (ln *Line) Boc() int {
	remain := ln.data
	b := 0
	for len(remain) > 0 {
		r, rlen := utf8.DecodeRuneInString(remain)
		remain = remain[rlen:]
		if !unicode.IsSpace(r) {
			break
		}
		b += rlen
	}
	return b
}

func (ln *Line) Insert(r string, b int) {
	ln.data = ln.data[:b] + r + ln.data[b:]
}

func (ln *Line) Remove(from, to int) string {
	deleted := ln.data[from:to]
	ln.data = ln.data[:from] + ln.data[to:]
	return deleted
}

func (ln *Line) RemoveFrom(from int) string {
	deleted := ln.data[from:]
	ln.data = ln.data[:from]
	return deleted
}

func (ln *Line) RemoveTo(to int) string {
	deleted := ln.data[:to]
	ln.data = ln.data[to:]
	return deleted
}

func (ln *Line) InsertTab() {
	ln.data = "\t" + ln.data
}

func (ln *Line) RemoveTab() error {
	if !strings.HasPrefix(ln.data, "\t") {
		return errors.New("first character is not a tab.")
	}
	ln.data = ln.data[1:]
	return nil
}

// Text
type Text struct {
	lines []Line
	tabToSpace bool
	tabWidth int
}

func (t *Text) JoinNextLine(l int) {
	t.lines = append(append(t.lines[:l], Line{t.lines[l].data + t.lines[l+1].data}), t.lines[l+2:]...)
}

func (t *Text) SplitLine(l, b int) {
	prev := t.lines[l].data[:b]
	next := t.lines[l].data[b:]
	t.lines[l].data = prev
	t.InsertLine(Line{next}, l)
}

func (t *Text) InsertLine(ln Line, l int) {
	t.lines = append(append(append([]Line{}, t.lines[:l+1]...), ln), t.lines[l+1:]...)
}

func (t *Text) RemoveLine(l int) string {
	deleted := t.lines[l]
	t.lines = append(append([]Line{}, t.lines[:l]...), t.lines[l+1:]...)
	return deleted.data + "\n"
}

func (t *Text) RemoveRange(min, max Point) string {
	deleted := ""
	if min.l == max.l {
		deleted = t.lines[min.l].data[min.o:max.o]
	} else {
		focusLines := t.lines[min.l : max.l+1]
		for i, line := range focusLines {
			if i == 0 {
				deleted += line.data[min.o:]
			} else if i == len(focusLines)-1 {
				deleted += "\n"
				deleted += line.data[:max.o]
			} else {
				deleted += "\n"
				deleted += line.data
			}
		}
	}
	t.lines = append(append(append([]Line{}, t.lines[:min.l]...), Line{t.lines[min.l].data[:min.o] + t.lines[max.l].data[max.o:]}), t.lines[max.l+1:]...)
	return deleted
}

func (t *Text) Insert(r string, l, b int) {
	t.lines[l].Insert(r, b)
}

func (t *Text) Remove(l, from, to int) string {
	return t.lines[l].Remove(from, to)
}

func (t *Text) RemoveFrom(l, from int) string {
	return t.lines[l].RemoveFrom(from)
}

func (t *Text) RemoveTo(l, to int) string {
	return t.lines[l].RemoveTo(to)
}

func (t *Text) DataInside(min, max Cursor) string {
	if min.l == max.l {
		return t.lines[min.l].data[min.b:max.b]
	}
	data := ""
	for l := min.l; l < max.l+1; l++ {
		if l == min.l {
			data += t.lines[l].data[min.b:]
		} else if l == max.l {
			data += t.lines[l].data[:max.b]
		} else {
			data += t.lines[l].data
		}
		if l != max.l {
			data += "\n"
		}
	}
	return data
}
