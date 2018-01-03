package main

import (
	"bufio"
	"os"
	"unicode/utf8"
)

// open read a file and return it as *Text.
// If the file not exist, it will return *Text with one empty line.
func open(f string) (*Text, error) {
	ex, err := exists(f)
	if err != nil {
		return nil, err
	}
	if !ex {
		return &Text{lines: []Line{Line{data: ""}}}, nil
	}
	file, err := os.Open(f)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	lines := make([]Line, 0)

	// tor use tab(shown as 4 space) for indentation as default.
	// But when parse an exsit file, follow the file's rule.
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

	return &Text{lines, tabToSpace, tabWidth, false}, nil
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
