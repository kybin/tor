package syntax

import (
	"regexp"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/kybin/tor/cell"
)

type Type int

const (
	TypeKeyword = Type(iota)
	TypeString
	TypeRune
	TypeInt
	TypeComment
	TypeTrailingSpaces
)

// Byter could converted to []bytes.
type Byter interface {
	Bytes() []byte
}

// Parser is syntax parser.
type Parser struct {
	text        Byter
	textChanged bool
	lang        *Language
	Matches     []Match

	nextStart cell.Pt // next parse starting point, exclusive.
}

// NewParser creates a new Parser.
func NewParser(text Byter, ext string) *Parser {
	p := &Parser{}
	p.lang = NewLanguage(ext)
	p.SetText(text)
	return p
}

// SetText set it's text.
// After done this, first ParseTo will clear current matches
// and calculate matches from start.
func (p *Parser) SetText(text Byter) {
	p.text = text
	p.textChanged = true
}

// ParseTo calculates it's matches to pt.
// If a match started but not ended when reached to pt,
// it will continue parsing to the match's end.
func (p *Parser) ParseTo(pt cell.Pt) {
	if p.textChanged {
		p.nextStart = cell.Pt{0, 0}
		p.Matches = []Match{}
		p.textChanged = false
	}

	// move cursor to start position.
	c := NewCursor(p.text.Bytes())
	for c.Pos().Compare(p.nextStart) < 0 {
		ok := c.Advance()
		if !ok {
			// already end of text. nothing to do.
			return
		}
	}

	matches := []Match{}
Loop:
	for {
		if c.Pos().Compare(pt) >= 0 {
			break
		}
		for _, syn := range p.lang.syntaxes {
			ms := syn.Re.FindSubmatch(c.Remain())
			if ms == nil {
				continue
			}
			m := ms[0]
			if len(ms) == 2 {
				// if the match has subgroup, use a first one.
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
	p.Matches = append(p.Matches, matches...)
	p.nextStart = cell.Pt{0, 0}
	if len(p.Matches) != 0 {
		p.nextStart = p.Matches[len(p.Matches)-1].Range.Max()
	}
}

// ClearFrom clears it's match from pt.
// If there is an overwrap with a match,
// it will clear that match too.
func (p *Parser) ClearFrom(pt cell.Pt) {
	clip := 0
	for i, m := range p.Matches {
		if m.Range.Max().Compare(pt) < 0 {
			continue
		}
		clip = i
		break
	}
	if clip == 0 {
		p.nextStart = cell.Pt{0, 0}
		p.Matches = []Match{}
		return
	}
	p.nextStart = p.Matches[clip-1].Range.Max()
	p.Matches = p.Matches[:clip]
}

type Syntax struct {
	Name string
	Type Type
	Re   *regexp.Regexp
}

func (s Syntax) NewMatch(start, end cell.Pt) Match {
	return Match{Name: s.Name, Type: s.Type, Range: cell.Range{start, end}}
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

func (c *Cursor) Pos() cell.Pt {
	return cell.Pt{c.l, c.o}
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
	Type  Type
	Range cell.Range
}

type Attr struct {
	Fg tcell.Color
	Bg tcell.Color
}

type Theme map[Type]Attr

var DefaultTheme = Theme{
	TypeKeyword:        Attr{Fg: tcell.ColorYellow, Bg: tcell.ColorReset},
	TypeString:         Attr{Fg: tcell.ColorRed, Bg: tcell.ColorReset},
	TypeRune:           Attr{Fg: tcell.ColorYellow, Bg: tcell.ColorReset},
	TypeInt:            Attr{Fg: tcell.ColorReset, Bg: tcell.ColorReset},
	TypeComment:        Attr{Fg: tcell.ColorPurple, Bg: tcell.ColorReset},
	TypeTrailingSpaces: Attr{Fg: tcell.ColorReset, Bg: tcell.ColorYellow},
}
