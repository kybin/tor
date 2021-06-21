package main

import (
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

func vlen(s string, tabWidth int) int {
	remain := s
	o := 0
	for len(remain) > 0 {
		r, rlen := utf8.DecodeRuneInString(remain)
		remain = remain[rlen:]
		if r == '\t' {
			o += tabWidth
		} else {
			w := runewidth.RuneWidth(r)
			if w == 0 {
				w = 1
			}
			o += w
		}
	}
	return o
}
