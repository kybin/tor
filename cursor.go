package main

import (
	"unicode"
	"unicode/utf8"
)

var (
	taboffset = 4
	pageoffset =8
)

//type place int
//
//var (
//	NONE place = iota
//	BOC
//	EOC
//	EOL
//)

type Cursor struct {
	l int // line offset
	o int // cursor offset - When MoveUp or MoveDown, it will calculated from visual offset.
	v int // visual offset - When MoveLeft of MoveRight, it will matched to cursor offset.
	b int // byte offset
	t *Text
	// stick place - will implement later
}

func NewCursor(t *Text) *Cursor {
	return &Cursor{0, 0, 0, 0, t}
}

func SetTermboxCursor(c *Cursor, w *Window, l *Layout) {
	view := l.MainViewerBound()
	p := c.PositionInWindow(w)
	SetCursor(view.min.l+p.l, view.min.o+p.o)
}

func (c *Cursor) SetOffsets(b int) {
	c.b = b
	c.v = c.VFromB(b)
	c.o = c.v
}

// Before shifting, visual offset will matched to cursor offset.
func (c *Cursor) ShiftOffsets(b, v int) {
	c.v = c.o
	c.b += b
	c.v += v
	c.o += v
}

// After MoveUp or MoveDown, we need reclaculate cursor offsets (except visual offset).
func (c *Cursor) RecalculateOffsets() {
	c.o = c.OFromV(c.v)
	c.b = c.BFromC(c.o)
}

func (c *Cursor) OFromV(v int) (o int) {
	// Cursor offset cannot go further than line's maximum visual length.
	maxv := c.LineVisualLength()
	if v >  maxv {
		return maxv
	}
	// It's not allowed the cursor is in the middle of multi-length(visual) character.
	// So we need recaculate the cursors offset.
	remain := c.LineData()
	lasto := 0
	for {
		r, rlen := utf8.DecodeRuneInString(remain)
		remain = remain[rlen:]
		lasto = o
		o += RuneVisualLength(r)
		if o==v {
			return o
		} else if o > v {
			return lasto
		}
	}
}


func (c *Cursor) BFromC(o int) (b int) {
	remain := c.LineData()
	for o>0 {
		r, rlen := utf8.DecodeRuneInString(remain)
		remain = remain[rlen:]
		b+= rlen
		o-= RuneVisualLength(r)
	}
	return
}

func (c *Cursor) VFromB(b int) (v int){
	remain := c.LineData()[:b]
	for len(remain) > 0 {
		r, rlen := utf8.DecodeRuneInString(remain)
		remain = remain[rlen:]
		v += RuneVisualLength(r)
	}
	return
}

func (c *Cursor) Position() Point {
	return Point{c.l, c.o}
}

// TODO : relativePosition(p Point) Point ?
func (c *Cursor) PositionInWindow(w *Window) Point {
	return c.Position().Sub(w.min)
}

func (c *Cursor) LineData() string {
	return c.t.lines[c.l].data
}

func (c *Cursor) LineDataUntilCursor() string {
	return c.LineData()[:c.b]
}

func (c *Cursor) LineDataFromCursor() string {
	return c.LineData()[c.b:]
}

func (c *Cursor) ExceededLineLimit() bool {
	return c.b > len(c.LineData())
}

func (c *Cursor) RuneAfter() (rune, int, int) {
	r, rlen := utf8.DecodeRuneInString(c.LineData()[c.b:])
	return r, rlen, RuneVisualLength(r)
}

func (c *Cursor) RuneBefore() (rune, int, int) {
	r, rlen := utf8.DecodeLastRuneInString(c.LineData()[:c.b])
	return r, rlen, RuneVisualLength(r)
}

// should refine after
// may be use dictionary??
func RuneVisualLength(r rune) int {
	if r=='\t' {
		return taboffset
	}
	return 1
}

func (c *Cursor) LineByteLength() int {
	return len(c.LineData())
}

func (c *Cursor) LineVisualLength() int {
	return c.VFromB(c.LineByteLength())
}

func (c *Cursor) AtBol() bool{
	return c.b == 0
}

func (c *Cursor) AtEol() bool{
	return c.b == c.LineByteLength()
}

func (c *Cursor) OnFirstLine() bool{
	return c.l == 0
}

func (c *Cursor) OnLastLine() bool {
	return c.l == len(c.t.lines)-1
}

func (c *Cursor) AtBof() bool {
	return c.OnFirstLine() && c.AtBol()
}

func (c *Cursor) AtEof() bool {
	return c.OnLastLine() && c.AtEol()
}

func (c *Cursor) MoveLeft() {
	if c.AtBof() {
		return
	} else if c.AtBol() {
		c.l--
		c.SetOffsets(c.LineByteLength())
		return
	}
	_, rlen, vlen := c.RuneBefore()
	c.ShiftOffsets(-rlen, -vlen)
}

