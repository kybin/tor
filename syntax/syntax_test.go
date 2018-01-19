package syntax

import (
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
				{Name: "package", Range: cell.Range{cell.Pt{0, 0}, cell.Pt{0, 8}}},
				{Name: "comment", Range: cell.Range{cell.Pt{2, 0}, cell.Pt{2, 21}}},
				{Name: "multi line comment", Range: cell.Range{cell.Pt{4, 0}, cell.Pt{8, 2}}},
				{Name: "string", Range: cell.Range{cell.Pt{11, 6}, cell.Pt{11, 10}}},
				{Name: "trailing spaces", Range: cell.Range{cell.Pt{11, 10}, cell.Pt{11, 13}}},
				{Name: "string", Range: cell.Range{cell.Pt{12, 6}, cell.Pt{12, 9}}},
				{Name: "string", Range: cell.Range{cell.Pt{12, 16}, cell.Pt{12, 37}}},
				{Name: "rune", Range: cell.Range{cell.Pt{13, 6}, cell.Pt{13, 12}}},
				{Name: "trailing spaces", Range: cell.Range{cell.Pt{14, 0}, cell.Pt{14, 1}}},
			},
		},
	}
	for _, c := range cases {
		p := NewParser(&B{c.text}, c.langName)
		p.ParseTo(cell.Pt{1000, 0})
		got := p.Matches
		if len(got) != len(c.want) {
			t.Fatalf("(%v).ParseTo(end): got %v, want %v", p, got, c.want)
		}
		for i := range got {
			if !sameMatch(got[i], c.want[i]) {
				t.Fatalf("(%v).ParseTo(end): got %v, want %v", p, got, c.want)
			}
		}
	}
}

// B implements Byter
type B struct {
	text []byte
}

func (b *B) Bytes() []byte {
	return b.text
}

// sameMatch returns whether those are same match.
// It does not care of the matches' color.
func sameMatch(m, n Match) bool {
	return m.Name == n.Name && m.Range == n.Range
}
