package main

import (
	"github.com/kybin/tor/cell"
	"github.com/kybin/tor/syntax"
)

// Area is an area of screen.
// An Area has it's matching Window that is same size.
type Area struct {
	min  cell.Pt
	size cell.Pt
	Win  *Window
}

// NewArea creates a new Area.
func NewArea(min cell.Pt, size cell.Pt) *Area {
	a := &Area{
		min:  min,
		size: size,
		Win:  NewWindow(size),
	}
	return a
}

func (a *Area) Set(min cell.Pt, size cell.Pt) {
	a.min = min
	a.size = size
	a.Win.size = size
}

// Resize resizes it and it's window size.
func (a *Area) Resize(size cell.Pt) {
	a.size = size
	a.Win.size = size
}

func (a *Area) Draw(tx Text, sel *Selection, matches []syntax.Match) {}

func (a *Area) DrawCursor(c Cursor) {}
