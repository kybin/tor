package main

import (
	"testing"

	"github.com/kybin/tor/cell"
)

func TestRemoveRange(t *testing.T) {
	cases := []struct {
		in       Text
		min, max cell.Pt
		want     Text
	}{
		{
			Text{lines: []Line{
				{"Hello, my name is yongbin."},
				{"This is the test string."},
				{"You are great."},
			}},
			cell.Pt{0, 6}, cell.Pt{2, 7},
			Text{lines: []Line{
				{"Hello, great."},
			}},
		},
		{
			Text{lines: []Line{
				{"blizzard"},
				{"	wow"},
				{"	Diablo"},
			}},
			cell.Pt{0, 0}, cell.Pt{1, 0},
			Text{lines: []Line{
				{"	wow"},
				{"	Diablo"},
			}},
		},
		{
			Text{lines: []Line{
				{"The delete built-in function"},
				{"deletes the element"},
				{"with the specified key (m[key]) from the map."},
				{"If m is nil or there is no such element,"},
				{"delete is a no-op."},
			}},
			cell.Pt{0, 0}, cell.Pt{4, 18},
			Text{lines: []Line{
				{""},
			}},
		},
		{
			Text{lines: []Line{
				{"Text is a set of lines."},
				{"Lines is a slice of bytes."},
			}},
			cell.Pt{0, 10}, cell.Pt{0, 10},
			Text{lines: []Line{
				{"Text is a set of lines."},
				{"Lines is a slice of bytes."},
			}},
		},
		{
			Text{lines: []Line{
				{"		for o := viewer.min.o ; o < viewer.max.o ; o++ {"},
				{"			SetCell(l, o, ' ', term.ColorDefault, term.ColorDefault)"},
			}},
			cell.Pt{0, BFromO("		for o := viewer.min.o ; o < viewer.max.o ; o++ {", 17, 4)}, cell.Pt{1, BFromO("			SetCell(l, o, ' ', term.ColorDefault, term.ColorDefault)", 19, 4)},
			Text{lines: []Line{
				{"		for o := (l, o, ' ', term.ColorDefault, term.ColorDefault)"},
			}},
		},
	}
	for _, c := range cases {
		in := c.in
		c.in.RemoveRange(c.min, c.max)
		got := c.in
		if len(c.in.lines) != len(c.in.lines) {
			t.Fatalf("len(got.lines) != len(want.lines), len(got.lines)==%q, len(want.lines)==%q", len(c.in.lines), len(c.want.lines))
		}
		var wantl Line
		for i, gotl := range c.in.lines {
			wantl = c.want.lines[i]
			if gotl.data != wantl.data {
				t.Fatalf("%v.RemoveRange(%v, %v) == %v, want %v", in, c.min, c.max, got, c.want)
			}
		}
	}
}
