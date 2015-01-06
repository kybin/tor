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
	line int
	boff int // byte offset.
	voff int // visual offset. It may be different with cursor offset. For that, use cursor.offset().
	t Text
	// stick place - will implement later
}

func NewCursor(t Text) *Cursor {
	return &Cursor{0, 0, 0, t}
}

func SetTermboxCursor(c *Cursor, w *Window, l *Layout) {
	viewbound := l.MainViewerBound()
	viewl, viewo := viewbound.min.l, viewbound.min.o
	cl, co := c.PositionInWindow(w)
	SetCursor(viewl+cl, viewo+co)
}

// cursor offset cannot go further than line's maximum visual length.
func (c *Cursor) Offset() (coff int) {
	maxlen := c.LineVisualLength()
	if c.voff >  maxlen {
		return maxlen
	}

	// nursor should not in the middle of multi-length(visual) character.
	// So we should recalculate cursor offset.
	remain := c.LineData()
	lastcoff := 0
	for {
		r, rlen := utf8.DecodeRune(remain)
		remain = remain[rlen:]
		lastcoff = coff
		coff += RuneVisualLength(r)
		if coff==c.voff {
			return coff
		} else if coff > c.voff {
			return lastcoff
		}
	}
	return
}

// non-instance version of cursor.Offset() 
func CursorOffset(l Line, boff int) int {
	remain := l
	coff := 0
	for boff > 0 {
		r, rlen := utf8.DecodeRune(remain)
		remain = remain[rlen:]
		boff -= rlen
		coff += RuneVisualLength(r)
	}
	return coff
}

func (c *Cursor) Position() (int, int) {
	return c.line, c.Offset()
}

func (c *Cursor) PositionInWindow(w *Window) (int, int) {
	return c.line-w.min.l, c.Offset()-w.min.o
}

// text on the line
func (c *Cursor) LineData() Line {
	return c.t[c.line]
}

func (c *Cursor) LineDataUntilCursor() Line {
	return c.LineData()[:c.boff]
}

func (c *Cursor) LineDataFromCursor() Line {
	return c.LineData()[c.boff:]
}

func (c *Cursor) ExceededLineLimit() bool {
	return c.boff > len(c.LineData())
}

func (c *Cursor) RuneAfter() (rune, int) {
	return utf8.DecodeRune(c.LineData()[c.boff:])
}

func (c *Cursor) RuneBefore() (rune, int) {
	return utf8.DecodeLastRune(c.LineData()[:c.boff])
}

// should refine after
// may be use dictionary??
func RuneVisualLength(r rune) int {
	if r=='\t' {
		return taboffset
	}
	return 1
}

func (c *Cursor) LineVisualLength() int {
	return c.VoffFromBoff(len(c.LineData()))
}

// Set boff and voff from current cursor position.
func (c *Cursor) ResetInternalOffsets() {
	c.boff = c.BoffFromCoff(c.Offset())
	c.voff = c.VoffFromBoff(c.boff)
}

func (c *Cursor) BoffFromCoff(coff int) (boff int) {
	remain := c.LineData()
	for coff>0 {
		r, rlen := utf8.DecodeRune(remain)
		remain = remain[rlen:]
		boff += rlen
		coff -= RuneVisualLength(r)
	}
	return
}

func (c *Cursor) VoffFromBoff(boff int) (voff int){
	remain := c.LineData()[:boff]
	for len(remain) > 0 {
		r, rlen := utf8.DecodeRune(remain)
		remain = remain[rlen:]
		voff += RuneVisualLength(r)
	}
	return
}

func (c *Cursor) SetOffsets(boff int) {
	c.boff = boff
	c.voff = c.VoffFromBoff(boff)
}

func (c *Cursor) AtBol() bool{
	return c.boff == 0
}

func (c *Cursor) AtEol() bool{
	return c.boff == len(c.LineData())
}

func (c *Cursor) OnFirstLine() bool{
	return c.line == 0
}

func (c *Cursor) OnLastLine() bool {
	return c.line == len(c.t)-1
}

func (c *Cursor) AtBof() bool {
	return c.OnFirstLine() && c.AtBol()
}

func (c *Cursor) AtEof() bool {
	return c.OnLastLine() && c.AtEol()
}

