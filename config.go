package main

import (
	"os"
	"os/user"
	"fmt"
	"strings"
	"strconv"
	"path"
	"io/ioutil"
)

func savePosition(workingfile string, l, b int) error {
	u, err := user.Current()
	if err != nil {
		return err
	}
	posfile := path.Join(u.HomeDir, ".config", "tor", "lastpos")
	if _, err = os.Stat(posfile); os.IsNotExist(err) {
		posdir := path.Join(u.HomeDir, ".config", "tor")
		os.MkdirAll(posdir, 0777)
		os.Create(posfile)
	}
	input, err := ioutil.ReadFile(posfile)
	if err != nil {
		return err
	}
	lines := strings.Split(string(input), "\n")
	find := false
	for i, ln := range lines {
		if strings.Contains(ln, workingfile + ":") {
			find = true
			lines[i] = fmt.Sprintf("%v:%v:%v", workingfile, l, b)
		}
	}
	if !find {
		lines = append(lines, fmt.Sprintf("%v:%v:%v", workingfile, l, b))
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(posfile, []byte(output), 0644)
	if err != nil {
		return err
	}
	return nil
}

func lastPosition(workingfile string) (int, int) {
	u, err := user.Current()
	if err != nil {
		return 0, 0
	}
	posfile := path.Join(u.HomeDir, ".config", "tor", "lastpos")
	input, err := ioutil.ReadFile(posfile)
	if err != nil {
		return 0, 0
	}
	find := false
	findline := ""
	lines := strings.Split(string(input), "\n")
	for _, ln := range lines {
		if strings.Contains(ln, workingfile + ":") {
			find = true
			findline = ln
		}
	}
	if !find {
		return 0, 0
	}
	tokens := strings.Split(findline, ":")
	if len(tokens) != 3 {
		return 0, 0
	}
	l, err := strconv.Atoi(tokens[1])
	if err != nil {
		return 0, 0
	}
	b, err := strconv.Atoi(tokens[2])
	if err != nil {
		return 0, 0
	}
	return l, b
}

