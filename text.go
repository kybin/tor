package main

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/kybin/tor/cell"
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
	writable   bool
	lineEnding string
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

func (t *Text) RemoveRange(min, max cell.Pt) string {
	deleted := ""
	if min.L == max.L {
		deleted = t.lines[min.L].data[min.O:max.O]
	} else {
		focusLines := t.lines[min.L : max.L+1]
		for i, line := range focusLines {
			if i == 0 {
				deleted += line.data[min.O:]
			} else if i == len(focusLines)-1 {
				deleted += "\n"
				deleted += line.data[:max.O]
			} else {
				deleted += "\n"
				deleted += line.data
			}
		}
	}
	t.lines = append(append(append([]Line{}, t.lines[:min.L]...), Line{t.lines[min.L].data[:min.O] + t.lines[max.L].data[max.O:]}), t.lines[max.L+1:]...)
	return deleted
}

func (t *Text) Insert(r string, l, b int) {
	t.lines[l].Insert(r, b)
}

func (t *Text) Remove(l, from, to int) string {
	return t.lines[l].Remove(from, to)
}

func (t *Text) DataInside(min, max cell.Pt) string {
	if min.L == max.L {
		return t.lines[min.L].data[min.O:max.O]
	}
	data := ""
	for l := min.L; l < max.L+1; l++ {
		if l == min.L {
			data += t.lines[l].data[min.O:]
		} else if l == max.L {
			data += t.lines[l].data[:max.O]
		} else {
			data += t.lines[l].data
		}
		if l != max.L {
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
