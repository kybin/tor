package main

import (
//term "github.com/nsf/termbox-go"
)

// Window is a area that follows a cursor.
// It used for clipping text.
// It's size should same as tor's main layout size.
type Window struct {
	min Point
	max Point
}

func NewWindow(size Point) *Window {
	w := Window{Point{0, 0}, size}
	return &w
}

func (w *Window) Resize(size Point) {
	w.max = Point{w.min.l + size.l, w.min.o + size.o}
}

func (w *Window) Set(min, max Point) {
	w.min = min
	w.max = max
}

func (w *Window) Size() (int, int) {
	size := w.max.Sub(w.min)
	return size.o, size.l
}

func (w *Window) Move(t Point) {
	w.min = w.min.Add(t)
	w.max = w.max.Add(t)
}

func (w *Window) Contains(c *Cursor) bool {
	cp := c.Position()
	if (w.min.o <= cp.o && cp.o < w.max.o) && (w.min.l <= cp.l && cp.l < w.max.l) {
		return true
	}
	return false
}

func (w *Window) Follow(c *Cursor, margin int) {
	var tl, to int
	cp := c.Position()

	minl := w.min.l + margin
	maxl := w.max.l - margin
	if cp.l < minl {
		tl = cp.l - minl
	} else if cp.l >= maxl {
		tl = cp.l - maxl + 1
	}
	// tl should not smaller than -w.min.l
	if tl < -w.min.l {
		tl = -w.min.l
	}

	mino := w.min.o + margin
	maxo := w.max.o - margin
	if cp.o < mino {
		to = cp.o - mino
	} else if cp.o >= maxo {
		to = cp.o - maxo + 1
	}
	// to should not smaller than -w.min.o
	if to < -w.min.o {
		to = -w.min.o
	}
	w.Move(Point{tl, to})
}
