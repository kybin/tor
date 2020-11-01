package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/kybin/tor/cell"
	"github.com/kybin/tor/syntax"
	"github.com/mattn/go-runewidth"
)

func SetCell(s tcell.Screen, l, o int, r rune, style tcell.Style) {
	s.SetContent(o, l, r, nil, style)
}

// draw text inside of window at mainarea.
func drawScreen(s tcell.Screen, norm *NormalMode) {
	w := norm.area.Win
	// parse syntax
	if norm.dirty {
		norm.parser.ClearFrom(cell.Pt{L: w.Min().L, O: 0})
		norm.dirty = false
	}
	norm.parser.ParseTo(cell.Pt{L: w.Max().L + 1, O: 0})

	// draw
	for l, ln := range norm.text.lines {
		if l < w.Min().L || l >= w.Max().L {
			continue
		}
		origStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
		o := 0
		for b, r := range ln.data {
			if o >= w.Max().O {
				break
			}

			style := origStyle
			for _, m := range norm.parser.Matches {
				if m.Range.Contains(cell.Pt{l, b}) {
					attr, ok := syntax.DefaultTheme[m.Type]
					if ok {
						style = tcell.StyleDefault.Background(attr.Bg).Foreground(attr.Fg)
					}
					break
				}
			}

			if norm.selection.Contains(cell.Pt{l, b}) {
				style = tcell.StyleDefault.Background(tcell.ColorGreen).Foreground(tcell.ColorReset)
			}
			if r == '\t' {
				for i := 0; i < norm.text.tabWidth; i++ {
					if o >= w.Min().O {
						SetCell(s, l-w.Min().L, o-w.Min().O+norm.area.min.O, rune(' '), style)
					}
					o += 1
				}
			} else {
				if o >= w.Min().O {
					SetCell(s, l-w.Min().L, o-w.Min().O+norm.area.min.O, rune(r), style)
				}
				o += runewidth.RuneWidth(r)
			}
		}
		// set original color to the last cell. (white and black)
		// if not set, the cursor's color will look different.
		SetCell(s, l-w.Min().L, o-w.Min().O+norm.area.min.O, rune(' '), origStyle)
	}
}

// drawStatus draws current status of m at bottom of terminal.
// If m has Error, it will printed with red background.
func drawStatus(s tcell.Screen, m Mode) {
	var style tcell.Style
	var status string
	if m.Error() != "" {
		style = tcell.StyleDefault.Background(tcell.ColorRed).Foreground(tcell.ColorBlack)
		status = m.Error()
	} else {
		style = tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack)
		status = m.Status()
	}

	termw, termh := s.Size()
	statusLine := termh - 1
	// clear and draw
	for i := 0; i < termw; i++ {
		SetCell(s, statusLine, i, ' ', style)
	}
	o := 0
	for _, r := range status {
		SetCell(s, statusLine, o, r, style)
		o += runewidth.RuneWidth(r)
	}
}
