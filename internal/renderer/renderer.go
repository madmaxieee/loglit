package renderer

import (
	"fmt"
	"regexp"
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

var keywordRegexCache map[string]*regexp.Regexp = make(map[string]*regexp.Regexp)

func New(cfg config.Config, th theme.Theme) (*Renderer, error) {
	renderer := &Renderer{
		Config: cfg,
		Theme:  th,
	}
	precompileKeywordRegex(cfg.BuiltInSyntax)
	precompileKeywordRegex(cfg.UserSyntax)
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
	AnsiEnd   string
}

func precompileKeywordRegex(syntaxList []proto.Syntax) {
	for _, syn := range syntaxList {
		for _, kw := range syn.Keywords {
			if _, exists := keywordRegexCache[kw]; !exists {
				re := regexp.MustCompile(`\b` + regexp.QuoteMeta(kw) + `\b`)
				keywordRegexCache[kw] = re
			}
		}
	}
}

func findMatches(syntaxList []proto.Syntax, highlights map[string]*style.Highlight, text string) (MatchLayer, error) {
	var matches MatchLayer

	// find matches for regex
	for _, syn := range syntaxList {
		p := syn.Pattern
		if !p.HasValue() {
			continue
		}
		for _, idx := range p.FindAllStringIndex(text, -1) {
			hl, ok := highlights[syn.Group]
			if !ok {
				return MatchLayer{}, fmt.Errorf("highlight group %s not found", syn.Group)
			}
			matches = append(matches, Match{
				Start:     idx[0],
				End:       idx[1],
				AnsiStart: hl.BuildAnsi(),
				AnsiEnd:   hl.BuildAnsiReset(),
			})
		}
	}

	// find matches for keywords
	for _, syn := range syntaxList {
		for _, kw := range syn.Keywords {
			re, ok := keywordRegexCache[kw]
			if !ok {
				return MatchLayer{}, fmt.Errorf("keyword regex for '%s' not found", kw)
			}
			for _, idx := range re.FindAllStringIndex(text, -1) {
				hl, ok := highlights[syn.Group]
				if !ok {
					return MatchLayer{}, fmt.Errorf("highlight group %s not found", syn.Group)
				}
				matches = append(matches, Match{
					Start:     idx[0],
					End:       idx[1],
					AnsiStart: hl.BuildAnsi(),
					AnsiEnd:   hl.BuildAnsiReset(),
				})
			}
		}
	}

	return matches, nil
}

func (r Renderer) Render(text string) (string, error) {
	builtInMatches, err := findMatches(r.Config.BuiltInSyntax, r.Theme.HighlightMap, text)
	if err != nil {
		return text, err
	}
	builtInMatches.removeOverlaps().Sort()

	userMatches, err := findMatches(r.Config.UserSyntax, r.Theme.HighlightMap, text)
	if err != nil {
		return text, err
	}
	userMatches.removeOverlaps().Sort()

	var matches MatchLayer
	if userMatches.Len() > 0 {
		matches = Stack(userMatches, builtInMatches)
		userBgHighlight, ok := r.Theme.HighlightMap["UserMatchLineBackground"]
		if !ok {
			return text, fmt.Errorf("highlight group %q not found", "UserMatchLineBackground")
		}
		matches = Stack(matches, MatchLayer{{
			Start:     0,
			End:       len(text),
			AnsiStart: userBgHighlight.BuildAnsi(),
			AnsiEnd:   userBgHighlight.BuildAnsiReset(),
		}})
	} else {
		matches = builtInMatches
	}

	if len(matches) == 0 {
		return text, nil
	}

	matches.Sort()

	// build final string
	var b strings.Builder
	b.Grow(len(text) * 2)

	b.WriteString(text[:matches[0].Start])
	for i := range len(matches) {
		match := matches[i]
		b.WriteString(match.AnsiStart)
		b.WriteString(text[match.Start:match.End])
		b.WriteString(match.AnsiEnd)
		if i == len(matches)-1 {
			b.WriteString(text[match.End:])
		} else {
			nextMatch := matches[i+1]
			b.WriteString(text[match.End:nextMatch.Start])
		}
	}
	b.WriteString(style.ResetAllAnsi)

	return b.String(), nil
}
