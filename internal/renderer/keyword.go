package renderer

import (
	"fmt"
	"regexp"
	"sync"

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
	words := findWordsInText(text)

	type result struct {
		matches []Match
		err     error
	}
	results := make([]result, len(syntaxList))

	var wg sync.WaitGroup
	wg.Add(len(syntaxList))

	// find matches for keywords
	for i, syn := range syntaxList {
		hl, ok := highlights[syn.Group]
		if !ok {
			results[i] = result{nil, fmt.Errorf("highlight group %s not found", syn.Group)}
			wg.Done()
			continue
		}
		go func() {
			defer wg.Done()
			for _, kw := range syn.Keywords {
				wordInstances, ok := words[kw]
				if !ok {
					continue
				}
				for _, word := range wordInstances {
					results[i].matches = append(results[i].matches, Match{
						Start:     word.Start,
						End:       word.End,
						AnsiStart: hl.BuildAnsi(),
						AnsiEnd:   hl.BuildAnsiReset(),
					})
				}
			}
		}()
	}

	wg.Wait()

	var matches MatchLayer
	for _, res := range results {
		if res.err != nil {
			return nil, res.err
		}
		matches = append(matches, res.matches...)
	}

	return matches, nil
}
