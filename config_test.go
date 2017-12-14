package main

import (
	"testing"
)

func TestSaveAndLoadLastPosition(t *testing.T) {
	err := saveLastPosition("/home/kybin/not-exist.file", 10, 3)
	if err != nil {
		t.Error(err)
	}
	l, b := loadLastPosition("/home/kybin/not-exist.file")
	if l != 10 || b != 3 {
		t.Error("Could not load last position properly.")
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	err := saveConfig("deleteme", "yay")
	if err != nil {
		t.Error(err)
	}
	copystr := loadConfig("deleteme")
	if copystr != "yay" {
		t.Error("Could not load copy string.")
	}
}
