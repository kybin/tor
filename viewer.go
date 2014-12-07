package main

import (
	"image"
	//term "github.com/nsf/termbox-go"
)

// Viewer is a window that follows a cursor.
// It's size should same as tor's main layout size.
// It will used for text buffer clipping.
type viewer struct {
	min image.Point
	max image.Point
}

func newViewer(l *layout) *viewer {
	maxpt := l.mainViewerSize()
	minpt := image.Pt(0, 0)
	v := viewer{minpt, maxpt}
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
	cy, cx := c.position()

	if (v.min.X <= cx && cx < v.max.X) && (v.min.Y <= cy && cy < v.max.Y) {
		return true
	}
	return false
}

func (v *viewer) moveToCursor(c *cursor) {
		cy, cx := c.position()
		tx, ty := 0, 0

		minx := v.min.X
		miny := v.min.Y
		// cursorMax = viewMax - 1
		maxx := v.max.X-1
		maxy := v.max.Y-1

		if cx < minx {
			tx = cx - minx
		} else if cx > maxx {
			tx = cx - maxx
		}
		if cy < miny {
			ty = cy - miny
		} else if cy > maxy {
			ty = cy - maxy
		}
		v.move(image.Pt(tx, ty))
}
