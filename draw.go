package main

import (
	"github.com/kybin/tor/syntax"
	"github.com/mattn/go-runewidth"
	term "github.com/nsf/termbox-go"
)

func SetCell(l, o int, r rune, fg, bg term.Attribute) {
	term.SetCell(o, l, r, fg, bg)
}

func clearScreen(ar *Area) {
	term.Clear(term.ColorDefault, term.ColorDefault)
}

func resizeScreen(ar *Area, win *Window, w, h int) {
	min := ar.min
	*ar = Area{min, Point{min.l + h, min.o + w}}
	win.size = ar.Size()
}

// draw text inside of window at mainarea.
func drawScreen(ar *Area, w *Window, t *Text, sel *Selection, lang syntax.Language) {
	var matches []syntax.Match
	if lang != nil {
		matches = lang.Parse(t.Bytes())
	}

	for l, ln := range t.lines {
		if l < w.min.l || l >= w.Max().l {
			continue
		}
		o := 0
		for b, r := range ln.data {
			if o >= w.Max().o {
				break
			}

			bg := term.ColorBlack
			fg := term.ColorWhite
			if matches != nil {
				for _, m := range matches {
					start := Point{m.Start.L, m.Start.O}
					end := Point{m.End.L, m.End.O}
					rng := &Range{start, end}
					if rng.Contains(Point{l, b}) {
						bg = m.Bg
						fg = m.Fg
						break
					}
				}
			}
			if sel.on && sel.Contains(Point{l, b}) {
				bg = term.ColorGreen
			}
			if r == '\t' {
				for i := 0; i < t.tabWidth; i++ {
					if o >= w.min.o {
						SetCell(l-w.min.l+ar.min.l, o-w.min.o+ar.min.o, rune(' '), fg, bg)
					}
					o += 1
				}
			} else {
				if o >= w.min.o {
					SetCell(l-w.min.l+ar.min.l, o-w.min.o+ar.min.o, rune(r), fg, bg)
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
