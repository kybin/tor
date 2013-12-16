package main

import (
	//	"fmt"
	//	"os"
	"io/ioutil"
)

type line []byte
type text []line

func open(f string) text {

	bytes, err := ioutil.ReadFile(f)
	if err != nil {
		panic(err)
	}

	txt := make(text, 0)
	lastidx := -1
	for idx, rune := range bytes {
		if rune == '\n' {
			newline := line(bytes[lastidx+1 : idx])
			lastidx = idx
			txt = append(txt, newline)
		}
	}
	//	fmt.Print(txt)
	return txt
}

//func main() {
//	readf := os.Args[1]
//	readf := "/home/kybin/go/src/github.com/coldmine/tor/tor.go"
//	contents := open(readf)
//	fmt.Println(contents)
//}
