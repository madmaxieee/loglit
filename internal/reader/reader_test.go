package reader

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"regexp"
	"testing"

	"github.com/madmaxieee/loglit/internal/config"
	"github.com/madmaxieee/loglit/internal/proto"
	"github.com/madmaxieee/loglit/internal/renderer"
	"github.com/madmaxieee/loglit/internal/theme"
)

func testBuffer(t *testing.T) (*LineBuffer, *bytes.Buffer, *bytes.Buffer, *bufio.Writer, *bufio.Writer) {
	t.Helper()
	r, err := renderer.New(config.GetDefaultConfig(), theme.GetDefaultTheme())
	if err != nil {
		t.Fatalf("create renderer: %v", err)
	}
	colored, raw := new(bytes.Buffer), new(bytes.Buffer)
	return NewLineBuffer(r), colored, raw, bufio.NewWriter(colored), bufio.NewWriter(raw)
}

func mustProcess(t *testing.T, lb *LineBuffer, coloredWriter, rawWriter *bufio.Writer) {
	t.Helper()
	if err := lb.ProcessCompleteLines(coloredWriter, rawWriter); err != nil {
		t.Fatal(err)
	}
}

func mustFlushPending(t *testing.T, lb *LineBuffer, coloredWriter, rawWriter *bufio.Writer) {
	t.Helper()
	if err := lb.FlushPending(coloredWriter, rawWriter); err != nil {
		t.Fatal(err)
	}
}

func mustFinalize(t *testing.T, lb *LineBuffer, coloredWriter, rawWriter *bufio.Writer) {
	t.Helper()
	if err := lb.Finalize(coloredWriter, rawWriter); err != nil {
		t.Fatal(err)
	}
}

func TestLineBufferProcessCompleteLines(t *testing.T) {
	tests := []struct {
		name   string
		chunks []string
		want   string
	}{
		{"LF", []string{"one\ntwo\n"}, "one\ntwo\n"},
		{"CRLF", []string{"one\r\ntwo\r\n"}, "one\r\ntwo\r\n"},
		{"split chunks", []string{"fi", "rst\nse", "cond\n"}, "first\nsecond\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lb, _, raw, coloredWriter, rawWriter := testBuffer(t)
			for _, chunk := range tt.chunks {
				lb.Append([]byte(chunk))
				mustProcess(t, lb, coloredWriter, rawWriter)
			}
			_ = coloredWriter.Flush()
			_ = rawWriter.Flush()
			if raw.String() != tt.want {
				t.Fatalf("raw output = %q, want %q", raw.String(), tt.want)
			}
		})
	}
}

func TestLineBufferFinalizeWithoutNewline(t *testing.T) {
	lb, _, raw, coloredWriter, rawWriter := testBuffer(t)
	lb.Append([]byte("final line"))
	mustFinalize(t, lb, coloredWriter, rawWriter)
	_ = rawWriter.Flush()
	if raw.String() != "final line" {
		t.Fatalf("raw output = %q, want %q", raw.String(), "final line")
	}
}

func TestLineBufferSkipsColoredOutputForCompleteLines(t *testing.T) {
	lb, _, raw, _, rawWriter := testBuffer(t)
	lb.Append([]byte("INFO\n"))
	mustProcess(t, lb, nil, rawWriter)
	_ = rawWriter.Flush()
	if raw.String() != "INFO\n" {
		t.Fatalf("raw output = %q, want %q", raw.String(), "INFO\n")
	}
}

func TestLineBufferSkipsColoredOutputWhenFinalizing(t *testing.T) {
	lb, _, raw, _, rawWriter := testBuffer(t)
	lb.Append([]byte("INFO"))
	mustFinalize(t, lb, nil, rawWriter)
	_ = rawWriter.Flush()
	if raw.String() != "INFO" {
		t.Fatalf("raw output = %q, want %q", raw.String(), "INFO")
	}
}

func TestLineBufferRepeatedPartialFlushes(t *testing.T) {
	lb, _, raw, coloredWriter, rawWriter := testBuffer(t)
	lb.Append([]byte("par"))
	mustFlushPending(t, lb, coloredWriter, rawWriter)
	lb.Append([]byte("tial"))
	mustFlushPending(t, lb, coloredWriter, rawWriter)
	lb.Append([]byte(" line"))
	mustFlushPending(t, lb, coloredWriter, rawWriter)
	mustFinalize(t, lb, coloredWriter, rawWriter)
	_ = rawWriter.Flush()
	if raw.String() != "partial line" {
		t.Fatalf("raw output = %q, want %q", raw.String(), "partial line")
	}
}

func TestLineBufferCRLFWithPartialFlush(t *testing.T) {
	lb, _, raw, coloredWriter, rawWriter := testBuffer(t)
	lb.Append([]byte("line\r"))
	mustFlushPending(t, lb, coloredWriter, rawWriter)
	lb.Append([]byte("\n"))
	mustProcess(t, lb, coloredWriter, rawWriter)
	if err := coloredWriter.Flush(); err != nil {
		t.Fatalf("flush colored writer: %v", err)
	}
	if err := rawWriter.Flush(); err != nil {
		t.Fatalf("flush raw writer: %v", err)
	}
	if raw.String() != "line\r\n" {
		t.Fatalf("raw output = %q, want %q", raw.String(), "line\r\n")
	}
}

