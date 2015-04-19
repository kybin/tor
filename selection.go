package main

type Selection struct {
	on bool
	start Cursor
	end Cursor
}

func NewSelection() *Selection {
	return &Selection{}
}

func (s *Selection) SetStart(c *Cursor) {
	s.start = *c
}

func (s *Selection) SetEnd(c *Cursor) {
	s.end = *c
}

func (s *Selection) MinMax() (Cursor, Cursor) {
	if (s.start.l > s.end.l) || (s.start.l == s.end.l && s.start.o > s.end.o) {
		return s.end, s.start
	}
	return s.start, s.end
}

func (s *Selection) Contains(p Point) bool {
	min, max := s.MinMax()
	if min.l <= p.l && p.l <= max.l {
		if p.l == min.l && p.o < min.o {
			return false
		} else if p.l == max.l && p.o >= max.o {
			return false
		}
		return true
	}
	return false
}


func withShift(r rune) bool {
	shifts := "QWERTYUIOP{}|ASDFGHJKL:ZXCVBNM<>?!@#$%^&*()_+"
	for _, s := range shifts {
		if s == r {
			return true
		}
	}
	return false
}
