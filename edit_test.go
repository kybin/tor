package main

import (
	"testing"
)

func TestRemoveRange(t *testing.T) {
	cases := []struct{
		in Text
		min, max Point
		want Text
	}{
		{
			Text{[]Line{
				Line{"Hello, my name is yongbin."},
				Line{"This is the test string."},
				Line{"You are great."},
			}},
			Point{0, 6}, Point{2, 7},
			Text{[]Line{
				Line{"Hello, great."},
			}},
		},
		{
			Text{[]Line{
				Line{"blizzard"},
				Line{"	wow"},
				Line{"	Diablo"},
			}},
			Point{0, 0}, Point{1, 0},
			Text{[]Line{
				Line{"	wow"},
				Line{"	Diablo"},
			}},
		},
		{
			Text{[]Line{
				Line{"The delete built-in function"},
				Line{"deletes the element"},
				Line{"with the specified key (m[key]) from the map."},
				Line{"If m is nil or there is no such element,"},
				Line{"delete is a no-op."},
			}},
			Point{0, 0}, Point{4, 18},
			Text{[]Line{
				Line{""},
			}},
		},
		{
			Text{[]Line{
				Line{"Text is a set of lines."},
				Line{"Lines is a slice of bytes."},
			}},
			Point{0, 10}, Point{0, 10},
			Text{[]Line{
				Line{"Text is a set of lines."},
				Line{"Lines is a slice of bytes."},
			}},
		},
		{
			Text{[]Line{
		        Line{"		for o := viewer.min.o ; o < viewer.max.o ; o++ {"},
				Line{"			SetCell(l, o, ' ', term.ColorDefault, term.ColorDefault)"},
			}},
			Point{0, BFromC("		for o := viewer.min.o ; o < viewer.max.o ; o++ {", 17)}, Point{1, BFromC("			SetCell(l, o, ' ', term.ColorDefault, term.ColorDefault)", 19)},
			Text{[]Line{
				Line{"		for o := (l, o, ' ', term.ColorDefault, term.ColorDefault)"},
			}},
		},
	}
	for _, c := range cases {
		in := c.in
		c.in.RemoveRange(c.min, c.max)
		got := c.in
		if len(c.in.lines) != len(c.in.lines) {
			t.Errorf("len(got.lines) != len(want.lines), len(got.lines)==%q, len(want.lines)==%q", len(c.in.lines), len(c.want.lines))
		}
		var wantl Line
		for i, gotl := range c.in.lines {
			wantl = c.want.lines[i]
			if gotl.data != wantl.data {
				t.Errorf("%q.RemoveRange(%v, %v) == %q, want %q", in, c.min, c.max, got, c.want)
			}
		}
	}
}
