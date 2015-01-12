package main

import (
	"bufio"
	"os"
	"log"
)

func open(f string) *Text {
	file, err := os.Open(f)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	lines := make([]Line, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ln := Line{scanner.Text()}
		lines = append(lines, ln)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return &Text{lines}
}
