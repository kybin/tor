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
	cl, co := c.Position()

	if (w.min.o <= co && co < w.max.o) && (w.min.l <= cl && cl < w.max.l) {
		return true
	}
	return false
}

func (w *Window) Follow(c *Cursor) {
		var tl, to int

		cl, co := c.Position()

		// mino := w.min.o
		// minl := w.min.l
		// maxo := w.max.o-1
		// maxl := w.max.l-1

		if co < w.min.o {
			to = co - w.min.o
		} else if co >= w.max.o {
			to = co - w.max.o + 1
		}
		if cl < w.min.l {
			tl = cl - w.min.l
		} else if cl >= w.max.l {
			tl = cl - w.max.l + 1
		}
		w.Move(Point{tl, to})
}
