package renderer

import (
	"fmt"
	"regexp"

	"github.com/madmaxieee/loglit/internal/proto"
	"github.com/madmaxieee/loglit/internal/style"
)

var unicodeWordRe = regexp.MustCompile(`[\p{L}\p{M}\p{N}]+`)

type Word struct {
	Value string
	Start int
	End   int
}

func IsValidKeyword(s string) bool {
	return unicodeWordRe.MatchString(s)
}

func findWordsInText(text string) map[string][]Word {
	var words = make(map[string][]Word)
	for _, idx := range unicodeWordRe.FindAllStringIndex(text, -1) {
		start, end := idx[0], idx[1]
		word := text[start:end]
		words[word] = append(words[word], Word{
			Value: word,
			Start: start,
			End:   end,
		})
	}
	return words
}

func findKeywordMatches(syntaxList []proto.Syntax, highlights map[string]*style.Highlight, text string) (MatchLayer, error) {
	var matches MatchLayer
	words := findWordsInText(text)

	// find matches for keywords
	for _, syn := range syntaxList {
		for _, kw := range syn.Keywords {
			wordInstances, ok := words[kw]
			if !ok {
				continue
			}
			hl, ok := highlights[syn.Group]
			if !ok {
				return MatchLayer{}, fmt.Errorf("highlight group %s not found", syn.Group)
			}
			for _, word := range wordInstances {
				matches = append(matches, Match{
					Start:     word.Start,
					End:       word.End,
					AnsiStart: hl.BuildAnsi(),
					AnsiEnd:   hl.BuildAnsiReset(),
				})
			}
		}
	}

	return matches, nil
}
