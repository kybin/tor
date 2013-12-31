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
//	BOC place = iota
//	EOC
//	EOL
//)

type cursor struct {
	txt text
	linenum int
	off int // byte offset.
	visoff int // visual offset. It may be different with actual cursor placement. For that, use cursor.cursorOffset().
	// stick place
}

func initializeCursor(t text) *cursor {
	term.SetCursor(0, 0)
	return &cursor{t, 0, 0, 0}
}

func setVisualCursor(c *cursor) {
	term.SetCursor(c.cursorOffset(), c.linenum)
}

func (c *cursor) linedata() line {
	return c.txt[c.linenum]
}

// when we move the cursor up or down, it will not recalculate it's offsets.
// This method tell us whether we should recalcuate offsets or not.
func (c *cursor) exceededLineLimit() bool {
	return c.off > len(c.linedata())
}

func (c *cursor) runeAfter() (rune, int) {
	return utf8.DecodeRune(c.linedata()[c.off:])
}

func (c *cursor) runeBefore() (rune, int) {
	return utf8.DecodeLastRune(c.linedata()[:c.off])
}

func runeVisualLength(r rune) int {
	if r=='\t' {
		return taboffset
	}
	return 1
}

func (c *cursor) lineVisualLength() int {
	return c.visualOffsetFromByteOffset(len(c.linedata()))
}

// Set byte and visual offsets from current cursor position.
func (c *cursor) resetInternalOffsets() {
	c.off = c.byteOffsetFromCursorOffset(c.cursorOffset())
	c.visoff = c.visualOffsetFromByteOffset(c.off)
}

func (c *cursor) byteOffsetFromCursorOffset(cursoroff int) (byteoff int) {
	remaintext := c.linedata()
	for cursoroff>0 {
		r, rlen := utf8.DecodeRune(remaintext)
		remaintext = remaintext[rlen:]
		byteoff += rlen
		cursoroff -= runeVisualLength(r)
	}
	return
}

func (c *cursor) visualOffsetFromByteOffset(byteoff int) (visoff int){
	remaintext := c.linedata()[:byteoff]
	for len(remaintext) > 0 {
		r, rlen := utf8.DecodeRune(remaintext)
		remaintext = remaintext[rlen:]
		visoff += runeVisualLength(r)
	}
	return
}

func (c *cursor) setOffsets(off int) {
	c.off = off
	c.visoff = c.visualOffsetFromByteOffset(off)
}

// cursor offset on terminal. It cannot exceeded line's maximum visual length.
func (c *cursor) cursorOffset() (cursoroff int) {
	linevisoff := c.lineVisualLength()
	if c.visoff >  linevisoff {
		return linevisoff
	}

	// Cursor should not in the middle of multi (v) length character.
	// So we should recalculate cursor offset.
	remaintext := c.linedata()
	lastcursor := 0
	for {
		r, rlen := utf8.DecodeRune(remaintext)
		remaintext = remaintext[rlen:]
		lastcursor = cursoroff
		cursoroff += runeVisualLength(r)
		if cursoroff==c.visoff {
			return cursoroff
		} else if cursoroff > c.visoff {
			return lastcursor
		}
	}
	return
}

func (c *cursor) atBol() bool{
	return c.off == 0
}

func (c *cursor) atEol() bool{
	return c.off == len(c.linedata())
}

func (c *cursor) onFirstline() bool{
	return c.linenum == 0
}

func (c *cursor) onLastline() bool {
	return c.linenum == len(c.txt)-1
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
		c.linenum -= 1
		c.setOffsets(len(c.linedata()))
		return
	} else if c.exceededLineLimit() {
		c.off = len(c.linedata())
	}
	r, rlen := c.runeBefore()
	c.off -= rlen
	c.visoff -= runeVisualLength(r)
}

func (c *cursor) moveRight() {
	c.resetInternalOffsets()
	if c.atEof() {
		return
	} else if c.atEol() || c.exceededLineLimit(){
		c.linenum += 1
		c.setOffsets(0)
		return
	}
	r, rlen := c.runeAfter()
	c.off += rlen
	c.visoff += runeVisualLength(r)
}

func (c *cursor) moveUp() {
	if c.onFirstline() {
		return
	}
	c.linenum--

}

func (c *cursor) moveDown() {
	if c.onLastline() {
		return
	}
	c.linenum++
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
	if c.off == 0 && !c.onFirstline() {
		c.linenum--
		return
	}
	remaintext := c.linedata()[:c.off]
	// if  prev data is spaces
	allspace := true
	for len(remaintext)>0 {
		r, rlen := utf8.DecodeRune(remaintext)
		remaintext = remaintext[rlen:]
		if !unicode.IsSpace(r) {
			allspace = false
			break
		}
	}
	if allspace {
		c.off = 0
		c.visoff = 0
		return
	}
	remaintext = c.linedata()
	byteoff := 0
        for len(remaintext)>0 {
		r, rlen := utf8.DecodeRune(remaintext)
		remaintext = remaintext[rlen:]
		if !unicode.IsSpace(r) {
			break
		}
		byteoff += rlen
	}
	c.off = byteoff
	c.visoff = c.visualOffsetFromByteOffset(byteoff)
}

func (c *cursor) moveEol() {
	if c.off == len(c.linedata()) && !c.onLastline() {
		c.linenum++
	}
	remaintext := c.linedata()[:c.off+1] // with itself
	allspace := true
	for len(remaintext)>0 {
		r, rlen := utf8.DecodeRune(remaintext)
		remaintext = remaintext[rlen:]
		if !unicode.IsSpace(r) {
			allspace = false
			break
		}
	}
	if !allspace { // NOT first.
		c.off = len(c.linedata())
		c.visoff = c.visualOffsetFromByteOffset(c.off)
		return
	}
	remaintext = c.linedata()
	byteoff := 0
	for len(remaintext)>0 { // will make this a function
		r, rlen := utf8.DecodeRune(remaintext)
		remaintext = remaintext[rlen:]
		if !unicode.IsSpace(r) {
			break
		}
		byteoff +=rlen
	}
	c.off = byteoff
	c.visoff = c.visualOffsetFromByteOffset(byteoff)
}
