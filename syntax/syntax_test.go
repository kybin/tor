package syntax

import (
	"reflect"
	"testing"

	"github.com/kybin/tor/cell"
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
				{"package", cell.Range{cell.Pt{0, 0}, cell.Pt{0, 8}}},
				{"comment", cell.Range{cell.Pt{2, 0}, cell.Pt{2, 21}}},
				{"multi line comment", cell.Range{cell.Pt{4, 0}, cell.Pt{8, 2}}},
				{"string", cell.Range{cell.Pt{11, 6}, cell.Pt{11, 10}}},
				{"trailing spaces", cell.Range{cell.Pt{11, 10}, cell.Pt{11, 13}}},
				{"string", cell.Range{cell.Pt{12, 6}, cell.Pt{12, 9}}},
				{"string", cell.Range{cell.Pt{12, 16}, cell.Pt{12, 37}}},
				{"rune", cell.Range{cell.Pt{13, 6}, cell.Pt{13, 12}}},
				{"trailing spaces", cell.Range{cell.Pt{14, 0}, cell.Pt{14, 1}}},
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
