package main

type Selection struct {
	on bool
	start Point
	end Point
}

func NewSelection() *Selection {
	return &Selection{}
}

func (s *Selection) SetStart(c *Cursor) {
	s.start = c.Position()
}

func (s *Selection) SetEnd(c *Cursor) {
	s.end = c.Position()
}

func (s *Selection) MinMax() (Point, Point) {
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


func withShift(ch rune) bool {
	shifts := "QWERTYUIOP{}|ASDFGHJKL:ZXCVBNM<>?!@#$%^&*()_+"
	for _, sch := range shifts {
		if ch == sch {
			return true
		}
	}
	return false
}
