package syntax

import (
	"regexp"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/kybin/tor/cell"
)

var Languages = make(map[string]*Language)

func color(fg, bg tcell.Color) tcell.Style {
	return tcell.StyleDefault.Foreground(fg).Background(bg)
}

type Type int

const (
	TypeKeyword = Type(iota)
	TypeString
	TypeRune
	TypeInt
	TypeComment
	TypeTrailingSpaces
)

func init() {
	def := NewLanguage(false, 4)
	def.AddSyntax(Syntax{"trailing spaces", TypeTrailingSpaces, regexp.MustCompile(`^(?m)[ \t]+$`)})
	Languages[""] = def

	golang := NewLanguage(false, 4)
	golang.AddSyntax(Syntax{"string", TypeString, regexp.MustCompile(`^(?m)".*?(?:[^\\]?"|$)`)})
	golang.AddSyntax(Syntax{"raw string", TypeString, regexp.MustCompile(`^(?s)` + "`" + `.*?` + "(?:`|$)")})
	golang.AddSyntax(Syntax{"rune", TypeRune, regexp.MustCompile(`^(?m)'.*?(?:[^\\]?'|$)`)})
	golang.AddSyntax(Syntax{"comment", TypeComment, regexp.MustCompile(`^(?m)//.*`)})
	golang.AddSyntax(Syntax{"multi line comment", TypeComment, regexp.MustCompile(`^(?s)/[*].*?(?:[*]/|$)`)})
	golang.AddSyntax(Syntax{"trailing spaces", TypeTrailingSpaces, regexp.MustCompile(`^(?m)[ \t]+$`)})
	golang.AddSyntax(Syntax{"package", TypeKeyword, regexp.MustCompile(`^package\s`)})
	Languages["go"] = golang

	py := NewLanguage(false, 4)
	py.AddSyntax(Syntax{"multi line string1", TypeString, regexp.MustCompile(`^(?s)""".*?(?:"""|$)`)})
	py.AddSyntax(Syntax{"multi line string2", TypeString, regexp.MustCompile(`^(?s)'''.*?(?:'''|$)`)})
	py.AddSyntax(Syntax{"string1", TypeString, regexp.MustCompile(`^(?m)".*?(?:[^\\]?"|$)`)})
	py.AddSyntax(Syntax{"string2", TypeString, regexp.MustCompile(`^(?m)'.*?(?:[^\\]?'|$)`)})
	py.AddSyntax(Syntax{"comment", TypeComment, regexp.MustCompile(`^(?m)#.*`)})
	py.AddSyntax(Syntax{"trailing spaces", TypeTrailingSpaces, regexp.MustCompile(`^(?m)[ \t]+$`)})
	Languages["py"] = py

	ts := NewLanguage(true, 2)
	ts.AddSyntax(Syntax{"raw string", TypeString, regexp.MustCompile(`^(?s)` + "`" + `.*?` + "(?:`|$)")})
	ts.AddSyntax(Syntax{"string1", TypeString, regexp.MustCompile(`^(?m)".*?(?:[^\\]?"|$)`)})
	ts.AddSyntax(Syntax{"string2", TypeString, regexp.MustCompile(`^(?m)'.*?(?:[^\\]?'|$)`)})
	ts.AddSyntax(Syntax{"comment", TypeComment, regexp.MustCompile(`^(?m)//.*`)})
	ts.AddSyntax(Syntax{"trailing spaces", TypeTrailingSpaces, regexp.MustCompile(`^(?m)[ \t]+$`)})
	ts.AddSyntax(Syntax{"keywords", TypeKeyword, regexp.MustCompile(`^(import|export)\s`)})
	Languages["ts"] = ts

	elm := NewLanguage(true, 2)
	elm.AddSyntax(Syntax{"trailing spaces", TypeTrailingSpaces, regexp.MustCompile(`^(?m)[ \t]+$`)})
	Languages["elm"] = elm
}

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
func NewParser(text Byter, langName string) *Parser {
	p := &Parser{}
	lang, ok := Languages[langName]
	if ok {
		p.lang = lang
	} else {
		p.lang = Languages[""]
	}
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

type Language struct {
	TabToSpace bool
	TabWidth   int
	syntaxes   []Syntax // should be ordered
}

func NewLanguage(tabToSpace bool, tabWidth int) *Language {
	return &Language{
		TabToSpace: tabToSpace,
		TabWidth:   tabWidth,
		syntaxes:   []Syntax{},
	}
}

func (l *Language) AddSyntax(s Syntax) {
	l.syntaxes = append(l.syntaxes, s)
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
