package main

import (
	"strings"
)

// toggleComment toggles line comment.
// It will comment lines when none of the lines are commented.
// It will uncomment lines when at least one of the lines are commented.
func toggleComment(comment string, t *Text, c *Cursor, sel *Selection) []*Action {
	lns := make([]int, 0)
	if sel.on {
		lns = sel.Lines()
	} else {
		lns = append(lns, c.l)
	}

	commentedLns := make([]int, 0)
	for _, l :=  range lns {
		if strings.HasPrefix(t.lines[l].data, comment + " ") {
			commentedLns = append(commentedLns, l)
			break
		}
	}

	if len(commentedLns) > 0 {
		for _, l := range commentedLns {
			t.lines[l].data = strings.Replace(t.lines[l].data, comment + " ", "", 1)
		}
	} else {
		for _, l := range lns {
			t.lines[l].data = comment + " " + t.lines[l].data
		}
	}

	// TODO: return actions instead modify it.
	return nil
}
