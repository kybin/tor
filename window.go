package main

import (
	//term "github.com/nsf/termbox-go"
)

// Window is a area that follows a cursor.
// It used for clipping text.
// It's size should same as tor's main layout size.
type window struct {
	min Point
	max Point
}

func NewWindow(l *layout) *window {
	maxpt := l.mainViewerBound().Size()
	minpt := Point{0, 0}
	w := window{minpt, maxpt}
	return &w
}

func (w *window) Set(min, max Point) {
	w.min = min
	w.max = max
}

func (w *window) Size() (int, int) {
	size := w.max.Sub(w.min)
	return size.o, size.l
}

func (w *window) Move(t Point) {
	w.min = w.min.Add(t)
	w.max = w.max.Add(t)
}

func (w *window) Contains(c *cursor) bool {
	cl, co := c.position()

	if (w.min.o <= co && co < w.max.o) && (w.min.l <= cl && cl < w.max.l) {
		return true
	}
	return false
}

func (w *window) Follow(c *cursor) {
		var tl, to int

		cl, co := c.position()

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
