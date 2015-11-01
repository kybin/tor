package main

import (
	"unicode"
	"unicode/utf8"
	"strings"
)

var (
	taboffset = 4
	pageoffset = 16
)

type Cursor struct {
	l int // line offset
	b int // byte offset
	o int // visual offset - not always matches with Cursor.O()
	t *Text
}

func NewCursor(t *Text) *Cursor {
	return &Cursor{0, 0, 0, t}
}

func (c *Cursor) Copy(c2 Cursor) {
	c.l = c2.l
	c.b = c2.b
	c.o = c2.o
}

func (c *Cursor) B() int {
	return c.b
}

func (c *Cursor) O() int {
	maxo := vlen(c.LineData())
	if c.o >  maxo {
		return maxo
	}
	// show cursor as well in mult-vis-character
	remain := c.LineData()
	o := 0
	for {
		r, rlen := utf8.DecodeRuneInString(remain)
		remain = remain[rlen:]
		lasto := o
		o += vlen(string(r))
		if o == c.o {
			return o
		} else if o > c.o {
			return lasto
		}
	}
	panic("should not reach here")
}

func (c *Cursor) SetB(b int) {
	o := 0
	remain := c.LineData()[:b]
	for len(remain) > 0 {
		r, rlen := utf8.DecodeRuneInString(remain)
		remain = remain[rlen:]
		o += vlen(string(r))
	}
	c.o = o
	c.b = b
}

func (c *Cursor) SetO(o int) {
	c.o = o
	c.RecalcB()
}

// if b is from lastpos file, it may less correct.
func (c *Cursor) SetCloseToB(tb int) {
	if tb > len(c.LineData()) {
		tb = len(c.LineData())
	}
	o, b := 0, 0
	remain := c.LineData()
	for len(remain) > 0 {
		r, rlen := utf8.DecodeRuneInString(remain)
		remain = remain[rlen:]
		lasto, lastb := o, b
		b += rlen
		o += vlen(string(r))
		if b == tb {
			c.b = b
			c.o = o
			return
		} else if b >= tb {
			c.b = lastb
			c.o = lasto
			return
		}
	}
}

// After MoveUp or MoveDown, we need reclaculate byte offset.
func (c *Cursor) RecalcB() {
	c.b = BFromO(c.LineData(), c.O())
}

func BFromO(line string, o int) (b int) {
	remain := line
	for o > 0 {
		r, rlen := utf8.DecodeRuneInString(remain)
		remain = remain[rlen:]
		b += rlen
		o -= vlen(string(r))
	}
	return
}

func (c *Cursor) Position() Point {
	return Point{c.l, c.O()}
}

func (c *Cursor) Line() *Line {
	return &(c.t.lines[c.l])
}

func (c *Cursor) LineData() string {
	return c.t.lines[c.l].data
}

func (c *Cursor) RuneAfter() (rune, int) {
	return utf8.DecodeRuneInString(c.LineData()[c.b:])
}

func (c *Cursor) RuneBefore() (rune, int) {
	return utf8.DecodeLastRuneInString(c.LineData()[:c.b])
}

func (c *Cursor) AtBol() bool{
	return c.b == 0
}

func (c *Cursor) AtEol() bool{
	return c.b == len(c.LineData())
}

func (c *Cursor) OnFirstLine() bool{
	return c.l == 0
}

func (c *Cursor) OnLastLine() bool {
	return c.l == len(c.t.lines)-1
}

func (c *Cursor) AtBow() bool {
	r, _ := c.RuneAfter()
	rb, _ := c.RuneBefore()
	if (unicode.IsLetter(r) || unicode.IsDigit(r))  && !(unicode.IsLetter(rb) || unicode.IsDigit(rb)) {
		return true
	}
	return false
}

func (c *Cursor) AtEow() bool {
	r, _ := c.RuneAfter()
	rb, _ := c.RuneBefore()
	if !(unicode.IsLetter(r) || unicode.IsDigit(r)) && (unicode.IsLetter(rb) || unicode.IsDigit(rb)) {
		return true
	}
	return false
}

func (c *Cursor) AtBof() bool {
	return c.OnFirstLine() && c.AtBol()
}

func (c *Cursor) AtEof() bool {
	return c.OnLastLine() && c.AtEol()
}

func (c *Cursor) InStrings() bool {
	instr := false
	var starter rune
	var old rune
	var oldold rune
	for _, r := range c.LineData()[:c.b] {
		if !instr && strings.ContainsAny(string(r), "'\"") && !(old == '\\' && oldold != '\\') {
			instr = true
			starter = r
		} else if instr && (r == starter) && !(old == '\\' && oldold != '\\') {
			instr = false
			starter = ' '
		}
		oldold = old
		old = r
	}
	return instr
}

