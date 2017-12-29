package cell

import "testing"

func TestPtCompare(t *testing.T) {
	cases := []struct {
		a    Pt
		b    Pt
		want int
	}{
		{
			a:    Pt{0, 0},
			b:    Pt{0, 0},
			want: 0,
		},
		{
			a:    Pt{0, 1},
			b:    Pt{0, 2},
			want: -1,
		},
		{
			a:    Pt{0, 2},
			b:    Pt{0, 1},
			want: 1,
		},
		{
			a:    Pt{0, 1},
			b:    Pt{1, 0},
			want: -1,
		},
		{
			a:    Pt{1, 0},
			b:    Pt{0, 1},
			want: 1,
		},
	}
	for _, c := range cases {
		got := c.a.Compare(c.b)
		if got != c.want {
			t.Fatalf("(%v).Compare(%v): got %v, want %v", c.a, c.b, got, c.want)
		}
	}
}
