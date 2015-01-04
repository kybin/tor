package main

import (
	//term "github.com/nsf/termbox-go"
)

// Viewer is a window that follows a cursor.
// It's size should same as tor's main layout size.
// It will used for text buffer clipping.
type viewer struct {
	min Point
	max Point
}

func newViewer(l *layout) *viewer {
	maxpt := l.mainViewerBound().Size()
	minpt := Point{0, 0}
	v := viewer{minpt, maxpt}
	return &v
}

func (v *viewer) set(min, max Point) {
	v.min = min
	v.max = max
}

func (v *viewer) size() (int, int) {
	size := v.max.Sub(v.min)
	return size.o, size.l
}

func (v *viewer) move(t Point) {
	v.min = v.min.Add(t)
	v.max = v.max.Add(t)
}

func (v *viewer) cursorInViewer(c *cursor) bool {
	cl, co := c.position()

	if (v.min.o <= co && co < v.max.o) && (v.min.l <= cl && cl < v.max.l) {
		return true
	}
	return false
}

func (v *viewer) moveToCursor(c *cursor) {
		cl, co := c.position()
		tl, to := 0, 0

		mino := v.min.o
		minl := v.min.l
		// cursorMax = viewMax - 1
		maxo := v.max.o-1
		maxl := v.max.l-1

		if co < mino {
			to = co - mino
		} else if co > maxo {
			to = co - maxo
		}
		if cl < minl {
			tl = cl - minl
		} else if cl > maxl {
			tl = cl - maxl
		}
		v.move(Point{tl, to})
}