func (c *Cursor) MoveLeft() {
	c.o = c.O()
	if c.AtBof() {
		return
	} else if c.AtBol() {
		c.l--
		c.SetB(len(c.LineData()))
		return
	}
	r, rlen := c.RuneBefore()
	c.b -= rlen
	c.o -= vlen(string(r))
}

func (c *Cursor) MoveRight() {
	c.o = c.O()
	if c.AtEof() {
		return
	} else if c.AtEol() {
		c.l++
		c.SetB(0)
		return
	}
	r, rlen := c.RuneAfter()
	c.b += rlen
	c.o += vlen(string(r))
}

func (c *Cursor) MoveUp() {
	if c.OnFirstLine() {
		return
	}
	c.l--
	c.RecalcB()
}

func (c *Cursor) MoveDown() {
	if c.OnLastLine() {
		return
	}
	c.l++
	c.RecalcB()
}

func (c *Cursor) MovePrevBowEow() {
	if c.AtBof() {
		return
	}
	if c.AtBow() {
		for {
			c.MoveLeft()
			if c.AtEow() || c.AtBof() {
				return
			}
		}
	} else if c.AtEow() || c.AtBof() {
		for {
			c.MoveLeft()
			if c.AtBow() {
				return
			}
		}
	} else {
		r, _ := c.RuneAfter()
		if !(unicode.IsLetter(r) || unicode.IsDigit(r)) {
			// we are in the middle of non-words. find eow.
			for {
				c.MoveLeft()
				if c.AtEow() || c.AtBof() {
					return
				}
			}
		} else {
			// we are in the middle of a word. find bow.
			for {
				c.MoveLeft()
				if c.AtBow() || c.AtBof() {
					return
				}
			}
		}
	}
}

func (c *Cursor) MoveNextBowEow() {
	if c.AtEof() {
		return
	}
	if c.AtBow() {
		for {
			c.MoveRight()
			if c.AtEow() || c.AtEof() {
				return
			}
		}
	} else if c.AtEow() {
		for {
			c.MoveRight()
			if c.AtBow() || c.AtEof() {
				return
			}
		}
	} else {
		r, _ := c.RuneAfter()
		if !(unicode.IsLetter(r) || unicode.IsDigit(r)) {
			// we are in the middle of non-words. find bow.
			for {
				c.MoveRight()
				if c.AtBow() || c.AtEof() {
					return
				}
			}
		} else {
			// we are in the middle of a word. find eow.
			for {
				c.MoveRight()
				if c.AtEow() || c.AtEof() {
					return
				}
			}
		}
	}
}

func (c *Cursor) MoveBol() {
	c.SetB(0)
}

func (c *Cursor) MoveBoc() {
	c.SetB(c.Line().Boc())
}

func (c *Cursor) MoveBocBolRepeat() {
	if c.AtBol() {
		c.MoveBoc()
	} else if c.b <= c.Line().Boc() {
		c.MoveBol()
	} else {
		c.MoveBoc()
	}
}

func (c *Cursor) MoveEol() {
	c.SetB(len(c.LineData()))
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
	c.l = 0
	c.b = 0
	c.o = 0
}

func (c *Cursor) MoveEof() {
	c.l = len(c.t.lines) - 1
	c.b = len(c.LineData())
	c.o = vlen(c.LineData())
}

func (c *Cursor) SplitLine() {
	c.t.SplitLine(c.l, c.b)
	c.MoveDown()
	c.SetB(0)
}

func (c *Cursor) Insert(str string) {
	for _, r := range str {
		if r == '\n' {
			c.SplitLine()
			continue
		}
		c.t.Insert(string(r), c.l, c.b)
		c.MoveRight()
	}
}

func (c *Cursor) Tab(sel *Selection) []int {
	tabed := make([]int, 0)
	if sel == nil {
		c.t.lines[c.l].InsertTab()
		tabed = append(tabed, c.l)
		c.SetB(c.b+1)
		return tabed
	}
	min, max := sel.MinMax()
	if min.b == len(c.t.lines[min.l].data) {
		min.l++
	}
	if max.b == 0 {
		max.l--
	}
	for l := min.l; l < max.l + 1; l++ {
		c.t.lines[l].InsertTab()
		tabed = append(tabed, l)
	}
	for _, l := range tabed {
		if l == c.l {
			c.SetB(c.b+1)
		}
	}
	return tabed
}

