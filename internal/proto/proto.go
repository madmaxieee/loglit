package proto

import (
	"regexp"

	"github.com/madmaxieee/loglit/internal/style"
	"github.com/madmaxieee/loglit/internal/utils"
	"github.com/pelletier/go-toml/v2"
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

type Highlight struct {
	Group     string
	Link      *string
	Fg        *string
	Bg        *string
	Italic    bool
	Bold      bool
	Underline bool
}

func (h *Highlight) UnmarshalText(text []byte) error {
	err := toml.Unmarshal(text, h)
	if err != nil {
		return err
	}
	if h.Link == nil {
		if h.Fg != nil {
			h.Fg = utils.Ptr(style.FgHex(*h.Fg))
		}
		if h.Bg != nil {
			h.Bg = utils.Ptr(style.BgHex(*h.Bg))
		}
	} else {
		h.Fg = nil
		h.Bg = nil
	}
	return nil
}
