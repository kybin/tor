package main

type Selection struct {
	on    bool
	start Cursor
	end   Cursor

	m *NormalMode
}

func (s *Selection) SetStart(c *Cursor) {
	s.start = *c
}

func (s *Selection) SetEnd(c *Cursor) {
	s.end = *c
}

// Lines return selected line numbers as int slice.
// Note it will not return last line number if last cursor's offset is 0.
func (s *Selection) Lines() []int {
	if !s.on {
		return nil
	}
	start, end := s.MinMax()

	endL := end.l
	if s.end.o == 0 {
		endL--
	}

	lns := make([]int, 0)
	for l := start.l; l <= endL; l++ {
		lns = append(lns, l)
	}
	return lns
}

func (s *Selection) Min() Cursor {
	if (s.start.l > s.end.l) || (s.start.l == s.end.l && s.start.b > s.end.b) {
		return s.end
	}
	return s.start
}

func (s *Selection) Max() Cursor {
	if (s.start.l > s.end.l) || (s.start.l == s.end.l && s.start.b > s.end.b) {
		return s.start
	}
	return s.end
}

func (s *Selection) MinMax() (Cursor, Cursor) {
	if (s.start.l > s.end.l) || (s.start.l == s.end.l && s.start.b > s.end.b) {
		return s.end, s.start
	}
	return s.start, s.end
}

func (s *Selection) Data() string {
	if !s.on {
		return ""
	}
	return s.m.text.DataInside(s.MinMax())
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
