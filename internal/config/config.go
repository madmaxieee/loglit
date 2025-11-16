package config

import "regexp"

type Config struct {
	Syntax    []Syntax
	Highlight []Highlight
}

type Syntax struct {
	Group    string
	Patterns []Pattern
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

type Highlight struct {
	Group string
	Color string
}
