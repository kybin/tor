
package main

import (
	"os"
	"bufio"
	"fmt"
	"log"
	_ "errors"
)

func main() {
	if len(os.Args) != 2 {
		err := fmt.Errorf("Invalid number of arguments %d", len(os.Args))
		log.Fatal(err)
	}
	fname := os.Args[1]
	f, err := os.Open(fname)
	if err != nil{
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
