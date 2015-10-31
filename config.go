package main

import (
	"os"
	"os/user"
	"fmt"
	"strings"
	"strconv"
	"path"
	"path/filepath"
	"io/ioutil"
)

func saveLastPosition(relpath string, l, b int) error {
	abspath, err := filepath.Abs(relpath)
	if err != nil {
		return err
	}
	u, err := user.Current()
	if err != nil {
		return err
	}
	f := path.Join(u.HomeDir, ".config", "tor", "lastpos")
	if _, err = os.Stat(f); os.IsNotExist(err) {
		d := path.Join(u.HomeDir, ".config", "tor")
		os.MkdirAll(d, 0777)
		os.Create(f)
	}
	input, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}
	lines := strings.Split(string(input), "\n")
	find := false
	for i, ln := range lines {
		if strings.Contains(ln, abspath + ":") {
			find = true
			lines[i] = fmt.Sprintf("%v:%v:%v", abspath, l, b)
		}
	}
	if !find {
		lines = append(lines, fmt.Sprintf("%v:%v:%v", abspath, l, b))
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(f, []byte(output), 0644)
	if err != nil {
		return err
	}
	return nil
}

func loadLastPosition(relpath string) (int, int) {
	abspath, err := filepath.Abs(relpath)
	if err != nil {
		return 0, 0
	}
	u, err := user.Current()
	if err != nil {
		return 0, 0
	}
	f := path.Join(u.HomeDir, ".config", "tor", "lastpos")
	input, err := ioutil.ReadFile(f)
	if err != nil {
		return 0, 0
	}
	find := false
	findline := ""
	lines := strings.Split(string(input), "\n")
	for _, ln := range lines {
		if strings.Contains(ln, abspath + ":") {
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

func saveCopyString(copystr string) error {
	u, err := user.Current()
	if err != nil {
		return err
	}
	f := path.Join(u.HomeDir, ".config", "tor", "copy")
	if _, err = os.Stat(f); os.IsNotExist(err) {
		d := path.Join(u.HomeDir, ".config", "tor")
		os.MkdirAll(d, 0777)
	}
	err = ioutil.WriteFile(f, []byte(copystr), 0644)
	if err != nil {
		return err
	}
	return nil
}

func loadCopyString() string {
	u, err := user.Current()
	if err != nil {
		return ""
	}
	f := path.Join(u.HomeDir, ".config", "tor", "copy")
	copystr, err := ioutil.ReadFile(f)
	if err != nil {
		return ""
	}
	return string(copystr)
}
