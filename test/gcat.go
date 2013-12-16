package main

import (
	"os"
	//"strings"
	"fmt"
	"io/ioutil"
)

func main() {
	// print command arguments
	//fmt.Printf(strings.Join(os.Args[1], " "))
	//fmt.Println()
	rf := os.Args[1]
	// read whole the file
	b, err := ioutil.ReadFile(rf)
	if err != nil {
		panic(err)
	}
	s := string(b)
	fmt.Print(s)
}
