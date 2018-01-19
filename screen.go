package main

import "github.com/kybin/tor/cell"

// Screen is terminal screen.
type Screen struct {
	size   cell.Pt
	main   *Area // where editing is done.
	status *Area // where status is shown.
}

// NewScreen creates a new Screen.
func NewScreen(size cell.Pt) *Screen {
	s := &Screen{size: size}
	s.Resize(size)
	return s
}

// Resize resizes it and it's sub areas.
func (s *Screen) Resize(size cell.Pt) {
	s.main = NewArea(cell.Pt{0, 0}, cell.Pt{size.L - 1, size.O})
	s.status = NewArea(cell.Pt{size.L - 1, 0}, cell.Pt{1, size.O})
}
