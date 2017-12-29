// cell provides basic cell geometries.
package cell

// Pt is a line oriented cell position.
type Pt struct {
	L int // line
	O int // offset
}

// Add moves a point to direction of q.
func (p Pt) Add(q Pt) Pt {
	return Pt{p.L + q.L, p.O + q.O}
}

// Add moves a point to opposite direction of q.
func (p Pt) Sub(q Pt) Pt {
	return Pt{p.L - q.L, p.O - q.O}
}

// Compare compares two Pt a and b.
// If a < b, it will return -1.
// If a > b, it will return 1.
// If a == b, it will return 0
func (a Pt) Compare(b Pt) int {
	if a.L < b.L {
		return -1
	}
	if a.L == b.L {
		if a.O < b.O {
			return -1
		}
		if a.O == b.O {
			return 0
		}
		return 1
	}
	return 1
}

// Rect is a rectangle area which exculdes Max.
//
// Start is not necessarily before than End.
// For that purpose, use MinMax.
type Rect struct {
	Start Pt
	End   Pt
}

// MinMax returns the Rect's caculated Min and Max positions.
func (r Rect) MinMax() (Pt, Pt) {
	minl := r.Start.L
	maxl := r.End.L
	if minl > maxl {
		minl, maxl = maxl, minl
	}
	mino := r.Start.O
	maxo := r.End.O
	if mino > maxo {
		mino, maxo = maxo, mino
	}
	return Pt{minl, mino}, Pt{maxl, maxo}
}

// Min returns the Rect's top left position.
func (r Rect) Min() Pt {
	min, _ := r.MinMax()
	return min
}

// Max returns the Rect's bottom right position.
func (r Rect) Max() Pt {
	_, max := r.MinMax()
	return max
}

// Size returns size of the Rect as a point.
func (a Rect) Size() Pt {
	min, max := a.MinMax()
	return Pt{max.L - min.L, max.O - min.O}
}

// Range is a range between two points which excludes max.
//
// Start is not necessarily before than End.
// For that purpose, use MinMax.
type Range struct {
	Start Pt
	End   Pt
}

// MinMax return the Range's caculated Min and Max points.
func (r *Range) MinMax() (Pt, Pt) {
	if (r.Start.L > r.End.L) || (r.Start.L == r.End.L && r.Start.O > r.End.O) {
		return r.End, r.Start
	}
	return r.Start, r.End
}

// Min returns the Range's minimum position.
func (r Range) Min() Pt {
	min, _ := r.MinMax()
	return min
}

// Max returns the Ranges's maximum position.
func (r Range) Max() Pt {
	_, max := r.MinMax()
	return max
}

// Lines returns line numbers the Range includes, as int slice.
// Note it will not return last line number if Max's offset is 0.
func (r Range) Lines() []int {
	min, max := r.MinMax()

	maxl := max.L
	if max.O == 0 {
		maxl--
	}

	lns := make([]int, 0)
	for l := min.L; l <= maxl; l++ {
		lns = append(lns, l)
	}
	return lns
}

// Contains checks whether it contain's p.
func (r Range) Contains(p Pt) bool {
	min, max := r.MinMax()
	if min.L <= p.L && p.L <= max.L {
		if p.L == min.L && p.O < min.O {
			return false
		} else if p.L == max.L && p.O >= max.O {
			return false
		}
		return true
	}
	return false
}
