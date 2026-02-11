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
	for _, syn := range syntaxList {
		p := syn.Pattern
		if !p.HasValue() {
			continue
		}
		hl, ok := highlights[syn.Group]
		if !ok {
			return fmt.Errorf("highlight group %s not found", syn.Group)
		}
		for _, idx := range p.FindAllStringIndex(text, -1) {
			*matches = append(*matches, Match{
				Start:     idx[0],
				End:       idx[1],
				AnsiStart: hl.BuildAnsi(),
				AnsiEnd:   hl.BuildAnsiReset(),
			})
		}
	}

	return nil
}
