package syntax

import (
	"regexp"
	"unicode/utf8"

	"github.com/kybin/tor/cell"
	term "github.com/nsf/termbox-go"
)

var Languages = make(map[string]*Language)

func init() {
	golang := NewLanguage()
	golang.AddSyntax(Syntax{"string", regexp.MustCompile(`^(?m)".*?(?:[^\\]?"|$)`)}, Color{term.ColorRed, term.ColorBlack})
	golang.AddSyntax(Syntax{"raw string", regexp.MustCompile(`^(?s)` + "`" + `.*?` + "(?:`|$)")}, Color{term.ColorRed, term.ColorBlack})
	golang.AddSyntax(Syntax{"rune", regexp.MustCompile(`^(?m)'.*?(?:[^\\]?'|$)`)}, Color{term.ColorYellow, term.ColorBlack})
	golang.AddSyntax(Syntax{"comment", regexp.MustCompile(`^(?m)//.*`)}, Color{term.ColorMagenta, term.ColorBlack})
	golang.AddSyntax(Syntax{"multi line comment", regexp.MustCompile(`^(?s)/[*].*?(?:[*]/|$)`)}, Color{term.ColorMagenta, term.ColorBlack})
	golang.AddSyntax(Syntax{"trailing spaces", regexp.MustCompile(`^(?m)[ \t]+$`)}, Color{term.ColorBlack, term.ColorYellow})
	golang.AddSyntax(Syntax{"package", regexp.MustCompile(`^package\s`)}, Color{term.ColorYellow, term.ColorBlack})
	Languages["go"] = golang

	py := NewLanguage()
	py.AddSyntax(Syntax{"multi line string1", regexp.MustCompile(`^(?s)""".*?(?:"""|$)`)}, Color{term.ColorRed, term.ColorBlack})
	py.AddSyntax(Syntax{"multi line string2", regexp.MustCompile(`^(?s)'''.*?(?:'''|$)`)}, Color{term.ColorYellow, term.ColorBlack})
	py.AddSyntax(Syntax{"string1", regexp.MustCompile(`^(?m)".*?(?:[^\\]?"|$)`)}, Color{term.ColorRed, term.ColorBlack})
	py.AddSyntax(Syntax{"string2", regexp.MustCompile(`^(?m)'.*?(?:[^\\]?'|$)`)}, Color{term.ColorYellow, term.ColorBlack})
	py.AddSyntax(Syntax{"comment", regexp.MustCompile(`^(?m)#.*`)}, Color{term.ColorMagenta, term.ColorBlack})
	py.AddSyntax(Syntax{"trailing spaces", regexp.MustCompile(`^(?m)[ \t]+$`)}, Color{term.ColorBlack, term.ColorYellow})
	Languages["py"] = py
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
}

// NewParser creates a new Parser.
func NewParser(text Byter, langName string) *Parser {
	p := &Parser{}
	lang, ok := Languages[langName]
	if ok {
		p.lang = lang
	}
	p.SetText(text)
	return p
}

// SetText set it's text.
// After doing this, it should re-Parse-d entirely.
// Until then, TextChanged will return true.
func (p *Parser) SetText(text Byter) {
	p.text = text
	p.textChanged = true
}

// TextChanged returns status whether
// it's text changed but not Parsed yet.
func (p *Parser) TextChanged() bool {
	return p.textChanged
}

// Parse calculates it's matches entirely.
func (p *Parser) Parse() {
	if p.lang != nil {
		p.Matches = p.lang.Parse(p.text.Bytes())
	}
	p.textChanged = false
}

// Parse calulate it's partial matches.
// It will re-parse text from min to max range and
// replace current matches if there is an overwrap.
func (p *Parser) ParseRange(min, max cell.Pt) {
	if p.lang != nil {
		p.Matches = p.lang.ParseRange(p.Matches, p.text.Bytes(), min, max)
	}
}

// Color returns any match's syntax color name.
func (p *Parser) Color(synName string) Color {
	return p.lang.Color(synName)
}

type Syntax struct {
	Name string
	Re   *regexp.Regexp
}

func (s Syntax) NewMatch(start, end cell.Pt) Match {
	return Match{Name: s.Name, Range: cell.Range{start, end}}
}

type Color struct {
	Fg term.Attribute
	Bg term.Attribute
}

type Language struct {
	syntaxes []Syntax // should be ordered
	colors   map[string]Color
}

func NewLanguage() *Language {
	return &Language{
		syntaxes: []Syntax{},
		colors:   make(map[string]Color),
	}
}

func (l *Language) AddSyntax(s Syntax, c Color) {
	l.syntaxes = append(l.syntaxes, s)
	l.colors[s.Name] = c
}

func (l *Language) Color(synName string) Color {
	c, ok := l.colors[synName]
	if !ok {
		return Color{term.ColorWhite, term.ColorBlack}
	}
	return c
}

func (l *Language) Parse(text []byte) []Match {
	c := NewCursor(text)
	matches := []Match{}
Loop:
	for {
		for _, syn := range l.syntaxes {
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
// When there is an overwrap between old matches and min position,
// it will recaculate matches from the overwrap begins.
func (l *Language) ParseRange(matches []Match, text []byte, min, max cell.Pt) []Match {
	// check where parse acutally started.
	last := -1
	overwrap := false
	for i, m := range matches {
		if m.Range.Min().Compare(min) < 0 {
			if m.Range.Max().Compare(min) < 0 {
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
			parseStart = matches[last].Range.Min()
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
		for _, syn := range l.syntaxes {
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
}
