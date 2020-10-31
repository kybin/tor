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

func init() {
	def := NewLanguage(false, 4)
	def.AddSyntax(Syntax{"trailing spaces", regexp.MustCompile(`^(?m)[ \t]+$`), color(tcell.ColorBlack, tcell.ColorYellow)})
	Languages[""] = def

	golang := NewLanguage(false, 4)
	golang.AddSyntax(Syntax{"string", regexp.MustCompile(`^(?m)".*?(?:[^\\]?"|$)`), color(tcell.ColorRed, tcell.ColorBlack)})
	golang.AddSyntax(Syntax{"raw string", regexp.MustCompile(`^(?s)` + "`" + `.*?` + "(?:`|$)"), color(tcell.ColorRed, tcell.ColorBlack)})
	golang.AddSyntax(Syntax{"rune", regexp.MustCompile(`^(?m)'.*?(?:[^\\]?'|$)`), color(tcell.ColorYellow, tcell.ColorBlack)})
	golang.AddSyntax(Syntax{"comment", regexp.MustCompile(`^(?m)//.*`), color(tcell.ColorPurple, tcell.ColorBlack)})
	golang.AddSyntax(Syntax{"multi line comment", regexp.MustCompile(`^(?s)/[*].*?(?:[*]/|$)`), color(tcell.ColorPurple, tcell.ColorBlack)})
	golang.AddSyntax(Syntax{"trailing spaces", regexp.MustCompile(`^(?m)[ \t]+$`), color(tcell.ColorBlack, tcell.ColorYellow)})
	golang.AddSyntax(Syntax{"package", regexp.MustCompile(`^package\s`), color(tcell.ColorYellow, tcell.ColorBlack)})
	Languages["go"] = golang

	py := NewLanguage(false, 4)
	py.AddSyntax(Syntax{"multi line string1", regexp.MustCompile(`^(?s)""".*?(?:"""|$)`), color(tcell.ColorRed, tcell.ColorBlack)})
	py.AddSyntax(Syntax{"multi line string2", regexp.MustCompile(`^(?s)'''.*?(?:'''|$)`), color(tcell.ColorYellow, tcell.ColorBlack)})
	py.AddSyntax(Syntax{"string1", regexp.MustCompile(`^(?m)".*?(?:[^\\]?"|$)`), color(tcell.ColorRed, tcell.ColorBlack)})
	py.AddSyntax(Syntax{"string2", regexp.MustCompile(`^(?m)'.*?(?:[^\\]?'|$)`), color(tcell.ColorYellow, tcell.ColorBlack)})
	py.AddSyntax(Syntax{"comment", regexp.MustCompile(`^(?m)#.*`), color(tcell.ColorPurple, tcell.ColorBlack)})
	py.AddSyntax(Syntax{"trailing spaces", regexp.MustCompile(`^(?m)[ \t]+$`), color(tcell.ColorBlack, tcell.ColorYellow)})
	Languages["py"] = py

	ts := NewLanguage(true, 2)
	ts.AddSyntax(Syntax{"raw string", regexp.MustCompile(`^(?s)` + "`" + `.*?` + "(?:`|$)"), color(tcell.ColorRed, tcell.ColorBlack)})
	ts.AddSyntax(Syntax{"string1", regexp.MustCompile(`^(?m)".*?(?:[^\\]?"|$)`), color(tcell.ColorRed, tcell.ColorBlack)})
	ts.AddSyntax(Syntax{"string2", regexp.MustCompile(`^(?m)'.*?(?:[^\\]?'|$)`), color(tcell.ColorYellow, tcell.ColorBlack)})
	ts.AddSyntax(Syntax{"comment", regexp.MustCompile(`^(?m)//.*`), color(tcell.ColorPurple, tcell.ColorBlack)})
	ts.AddSyntax(Syntax{"trailing spaces", regexp.MustCompile(`^(?m)[ \t]+$`), color(tcell.ColorBlack, tcell.ColorYellow)})
	ts.AddSyntax(Syntax{"keywords", regexp.MustCompile(`^(import|export)\s`), color(tcell.ColorYellow, tcell.ColorBlack)})
	Languages["ts"] = ts

	elm := NewLanguage(true, 2)
	elm.AddSyntax(Syntax{"trailing spaces", regexp.MustCompile(`^(?m)[ \t]+$`), color(tcell.ColorBlack, tcell.ColorYellow)})
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
	Name  string
	Re    *regexp.Regexp
	Style tcell.Style
}

func (s Syntax) NewMatch(start, end cell.Pt) Match {
	return Match{Name: s.Name, Range: cell.Range{start, end}, Style: s.Style}
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
	Range cell.Range
	Style tcell.Style
}
