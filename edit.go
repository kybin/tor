package main

import (
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

// Text
type Text struct {
	lines      []Line
	tabToSpace bool
	tabWidth   int
	edited     bool
}

func (t *Text) Line(l int) *Line {
	return &t.lines[l]
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

func (t *Text) DataInside(min, max Point) string {
	if min.l == max.l {
		return t.lines[min.l].data[min.o:max.o]
	}
	data := ""
	for l := min.l; l < max.l+1; l++ {
		if l == min.l {
			data += t.lines[l].data[min.o:]
		} else if l == max.l {
			data += t.lines[l].data[:max.o]
		} else {
			data += t.lines[l].data
		}
		if l != max.l {
			data += "\n"
		}
	}
	return data
}

func (t *Text) Bytes() []byte {
	datas := []string{}
	for _, l := range t.lines {
		datas = append(datas, l.data)
	}
	return []byte(strings.Join(datas, "\n"))
}
