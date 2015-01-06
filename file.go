package main

import (
	"io/ioutil"
)

type Line []byte
type Text []Line

func open(f string) Text {

	bytes, err := ioutil.ReadFile(f)
	if err != nil {
		panic(err)
	}

	txt := make(Text, 0)
	lastidx := -1
	for idx, rune := range bytes {
		if rune == '\n' {
			newline := Line(bytes[lastidx+1 : idx])
			lastidx = idx
			txt = append(txt, newline)
		}
	}
	//	fmt.Print(txt)
	return txt
}
