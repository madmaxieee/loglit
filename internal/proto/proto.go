package proto

import (
	"regexp"
)

type Syntax struct {
	Group    string
	Pattern  Pattern
	Keywords []string
}

type Pattern struct {
	*regexp.Regexp
}

func (p *Pattern) UnmarshalText(text []byte) error {
	var err error
	p.Regexp, err = regexp.Compile(string(text))
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