func (c *Cursor) UnTab(sel *Selection) []int {
	untabed := make([]int, 0)
	if sel == nil {
		if err := c.t.lines[c.l].RemoveTab(); err != nil {
			return untabed
		}
		if c.b != 0 {
			c.SetB(c.b-1)
		}
		untabed = append(untabed, c.l)
		return untabed
	}
	min, max := sel.MinMax()
	if min.b == len(c.t.lines[min.l].data) {
		min.l++
	}
	if max.b == 0 {
		max.l--
	}
	for l := min.l; l < max.l+1; l++ {
		if err := c.t.lines[l].RemoveTab(); err == nil {
			untabed = append(untabed, l)
		}
	}
	for _, l := range untabed {
		if l == c.l && !c.AtBol() {
			c.SetB(c.b-1)
		}
	}
	return untabed
}

func (c *Cursor) Delete() string {
	if c.AtEof() {
		return ""
	}
	if c.AtEol() {
		c.t.JoinNextLine(c.l)
		return "\n"
	}
	_, rlen := c.RuneAfter()
	return c.t.Remove(c.l, c.b, c.b+rlen)
}

func (c *Cursor) Backspace() string {
	if c.AtBof() {
		return ""
	}
	c.MoveLeft()
	return c.Delete()
}

func (c *Cursor) DeleteSelection(sel *Selection) string {
	min, max := sel.MinMax()
	bmin := Point{min.l, BFromO(c.t.lines[min.l].data, min.o)}
	bmax := Point{max.l, BFromO(c.t.lines[max.l].data, max.o)}
	deleted := c.t.RemoveRange(bmin, bmax)
	c.l = min.l
	c.SetB(bmin.o)
	return deleted
}

func (c *Cursor) GotoNext(find string) bool {
	if find == "" {
		return true
	}
	for l := c.l; l < len(c.t.lines); l++ {
		linedata := string(c.t.lines[l].data)
		offset := 0
		if l == c.l {
			if c.b == len(linedata) {
				continue
			}
			linedata = linedata[c.b+1:]
			offset = c.b+1
		}
		b := strings.Index(linedata, find)
		if b != -1 {
			c.l = l
			c.SetB(b+offset)
			return true
		}
	}
	return false
}

func (c *Cursor) GotoPrev(find string) bool {
	if find == "" {
		return true
	}
	for l := c.l; l >= 0; l-- {
		linedata := string(c.t.lines[l].data)
		if l == c.l {
			linedata = linedata[:c.b]
		}
		b := strings.LastIndex(linedata, find)
		if b != -1 {
			c.l = l
			c.SetB(b)
			return true
		}
	}
	return false
}

func (c *Cursor) GotoNextWord(find string) bool {
	oldc := *c
	for l := c.l; l < len(c.t.lines); l++ {
		linedata := string(c.t.lines[l].data)
		offset := 0
		if l == c.l {
			if c.b == len(linedata) {
				continue
			}
			linedata = linedata[c.b+1:]
			offset = c.b+1
		}
		b := strings.Index(linedata, find)
		if b != -1 {
			c.l = l
			c.SetB(b+offset)
			if c.Word() == find {
				return true
			}
		}
	}
	c.Copy(oldc)
	return false
}

func (c *Cursor) GotoPrevWord(find string) bool {
	oldc := *c
	for l := c.l; l >= 0; l-- {
		linedata := string(c.t.lines[l].data)
		if l == c.l {
			linedata = linedata[:c.b]
		}
		b := strings.LastIndex(linedata, find)
		if b != -1 {
			c.l = l
			c.SetB(b)
			if c.Word() == find {
				return true
			}
		}
	}
	c.Copy(oldc)
	return false
}

func (c *Cursor) GotoFirst(find string) bool {
	for l := 0; l < len(c.t.lines); l++ {
		linedata := string(c.t.lines[l].data)
		b := strings.Index(linedata, find)
		if b != -1 {
			c.l = l
			c.SetB(b)
			return true
		}
	}
	return false
}

func (c *Cursor) GotoLast(find string) bool {
	for l := len(c.t.lines)-1; l >= 0; l-- {
		linedata := string(c.t.lines[l].data)
		b := strings.LastIndex(linedata, find)
		if b != -1 {
			c.l = l
			c.SetB(b)
			return true
		}
	}
	return false
}

func (c *Cursor) GotoNextAny(chars string) bool {
	for l := c.l; l < len(c.t.lines); l++ {
		linedata := string(c.t.lines[l].data)
		offset := 0
		if l == c.l {
			if c.b == len(linedata) {
				continue
			}
			linedata = linedata[c.b+1:]
			offset = c.b+1
		}
		b := strings.IndexAny(linedata, chars)
		if b != -1 {
			c.l = l
			c.SetB(b+offset)
			return true
		}
	}
	return false
}

func (c *Cursor) GotoPrevAny(chars string) bool {
	for l := c.l; l >= 0; l-- {
		linedata := string(c.t.lines[l].data)
		if l == c.l {
			linedata = linedata[:c.b]
		}
		b := strings.LastIndexAny(linedata, chars)
		if b != -1 {
			c.l = l
			c.SetB(b)
			return true
		}
	}
	return false
}

