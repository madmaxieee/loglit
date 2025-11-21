package renderer

import (
	"fmt"
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
	AnsiEnd   string
}

func (r Renderer) Render(text string) (string, error) {
	builtInLowerMatches, err := findMatches(r.Config.BuiltInSyntaxLower, r.Theme.HighlightMap, text)
	if err != nil {
		return text, err
	}
	builtInLowerMatches.removeOverlaps().Sort()

	builtInMatches, err := findMatches(r.Config.BuiltInSyntax, r.Theme.HighlightMap, text)
	if err != nil {
		return text, err
	}
	builtInMatches.removeOverlaps().Sort()

	builtinMatchesCombined := Stack(builtInMatches, builtInLowerMatches)

	userMatches, err := findMatches(r.Config.UserSyntax, r.Theme.HighlightMap, text)
	if err != nil {
		return text, err
	}
	userMatches.removeOverlaps().Sort()

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

func findMatches(syntaxList []proto.Syntax, highlights map[string]*style.Highlight, text string) (MatchLayer, error) {
	patternMatches, err := findPatternMatches(syntaxList, highlights, text)
	if err != nil {
		return nil, err
	}

	keywordMatches, err := findKeywordMatches(syntaxList, highlights, text)
	if err != nil {
		return nil, err
	}

	matches := append(patternMatches, keywordMatches...)
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
