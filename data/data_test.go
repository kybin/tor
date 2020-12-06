package data

import (
	"reflect"
	"testing"
)

// Pos represents position of a cursor.
type Pos struct {
	i int // clip index
	o int // offset in the clip
}

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

func TestCursorCut(t *testing.T) {
	cases := []struct {
		label string
		cs    *Cursor
		want  *Cursor
	}{
		{
			label: "middle",
			cs: &Cursor{
				clips:     []Clip{DataClip([]byte("this is sparta."))},
				i:         0,
				o:         7,
				appending: false,
			},
			want: &Cursor{
				clips:     []Clip{DataClip([]byte("this is")), DataClip([]byte(" sparta."))},
				i:         1,
				o:         0,
				appending: false,
			},
		},
		{
			label: "first",
			cs: &Cursor{
				clips:     []Clip{DataClip([]byte("this is sparta."))},
				i:         0,
				o:         0,
				appending: false,
			},
			want: &Cursor{
				clips:     []Clip{DataClip([]byte("this is sparta."))},
				i:         0,
				o:         0,
				appending: false,
			},
		},
		{
			label: "last",
			cs: &Cursor{
				clips:     []Clip{DataClip([]byte("this is sparta."))},
				i:         1,
				o:         0,
				appending: false,
			},
			want: &Cursor{
				clips:     []Clip{DataClip([]byte("this is sparta."))},
				i:         1,
				o:         0,
				appending: false,
			},
		},
	}
	for _, c := range cases {
		c.cs.Cut()
		if !reflect.DeepEqual(c.cs, c.want) {
			t.Fatalf("%s: got %v, want %v", c.label, c.cs, c.want)
		}
	}
}

func TestCursorWrite(t *testing.T) {
	cases := []struct {
		label  string
		cs     *Cursor
		writes []rune
		want   *Cursor
	}{
		{
			label: "middle",
			cs: &Cursor{
				clips:     []Clip{DataClip([]byte("this is sparta."))},
				i:         0,
				o:         7,
				appending: false,
			},
			writes: []rune("n't"),
			want: &Cursor{
				clips:     []Clip{DataClip([]byte("this is")), DataClip([]byte("n't")), DataClip([]byte(" sparta."))},
				i:         2,
				o:         0,
				appending: true,
			},
		},
		{
			label: "first",
			cs: &Cursor{
				clips:     []Clip{DataClip([]byte("this is sparta."))},
				i:         0,
				o:         0,
				appending: false,
			},
			writes: []rune("hey, "),
			want: &Cursor{
				clips:     []Clip{DataClip([]byte("hey, ")), DataClip([]byte("this is sparta."))},
				i:         1,
				o:         0,
				appending: true,
			},
		},
		{
			label: "last",
			cs: &Cursor{
				clips:     []Clip{DataClip([]byte("this is sparta."))},
				i:         1,
				o:         0,
				appending: false,
			},
			writes: []rune(" isn't it?"),
			want: &Cursor{
				clips:     []Clip{DataClip([]byte("this is sparta.")), DataClip([]byte(" isn't it?"))},
				i:         2,
				o:         0,
				appending: true,
			},
		},
	}
	for _, c := range cases {
		for _, r := range c.writes {
			c.cs.Write(r)
		}
		if !reflect.DeepEqual(c.cs, c.want) {
			t.Fatalf("%s: got %v, want %v", c.label, c.cs, c.want)
		}
	}
}

