package main

import (
	"fmt"
	"errors"
	"strconv"
)

type Action struct {
	kind string
	value string
	beforeCursor Cursor
	afterCursor Cursor
}

func (a Action) String() string {
	bc := strconv.Itoa(a.beforeCursor.l)+":"+strconv.Itoa(a.beforeCursor.b)
	ac := strconv.Itoa(a.afterCursor.l)+":"+strconv.Itoa(a.afterCursor.b)
	return fmt.Sprintf("%v, %v, %v, %v", a.kind, a.value, bc, ac)
}

type History struct {
	head int
	actions []*Action
}

func newHistory() *History{
	return &History{
		head:0,
		actions:make([]*Action, 0),
	}
}

func (h *History) Cut(to int) {
	h.actions = h.actions[:to]
}

func (h *History) Len() int {
	return len(h.actions)
}

func (h *History) At(i int) *Action {
	return h.actions[i]
}

func (h *History) Last() *Action {
	// TODO : last from head? or from h?
	if len(h.actions) == 0 {
		return nil
	}
	return h.actions[len(h.actions)-1]
}

func (h *History) Pop() (*Action, error) {
	// TODO : last from head? or from h?
	if len(h.actions) == 0 {
		return nil, errors.New("empty undo stack")
	}
	last := h.actions[len(h.actions)-1]
	h.actions = h.actions[0:len(h.actions)-1]
	return last, nil
}

func (h *History) RemoveLast() (error) {
	// TODO : last from head? or from h?
	if len(h.actions) == 0 {
		return errors.New("empty undo stack")
	}
	h.actions = h.actions[0:len(h.actions)-1]
	return nil
}

func (h *History) Add(action *Action) {
	h.actions = append(h.actions, action)
}
