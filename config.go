package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

// configDir is where config files will saved.
// It is $HOME/.config/tor
var configDir = ""

func init() {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	configDir = path.Join(u.HomeDir, ".config", "tor")
	if err := os.MkdirAll(configDir, 0755); err != nil && !os.IsExist(err) {
		panic(err)
	}
}

// saveLastPosition saves a cursor position to
// the 'lastpos' config file.
// Each line is formatted as {filepath}:{line}:{offset}
func saveLastPosition(relpath string, l, b int) error {
	abspath, err := filepath.Abs(relpath)
	if err != nil {
		return err
	}
	f := path.Join(configDir, "lastpos")
	if _, err = os.Stat(f); os.IsNotExist(err) {
		os.Create(f)
	}
	input, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}
	lines := strings.Split(string(input), "\n")
	find := false
	for i, ln := range lines {
		if strings.Contains(ln, abspath+":") {
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

// loadLastPosition loads a cursor position from
// the 'lastpos' config file.
// If there is no information about the file in 'lastpos',
// it will return 0, 0.
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
		if strings.Contains(ln, abspath+":") {
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

// saveConfig saves a string to ~/.config/tor/{fname} file.
// It will return error if exists.
func saveConfig(fname, s string) error {
	f := path.Join(configDir, fname)
	return ioutil.WriteFile(f, []byte(s), 0644)
}

// loadConfig loads a string from ~/.config/tor/{fname} file.
// On any error, it will return empty string.
func loadConfig(fname string) string {
	f := path.Join(configDir, fname)
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return ""
	}
	return string(b)
}
