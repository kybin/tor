package syntax

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestUsage(t *testing.T) {
	testText, err := ioutil.ReadFile("testdata/test.txt")
	if err != nil {
		t.Fatal(err)
	}
	lang, ok := Languages["go"]
	if !ok {
		return
	}
	matches := lang.Parse(testText)
	for _, m := range matches {
		fmt.Println(m)
	}
}