func TestLineBufferSplitCRLFWithoutPartialFlush(t *testing.T) {
	lb, _, raw, coloredWriter, rawWriter := testBuffer(t)
	lb.Append([]byte("line\r"))
	lb.Append([]byte("\n"))
	mustProcess(t, lb, coloredWriter, rawWriter)
	if err := rawWriter.Flush(); err != nil {
		t.Fatalf("flush raw writer: %v", err)
	}
	if raw.String() != "line\r\n" {
		t.Fatalf("raw output = %q, want %q", raw.String(), "line\r\n")
	}
}

func TestLineBufferLFPartialFlushPreserved(t *testing.T) {
	lb, _, raw, coloredWriter, rawWriter := testBuffer(t)
	lb.Append([]byte("line"))
	mustFlushPending(t, lb, coloredWriter, rawWriter)
	lb.Append([]byte("\n"))
	mustProcess(t, lb, coloredWriter, rawWriter)
	if err := rawWriter.Flush(); err != nil {
		t.Fatalf("flush raw writer: %v", err)
	}
	if raw.String() != "line\n" {
		t.Fatalf("raw output = %q, want %q", raw.String(), "line\n")
	}
}

func TestLineBufferStandaloneCRPreserved(t *testing.T) {
	lb, _, raw, coloredWriter, rawWriter := testBuffer(t)
	lb.Append([]byte("line\r"))
	mustFlushPending(t, lb, coloredWriter, rawWriter)
	if err := rawWriter.Flush(); err != nil {
		t.Fatalf("flush raw writer: %v", err)
	}
	if raw.String() != "line\r" {
		t.Fatalf("raw output = %q, want %q", raw.String(), "line\r")
	}
}

func TestLineBufferTrailingCRFinalization(t *testing.T) {
	lb, _, raw, coloredWriter, rawWriter := testBuffer(t)
	lb.Append([]byte("line\r"))
	mustFinalize(t, lb, coloredWriter, rawWriter)
	if err := coloredWriter.Flush(); err != nil {
		t.Fatalf("flush colored writer: %v", err)
	}
	if err := rawWriter.Flush(); err != nil {
		t.Fatalf("flush raw writer: %v", err)
	}
	if raw.String() != "line\r" {
		t.Fatalf("raw output = %q, want %q", raw.String(), "line\r")
	}
}

func TestLineBufferFlushPendingAllowsNilWriters(t *testing.T) {
	lb, _, raw, _, rawWriter := testBuffer(t)
	lb.Append([]byte("pending"))
	mustFlushPending(t, lb, nil, rawWriter)
	mustFlushPending(t, lb, nil, nil)
	_ = rawWriter.Flush()
	if raw.String() != "pending" {
		t.Fatalf("raw output = %q, want %q", raw.String(), "pending")
	}
}

var errLineBufferWriter = errors.New("line buffer writer failed")

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errLineBufferWriter }

func TestLineBufferPropagatesWriterErrors(t *testing.T) {
	tests := []struct {
		name string
		call func(*LineBuffer, *bufio.Writer, *bufio.Writer) error
	}{
		{"process complete lines", func(lb *LineBuffer, colored, raw *bufio.Writer) error {
			lb.Append([]byte("line\n"))
			return lb.ProcessCompleteLines(colored, raw)
		}},
		{"flush pending", func(lb *LineBuffer, colored, raw *bufio.Writer) error {
			lb.Append([]byte("pending"))
			return lb.FlushPending(colored, raw)
		}},
		{"finalize", func(lb *LineBuffer, colored, raw *bufio.Writer) error {
			lb.Append([]byte("final"))
			return lb.Finalize(colored, raw)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := renderer.New(config.GetDefaultConfig(), theme.GetDefaultTheme())
			if err != nil {
				t.Fatal(err)
			}
			lb := NewLineBuffer(r)
			writer := bufio.NewWriterSize(failingWriter{}, 1)
			if err := tt.call(lb, bufio.NewWriter(bytes.NewBuffer(nil)), writer); !errors.Is(err, errLineBufferWriter) {
				t.Fatalf("error = %v, want %v", err, errLineBufferWriter)
			}
		})
	}
}

func TestLineBufferPropagatesRenderErrors(t *testing.T) {
	cfg := config.GetDefaultConfig()
	pattern := regexp.MustCompile("match")
	cfg.UserSyntax = []proto.Syntax{{
		Group:   "UserPattern",
		Pattern: proto.Pattern{Regexp: pattern},
	}}
	r, err := renderer.New(cfg, theme.GetDefaultTheme())
	if err != nil {
		t.Fatal(err)
	}
	delete(r.Theme.HighlightMap, "UserMatchLineBackground")

	tests := []struct {
		name string
		call func(*LineBuffer, *bufio.Writer, *bufio.Writer) error
	}{
		{"process complete lines", func(lb *LineBuffer, colored, raw *bufio.Writer) error {
			lb.Append([]byte("match\n"))
			return lb.ProcessCompleteLines(colored, raw)
		}},
		{"flush pending", func(lb *LineBuffer, colored, raw *bufio.Writer) error {
			lb.Append([]byte("match"))
			return lb.FlushPending(colored, raw)
		}},
		{"finalize", func(lb *LineBuffer, colored, raw *bufio.Writer) error {
			lb.Append([]byte("match"))
			return lb.Finalize(colored, raw)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lb := NewLineBuffer(r)
			if err := tt.call(lb, bufio.NewWriter(io.Discard), bufio.NewWriter(io.Discard)); err == nil {
				t.Fatal("render error was not propagated")
			}
		})
	}
}
