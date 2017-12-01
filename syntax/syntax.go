package syntax

import (
	"regexp"
	"unicode/utf8"

	termbox "github.com/nsf/termbox-go"
)

func init() {
	Languages["go"] = Language{
		Syntax{"string", regexp.MustCompile(`^(?m)".*?(?:[^\\]"|$)`), termbox.ColorRed, termbox.ColorBlack},
		Syntax{"raw string", regexp.MustCompile(`^(?s)` + "`" + `.*?` + "(?:`|$)"), termbox.ColorRed, termbox.ColorBlack},
		Syntax{"rune", regexp.MustCompile(`^(?m)'.*?(?:[^\\]'|$)`), termbox.ColorYellow, termbox.ColorBlack},
		Syntax{"comment", regexp.MustCompile(`^(?m)//.*`), termbox.ColorMagenta, termbox.ColorBlack},
		Syntax{"multi line comment", regexp.MustCompile(`^(?s)/[*].*?(?:[*]/|$)`), termbox.ColorMagenta, termbox.ColorBlack},
		Syntax{"trailing spaces", regexp.MustCompile(`^(?m)[ \t]+$`), termbox.ColorBlack, termbox.ColorYellow},
		Syntax{"package", regexp.MustCompile(`^package\s`), termbox.ColorYellow, termbox.ColorBlack},
	}

	Languages["py"] = Language{
		Syntax{"multi line string1", regexp.MustCompile(`^(?s)""".*?(?:"""|$)`), termbox.ColorRed, termbox.ColorBlack},
		Syntax{"multi line string2", regexp.MustCompile(`^(?s)'''.*?(?:'''|$)`), termbox.ColorYellow, termbox.ColorBlack},
		Syntax{"string1", regexp.MustCompile(`^(?m)".*?(?:[^\\]"|$)`), termbox.ColorRed, termbox.ColorBlack},
		Syntax{"string2", regexp.MustCompile(`^(?m)'.*?(?:[^\\]'|$)`), termbox.ColorYellow, termbox.ColorBlack},
		Syntax{"comment", regexp.MustCompile(`^(?m)#.*`), termbox.ColorMagenta, termbox.ColorBlack},
		Syntax{"trailing spaces", regexp.MustCompile(`^(?m)[ \t]+$`), termbox.ColorBlack, termbox.ColorYellow},
	}
}

type Syntax struct {
	Name string
	Re   *regexp.Regexp
	Fg   termbox.Attribute
	Bg   termbox.Attribute
}

func (s Syntax) NewMatch(start, end Pos) Match {
	return Match{Name: s.Name, Start: start, End: end, Fg: s.Fg, Bg: s.Bg}
}

type Language []Syntax

var Languages = make(map[string]Language)

func (l Language) Parse(text []byte) []Match {
	c := NewCursor(text)
	matches := []Match{}
Loop:
	for {
		for _, syn := range l {
			ms := syn.Re.FindSubmatch(c.Remain())
			if ms == nil {
				continue
			}
			m := ms[0]
			if len(ms) == 2 {
				m = ms[1]
			}
			if string(m) == "" {
				continue
			}
			start := c.Pos()
			c.Skip(len(m))
			end := c.Pos()
			matches = append(matches, syn.NewMatch(start, end))
			continue Loop
		}
		if !c.Advance() {
			break
		}
	}
	return matches
}

// ParseRange checks and replace matches from min to max.
// When there is an overwrap between old matches and min Pos,
// it will recaculate matches from the overwrap begins.
func (l Language) ParseRange(matches []Match, text []byte, min, max Pos) []Match {
	// check where parse acutally started.
	last := -1
	overwrap := false
	for i, m := range matches {
		if m.Min().Compare(min) < 0 {
			if m.Max().Compare(min) < 0 {
				continue
			}
			overwrap = true
			last = i
			break
		}
		last = i
		break
	}

	parseStart := min
	if last != -1 {
		if overwrap {
			parseStart = matches[last].Min()
		}
		matches = matches[:last]
	}

	// move cursor to start position.
	c := NewCursor(text)
	for c.Pos().Compare(parseStart) < 0 {
		ok := c.Advance()
		if !ok {
			// already end of text. nothing to do.
			return matches
		}
	}

Loop:
	for {
		if c.Pos().Compare(max) >= 0 {
			break
		}
		for _, syn := range l {
			ms := syn.Re.FindSubmatch(c.Remain())
			if ms == nil {
				continue
			}
			m := ms[0]
			if len(ms) == 2 {
				m = ms[1]
			}
			if string(m) == "" {
				continue
			}
			start := c.Pos()
			c.Skip(len(m))
			end := c.Pos()
			matches = append(matches, syn.NewMatch(start, end))
			continue Loop
		}
		if !c.Advance() {
			break
		}
	}
	return matches
}

type Cursor struct {
	text []byte
	b    int // byte offset
	l    int // line offset
	o    int // byte in line offset
}

func NewCursor(text []byte) *Cursor {
	return &Cursor{text: text}
}

func (c *Cursor) Pos() Pos {
	return Pos{c.l, c.o}
}

func (c *Cursor) Remain() []byte {
	if c.l == len(c.text) {
		return []byte("")
	}
	return c.text[c.b:]
}

func (c *Cursor) Advance() bool {
	if c.b == len(c.text) {
		return false
	}
	c.next()
	return true
}

func (c *Cursor) Skip(b int) {
	i := 0
	for i < b {
		_, size := c.next()
		i += size
	}
}

func (c *Cursor) next() (r rune, size int) {
	r, size = utf8.DecodeRune(c.Remain())
	c.b += size
	c.o += size
	if r == '\n' {
		c.l += 1
		c.o = 0
	}
	return r, size
}

type Match struct {
	Name  string
	Start Pos
	End   Pos
	Fg    termbox.Attribute
	Bg    termbox.Attribute
}

func (m *Match) MinMax() (Pos, Pos) {
	if m.Start.L < m.End.L {
		return m.Start, m.End
	}
	if m.Start.L == m.End.L && m.Start.O <= m.End.O {
		return m.Start, m.End
	}
	return m.End, m.Start
}

func (m *Match) Min() Pos {
	min, _ := m.MinMax()
	return min
}

func (m *Match) Max() Pos {
	_, max := m.MinMax()
	return max
}

type Pos struct {
	L int
	O int
}

// Compare compares two Pos a and b.
// If a < b, it will return -1.
// If a > b, it will return 1.
// If a == b, it will return 0
func (a Pos) Compare(b Pos) int {
	if a.L < b.L {
		return -1
	}
	if a.L == b.L {
		if a.O < b.O {
			return -1
		}
		if a.O == b.O {
			return 0
		}
		return 1
	}
	return 1
}
