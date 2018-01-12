package main

import (
	"bufio"
	"fmt"
	"os"
	"sync"

	term "github.com/nsf/termbox-go"
)

func debug(args ...interface{}) {
	// there is another goroutine interacting with terminal.
	mu := &sync.Mutex{}
	mu.Lock()
	defer mu.Unlock()

	term.Close()

	fmt.Println(args)

	// wait for enter
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')

	err := term.Init()
	if err != nil {
		panic(err)
	}
	term.SetInputMode(term.InputAlt)
}
