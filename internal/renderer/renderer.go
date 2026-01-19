package renderer

import (
	"fmt"
	"strings"

	"github.com/madmaxieee/loglit/internal/config"
	"github.com/madmaxieee/loglit/internal/proto"
	"github.com/madmaxieee/loglit/internal/style"
	"github.com/madmaxieee/loglit/internal/theme"
)

type keywordMap map[string]*style.Highlight

type Renderer struct {
	Config config.Config
	Theme  theme.Theme

	builtinLowerKeywordMap keywordMap
	builtinKeywordMap      keywordMap
	userKeywordMap         keywordMap
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

	// precompile keyword map
	renderer.builtinLowerKeywordMap = make(keywordMap)
	for _, syntax := range renderer.Config.BuiltInSyntaxLower {
		for _, keyword := range syntax.Keywords {
			hl, ok := th.HighlightMap[syntax.Group]
			if !ok {
				return nil, fmt.Errorf("highlight group %q not found", syntax.Group)
			}
			renderer.builtinLowerKeywordMap[keyword] = hl
		}
	}

	renderer.builtinKeywordMap = make(keywordMap)
	for _, syntax := range renderer.Config.BuiltInSyntax {
		for _, keyword := range syntax.Keywords {
			hl, ok := th.HighlightMap[syntax.Group]
			if !ok {
				return nil, fmt.Errorf("highlight group %q not found", syntax.Group)
			}
			renderer.builtinKeywordMap[keyword] = hl
		}
	}

	renderer.userKeywordMap = make(keywordMap)
	for _, syntax := range renderer.Config.UserSyntax {
		for _, keyword := range syntax.Keywords {
			hl, ok := th.HighlightMap[syntax.Group]
			if !ok {
				return nil, fmt.Errorf("highlight group %q not found", syntax.Group)
			}
			renderer.userKeywordMap[keyword] = hl
		}
	}

	return renderer, nil
}

type Match struct {
	Start     int
	End       int
	AnsiStart string
	AnsiEnd   string
}

func (r Renderer) Render(text string) (string, error) {
	builtInLowerMatches, err := findMatches(
		r.Config.BuiltInSyntaxLower,
		r.Theme.HighlightMap,
		r.builtinLowerKeywordMap,
		text,
	)
	if err != nil {
		return text, err
	}

	builtInMatches, err := findMatches(
		r.Config.BuiltInSyntax,
		r.Theme.HighlightMap,
		r.builtinKeywordMap,
		text,
	)
	if err != nil {
		return text, err
	}

	builtinMatchesCombined := Stack(builtInMatches, builtInLowerMatches)

	userMatches, err := findMatches(
		r.Config.UserSyntax,
		r.Theme.HighlightMap,
		r.userKeywordMap,
		text,
	)
	if err != nil {
		return text, err
	}

	matches := Stack(userMatches, builtinMatchesCombined)

	if userMatches.Len() > 0 {
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
	}

	if len(matches) == 0 {
		return text, nil
	}

	matches.Sort()

	return buildHighlightedString(text, matches), nil
}

func findMatches(
	syntaxList []proto.Syntax,
	highlights map[string]*style.Highlight,
	keywordMap map[string]*style.Highlight,
	text string,
) (MatchLayer, error) {
	var matches MatchLayer
	var err error

	err = findPatternMatches(&matches, syntaxList, highlights, text)
	if err != nil {
		return nil, err
	}

	err = findKeywordMatches(&matches, keywordMap, text)
	if err != nil {
		return nil, err
	}

	matches.removeOverlaps()
	matches.Sort()

	return matches, nil
}

func buildHighlightedString(text string, matches MatchLayer) string {
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

	return b.String()
}