func (c *Cursor) MoveRight() {
	if c.AtEof() {
		return
	} else if c.AtEol() || c.ExceededLineLimit(){
		c.l++
		c.SetOffsets(0)
		return
	}
	_, rlen, vlen := c.RuneAfter()
	c.ShiftOffsets(rlen, vlen)
}

func (c *Cursor) MoveUp() {
	if c.OnFirstLine() {
		return
	}
	c.l--
	c.RecalculateOffsets()
}

func (c *Cursor) MoveDown() {
	if c.OnLastLine() {
		return
	}
	c.l++
	c.RecalculateOffsets()
}

func (c *Cursor) MoveBow() {
	if c.AtBof() {
		return
	}
	// First we should pass every space character.
	for {
		r, _, _ := c.RuneBefore()
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			break
		}
		c.MoveLeft()
		if c.AtBof() {
			return
		}
	}
	// Then we will find first space charactor and stop.
	for {
		r, _, _ := c.RuneBefore()
		if !(unicode.IsLetter(r) || unicode.IsDigit(r)) {
			return
		}
		c.MoveLeft()
		if c.AtBof() {
			return
		}
	}
}

// See moveEow for the algorithm.
func (c *Cursor) MoveEow() {
	if c.AtEof() {
		return
	}
	for {
		r, _, _ := c.RuneAfter()
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			break
		}
		c.MoveRight()
		if c.AtEof() {
			return
		}
	}
	for {
		r, _, _ := c.RuneAfter()
		if !(unicode.IsLetter(r) || unicode.IsDigit(r)) {
			return
		}
		c.MoveRight()
		if c.AtEof() {
			return
		}
	}
}

func (c *Cursor) MoveBol() {
	// if already bol, move cursor to prev line
	if c.AtBol() && !c.OnFirstLine() {
		c.MoveUp()
		return
	}

	// if  prev data is all spaces, move cursor to beginning of line
	remain := c.LineDataUntilCursor()
	allspace := true
	for len(remain)>0 {
		r, rlen := utf8.DecodeRuneInString(remain)
		remain = remain[rlen:]
		if !unicode.IsSpace(r) {
			allspace = false
			break
		}
	}
	if allspace {
		c.SetOffsets(0)
		return
	}

	// or, move cursor to first character of text on line.
	remain = c.LineData()
	b := 0
        for len(remain)>0 {
		r, rlen := utf8.DecodeRuneInString(remain)
		remain = remain[rlen:]
		if !unicode.IsSpace(r) {
			break
		}
		b += rlen
	}
	c.SetOffsets(b)
}

func (c *Cursor) MoveEol() {
	// if already eol, move to next line
	if c.b == len(c.LineData()) && !c.OnLastLine() {
		c.MoveDown()
	}

	// if  prev data is not all spaces move cursor to eol.
	remain := c.LineData()[:c.b+1] // we should use runelength instead of 1
	allspace := true
	for len(remain)>0 {
		r, rlen := utf8.DecodeRuneInString(remain)
		remain = remain[rlen:]
		if !unicode.IsSpace(r) {
			allspace = false
			break
		}
	}
	if !allspace {
		c.SetOffsets(c.LineByteLength())
		return
	}

	// or, move it to first chararacter of text on line.
	remain = c.LineData()
	b := 0
	for len(remain)>0 { // will make this a function
		r, rlen := utf8.DecodeRuneInString(remain)
		remain = remain[rlen:]
		if !unicode.IsSpace(r) {
			break
		}
		b +=rlen
	}
	c.SetOffsets(b)
}

func (c *Cursor) PageUp() {
	for i:=0; i < pageoffset; i++ {
		if c.OnFirstLine() {
			break
		}
		c.MoveUp()
	}
}

func (c *Cursor) PageDown() {
	for i:=0; i < pageoffset; i++ {
		if c.OnLastLine() {
			break
		}
		c.MoveDown()
	}
}

func (c *Cursor) MoveBof() {
	for {
		if c.OnFirstLine() {
			break
		}
		c.MoveUp()
	}
	c.MoveBol()
}

func (c *Cursor) MoveEof() {
	for {
		if c.OnLastLine() {
			break
		}
		c.MoveDown()
	}
	c.MoveEol()
}

func (c *Cursor) SplitLine() {
	c.t.SplitLine(c.l, c.b)
	c.MoveDown()
	c.SetOffsets(0)
}

func (c *Cursor) Insert(r rune) {
	c.t.Insert(r, c.l, c.b)
}

func (c *Cursor) Delete() {
	if c.AtEof() {
		return
	}
	remain := c.LineDataFromCursor()
	if len(remain) == 0 {
		// reach at end of line. join with bottom line.
		c.t.lines = append(append(c.t.lines[:c.l], Line{c.LineData()+c.t.lines[c.l+1].data}), c.t.lines[c.l+2:]...)
		return
	}
	_, rlen := utf8.DecodeRuneInString(remain)
	c.t.Remove(c.l, c.b, c.b+rlen)
}

func (c *Cursor) Backspace() {
	if c.AtBof() {
		return
	}
	c.MoveLeft()
	c.Delete()
}
