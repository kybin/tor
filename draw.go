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

func resizeScreen(win *Window, w, h int) {
	win.size = cell.Pt{h, w}
}

// draw text inside of window at mainarea.
func drawScreen(w *Window, t *Text, sel *Selection, lang *syntax.Language, syntaxMatches []syntax.Match) {
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
			if syntaxMatches != nil {
				for _, m := range syntaxMatches {
					if m.Range.Contains(cell.Pt{l, b}) {
						c := lang.Color(m.Name)
						bg = c.Bg
						fg = c.Fg
						break
					}
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

func printStatus(status string) {
	termw, termh := term.Size()
	statusLine := termh - 1
	// clear
	for i := 0; i < termw; i++ {
		SetCell(statusLine, i, ' ', term.ColorBlack, term.ColorWhite)
	}
	// draw
	o := 0
	for _, r := range status {
		SetCell(statusLine, o, r, term.ColorBlack, term.ColorWhite)
		o += runewidth.RuneWidth(r)
	}
}

func printErrorStatus(err string) {
	termw, termh := term.Size()
	statusLine := termh - 1
	// clear
	for i := 0; i < termw; i++ {
		SetCell(statusLine, i, ' ', term.ColorBlack, term.ColorRed)
	}
	// draw
	o := 0
	for _, r := range err {
		SetCell(statusLine, o, r, term.ColorBlack, term.ColorRed)
		o += runewidth.RuneWidth(r)
	}
}
