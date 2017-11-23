package main

type Selection struct {
	on  bool
	rng Range

	text *Text
}

func (s *Selection) SetStart(p Point) {
	s.rng.SetStart(p)
}

func (s *Selection) SetEnd(p Point) {
	s.rng.SetEnd(p)
}

// Lines return selected line numbers as int slice.
// Note it will not return last line number if last cursor's offset is 0.
func (s *Selection) Lines() []int {
	if !s.on {
		return nil
	}
	return s.rng.Lines()
}

func (s *Selection) Min() Point {
	return s.rng.Min()
}

func (s *Selection) Max() Point {
	return s.rng.Max()
}

func (s *Selection) MinMax() (Point, Point) {
	return s.rng.MinMax()
}

func (s *Selection) Contains(p Point) bool {
	return s.rng.Contains(p)
}

func (s *Selection) Data() string {
	if !s.on {
		return ""
	}
	return s.text.DataInside(s.MinMax())
}
