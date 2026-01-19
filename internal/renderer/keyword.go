package renderer

import (
	"regexp"

	"github.com/madmaxieee/loglit/internal/style"
)

var unicodeWordRe = regexp.MustCompile(`[\p{L}\p{M}\p{N}_]+`)

func IsValidKeyword(s string) bool {
	return unicodeWordRe.MatchString(s)
}

func findKeywordMatches(
	matches *MatchLayer,
	keywordMap map[string]*style.Highlight,
	text string,
) error {
	for _, idx := range unicodeWordRe.FindAllStringIndex(text, -1) {
		start, end := idx[0], idx[1]
		word := text[start:end]
		if hl, ok := keywordMap[word]; ok {
			*matches = append(*matches, Match{
				Start:     start,
				End:       end,
				AnsiStart: hl.BuildAnsi(),
				AnsiEnd:   hl.BuildAnsiReset(),
			})
		}
	}

	return nil
}
