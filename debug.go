package main

import (
	"bufio"
	"fmt"
	"os"

	term "github.com/nsf/termbox-go"
)

func debug(args ...interface{}) {
	term.Close()

	for _, a := range args {
		fmt.Println(a)
	}
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')

	err := term.Init()
	if err != nil {
		panic(err)
	}
	term.SetInputMode(term.InputAlt)
}
