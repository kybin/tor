package syntax

import (
	"regexp"
)

// langGenerator is a collection of language generators.
var langGenerator map[string]func() *Language

// NewLanguage finds a language from given extension.
// It will return "unknown" language if it didn't find language for the extension.
func NewLanguage(ext string) *Language {
	gen, ok := langGenerator[ext]
	if ok {
		lang := gen()
		return lang
	}
	return unknownLanguage()
}

// Language is a set of tab configurations and syntaxes.
type Language struct {
	TabToSpace bool
	TabWidth   int
	// syntaxes is not a map, because highlighting is affected by syntax order
	syntaxes []Syntax
}

// newLanguage creates a new language with the tab configurations.
// Syntax should be added with AddSyntax to this language.
func newLanguage(tabToSpace bool, tabWidth int) *Language {
	return &Language{
		TabToSpace: tabToSpace,
		TabWidth:   tabWidth,
		syntaxes:   []Syntax{},
	}
}

// AddSyntax adds a syntax to the language.
// It will become more useful when we really split syntax for languages
// into their own configuration files.
func (l *Language) AddSyntax(s Syntax) {
	l.syntaxes = append(l.syntaxes, s)
}

// unknownLanguage is a fallback language.
func unknownLanguage() *Language {
	def := newLanguage(false, 4)
	def.AddSyntax(Syntax{"trailing spaces", TypeTrailingSpaces, regexp.MustCompile(`^(?m)[ \t]+$`)})
	return def
}

func init() {
	langGenerator = make(map[string]func() *Language)

	langGenerator["go"] = func() *Language {
		golang := newLanguage(false, 4)
		golang.AddSyntax(Syntax{"string", TypeString, regexp.MustCompile(`^(?m)".*?(?:[^\\]?"|$)`)})
		golang.AddSyntax(Syntax{"raw string", TypeString, regexp.MustCompile(`^(?s)` + "`" + `.*?` + "(?:`|$)")})
		golang.AddSyntax(Syntax{"rune", TypeRune, regexp.MustCompile(`^(?m)'.*?(?:[^\\]?'|$)`)})
		golang.AddSyntax(Syntax{"comment", TypeComment, regexp.MustCompile(`^(?m)//.*`)})
		golang.AddSyntax(Syntax{"multi line comment", TypeComment, regexp.MustCompile(`^(?s)/[*].*?(?:[*]/|$)`)})
		golang.AddSyntax(Syntax{"trailing spaces", TypeTrailingSpaces, regexp.MustCompile(`^(?m)[ \t]+$`)})
		golang.AddSyntax(Syntax{"package", TypeKeyword, regexp.MustCompile(`^package\s`)})
		return golang
	}

	langGenerator["py"] = func() *Language {
		py := newLanguage(false, 4)
		py.AddSyntax(Syntax{"multi line string1", TypeString, regexp.MustCompile(`^(?s)""".*?(?:"""|$)`)})
		py.AddSyntax(Syntax{"multi line string2", TypeString, regexp.MustCompile(`^(?s)'''.*?(?:'''|$)`)})
		py.AddSyntax(Syntax{"string1", TypeString, regexp.MustCompile(`^(?m)".*?(?:[^\\]?"|$)`)})
		py.AddSyntax(Syntax{"string2", TypeString, regexp.MustCompile(`^(?m)'.*?(?:[^\\]?'|$)`)})
		py.AddSyntax(Syntax{"comment", TypeComment, regexp.MustCompile(`^(?m)#.*`)})
		py.AddSyntax(Syntax{"trailing spaces", TypeTrailingSpaces, regexp.MustCompile(`^(?m)[ \t]+$`)})
		return py
	}

	langGenerator["ts"] = func() *Language {
		ts := newLanguage(true, 2)
		ts.AddSyntax(Syntax{"raw string", TypeString, regexp.MustCompile(`^(?s)` + "`" + `.*?` + "(?:`|$)")})
		ts.AddSyntax(Syntax{"string1", TypeString, regexp.MustCompile(`^(?m)".*?(?:[^\\]?"|$)`)})
		ts.AddSyntax(Syntax{"string2", TypeString, regexp.MustCompile(`^(?m)'.*?(?:[^\\]?'|$)`)})
		ts.AddSyntax(Syntax{"comment", TypeComment, regexp.MustCompile(`^(?m)//.*`)})
		ts.AddSyntax(Syntax{"trailing spaces", TypeTrailingSpaces, regexp.MustCompile(`^(?m)[ \t]+$`)})
		ts.AddSyntax(Syntax{"keywords", TypeKeyword, regexp.MustCompile(`^(import|export)\s`)})
		return ts
	}

	langGenerator["elm"] = func() *Language {
		elm := newLanguage(true, 2)
		elm.AddSyntax(Syntax{"trailing spaces", TypeTrailingSpaces, regexp.MustCompile(`^(?m)[ \t]+$`)})
		return elm
	}
}
