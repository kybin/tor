package main

type Range struct {
	start Point
	end   Point
}

func (r *Range) SetStart(p Point) {
	r.start = p
}

func (r *Range) SetEnd(p Point) {
	r.end = p
}

// Lines return selected line numbers as int slice.
// Note it will not return last line number if last cursor's offset is 0.
func (r *Range) Lines() []int {
	start, end := r.MinMax()

	endL := end.l
	if r.end.o == 0 {
		endL--
	}

	lns := make([]int, 0)
	for l := start.l; l <= endL; l++ {
		lns = append(lns, l)
	}
	return lns
}

func (r *Range) Min() Point {
	if (r.start.l > r.end.l) || (r.start.l == r.end.l && r.start.o > r.end.o) {
		return r.end
	}
	return r.start
}

func (r *Range) Max() Point {
	if (r.start.l > r.end.l) || (r.start.l == r.end.l && r.start.o > r.end.o) {
		return r.start
	}
	return r.end
}

func (r *Range) MinMax() (Point, Point) {
	if (r.start.l > r.end.l) || (r.start.l == r.end.l && r.start.o > r.end.o) {
		return r.end, r.start
	}
	return r.start, r.end
}

func (r *Range) Contains(p Point) bool {
	min, max := r.MinMax()
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
