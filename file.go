package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"unicode/utf8"
)

// isCreatable check if users has permission to create.
// If the file already exists, it will return error.
//
// NOTE: It actually creates the file then delete.
func isCreatable(f string) (bool, error) {
	// ensure the file does not exist.
	_, err := os.Stat(f)
	if err == nil {
		return false, fmt.Errorf("file should not exists: %v", f)
	}
	if !os.IsNotExist(err) {
		return false, err
	}
	// file creation checking.
	file, err := os.Create(f)
	if err != nil {
		if !os.IsPermission(err) {
			return false, err
		}
		return false, nil
	}
	if err := file.Close(); err != nil {
		return false, err
	}
	if err := os.Remove(f); err != nil {
		// TODO: better finalization for remove failure?
		return false, err
	}
	return true, nil
}

// isWritable checks whether f is writable file or not.
// If it couldn't open the file for check, it will return error.
func isWritable(f string) (bool, error) {
	file, err := os.OpenFile(f, os.O_WRONLY, 0666)
	if err != nil {
		if os.IsPermission(err) {
			return false, nil
		}
		return false, err
	}
	file.Close()
	return true, nil
}

// readOrCreate open f and read it's Text.
// When f doesn't exist and allow to create, it will create a new Text.
func readOrCreate(f string, allowCreate bool) (*Text, error) {
	if _, err := os.Stat(f); err != nil {
		if os.IsNotExist(err) {
			if allowCreate {
				return create(f)
			}
			return nil, errors.New("file not exist. please retry with -new flag.")
		}
		return nil, err
	}
	return read(f)
}

// create creates a new Text for f.
// When f is not creatable, it will return error.
func create(f string) (*Text, error) {
	writable, err := isCreatable(f)
	if err != nil {
		return nil, err
	}
	if !writable {
		return nil, errors.New("could not create the file. please check the directory permission.")
	}
	return &Text{lines: []Line{{""}}, tabToSpace: false, tabWidth: 4, writable: writable, lineEnding: "\n"}, nil
}

// read reads a file and returns it as *Text.
// When the file is not exists, it will return error with nil *Text.
func read(f string) (*Text, error) {
	file, err := os.Open(f)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	writable, err := isWritable(f)
	if err != nil {
		return nil, err
	}

	// aggregate the text info.
	// tor uses tab (4 space) for indentation.
	// but when parse an exist file, follow the file's rule.
	lines := make([]Line, 0)
	tabToSpace := false
	tabWidth := 4

	findIndentLine := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		t := scanner.Text()
		if !findIndentLine {
			r, _ := utf8.DecodeRuneInString(t)
			if r == ' ' || r == '\t' {
				findIndentLine = true
				if r == ' ' {
					tabToSpace = true
					// calculate tab width
					tabWidth = 0
					remain := t
					for len(remain) != 0 {
						r, rlen := utf8.DecodeRuneInString(remain)
						remain = remain[rlen:]
						if r != ' ' {
							break
						}
						tabWidth++
					}
				}
			}
		}
		lines = append(lines, Line{t})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// check line ending
	lineEnding := "\n"
	if len(lines) != 0 {
		file.Seek(0, 0)
		reader := bufio.NewReader(file)
		firstLine, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, err
		}
		if len(firstLine) >= 2 && firstLine[len(firstLine)-2:] == "\r\n" {
			lineEnding = "\r\n"
		}
	}

	// `touch` cmd creates a file with no content.
	// avoid program panic from empty text.
	if len(lines) == 0 {
		lines = []Line{{""}}
	}

	return &Text{lines: lines, tabToSpace: tabToSpace, tabWidth: tabWidth, writable: writable, lineEnding: lineEnding}, nil
}

// save saves Text to a file.
func save(f string, t *Text) error {
	file, err := os.Create(f)
	if err != nil {
		return err
	}
	defer file.Close()
	for _, line := range t.lines {
		file.WriteString(line.data)
		file.WriteString(t.lineEnding)
	}
	return nil
}
