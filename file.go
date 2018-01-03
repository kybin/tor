package main

import (
	"bufio"
	"os"
	"unicode/utf8"
)

// open reads a file and returns it as *Text.
// When the file is not exists, it will return error with nil *Text.
func open(f string) (*Text, error) {
	file, err := os.Open(f)
	if err != nil {
		return nil, err
	}
	defer file.Close()

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

	// `touch` cmd creates a file with no content.
	// avoid program panic from empty text.
	if len(lines) == 0 {
		lines = []Line{{""}}
	}

	return &Text{lines: lines, tabToSpace: tabToSpace, tabWidth: tabWidth}, nil
}

func save(f string, t *Text) error {
	file, err := os.Create(f)
	if err != nil {
		return err
	}
	defer file.Close()
	for _, line := range t.lines {
		file.WriteString(line.data + "\n")
	}
	return nil
}

// exists returns whether the given file or directory exists or not
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
