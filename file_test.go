package main

import (
	"testing"
)

func TestParseFileArg(t *testing.T) {
	f, l, o, err := parseFileArg("hello.go:27:3")
	if err != nil {
		t.Error(err)
	}
	if f != "hello.go" || l != 27 || o != 3 {
		t.Errorf("expect hello:27:3, got %v:%v:%v", f, l, o)
	}

	f, l, o, err = parseFileArg("hello.go:1:0")
	if err != nil {
		t.Error(err)
	}
	if f != "hello.go" || l != 1 || o != 0 {
		t.Errorf("expect hello:1:0, got %v:%v:%v", f, l, o)
	}

	f, l, o, err = parseFileArg("hello.go:-1:-1")
	if err != nil {
		t.Error(err)
	}
	if f != "hello.go" || l != 0 || o != 0 {
		t.Errorf("expect hello:0:0, got %v:%v:%v", f, l, o)
	}

	f, l, o, err = parseFileArg("hello.go:30:-5")
	if err != nil {
		t.Error(err)
	}
	if f != "hello.go" || l != 30 || o != 0 {
		t.Errorf("expect hello:30:0, got %v:%v:%v", f, l, o)
	}

	f, l, o, err = parseFileArg("hello.go:-10:15")
	if err != nil {
		t.Error(err)
	}
	if f != "hello.go" || l != 0 || o != 15 {
		t.Errorf("expect hello:0:15, got %v:%v:%v", f, l, o)
	}

	f, l, o, err = parseFileArg("hello.go:")
	if err == nil {
		t.Error("shold return parse arugment error")
	}

	f, l, o, err = parseFileArg("hello.go:-1:")
	if err == nil {
		t.Error("shold return parse arugment error")
	}

	f, l, o, err = parseFileArg("hello.go:-1:-1:")
	if err == nil {
		t.Error("shold return too many colon error")
	}
}

func TestExtendFileName(t *testing.T) {
	cases := []struct {
		f, e, want string
	}{
		{"/home/yongbin/test.txt", "_tor", "/home/yongbin/test_tor.txt"},
		{"D:\\undercity.bgeo", ".raise", "D:\\undercity.raise.bgeo"},
		{"hi.txt", "Hello", "hiHello.txt"},
		{"a", "b", "ab"},
	}
	for _, c := range cases {
		got := extendFileName(c.f, c.e)
		if got != c.want {
			t.Errorf("extendFileName(%q, %q) == %q, want %q", c.f, c.e, c.want)
		}
	}
}
