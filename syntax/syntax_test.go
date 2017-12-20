package syntax

import (
	"reflect"
	"testing"
)

func TestUsage(t *testing.T) {
	cases := []struct {
		text     []byte
		langName string
		want     []Match
	}{
		{
			text: []byte(`package main

// this is a comment.

/*
	this is a multi-line comment.
	that has multiple lines.
		with irregular indents.
*/

func main() {
	s := "yo"	  
	t := "\"string\" inside of a string"
	r := 'rune'
	
}
`),
			langName: "go",
			want: []Match{
				{"package", Pos{0, 0}, Pos{0, 8}},
				{"comment", Pos{2, 0}, Pos{2, 21}},
				{"multi line comment", Pos{4, 0}, Pos{8, 2}},
				{"string", Pos{11, 6}, Pos{11, 10}},
				{"trailing spaces", Pos{11, 10}, Pos{11, 13}},
				{"string", Pos{12, 6}, Pos{12, 9}},
				{"string", Pos{12, 16}, Pos{12, 37}},
				{"rune", Pos{13, 6}, Pos{13, 12}},
				{"trailing spaces", Pos{14, 0}, Pos{14, 1}},
			},
		},
	}
	for _, c := range cases {
		lang, ok := Languages[c.langName]
		if !ok {
			return
		}
		got := lang.Parse(c.text)
		if !reflect.DeepEqual(got, c.want) {
			t.Fatalf("(%v).Parse(%v): got %v, want %v", lang, c.text, got, c.want)
		}

	}
}

func TestPosCompare(t *testing.T) {
	cases := []struct {
		a    Pos
		b    Pos
		want int
	}{
		{
			a:    Pos{0, 0},
			b:    Pos{0, 0},
			want: 0,
		},
		{
			a:    Pos{0, 1},
			b:    Pos{0, 2},
			want: -1,
		},
		{
			a:    Pos{0, 2},
			b:    Pos{0, 1},
			want: 1,
		},
		{
			a:    Pos{0, 1},
			b:    Pos{1, 0},
			want: -1,
		},
		{
			a:    Pos{1, 0},
			b:    Pos{0, 1},
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
