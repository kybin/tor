package main

import (
	"github.com/kybin/tor/cell"
	"github.com/kybin/tor/syntax"
	"github.com/mattn/go-runewidth"
	term "github.com/nsf/termbox-go"
)

func SetCell(l, o int, r rune, fg, bg term.Attribute) {
	term.SetCell(o, l, r, fg, bg)
}

// draw text inside of window at mainarea.
func drawScreen(w *Window, t *Text, sel *Selection, parser *syntax.Parser) {
	for l, ln := range t.lines {
		if l < w.min.L || l >= w.Max().L {
			continue
		}
		o := 0
		for b, r := range ln.data {
			if o >= w.Max().O {
				break
			}

			bg := term.ColorBlack
			fg := term.ColorWhite
			for _, m := range parser.Matches {
				if m.Range.Contains(cell.Pt{l, b}) {
					c := parser.Color(m.Name)
					bg = c.Bg
					fg = c.Fg
					break
				}
			}

			if sel.on && sel.Contains(cell.Pt{l, b}) {
				bg = term.ColorGreen
			}
			if r == '\t' {
				for i := 0; i < t.tabWidth; i++ {
					if o >= w.min.O {
						SetCell(l-w.min.L, o-w.min.O, rune(' '), fg, bg)
					}
					o += 1
				}
			} else {
				if o >= w.min.O {
					SetCell(l-w.min.L, o-w.min.O, rune(r), fg, bg)
				}
				o += runewidth.RuneWidth(r)
			}
		}
	}
}

// drawStatus draws current status of m at bottom of terminal.
// If m has Error, it will printed with red background.
func drawStatus(m Mode) {
	var bg term.Attribute
	var status string
	if m.Error() != "" {
		bg = term.ColorRed
		status = m.Error()
	} else {
		bg = term.ColorWhite
		status = m.Status()
	}

	termw, termh := term.Size()
	statusLine := termh - 1
	// clear and draw
	for i := 0; i < termw; i++ {
		SetCell(statusLine, i, ' ', term.ColorBlack, bg)
	}
	o := 0
	for _, r := range status {
		SetCell(statusLine, o, r, term.ColorBlack, bg)
		o += runewidth.RuneWidth(r)
	}
}
