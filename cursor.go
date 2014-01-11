package main

import (
	"unicode"
	"unicode/utf8"
	term "github.com/nsf/termbox-go"
)

var (
	taboffset = 4
)

//type place int
//
//var (
//	NONE place = iota
//	BOC
//	EOC
//	EOL
//)

type cursor struct {
	line int
	boff int // byte offset.
	voff int // visual offset. It may be different with cursor offset. For that, use cursor.offset().
	t text
	// stick place - will implement later
}

func newCursor(t text) *cursor {
	return &cursor{0, 0, 0, t}
}

func setTermboxCursor(c *cursor, v *viewer, l *layout) {
	viewbound := l.mainViewerBound()
	viewx, viewy := viewbound.Min.X, viewbound.Min.Y
	cy, cx := c.positionInViewer(v)
	term.SetCursor(viewx+cx, viewy+cy)
}

// cursor offset cannot go further than line's maximum visual length.
func (c *cursor) offset() (coff int) {
	maxlen := c.lineVisualLength()
	if c.voff >  maxlen {
		return maxlen
	}

	// Cursor should not in the middle of multi-length(visual) character.
	// So we should recalculate cursor offset.
	remain := c.lineData()
	lastcoff := 0
	for {
		r, rlen := utf8.DecodeRune(remain)
		remain = remain[rlen:]
		lastcoff = coff
		coff += runeVisualLength(r)
		if coff==c.voff {
			return coff
		} else if coff > c.voff {
			return lastcoff
		}
	}
	return
}

func (c *cursor) position() (int, int) {
	return c.line, c.offset()
}

func (c *cursor) positionInViewer(v *viewer) (int, int) {
	return c.line-v.min.Y, c.offset()-v.min.X
}

// text on the line
func (c *cursor) lineData() line {
	return c.t[c.line]
}

func (c *cursor) lineDataUntilCursor() line {
	return c.lineData()[:c.boff]
}

func (c *cursor) lineDataFromCursor() line {
	return c.lineData()[c.boff:]
}

func (c *cursor) exceededLineLimit() bool {
	return c.boff > len(c.lineData())
}

func (c *cursor) runeAfter() (rune, int) {
	return utf8.DecodeRune(c.lineData()[c.boff:])
}

func (c *cursor) runeBefore() (rune, int) {
	return utf8.DecodeLastRune(c.lineData()[:c.boff])
}

// should refine after
// may be use dictionary??
func runeVisualLength(r rune) int {
	if r=='\t' {
		return taboffset
	}
	return 1
}

func (c *cursor) lineVisualLength() int {
	return c.voffFromBoff(len(c.lineData()))
}

// Set boff and voff from current cursor position.
func (c *cursor) resetInternalOffsets() {
	c.boff = c.boffFromCoff(c.offset())
	c.voff = c.voffFromBoff(c.boff)
}

func (c *cursor) boffFromCoff(coff int) (boff int) {
	remain := c.lineData()
	for coff>0 {
		r, rlen := utf8.DecodeRune(remain)
		remain = remain[rlen:]
		boff += rlen
		coff -= runeVisualLength(r)
	}
	return
}

func (c *cursor) voffFromBoff(boff int) (voff int){
	remain := c.lineData()[:boff]
	for len(remain) > 0 {
		r, rlen := utf8.DecodeRune(remain)
		remain = remain[rlen:]
		voff += runeVisualLength(r)
	}
	return
}

func (c *cursor) setOffsets(boff int) {
	c.boff = boff
	c.voff = c.voffFromBoff(boff)
}

func (c *cursor) atBol() bool{
	return c.boff == 0
}

func (c *cursor) atEol() bool{
	return c.boff == len(c.lineData())
}

func (c *cursor) onFirstline() bool{
	return c.line == 0
}

func (c *cursor) onLastline() bool {
	return c.line == len(c.t)-1
}

func (c *cursor) atBof() bool {
	return c.onFirstline() && c.atBol()
}

func (c *cursor) atEof() bool {
	return c.onLastline() && c.atEol()
}

func (c *cursor) moveLeft() {
	c.resetInternalOffsets()
	if c.atBof() {
		return
	} else if c.atBol() {
		c.line -= 1
		c.boff = len(c.lineData())
		c.voff = c.voffFromBoff(c.boff)
		return
	}
	r, rlen := c.runeBefore()
	c.boff -= rlen
	c.voff -= runeVisualLength(r)
}

func (c *cursor) moveRight() {
	c.resetInternalOffsets()
	if c.atEof() {
		return
	} else if c.atEol() || c.exceededLineLimit(){
		c.line += 1
		c.setOffsets(0)
		return
	}
	r, rlen := c.runeAfter()
	c.boff += rlen
	c.voff += runeVisualLength(r)
}

// With move up and down, we will lazily evaluate internal offsets.
func (c *cursor) moveUp() {
	if c.onFirstline() {
		return
	}
	c.line--

}

func (c *cursor) moveDown() {
	if c.onLastline() {
		return
	}
	c.line++
}

func (c *cursor) moveBow() {
	c.resetInternalOffsets()
	if c.atBof() {
		return
	}
	// First we should pass every space character.
	for {
		r, _ := c.runeBefore()
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			break
		}
		c.moveLeft()
	}
	// Then we will find first space charactor and stop.
	for {
		r, _ := c.runeBefore()
		if !(unicode.IsLetter(r) || unicode.IsDigit(r)) {
			return
		}
		c.moveLeft()
	}
}

// See moveEow for the algorithm.
func (c *cursor) moveEow() {
	c.resetInternalOffsets()
	if c.atEof() {
		return
	}
	for {
		r, _ := c.runeAfter()
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			break
		}
		c.moveRight()
	}
	for {
		r, _ := c.runeAfter()
		if !(unicode.IsLetter(r) || unicode.IsDigit(r)) {
			return
		}
		c.moveRight()
	}
}

func (c *cursor) moveBol() {
	// if already bol, move cursor to prev line
	if c.boff == 0 && !c.onFirstline() {
		c.line--
		return
	}

	// if  prev data is all spaces, move cursor to beginning of line
	remain := c.lineDataUntilCursor()
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
	remain = c.lineData()
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
	c.voff = c.voffFromBoff(boff)
}

func (c *cursor) moveEol() {
	// if already eol, move to next line
	if c.boff == len(c.lineData()) && !c.onLastline() {
		c.line++
	}

	// if  prev data is not all spaces move cursor to eol.
	remain := c.lineData()[:c.boff+1] // we should use runelength instead of 1
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
		c.boff = len(c.lineData())
		c.voff = c.voffFromBoff(c.boff)
		return
	}

	// or, move it to first chararacter of text on line.
	remain = c.lineData()
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
	c.voff = c.voffFromBoff(boff)
}
