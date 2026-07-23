package renderer

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/madmaxieee/loglit/internal/config"
	"github.com/madmaxieee/loglit/internal/proto"
	"github.com/madmaxieee/loglit/internal/style"
	"github.com/madmaxieee/loglit/internal/theme"
)

func TestBuiltInPatternGatesMatchUnconditionalResults(t *testing.T) {
	cfg := config.GetDefaultConfig()
	th := theme.GetDefaultTheme()
	r, err := New(cfg, th)
	if err != nil {
		t.Fatal(err)
	}
	lines := []string{
		"", "plain message", "INFO count=123", "'quoted' \"text\" \\n",
		"2024-01-02T03:04:05Z https://example.test 10.5s",
		"AA:BB:CC:DD:EE:FF 127.0.0.1 2001:db8:0:0:0:0:0:1",
		"SERVICE_ERROR and FATAL, Jan 2, 2024",
		"--- === ### *** <<< >>> \\t 12:34:56.123",
		"0b101 0o755 0xdead beef 12345 1.25e+3",
		"01/02 2024/01/02 02-Jan-2024",
		"0123456789abcdef0123456789abcdef0123456789",
	}
	rng := rand.New(rand.NewSource(1))
	for i := 0; i < 256; i++ {
		line := make([]byte, rng.Intn(160))
		for j := range line {
			line[j] = byte(32 + rng.Intn(95))
		}
		lines = append(lines, string(line))
	}

	for _, line := range lines {
		facts := proto.MakeLineFacts(line)
		for syntaxIndex, syntax := range [][]proto.Syntax{cfg.BuiltInSyntaxLower, cfg.BuiltInSyntax} {
			gated, err := findPatternMatchesForTest(syntax, r.Theme.HighlightMap, line, facts)
			if err != nil {
				t.Fatalf("gated matching failed for %q: %v", line, err)
			}
			plain, err := findPatternMatchesUnconditionalForTest(syntax, r.Theme.HighlightMap, line)
			if err != nil {
				t.Fatalf("unconditional matching failed for %q: %v", line, err)
			}
			if !sameMatches(gated, plain) {
				t.Fatalf("gated results differ for %q (syntax %d)\ngated=%#v\nplain=%#v", line, syntaxIndex, gated, plain)
			}
		}
	}
}

func TestPatternsWithoutGatesRemainUnconditional(t *testing.T) {
	pattern := proto.MustCompile(`[!@#$%^&*;:?=]`)
	if !pattern.CanRun(proto.MakeLineFacts("ordinary=message")) {
		t.Fatal("ordinary patterns must be unconditional")
	}

	syntax := []proto.Syntax{{Group: "UserPattern", Pattern: pattern}}
	got, err := findPatternMatchesForTest(syntax, theme.GetDefaultTheme().HighlightMap, "ordinary=message", proto.MakeLineFacts("ordinary=message"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Start != 8 || got[0].End != 9 {
		t.Fatalf("default-unconditional pattern was not matched: %#v", got)
	}
}

func findPatternMatchesForTest(syntax []proto.Syntax, highlights map[string]*style.Highlight, text string, facts proto.LineFacts) (MatchLayer, error) {
	var matches MatchLayer
	if err := findPatternMatches(&matches, syntax, highlights, text, facts); err != nil {
		return nil, err
	}
	matches.removeOverlaps()
	matches.Sort()
	return matches, nil
}

func findPatternMatchesUnconditionalForTest(syntax []proto.Syntax, highlights map[string]*style.Highlight, text string) (MatchLayer, error) {
	var matches MatchLayer
	for _, syn := range syntax {
		if !syn.Pattern.HasValue() {
			continue
		}
		hl, ok := highlights[syn.Group]
		if !ok {
			return nil, fmt.Errorf("highlight group %s not found", syn.Group)
		}
		for _, idx := range syn.Pattern.FindAllStringIndex(text, -1) {
			matches = append(matches, Match{Start: idx[0], End: idx[1], AnsiStart: hl.BuildAnsi(), AnsiEnd: hl.BuildAnsiReset()})
		}
	}
	matches.removeOverlaps()
	matches.Sort()
	return matches, nil
}

func sameMatches(a, b MatchLayer) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
