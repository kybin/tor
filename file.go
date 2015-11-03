package main

import (
	"bufio"
	"os"
	"path/filepath"
	// "log"
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
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ln := Line{scanner.Text()}
		lines = append(lines, ln)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return &Text{lines}, nil
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
