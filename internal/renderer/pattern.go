package renderer

import (
	"fmt"

	"github.com/madmaxieee/loglit/internal/proto"
	"github.com/madmaxieee/loglit/internal/style"
)

func findPatternMatches(
	matches *MatchLayer,
	syntaxList []proto.Syntax,
	highlights map[string]*style.Highlight,
	text string,
) error {
	type result struct {
		matches []Match
		err     error
	}
	results := make([]result, len(syntaxList))

	// find matches for regex
	for i, syn := range syntaxList {
		p := syn.Pattern
		if !p.HasValue() {
			continue
		}
		hl, ok := highlights[syn.Group]
		if !ok {
			results[i] = result{nil, fmt.Errorf("highlight group %s not found", syn.Group)}
			continue
		}
		for _, idx := range p.FindAllStringIndex(text, -1) {
			results[i].matches = append(results[i].matches, Match{
				Start:     idx[0],
				End:       idx[1],
				AnsiStart: hl.BuildAnsi(),
				AnsiEnd:   hl.BuildAnsiReset(),
			})
		}
	}

	for _, res := range results {
		if res.err != nil {
			return res.err
		}
		*matches = append(*matches, res.matches...)
	}

	return nil
}
