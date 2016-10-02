package main

import (
	"fmt"
	"strconv"
)

// Action is a user action.
// It also remember what are done by given action (Action.value).
type Action struct {
	kind         string
	value        string
	beforeCursor Cursor
	afterCursor  Cursor
}

func (a Action) String() string {
	bc := strconv.Itoa(a.beforeCursor.l) + ":" + strconv.Itoa(a.beforeCursor.b)
	ac := strconv.Itoa(a.afterCursor.l) + ":" + strconv.Itoa(a.afterCursor.b)
	return fmt.Sprintf("(%v, %v, %v, %v)", a.kind, a.value, bc, ac)
}

// History remembers what actions are done by user.
type History struct {
	head    int
	actions [][]*Action
}

// newHistory create a new History.
func newHistory() *History {
	return &History{
		head:    0,
		actions: make([][]*Action, 0),
	}
}

// Add adds action group to history.
func (h *History) Add(action []*Action) {
	h.actions = append(h.actions, action)
	h.head++
}

// Cut cuts history slice to given number and return how many were cut.
func (h *History) Cut(to int) int {
	b := len(h.actions)
	h.actions = h.actions[:to]
	h.head = len(h.actions)
	a := len(h.actions)
	return b - a
}

// Len returns length of action group slice.
func (h *History) Len() int {
	return len(h.actions)
}

// At returns action group of given index.
func (h *History) At(i int) []*Action {
	return h.actions[i]
}

// Last return last action group of history.
// If history doesn't have any action group, it will return nil.
func (h *History) Last() []*Action {
	if len(h.actions) == 0 {
		return nil
	}
	return h.actions[len(h.actions)-1]
}
