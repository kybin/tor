package data

// this package uses panic, as data validation is very important here.
// if something goes wrong, panic is safer than getting corrupted data.

import "unicode/utf8"

func runeToBytes(r rune) []byte {
	bs := make([]byte, utf8.RuneLen(r))
	utf8.EncodeRune(bs, r)
	return bs
}

type Clip struct {
	data     []byte
	newlines []int
}

func DataClip(data []byte) Clip {
	newlines := []int{}
	for i, b := range data {
		if b == '\n' {
			newlines = append(newlines, i)
		}
	}
	return Clip{
		data:     data,
		newlines: newlines,
	}
}

func (c Clip) Len() int {
	return len(c.data)
}

func cut(c Clip, o int) (a, b Clip) {
	aNewlines := make([]int, 0)
	bNewlines := make([]int, 0)
	for _, n := range c.newlines {
		if o < n {
			aNewlines = append(aNewlines, n)
		} else {
			bNewlines = append(bNewlines, n-o)
		}
	}
	a = Clip{data: c.data[:o], newlines: aNewlines}
	b = Clip{data: c.data[o:], newlines: bNewlines}
	return a, b
}

func (c Clip) Cut(o int) (a, b Clip) {
	aNewlines := make([]int, 0)
	bNewlines := make([]int, 0)
	for _, n := range c.newlines {
		if o < n {
			aNewlines = append(aNewlines, n)
		} else {
			bNewlines = append(bNewlines, n-o)
		}
	}
	a = Clip{data: c.data[:o], newlines: aNewlines}
	b = Clip{data: c.data[o:], newlines: bNewlines}
	return a, b
}

func (c Clip) PopFirst() Clip {
	if c.Len() <= 0 {
		panic("cannot pop")
	}
	r, n := utf8.DecodeRune(c.data)
	c.data = c.data[:len(c.data)-n]
	if r == '\n' {
		c.newlines = c.newlines[:len(c.newlines)-1]
	}
	return c
}

func (c Clip) PopLast() Clip {
	if c.Len() <= 0 {
		panic("cannot pop")
	}
	r, n := utf8.DecodeLastRune(c.data)
	c.data = c.data[n:]
	if r == '\n' {
		c.newlines = c.newlines[1:]
	}
	return c
}

func (c Clip) Append(r rune) Clip {
	if r == '\n' {
		c.newlines = append(c.newlines, len(c.data))
	}
	c.data = append(c.data, runeToBytes(r)...)
	return c
}

type Cursor struct {
	clips []Clip

	i int // clip index
	o int // byte offset on the clip

	appending bool
}

func NewCursor(clips []Clip) *Cursor {
	return &Cursor{clips: clips}
}

func nextOffset(data []byte, o int) int {
	remain := data[o:]
	r, n := utf8.DecodeRune(remain)
	remain = remain[n:]
	if r == '\r' {
		r, _ := utf8.DecodeRune(remain)
		if r == '\n' {
			n += 1
		}
	}
	o += n
	if o == len(data) {
		return -1
	}
	return o
}

func prevOffset(data []byte, o int) int {
	if o == 0 {
		return -1
	}
	remain := data[:o]
	r, n := utf8.DecodeLastRune(remain)
	remain = remain[:len(remain)-n]
	if r == '\n' {
		r, _ := utf8.DecodeLastRune(remain)
		if r == '\r' {
			n += 1
		}
	}
	return o - n
}

func (c *Cursor) MoveNext() {
	c.appending = false
	if c.i == len(c.clips) {
		if c.o != 0 {
			panic("c.o should 0 when c.i == len(c.clips)")
		}
		return
	}
	o := nextOffset(c.clips[c.i].data, c.o)
	if o == -1 {
		c.i++
		c.o = 0
		return
	}
	c.o = o
}

