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
