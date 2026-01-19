package renderer

import (
	"strings"
	"testing"

	"github.com/madmaxieee/loglit/internal/config"
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
