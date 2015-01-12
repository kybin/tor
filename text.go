package main

// Line
type Line struct {
	data string
}

func (ln *Line) Insert(r rune, b int) {
	ln.data = ln.data[:b] + string(r) + ln.data[b:]
}

func (ln *Line) Remove(from, to int) {
	ln.data = ln.data[:from] + ln.data[to:]
}

// Text
type Text struct {
	lines []Line
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

func (t *Text) Insert(r rune, l, b int) {
	t.lines[l].Insert(r, b)
}

func (t *Text) Remove(l, from, to int) {
	t.lines[l].Remove(from, to)
}
