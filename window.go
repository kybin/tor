package main

import "github.com/kybin/tor/cell"

// Window is a area that follows a cursor.
// It used for clipping text.
// It's size should same as tor's main layout size.
type Window struct {
	min  cell.Pt
	size cell.Pt
}

func NewWindow(size cell.Pt) *Window {
	w := Window{cell.Pt{0, 0}, size}
	return &w
}

func (w *Window) Min() cell.Pt {
	return w.min
}

func (w *Window) Max() cell.Pt {
	return w.min.Add(w.size)
}

func (w *Window) Move(t cell.Pt) {
	w.min = w.min.Add(t)
}

func (w *Window) Contains(c *Cursor) bool {
	cp := c.Position()
	if (w.Min().O <= cp.O && cp.O < w.Max().O) && (w.Min().L <= cp.L && cp.L < w.Max().L) {
		return true
	}
	return false
}

// Follow makes Window follows to Cursor c.
// It returns true if Window is really moved, or false.
func (w *Window) Follow(c *Cursor, margin int) bool {
	var tl, to int
	cp := c.Position()

	minl := w.Min().L + margin
	maxl := w.Max().L - margin
	if cp.L < minl {
		tl = cp.L - minl
	} else if cp.L >= maxl {
		tl = cp.L - maxl + 1
	}
	// tl should not smaller than -w.Min().l
	if tl < -w.Min().L {
		tl = -w.Min().L
	}

	mino := w.Min().O + margin
	maxo := w.Max().O - margin
	if cp.O < mino {
		to = cp.O - mino
	} else if cp.O >= maxo {
		to = cp.O - maxo + 1
	}
	// to should not smaller than -w.Min().o
	if to < -w.Min().O {
		to = -w.Min().O
	}

	if tl == 0 && to == 0 {
		return false
	}
	w.Move(cell.Pt{tl, to})
	return true
}
