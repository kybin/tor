package main

type selection struct {
	on bool
	start Point
	end Point
}

func NewSelection() *selection {
	return &selection{}
}

func (s *selection) SetStart(c *cursor) {
	s.start = Point{c.line, c.offset()}
}

func (s *selection) SetEnd(c *cursor) {
	s.end = Point{c.line, c.offset()}
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