func (c *Cursor) MoveLeft() {
	c.ResetInternalOffsets()
	if c.AtBof() {
		return
	} else if c.AtBol() {
		c.line -= 1
		c.boff = len(c.LineData())
		c.voff = c.VoffFromBoff(c.boff)
		return
	}
	r, rlen := c.RuneBefore()
	c.boff -= rlen
	c.voff -= RuneVisualLength(r)
}

func (c *Cursor) MoveRight() {
	c.ResetInternalOffsets()
	if c.AtEof() {
		return
	} else if c.AtEol() || c.ExceededLineLimit(){
		c.line += 1
		c.SetOffsets(0)
		return
	}
	r, rlen := c.RuneAfter()
	c.boff += rlen
	c.voff += RuneVisualLength(r)
}

// With move up and down, we will lazily evaluate internal offsets.
func (c *Cursor) MoveUp() {
	if c.OnFirstLine() {
		return
	}
	c.line--

}

func (c *Cursor) MoveDown() {
	if c.OnLastLine() {
		return
	}
	c.line++
}

func (c *Cursor) MoveBow() {
	c.ResetInternalOffsets()
	if c.AtBof() {
		return
	}
	// First we should pass every space character.
	for {
		r, _ := c.RuneBefore()
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
		r, _ := c.RuneBefore()
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
	c.ResetInternalOffsets()
	if c.AtEof() {
		return
	}
	for {
		r, _ := c.RuneAfter()
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			break
		}
		c.MoveRight()
		if c.AtEof() {
			return
		}
	}
	for {
		r, _ := c.RuneAfter()
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
	if c.boff == 0 && !c.OnFirstLine() {
		c.line--
		return
	}

	// if  prev data is all spaces, move cursor to beginning of line
	remain := c.LineDataUntilCursor()
	allspace := true
	for len(remain)>0 {
		r, rlen := utf8.DecodeRune(remain)
		remain = remain[rlen:]
		if !unicode.IsSpace(r) {
			allspace = false
			break
		}
	}
	if allspace {
		c.boff = 0
		c.voff = 0
		return
	}

	// or, move cursor to first character of text on line.
	remain = c.LineData()
	boff := 0
        for len(remain)>0 {
		r, rlen := utf8.DecodeRune(remain)
		remain = remain[rlen:]
		if !unicode.IsSpace(r) {
			break
		}
		boff += rlen
	}
	c.boff = boff
	c.voff = c.VoffFromBoff(boff)
}

func (c *Cursor) MoveEol() {
	// if already eol, move to next line
	if c.boff == len(c.LineData()) && !c.OnLastLine() {
		c.line++
	}

	// if  prev data is not all spaces move cursor to eol.
	remain := c.LineData()[:c.boff+1] // we should use runelength instead of 1
	allspace := true
	for len(remain)>0 {
		r, rlen := utf8.DecodeRune(remain)
		remain = remain[rlen:]
		if !unicode.IsSpace(r) {
			allspace = false
			break
		}
	}
	if !allspace {
		c.boff = len(c.LineData())
		c.voff = c.VoffFromBoff(c.boff)
		return
	}

	// or, move it to first chararacter of text on line.
	remain = c.LineData()
	boff := 0
	for len(remain)>0 { // will make this a function
		r, rlen := utf8.DecodeRune(remain)
		remain = remain[rlen:]
		if !unicode.IsSpace(r) {
			break
		}
		boff +=rlen
	}
	c.boff = boff
	c.voff = c.VoffFromBoff(boff)
}

func (c *Cursor) PageUp() {
	for i:=0; i < pageoffset; i++ {
		if c.OnFirstLine() {
			break
		}
		c.line--
	}
}

func (c *Cursor) PageDown() {
	for i:=0; i < pageoffset; i++ {
		if c.OnLastLine() {
			break
		}
		c.line++
	}
}

func (c *Cursor) MoveBof() {
	for {
		if c.OnFirstLine() {
			break
		}
		c.line--
	}
	c.MoveBol()
}

func (c *Cursor) MoveEof() {
	for {
		if c.OnLastLine() {
			break
		}
		c.line++
	}
	c.MoveEol()
}
