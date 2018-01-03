package main

import (
	"reflect"
	"testing"
)

func TestSortArgs(t *testing.T) {
	cases := []struct {
		args []string
		want []string
	}{
		{
			args: []string{"file"},
			want: []string{"file"},
		},
		{
			args: []string{"-new", "file"},
			want: []string{"-new", "file"},
		},
		{
			args: []string{"file", "-new"},
			want: []string{"-new", "file"},
		},
		{
			args: []string{"-a", "file", "-b"},
			want: []string{"-a", "-b", "file"},
		},
		{
			args: []string{"file", "-a", "-b"},
			want: []string{"-a", "-b", "file"},
		},
	}
	for _, c := range cases {
		args := c.args
		sortArgs(args)
		if !reflect.DeepEqual(args, c.want) {
			t.Fatalf("sortArgs(%v): got %v, want %v", c.args, args, c.want)
		}
	}
}

func TestParseFileArg(t *testing.T) {
	cases := []struct {
		arg      string
		wantFile string
		wantL    int
		wantO    int
	}{
		{
			arg:      "hello.go:1:1",
			wantFile: "hello.go",
			wantL:    0,
			wantO:    0,
		},
		{
			arg:      "hello.go:27:3",
			wantFile: "hello.go",
			wantL:    26,
			wantO:    2,
		},
		{
			// make negative offsets to 0.
			arg:      "hello.go:-1:-1",
			wantFile: "hello.go",
			wantL:    0,
			wantO:    0,
		},
		{
			arg:      "hello.go",
			wantFile: "hello.go",
			wantL:    -1,
			wantO:    -1,
		},
		{
			arg:      "hello.go:",
			wantFile: "hello.go",
			wantL:    -1,
			wantO:    -1,
		},
		{
			arg:      "hello.go:2",
			wantFile: "hello.go",
			wantL:    1,
			wantO:    0,
		},
		{
			arg:      "hello.go:2:",
			wantFile: "hello.go",
			wantL:    1,
			wantO:    0,
		},
		{
			arg:      "hello.go:2:2",
			wantFile: "hello.go",
			wantL:    1,
			wantO:    1,
		},
		{
			arg:      "hello.go:2:2:",
			wantFile: "hello.go",
			wantL:    1,
			wantO:    1,
		},
		{
			arg:      "hello.go:a",
			wantFile: "hello.go",
			wantL:    0,
			wantO:    0,
		},
		{
			arg:      "hello.go:2:b",
			wantFile: "hello.go",
			wantL:    1,
			wantO:    0,
		},
	}
	for _, c := range cases {
		gotFile, gotL, gotO := parseFileArg(c.arg)
		if gotFile != c.wantFile || gotL != c.wantL || gotO != c.wantO {
			t.Fatalf("parseFileArg(%v): got %v:%v:%v, want %v:%v:%v", c.arg, gotFile, gotL, gotO, c.wantFile, c.wantL, c.wantO)
		}
	}
}
