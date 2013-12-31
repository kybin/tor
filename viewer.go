package main

import (
	"image"
	term "github.com/nsf/termbox-go"
)

type viewer struct {
	min image.Point
	max image.Point
}

func newViewer() *viewer {
	v := viewer{image.Pt(0,0), image.Pt(term.Size())}
	return &v
}

func (v *viewer) set(min, max image.Point) {
	v.min = min
	v.max = max
}

func (v *viewer) size() (int, int) {
	size := v.max.Sub(v.min)
	return size.X, size.Y
}

func (v *viewer) move(t image.Point) {
	v.min = v.min.Add(t)
	v.max = v.max.Add(t)
}

func (v *viewer) cursorInViewer(c *cursor) bool {
	cx := c.cursorOffset()
	cy := c.linenum

	if (v.min.X <= cx && cx <= v.max.X) && (v.min.Y <= cy && cy <= v.max.Y) {
		return true
	}
	return false
}

func (v *viewer) moveToCursor(c *cursor) {
	if !v.cursorInViewer(c) {
		cx := c.cursorOffset()
		cy := c.linenum
		tx, ty := 0, 0
		if cx < v.min.X {
			tx = cx - v.min.X
		} else if cx > v.max.X {
			tx = cx - v.max.X
		}
		if cy < v.min.Y {
			ty = cy - v.min.Y
		} else if cy > v.max.Y {
			ty = cy - v.max.Y
		}
		v.move(image.Pt(tx, ty))
	}
}
