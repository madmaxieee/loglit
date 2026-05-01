package renderer

import (
	"strings"
	"testing"

	"github.com/madmaxieee/loglit/internal/config"
	"github.com/madmaxieee/loglit/internal/proto"
	"github.com/madmaxieee/loglit/internal/theme"
)

func BenchmarkRender(b *testing.B) {
	cfg := config.GetDefaultConfig()
	th := theme.GetDefaultTheme()
	r, err := New(cfg, th)
	if err != nil {
		b.Fatalf("failed to create renderer: %v", err)
	}

	line := "2023-10-27 10:00:00 INFO [main] This is a test log message with some numbers 12345 and a url https://example.com"

	for b.Loop() {
		_, _ = r.Render(line)
	}
}

func TestRender(t *testing.T) {
	cfg := config.GetDefaultConfig()
	th := theme.GetDefaultTheme()
	r, err := New(cfg, th)
	if err != nil {
		t.Fatalf("failed to create renderer: %v", err)
	}

	line := "INFO"
	out, err := r.Render(line)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	// We expect ANSI codes. Just a basic check that it's not empty and different from input.
	if out == "" {
		t.Error("expected output, got empty string")
	}
	if out == line {
		t.Error("expected highlighting, got raw string")
	}
	if !strings.Contains(out, "\x1b[") {
		t.Error("expected ANSI escape codes")
	}
}

func TestRender_UserMatchBackground(t *testing.T) {
	cfg := config.GetDefaultConfig()
	th := theme.GetDefaultTheme()
	cfg.UserSyntax = []proto.Syntax{{Group: "UserPattern", Pattern: proto.MustCompile(`USER\d+`)}}
	r, err := New(cfg, th)
	if err != nil {
		t.Fatalf("failed to create renderer: %v", err)
	}

	line := "prefix USER123 suffix"
	out, err := r.Render(line)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if !strings.Contains(out, "\x1b[") {
		t.Error("expected ANSI escape codes")
	}
	if out == line {
		t.Error("expected highlighting, got raw string")
	}

	userBgHighlight := th.HighlightMap["UserMatchLineBackground"]
	userBgAnsi := userBgHighlight.BuildAnsi()
	userBgReset := userBgHighlight.BuildAnsiReset()

	if !strings.HasPrefix(out, userBgAnsi) {
		t.Errorf("expected output to start with UserMatchLineBackground ANSI %q, got %q", userBgAnsi, out)
	}
	if !strings.HasSuffix(out, userBgReset) {
		t.Errorf("expected output to end with UserMatchLineBackground reset %q, got %q", userBgReset, out)
	}

	// The background is applied as a prefix and re-applied inside any match
	// that would otherwise reset the background, so it occurs at least twice.
	count := strings.Count(out, userBgAnsi)
	if count < 2 {
		t.Errorf("expected UserMatchLineBackground ANSI to occur at least twice, got %d occurrence(s)", count)
	}
}

func TestBuildHighlightedString(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		matches  MatchLayer
		expected string
	}{
		{
			name: "single match at start",
			text: "hello world",
			matches: MatchLayer{
				{Start: 0, End: 5, AnsiStart: "<red>", AnsiEnd: "</red>"},
			},
			expected: "<red>hello</red> world",
		},
		{
			name: "single match in middle",
			text: "hello world",
			matches: MatchLayer{
				{Start: 6, End: 11, AnsiStart: "<blue>", AnsiEnd: "</blue>"},
			},
			expected: "hello <blue>world</blue>",
		},
		{
			name: "single match at end",
			text: "hello world",
			matches: MatchLayer{
				{Start: 6, End: 11, AnsiStart: "<green>", AnsiEnd: "</green>"},
			},
			expected: "hello <green>world</green>",
		},
		{
			name: "multiple matches disjoint",
			text: "foo bar baz",
			matches: MatchLayer{
				{Start: 0, End: 3, AnsiStart: "<r>", AnsiEnd: "</r>"},
				{Start: 8, End: 11, AnsiStart: "<b>", AnsiEnd: "</b>"},
			},
			expected: "<r>foo</r> bar <b>baz</b>",
		},
		{
			name: "matches touching",
			text: "foobar",
			matches: MatchLayer{
				{Start: 0, End: 3, AnsiStart: "<r>", AnsiEnd: "</r>"},
				{Start: 3, End: 6, AnsiStart: "<b>", AnsiEnd: "</b>"},
			},
			expected: "<r>foo</r><b>bar</b>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildHighlightedString(tt.text, tt.matches)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
