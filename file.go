package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"
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

	// if file created with `touch` cmd, scanner could not scan anything,
	// which cause no line in text that makes program panic.
	if len(lines) == 0 {
		lines = append(lines, Line{""})
	}

	return &Text{lines, tabToSpace, tabWidth, false}, nil
}

// parseFileArg parses farg. Which looks like "filepath:line:offset"
// The correspond return values are (filepath, linenum, offset).
//
// If linenum or offset is given but invalid, it will be 0.
// If both are ungiven, those will be -1.
func parseFileArg(farg string) (string, int, int) {
	// final ":" is invalid but ignore it is sufficient.
	if strings.HasSuffix(farg, ":") {
		farg = farg[:len(farg)-1]
	}

	finfo := strings.Split(farg, ":")
	f := finfo[0]
	if len(finfo) == 1 {
		return f, -1, -1
	}

	l, err := strconv.Atoi(finfo[1])
	if err != nil || l < 0 {
		l = 0
	}
	if len(finfo) == 2 {
		return f, l, 0
	}

	o, err := strconv.Atoi(finfo[2])
	if err != nil || o < 0 {
		o = 0
	}
	return f, l, o
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
