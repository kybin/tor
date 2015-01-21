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

func (ln *Line) RemoveFrom(from int) {
	ln.data = ln.data[:from]
}

func (ln *Line) RemoveTo(to int) {
	ln.data = ln.data[to:]
}

// Text
type Text struct {
	lines []Line
}

func (t *Text) JoinNextLine(l int) {
	t.lines = append(append(t.lines[:l], Line{t.lines[l].data+t.lines[l+1].data}), t.lines[l+2:]...)
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

func (t *Text) RemoveLine(l int) {
	t.lines = append(append([]Line{}, t.lines[:l]...), t.lines[l+1:]...)
}

func (t *Text) RemoveRange(min, max Point) {
	t.lines = append(append(append([]Line{}, t.lines[:min.l]...), Line{t.lines[min.l].data[:min.o]+t.lines[max.l].data[max.o:]}), t.lines[max.l+1:]...)
}

func (t *Text) Insert(r rune, l, b int) {
	t.lines[l].Insert(r, b)
}

func (t *Text) Remove(l, from, to int) {
	t.lines[l].Remove(from, to)
}

func (t *Text) RemoveFrom(l, from int) {
	t.lines[l].RemoveFrom(from)
}

func (t *Text) RemoveTo(l, to int) {
	t.lines[l].RemoveTo(to)
}
