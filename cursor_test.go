package main

import (
	"testing"
)

func TestBFromO(t *testing.T) {
	cases := []struct {
		l    string
		in   int
		want int
	}{
		{
			"		yo, how you doin?",
			9,
			3,
		},
		{
			"	func() xxx {",
			10,
			7,
		},
		{
			"for o := viewer.min.o ; o < viewer.max.o ; o++ {",
			10,
			10,
		},
		{
			"	SetCell(l, o, ' ', term.ColorDefault, term.ColorDefault)",
			11,
			8,
		},
	}
	for _, c := range cases {
		got := BFromO(c.l, c.in)
		if got != c.want {
			t.Errorf("BFromO(%v, %v) == %v, want %v", c.l, c.in, got, c.want)
		}
	}
}