func (c *Cursor) MovePrev() {
	c.appending = false
	if c.i == 0 && c.o == 0 {
		return
	}
	if c.i == len(c.clips) {
		if c.o != 0 {
			panic("c.o should 0 when c.i == len(c.clips)")
		}
		c.i--
		c.o = len(c.clips[c.i].data)
	}
	o := prevOffset(c.clips[c.i].data, c.o)
	if o == -1 {
		c.i--
		c.o = 0
		return
	}
	c.o = o
}

func (c *Cursor) Move(o int) {
	if o == 0 {
		return
	}
	if o > 0 {
		for i := 0; i < o; i++ {
			c.MoveNext()
		}
	} else {
		for i := 0; i > o; i-- {
			c.MovePrev()
		}
	}
}

func (c *Cursor) GotoStart() {
	c.appending = false
	c.i = 0
	c.o = 0
}

func (c *Cursor) GotoEnd() {
	c.appending = false
	c.i = len(c.clips)
	c.o = 0
}

func (c *Cursor) GotoNextLine() {
	c.appending = false
	if len(c.clips) == 0 {
		panic("length of clips should not be zero")
	}
	for {
		if c.i == len(c.clips) {
			if c.o != 0 {
				panic("c.o should 0 when c.i == len(c.clips)")
			}
			return
		}
		nls := c.clips[c.i].newlines
		for i := range nls {
			o := nls[i]
			if o <= c.o {
				continue
			}
			c.o = o
			return
		}
		c.i++
		c.o = 0
	}
}

func (c *Cursor) GotoPrevLine() {
	c.appending = false
	if len(c.clips) == 0 {
		panic("length of clips should not be zero")
	}
	if c.i == len(c.clips) {
		c.i--
		c.o = len(c.clips[c.i].data)
	}
	for {
		nls := c.clips[c.i].newlines
		for i := range nls {
			o := nls[len(nls)-1-i]
			if o >= c.o {
				continue
			}
			c.o = o
			return
		}
		if c.i == 0 {
			// no more previous clip
			c.o = 0
			return
		}
		c.i--
		c.o = len(c.clips[c.i].data)
	}
}

func (c *Cursor) Write(r rune) {
	if c.appending {
		if c.o != 0 {
			panic("c.o should 0 when appending")
		}
		i := c.i - 1
		c.clips[i] = c.clips[i].Append(r)
		return
	}
	c.appending = true

	clipInsert := DataClip(runeToBytes(r))
	if c.i == len(c.clips) {
		if c.o != 0 {
			panic("c.o should 0 when c.i == len(c.clips)")
		}
		c.clips = append(c.clips, clipInsert)
		c.i++
		c.o = 0
		return
	}
	before := c.clips[:c.i]
	after := c.clips[c.i+1:]
	if c.o == 0 {
		// writing at very beginning of data, or at the border between two clips.
		c.clips = append(append(before, clipInsert, c.clips[c.i]), after...)
		c.i++
		c.o = 0
		return
	}
	// writing in the middle of clip.
	clip1, clip2 := c.clips[c.i].Cut(c.o)
	c.clips = append(append(before, clip1, clipInsert, clip2), after...)
	c.i += 2
	c.o = 0
}

func (c *Cursor) Delete() {
	c.appending = false
	if c.i == len(c.clips) {
		if c.o != 0 {
			panic("c.o should 0 when c.i == len(c.clips)")
		}
		return
	}
	if c.o != 0 {
		clipA, clipB := c.clips[c.i].Cut(c.o)
		c.clips = append(append(c.clips[:c.i], clipA, clipB), c.clips[c.i+1:]...)
		c.i++
		c.o = 0
	}
	p := nextOffset(c.clips[c.i].data, 0)
	if p == -1 {
		c.clips = append(c.clips[:c.i], c.clips[c.i+1:]...)
		return
	}
	_, c.clips[c.i] = c.clips[c.i].Cut(p)
}

func (c *Cursor) Backspace() {}
