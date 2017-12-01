package syntax

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestUsage(t *testing.T) {
	testText, err := ioutil.ReadFile("testdata/test.txt")
	if err != nil {
		t.Fatal(err)
	}
	lang, ok := Languages["go"]
	if !ok {
		return
	}
	matches := lang.Parse(testText)
	for _, m := range matches {
		fmt.Println(m)
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
