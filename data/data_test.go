package data

import (
	"reflect"
	"testing"
)

func TestDataClip(t *testing.T) {
	cases := []struct {
		data []byte
		want Clip
	}{
		{
			data: []byte("this is sparta!"),
			want: Clip{data: []byte("this is sparta!"), newlines: []int{}},
		},
		{
			data: []byte("this is sparta!\nright?"),
			want: Clip{data: []byte("this is sparta!\nright?"), newlines: []int{15}},
		},
		{
			data: []byte("\n\n\n\n"),
			want: Clip{data: []byte("\n\n\n\n"), newlines: []int{0, 1, 2, 3}},
		},
	}
	for _, c := range cases {
		got := DataClip(c.data)
		if !reflect.DeepEqual(got, c.want) {
			t.Fatalf("got %q, want %q", got, c.want)
		}
	}
}

func TestCursorWrite(t *testing.T) {
	cases := []struct {
		label  string
		cs     *Cursor
		writes []rune
		at     int
		want   *Cursor
	}{
		{
			label:  "middle",
			cs:     NewCursor([]Clip{DataClip([]byte("this is sparta."))}),
			writes: []rune("n't"),
			at:     7,
			want: &Cursor{
				clips:     []Clip{DataClip([]byte("this is")), DataClip([]byte("n't")), DataClip([]byte(" sparta."))},
				i:         2,
				o:         0,
				appending: true,
			},
		},
		{
			label:  "first",
			cs:     NewCursor([]Clip{DataClip([]byte("this is sparta."))}),
			writes: []rune("hey, "),
			at:     0,
			want: &Cursor{
				clips:     []Clip{DataClip([]byte("hey, ")), DataClip([]byte("this is sparta."))},
				i:         1,
				o:         0,
				appending: true,
			},
		},
		{
			label:  "last",
			cs:     NewCursor([]Clip{DataClip([]byte("this is sparta."))}),
			writes: []rune(" isn't it?"),
			at:     15,
			want: &Cursor{
				clips:     []Clip{DataClip([]byte("this is sparta.")), DataClip([]byte(" isn't it?"))},
				i:         2,
				o:         0,
				appending: true,
			},
		},
	}
	for _, c := range cases {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("%q panicked: %s", c.label, r)
			}
		}()
		c.cs.Shift(c.at)
		for _, r := range c.writes {
			c.cs.Write(r)
		}
		if !reflect.DeepEqual(c.cs, c.want) {
			t.Fatalf("got %v, want %v", c.cs, c.want)
		}
	}
}