func (c *Cursor) GotoNextGlobalLineWithout(str string) bool {
	findLine := -1
	for l := c.l+1; l < len(c.t.lines); l++ {
		find := true
		for _, r := range str {
			if c.t.lines[l].data == "" {
				find = false
			} else if strings.HasPrefix(c.t.lines[l].data, string(r)) {
				find = false
			}
		}
		if find {
			findLine = l
			break
		}
	}
	if findLine != -1 {
		c.l = findLine
		c.SetB(0)
		return true
	}
	return false
}

func (c *Cursor) GotoPrevGlobalLineWithout(str string) bool {
	var startLine int
	if c.b == 0 {
		startLine = c.l - 1
	} else {
		startLine = c.l
	}
	findLine := -1
	for l := startLine; l >= 0; l-- {
		find := true
		for _, r := range str {
			if c.t.lines[l].data == "" {
				find = false
			} else if strings.HasPrefix(c.t.lines[l].data, string(r)) {
				find = false
			}
		}
		if find {
			findLine = l
			break
		}
	}
	if findLine != -1 {
		c.l = findLine
		c.SetB(0)
		return true
	}
	return false
}

func (c *Cursor) GotoNextDefinition(defn []string) bool {
	nextLines := c.t.lines[c.l+1:]
	for i, line := range nextLines {
		l := c.l + 1 + i
		find := false
		for _, d := range defn {
			if strings.HasPrefix(string(line.data), d) {
				find = true
				break
			}
		}
		if find {
			c.l = l
			c.SetB(0)
			return true
		}
	}
	return false
}

func (c *Cursor) GotoPrevDefinition(defn []string) bool {
	var startLine int
	if c.b == 0 {
		startLine = c.l - 1
	} else {
		startLine = c.l
	}
	find := false
	for l := startLine; l >= 0; l-- {
		for _, d := range defn {
			if strings.HasPrefix(string(c.t.lines[l].data), d) {
				find = true
				break
			}
		}
		if find {
			c.l = l
			c.SetB(0)
			return true
		}
	}
	return false
}

func (c *Cursor) GotoMatchingBracket() bool {
	rb, _ := c.RuneBefore()
	ra, _ := c.RuneAfter()
	var r rune
	dir := ""
	if strings.Contains("{[(", string(rb)) {
		r = rb
		dir = "right"
	}
	if strings.Contains("}])", string(ra)) {
		r = ra
		dir = "left"
	}
	if dir == "" {
		return true
	}
	// rune for matching.
	var m rune
	switch r {
	case '{':
		m = '}'
	case '}':
		m = '{'
	case '[':
		m = ']'
	case ']':
		m = '['
	case '(':
		m = ')'
	case ')':
		m = '('
	}
	if dir == "left" && rb == m {
		return true
	} else if dir == "right" && ra == m {
		return true
	}
	set := string(r) + string(m)
	depth := 0
	origc := *c
	for {
		bc := *c
		if dir == "right" {
			c.GotoNextAny(set)
		} else {
			c.GotoPrevAny(set)
		}
		if c.l == bc.l && c.o == bc.o {
			// did not find next set.
			c.Copy(origc)
			return false
		}
		if c.InStrings() {
			continue
		}
		cr, _ := c.RuneAfter()
		if cr == r {
			depth++
		} else if cr == m {
			if depth == 0 {
				if dir == "left" {
					c.MoveRight()
				}
				return true
			}
			depth--
		}
	}
	return false
}

func (c *Cursor) GotoLine(l int) {
	if l >= len(c.t.lines) {
		l = len(c.t.lines)-1
	}
	c.l = l
	c.SetB(0)
}

func (c *Cursor) Word() string {
	// check cursor is on a word
	r, _ := c.RuneAfter()
	if !(unicode.IsLetter(r) || unicode.IsDigit(r)) {
		return ""
	}
	// find min byte offset
	bmin := c.b
	remain := c.LineData()[:c.b]
	for {
		if len(remain) == 0 {
			break
		}
		r, rlen := utf8.DecodeLastRuneInString(remain)
		remain = remain[:len(remain)-rlen]
		if !(unicode.IsLetter(r) || unicode.IsDigit(r)) {
			break
		}
		bmin -= rlen
	}
	// find max byte offset
	bmax := c.b
	remain = c.LineData()[c.b:]
	for {
		if len(remain) == 0 {
			break
		}
		r, rlen := utf8.DecodeRuneInString(remain)
		remain = remain[rlen:]
		if !(unicode.IsLetter(r) || unicode.IsDigit(r)) {
			break
		}
		bmax += rlen
	}
	return c.LineData()[bmin:bmax]
}
