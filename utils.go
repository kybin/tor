package main

import (
	"path/filepath"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// extend file name at very before the file extension.
// extendFileName("/home/yongbin/test.txt", "_tor") -> "/home/yongbin/test_tor.txt"
func extendFileName(f, e string) string {
	fdir := filepath.Dir(f)
	fname := filepath.Base(f)
	fext := filepath.Ext(f)
	var froot string
	if fext == "" {
		froot = fname
	} else {
		froot = fname[:len(fname)-len(fext)]
	}
	return filepath.Join(fdir, froot+e+fext)
}
