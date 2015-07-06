package main

import (
	"testing"
)

func TestSaveAndLoadPosition(t *testing.T) {
	err := savePosition("/home/kybin/not-exist.file", 10, 3)
	if err != nil {
		t.Error(err)
	}
	l, b := lastPosition("/home/kybin/not-exist.file")
	if l != 10 || b != 3 {
		t.Error("Could not load last position properly.")
	}
}
