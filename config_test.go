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

func TestSaveAndLoadCopyString(t *testing.T) {
	err := saveCopyString("yay")
	if err != nil {
		t.Error(err)
	}
	copystr := loadCopyString()
	if copystr != "yay" {
		t.Error("Could not load copy string.")
	}
}
