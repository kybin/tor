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

func Clips(datas ...[]byte) []Clip {
	clips := make([]Clip, 0)
	for _, data := range datas {
		c := DataClip(data)
		clips = append(clips, c)
	}
	return clips
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

// next returns bytes for the next rune(s) of cursor.
// Note: It could return "\r\n" if it exists in data.
func (c *Cursor) next() []byte {
	data := c.clips[c.i].data[c.o:]
	r, n := utf8.DecodeRune(data)
	if n == 0 {
		panic("data empty")
	}
	if n == -1 {
		panic("rune error")
	}
	if r == '\r' {
		r, _ := utf8.DecodeRune(data[n:])
		if r == '\n' {
			n++
		}
	}
	return data[:n]
}

// prev returns bytes for the prev rune(s) of cursor.
// Note: It could return "\r\n" if it exists in data.
func (c *Cursor) prev() []byte {
	data := c.clips[c.i].data[:c.o]
	r, n := utf8.DecodeLastRune(data)
	if n == 0 {
		panic("data empty")
	}
	if n == -1 {
		panic("rune error")
	}
	if r == '\n' {
		r, _ := utf8.DecodeLastRune(data[:len(data)-n])
		if r == '\r' {
			n++
		}
	}
	return data[len(data)-n:]
}

func (c *Cursor) MoveNext() {
	c.appending = false
	if c.i == len(c.clips) && c.o == 0 {
		return
	}
	next := c.next()
	c.o += len(next)
	if c.o == len(c.clips[c.i].data) {
		c.i++
		c.o = 0
	}
}

func (c *Cursor) MovePrev() {
	c.appending = false
	if c.i == 0 && c.o == 0 {
		return
	}
	if c.o == 0 {
		c.i--
		c.o = len(c.clips[c.i].data)
	}
	prev := c.prev()
	c.o -= len(prev)
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

// Cut cuts the underlying clip it stands.
// If it is standing at edge of clip(s), it doesn't do anything.
func (c *Cursor) Cut() {
	if c.i == len(c.clips) {
		if c.o != 0 {
			panic("c.o should 0 when c.i == len(c.clips)")
		}
		return
	}
	if c.o == 0 {
		// edge
		return
	}
	clipA, clipB := c.clips[c.i].Cut(c.o)
	c.clips = append(c.clips[:c.i], append([]Clip{clipA, clipB}, c.clips[c.i+1:]...)...)
	c.i++
	c.o = 0
}

func (c *Cursor) Write(r rune) {
	if c.appending {
		if c.o != 0 {
			panic("c.o should 0 when appending")
		}
		c.clips[c.i-1] = c.clips[c.i-1].Append(r)
		return
	}
	c.appending = true
	c.Cut()
	clipInsert := DataClip(runeToBytes(r))
	c.clips = append(c.clips[:c.i], append([]Clip{clipInsert}, c.clips[c.i:]...)...)
	c.i++
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
	c.Cut()
	c.MoveNext()
	c.Cut()
	c.clips = append(c.clips[:c.i-1], c.clips[c.i:]...)
	c.i--
}

func (c *Cursor) Backspace() {
	c.appending = false
	if c.i == 0 && c.o == 0 {
		return
	}
	c.Cut()
	c.MovePrev()
	c.Cut()
	c.clips = append(c.clips[:c.i], c.clips[c.i+1:]...)
}