func TestCursorGotoNextLine(t *testing.T) {
	cases := []struct {
		label string
		clips []Clip
		wants []Pos
	}{
		{
			label: "simple",
			clips: []Clip{
				DataClip([]byte("what a nice day\n")),
				DataClip([]byte("do you have breakfast?\n or shall we?")),
			},
			wants: []Pos{{0, 15}, {1, 22}, {2, 0}, {2, 0}},
		},
		{
			label: "korean",
			clips: []Clip{
				DataClip([]byte("이 건\n 한글")), DataClip([]byte("테스트 입니다.\n")), // each hangul character is 3 bytes
			},
			wants: []Pos{{0, 7}, {1, 20}, {2, 0}, {2, 0}},
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
	cases := []struct {
		label string
		clips []Clip
		wants []Pos
	}{
		{
			label: "simple",
			clips: []Clip{
				DataClip([]byte("what a nice day\n")),
				DataClip([]byte("do you have breakfast?\n or shall we?")),
			},
			wants: []Pos{{1, 22}, {0, 15}, {0, 0}, {0, 0}},
		},
		{
			label: "korean",
			clips: []Clip{
				DataClip([]byte("이 건\n 한글")), DataClip([]byte("테스트 입니다.\n")), // each hangul character is 3 bytes
			},
			wants: []Pos{{1, 20}, {0, 7}, {0, 0}, {0, 0}},
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

func TestCursorDelete(t *testing.T) {
	cases := []struct {
		label  string
		clips  []Clip
		from   Pos
		nsteps int
		want   []Clip
	}{
		{
			label: "basic",
			clips: []Clip{
				DataClip([]byte("a")),
				DataClip([]byte("bc")),
				DataClip([]byte("d")),
				DataClip([]byte("e")),
				DataClip([]byte("fgh")),
			},
			from:   Pos{0, 0},
			nsteps: 8,
			want:   []Clip{},
		},
		{
			label: "more steps",
			clips: []Clip{
				DataClip([]byte("a")),
				DataClip([]byte("bc")),
				DataClip([]byte("d")),
				DataClip([]byte("e")),
				DataClip([]byte("fgh")),
			},
			from:   Pos{0, 0},
			nsteps: 12,
			want:   []Clip{},
		},
		{
			label: "in the middle",
			clips: []Clip{
				DataClip([]byte("a")),
				DataClip([]byte("bc")),
				DataClip([]byte("d")),
				DataClip([]byte("e")),
				DataClip([]byte("fgh")),
			},
			from:   Pos{1, 1},
			nsteps: 6,
			want: []Clip{
				DataClip([]byte("a")),
				DataClip([]byte("b")),
			},
		},
		{
			label: "from the end",
			clips: []Clip{
				DataClip([]byte("a")),
			},
			from:   Pos{1, 0},
			nsteps: 3,
			want: []Clip{
				DataClip([]byte("a")),
			},
		},
		{
			label:  "empty clip",
			clips:  []Clip{},
			from:   Pos{0, 0},
			nsteps: 1,
			want:   []Clip{},
		},
	}
	for _, c := range cases {
		cs := NewCursor(c.clips)
		cs.i = c.from.i
		cs.o = c.from.o
		for i := 0; i < c.nsteps; i++ {
			cs.Delete()
		}
		if !reflect.DeepEqual(cs.clips, c.want) {
			t.Fatalf("%s: got %v, want %v", c.label, cs.clips, c.want)
		}
	}
}

func TestCursorBackspace(t *testing.T) {
	cases := []struct {
		label  string
		clips  []Clip
		from   Pos
		nsteps int
		want   []Clip
	}{
		{
			label: "basic",
			clips: []Clip{
				DataClip([]byte("a")),
				DataClip([]byte("bc")),
				DataClip([]byte("d")),
				DataClip([]byte("e")),
				DataClip([]byte("fgh")),
			},
			from:   Pos{5, 0},
			nsteps: 8,
			want:   []Clip{},
		},
		{
			label: "more steps",
			clips: []Clip{
				DataClip([]byte("a")),
				DataClip([]byte("bc")),
				DataClip([]byte("d")),
				DataClip([]byte("e")),
				DataClip([]byte("fgh")),
			},
			from:   Pos{5, 0},
			nsteps: 12,
			want:   []Clip{},
		},
		{
			label: "in the middle",
			clips: []Clip{
				DataClip([]byte("a")),
				DataClip([]byte("bc")),
				DataClip([]byte("d")),
				DataClip([]byte("e")),
				DataClip([]byte("fgh")),
			},
			from:   Pos{4, 1},
			nsteps: 6,
			want: []Clip{
				DataClip([]byte("gh")),
			},
		},
		{
			label: "from the start",
			clips: []Clip{
				DataClip([]byte("a")),
			},
			from:   Pos{0, 0},
			nsteps: 3,
			want: []Clip{
				DataClip([]byte("a")),
			},
		},
		{
			label:  "empty clip",
			clips:  []Clip{},
			from:   Pos{0, 0},
			nsteps: 1,
			want:   []Clip{},
		},
	}
	for _, c := range cases {
		cs := NewCursor(c.clips)
		cs.i = c.from.i
		cs.o = c.from.o
		for i := 0; i < c.nsteps; i++ {
			cs.Backspace()
		}
		if !reflect.DeepEqual(cs.clips, c.want) {
			t.Fatalf("%s: got %v, want %v", c.label, cs.clips, c.want)
		}
	}
}
