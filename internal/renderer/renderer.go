package renderer

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/madmaxieee/loglit/internal/config"
	"github.com/madmaxieee/loglit/internal/proto"
	"github.com/madmaxieee/loglit/internal/style"
	"github.com/madmaxieee/loglit/internal/theme"
)

type Renderer struct {
	Config config.Config
	Theme  theme.Theme
}

func New(cfg config.Config, th theme.Theme) (*Renderer, error) {
	renderer := &Renderer{
		Config: cfg,
		Theme:  th,
	}
	for _, hl := range cfg.Highlight {
		th.Insert(hl)
	}
	err := th.ResolveAllLinks()
	if err != nil {
		return nil, err
	}
	return renderer, nil
}

type Match struct {
	Start     int
	End       int
	AnsiStart string
}

func findMatches(syntaxList []proto.Syntax, highlights map[string]*style.Highlight, text string) ([]Match, error) {
	var matches []Match

	// find matches for regex
	for _, syn := range syntaxList {
		p := syn.Pattern
		if !p.HasValue() {
			continue
		}
		for _, idx := range p.FindAllStringIndex(text, -1) {
			hl, ok := highlights[syn.Group]
			if !ok {
				return []Match{}, fmt.Errorf("highlight group %s not found", syn.Group)
			}
			matches = append(matches, Match{
				Start:     idx[0],
				End:       idx[1],
				AnsiStart: hl.BuildAnsi(),
			})
		}
	}

	// find matches for keywords
	for _, syn := range syntaxList {
		for _, kw := range syn.Keywords {
			re := regexp.MustCompile(`\b` + regexp.QuoteMeta(kw) + `\b`)
			for _, idx := range re.FindAllStringIndex(text, -1) {
				hl, ok := highlights[syn.Group]
				if !ok {
					return []Match{}, fmt.Errorf("highlight group %s not found", syn.Group)
				}
				matches = append(matches, Match{
					Start:     idx[0],
					End:       idx[1],
					AnsiStart: hl.BuildAnsi(),
				})
			}
		}
	}

	return matches, nil
}

func (r Renderer) Render(text string) (string, error) {
	matches, err := findMatches(r.Config.BuiltInSyntax, r.Theme.HighlightMap, text)
	if err != nil {
		return text, err
	}

	// TODO: allow user matches to overlay on top of built-in matches
	userMatches, err := findMatches(r.Config.UserSyntax, r.Theme.HighlightMap, text)
	if err != nil {
		return text, err
	}
	matches = append(matches, userMatches...)

	if len(matches) == 0 {
		return text, nil
	}

	// resolve collisions
	var validMatches []Match
	// NOTE: later matches have higher priority
	for i := len(matches) - 1; i >= 0; i-- {
		match := matches[i]
		collision := false
		for _, existingMatch := range validMatches {
			if !(match.End <= existingMatch.Start || match.Start >= existingMatch.End) {
				collision = true
				break
			}
		}
		if !collision {
			validMatches = append(validMatches, match)
		}
	}

	// sort by start position
	sort.Slice(validMatches, func(i, j int) bool {
		return validMatches[i].Start < validMatches[j].Start
	})

	// build final string
	var b strings.Builder
	b.Grow(len(text) * 2)

	b.WriteString(text[:validMatches[0].Start])
	for i := range len(validMatches) {
		match := validMatches[i]
		b.WriteString(match.AnsiStart)
		b.WriteString(text[match.Start:match.End])
		b.WriteString(style.Reset)
		if i == len(validMatches)-1 {
			b.WriteString(text[match.End:])
		} else {
			nextMatch := validMatches[i+1]
			b.WriteString(text[match.End:nextMatch.Start])
		}
	}

	return b.String(), nil
}
