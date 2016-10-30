package main

// ~/.config/tor 디렉토리 안의 설정파일을 저장하거나 부를때 사용하는 함수들

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

// saveLastPosition는 해당 파일을 닫을 때의 커서위치를 저장한다.
// 이 정보는 ~/.config/tor/lastpos 파일에 쌓이며 그 형식은
// {filepath}:{line}:{offset} 이다.
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

// loadLastPosition는 마지막으로 해당 파일을 닫을 때의 커서위치를 가지고 온다.
// 이 정보는 ~/.config/tor/lastpos 파일에 쌓이며 그 형식은
// {filepath}:{line}:{offset} 이다.
// 만일 이 파일을 전에 닫은 적이 없다면 (0, 0)을 반환한다.
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

// saveConfig save string to ~/.config/tor/{fname} file.
func saveConfig(fname, s string) error {
	u, err := user.Current()
	if err != nil {
		return err
	}
	f := path.Join(u.HomeDir, ".config", "tor", fname)
	if _, err = os.Stat(f); os.IsNotExist(err) {
		d := path.Join(u.HomeDir, ".config", "tor")
		os.MkdirAll(d, 0777)
	}
	err = ioutil.WriteFile(f, []byte(s), 0644)
	if err != nil {
		return err
	}
	return nil
}

// saveConfig load string from ~/.config/tor/{fname} file.
// On any error, it will return empty string.
func loadConfig(fname string) string {
	u, err := user.Current()
	if err != nil {
		return ""
	}
	f := path.Join(u.HomeDir, ".config", "tor", fname)
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return ""
	}
	return string(b)
}
