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

func TestCursorGotoNextLine(t *testing.T) {
	type Step struct {
		i int
		o int
	}
	cases := []struct {
		label string
		clips []Clip
		wants []Step
	}{
		{
			label: "simple",
			clips: []Clip{
				DataClip([]byte("what a nice day\n")),
				DataClip([]byte("do you have breakfast?\n or shall we?")),
			},
			wants: []Step{{0, 15}, {1, 22}, {2, 0}, {2, 0}},
		},
		{
			label: "korean",
			clips: []Clip{
				DataClip([]byte("이 건\n 한글")), DataClip([]byte("테스트 입니다.\n")), // each hangul character is 3 bytes
			},
			wants: []Step{{0, 7}, {1, 20}, {2, 0}, {2, 0}},
		},
	}

	for _, c := range cases {
		cs := NewCursor(c.clips)
		for i, want := range c.wants {
			cs.GotoNextLine()
			if cs.i != want.i || cs.o != want.o {
				t.Fatalf("%s: step %d: got [%d:%d], want [%d:%d]", c.label, i, cs.i, cs.o, want.i, want.o)
			}
		}
	}
}

func TestCursorGotoPrevLine(t *testing.T) {
	type Step struct {
		i int
		o int
	}
	cases := []struct {
		label string
		clips []Clip
		wants []Step
	}{
		{
			label: "simple",
			clips: []Clip{
				DataClip([]byte("what a nice day\n")),
				DataClip([]byte("do you have breakfast?\n or shall we?")),
			},
			wants: []Step{{1, 22}, {0, 15}, {0, 0}, {0, 0}},
		},
		{
			label: "korean",
			clips: []Clip{
				DataClip([]byte("이 건\n 한글")), DataClip([]byte("테스트 입니다.\n")), // each hangul character is 3 bytes
			},
			wants: []Step{{1, 20}, {0, 7}, {0, 0}, {0, 0}},
		},
	}

	for _, c := range cases {
		cs := &Cursor{clips: c.clips, i: len(c.clips), o: 0}
		for i, want := range c.wants {
			cs.GotoPrevLine()
			if cs.i != want.i || cs.o != want.o {
				t.Fatalf("%s: step %d: got [%d:%d], want [%d:%d]", c.label, i, cs.i, cs.o, want.i, want.o)
			}
		}
	}
}

func TestCursorGotoNext(t *testing.T) {
	cases := []struct {
		label string
		clips []Clip
		nstep int
	}{
		{
			label: "numbers and newlines",
			clips: []Clip{
				DataClip([]byte("1")),
				DataClip([]byte("23\n")),
				DataClip([]byte(" 4\r\n")),
				DataClip([]byte("56 7")),
				DataClip([]byte("8\n\n\n90")),
			},
			nstep: 17,
		},
		{
			label: "korean",
			clips: []Clip{
				DataClip([]byte("한글 테스트\n")),
				DataClip([]byte("english test\n")),
			},
			nstep: 20,
		},
	}
	for _, c := range cases {
		cs := NewCursor(c.clips)
		for i := 0; i < c.nstep; i++ {
			cs.MoveNext()
		}
		if cs.i != len(c.clips) || cs.o != 0 {
			t.Fatalf("%s: got [%d:%d], want [%d:%d]", c.label, cs.i, cs.o, len(c.clips), 0)
		}
	}
}

func TestCursorGotoPrev(t *testing.T) {
	cases := []struct {
		label string
		clips []Clip
		nstep int
	}{
		{
			label: "numbers and newlines",
			clips: []Clip{
				DataClip([]byte("1")),
				DataClip([]byte("23\n")),
				DataClip([]byte(" 4\r\n")),
				DataClip([]byte("56 7")),
				DataClip([]byte("8\n\n\n90")),
			},
			nstep: 17,
		},
		{
			label: "korean",
			clips: []Clip{
				DataClip([]byte("한글 테스트\n")),
				DataClip([]byte("english test\n")),
			},
			nstep: 20,
		},
	}
	for _, c := range cases {
		cs := NewCursor(c.clips)
		cs.i = len(c.clips)
		for i := 0; i < c.nstep; i++ {
			cs.MovePrev()
		}
		if cs.i != 0 || cs.o != 0 {
			t.Fatalf("%s: got [%d:%d], want [%d:%d]", c.label, cs.i, cs.o, 0, 0)
		}
	}
}
