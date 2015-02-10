package main

import (
	"testing"
)

func TestExtendFileName(t *testing.T) {
	cases := []struct{
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
