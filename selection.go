package main

import "github.com/kybin/tor/cell"

type Selection struct {
	on  bool
	rng cell.Range

	text *Text
}

func NewSelection(text *Text) *Selection {
	return &Selection{text: text}
}

func (s *Selection) SetStart(p cell.Pt) {
	s.rng.Start = p
}

func (s *Selection) SetEnd(p cell.Pt) {
	s.rng.End = p
}

// Lines return selected line numbers as int slice.
// Note it will not return last line number if last cursor's offset is 0.
func (s *Selection) Lines() []int {
	if !s.on {
		return nil
	}
	return s.rng.Lines()
}

func (s *Selection) Min() cell.Pt {
	return s.rng.Min()
}

func (s *Selection) Max() cell.Pt {
	return s.rng.Max()
}

func (s *Selection) MinMax() (cell.Pt, cell.Pt) {
	return s.rng.MinMax()
}

func (s *Selection) Contains(p cell.Pt) bool {
	return s.rng.Contains(p)
}

func (s *Selection) Data() string {
	if !s.on {
		return ""
	}
	return s.text.DataInside(s.MinMax())
}
