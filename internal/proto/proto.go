package proto

import (
	"regexp"
)

type Syntax struct {
	Group    string
	Pattern  Pattern
	Keywords []string
	IsUser   bool
}

type Pattern struct {
	*regexp.Regexp
	canRun func(LineFacts) bool
}

// LineFacts contains the byte-level facts shared by all patterns on one line.
// Text is retained so predicates needing a literal marker do not rescan it.
type LineFacts struct {
	Text                                            string
	HasDigit, HasDot, HasSlash, HasColon, HasHyphen bool
	HasUnderscore, HasUpper, HasBackslash           bool
	HasSingleQuote, HasDoubleQuote                  bool
	HasSymbol                                       bool // !, @, #, $, %, ^, &, *, ;, ?, =
	DotCount, ColonCount, HyphenCount               int
	MaxSeparatorRun                                 int // -, =, #, *, <, >
	MaxHexRun                                       int // 0-9 a-f A-F
	MaxDigitRun                                     int
}

// MakeLineFacts computes the shared prerequisites for built-in pattern gates.
func MakeLineFacts(text string) LineFacts {
	var f LineFacts
	f.Text = text
	run, runChar := 0, byte(0)
	hexRun := 0
	digitRun := 0
	for i := 0; i < len(text); i++ {
		c := text[i]
		if c >= 'A' && c <= 'Z' {
			f.HasUpper = true
		}
		if c >= '0' && c <= '9' {
			f.HasDigit = true
			digitRun++
			if digitRun > f.MaxDigitRun {
				f.MaxDigitRun = digitRun
			}
		} else {
			digitRun = 0
		}
		if c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F' {
			hexRun++
			if hexRun > f.MaxHexRun {
				f.MaxHexRun = hexRun
			}
		} else {
			hexRun = 0
		}
		switch c {
		case '.':
			f.HasDot, f.DotCount = true, f.DotCount+1
		case '/':
			f.HasSlash = true
		case ':':
			f.HasColon, f.ColonCount = true, f.ColonCount+1
			f.HasSymbol = true
		case '-':
			f.HasHyphen, f.HyphenCount = true, f.HyphenCount+1
		case '_':
			f.HasUnderscore = true
		case '\\':
			f.HasBackslash = true
		case '\'':
			f.HasSingleQuote = true
		case '"':
			f.HasDoubleQuote = true
		case '!', '@', '#', '$', '%', '^', '&', '*', ';', '?', '=':
			f.HasSymbol = true
		}
		if c == runChar {
			run++
		} else {
			run, runChar = 1, c
		}
		for _, separator := range [...]byte{'-', '=', '#', '*', '<', '>'} {
			if c == separator && run > f.MaxSeparatorRun {
				f.MaxSeparatorRun = run
			}
		}
	}
	return f
}

type PatternGate func(LineFacts) bool

func (p Pattern) CanRun(facts LineFacts) bool {
	return p.canRun == nil || p.canRun(facts)
}

func MustCompileWithGate(pattern string, gate PatternGate) Pattern {
	return Pattern{Regexp: regexp.MustCompile(pattern), canRun: gate}
}

func (p *Pattern) UnmarshalText(text []byte) error {
	var err error
	p.Regexp, err = regexp.Compile(string(text))
	p.canRun = nil
	return err
}

func (p Pattern) HasValue() bool {
	return p.Regexp != nil
}

func MustCompile(pattern string) Pattern {
	return Pattern{Regexp: regexp.MustCompile(pattern)}
}

func MustCompileAll(patterns ...string) []Pattern {
	result := make([]Pattern, 0, len(patterns))
	for _, p := range patterns {
		result = append(result, MustCompile(p))
	}
	return result
}
