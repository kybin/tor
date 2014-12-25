package main

import (
	"image"
)

type selection struct {
	on bool
	start image.Point
	end image.Point
}

func NewSelection() *selection {
	return &selection{}
}

func (s *selection) SetStart(c *cursor) {
	s.start = image.Point{c.offset(), c.line}
}

func (s *selection) SetEnd(c *cursor) {
	s.end = image.Point{c.offset(), c.line}
}

func withShift(ch rune) bool {
	shifts := "QWERTYUIOP{}|ASDFGHJKL:ZXCVBNM<>?!@#$%^&*()_+"
	for _, sch := range shifts {
		if ch == sch {
			return true
		}
	}
	return false
}
