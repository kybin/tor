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

func NewWindow(l *Layout) *Window {
	maxpt := l.MainViewerBound().Size()
	minpt := Point{0, 0}
	w := Window{minpt, maxpt}
	return &w
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

func (w *Window) Follow(c *Cursor) {
		var tl, to int

		cp := c.Position()

		// mino := w.min.o
		// minl := w.min.l
		// maxo := w.max.o-1
		// maxl := w.max.l-1

		if cp.o < w.min.o {
			to = cp.o - w.min.o
		} else if cp.o >= w.max.o {
			to = cp.o - w.max.o + 1
		}
		if cp.l < w.min.l {
			tl = cp.l - w.min.l
		} else if cp.l >= w.max.l {
			tl = cp.l - w.max.l + 1
		}
		w.Move(Point{tl, to})
}
