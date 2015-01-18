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

func save(f string, t *Text) error {
	file, err := os.Create(extendFileName(f, "_tor"))
	if err != nil {
		return err
	}
	defer file.Close()
	for _, line := range t.lines {
		file.WriteString(line.data+"\n")
	}
	return nil
}
