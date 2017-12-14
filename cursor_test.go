package main

import (
	"testing"
)

func TestBFromO(t *testing.T) {
	cases := []struct {
		line     string
		o        int
		tabWidth int
		want     int
	}{
		{
			line: "		yo, how you doin?",
			o:        9,
			tabWidth: 4,
			want:     3,
		},
		{
			line: "	func() xxx {",
			o:        10,
			tabWidth: 4,
			want:     7,
		},
		{
			line:     "for o := viewer.min.o ; o < viewer.max.o ; o++ {",
			o:        10,
			tabWidth: 4,
			want:     10,
		},
		{
			line: "	SetCell(l, o, ' ', term.ColorDefault, term.ColorDefault)",
			o:        11,
			tabWidth: 4,
			want:     8,
		},
	}
	for _, c := range cases {
		got := BFromO(c.line, c.o, c.tabWidth)
		if got != c.want {
			t.Errorf("BFromO(%v, %v): got %v, want %v", c.line, c.o, got, c.want)
		}
	}
}
