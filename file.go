package main

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"
)

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
	tabToSpace := false
	tabWidth := 4 // the default tab width

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

	// if file created with `touch` cmd, scanner could not scan anything,
	// which cause no line in text that makes program panic.
	if len(lines) == 0 {
		lines = append(lines, Line{""})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return &Text{lines, tabToSpace, tabWidth, false}, nil
}

// parseFileArg returns (filepath, linenum, offset, error).
// if the linenum is given, but 0 or negative, it will be 1.
// if the offset is given, but negative, it will be 0.
// when only filepath is given, linenum and offset will be set to -1.
func parseFileArg(farg string) (string, int, int, error) {
	finfo := strings.Split(farg, ":")
	f := finfo[0]
	l, o := -1, -1
	err := error(nil)

	if len(finfo) >= 4 {
		return "", -1, -1, errors.New("too many colons in file argument")
	}

	if len(finfo) == 1 {
		return f, -1, -1, nil
	}
	if len(finfo) >= 2 {
		l, err = strconv.Atoi(finfo[1])
		if err != nil {
			return "", -1, -1, err
		}
		if len(finfo) == 3 {
			o, err = strconv.Atoi(finfo[2])
			if err != nil {
				return "", -1, -1, err
			}
		}
	}

	if l < 0 {
		l = 0
	}
	if o < 0 {
		o = 0
	}
	return f, l, o, nil
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

// extend file name at very before the file extension.
// extendFileName("/home/yongbin/test.txt", "_tor") -> "/home/yongbin/test_tor.txt"
func extendFileName(f, e string) string {
	fdir := filepath.Dir(f)
	fname := filepath.Base(f)
	fext := filepath.Ext(f)
	var froot string
	if fext == "" {
		froot = fname
	} else {
		froot = fname[:len(fname)-len(fext)]
	}
	return filepath.Join(fdir, froot+e+fext)
}
