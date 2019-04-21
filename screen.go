package main

import "github.com/kybin/tor/cell"

// Screen is terminal screen.
type Screen struct {
	size   cell.Pt
	main   *Area // where editing is done.
	status *Area // where status is shown.
}

// mainAreaOffset is offset for main area.
// The area's content will look like centered if it's width is 80.
func mainAreaOffset(screenWidth int) int {
	center := screenWidth / 2
	width := 80
	left := center - width/2
	if left < 0 {
		left = 0
	}
	return left
}

// NewScreen creates a new Screen.
func NewScreen(size cell.Pt) *Screen {
	off := mainAreaOffset(size.O)
	s := &Screen{
		size:   size,
		main:   NewArea(cell.Pt{0, off}, cell.Pt{size.L - 1, size.O - off}),
		status: NewArea(cell.Pt{size.L - 1, 0}, cell.Pt{1, size.O}),
	}
	return s
}

// Resize resizes it and it's sub areas.
func (s *Screen) Resize(size cell.Pt) {
	s.size = size
	off := mainAreaOffset(size.O)
	s.main.Set(cell.Pt{0, off}, cell.Pt{size.L - 1, size.O - off})
	s.status.Set(cell.Pt{size.L - 1, 0}, cell.Pt{1, size.O})
}
