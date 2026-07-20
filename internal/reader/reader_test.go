package reader

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/madmaxieee/loglit/internal/config"
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
				lb.ProcessCompleteLines(coloredWriter, rawWriter)
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
	lb.Finalize(coloredWriter, rawWriter)
	_ = rawWriter.Flush()
	if raw.String() != "final line" {
		t.Fatalf("raw output = %q, want %q", raw.String(), "final line")
	}
}

func TestLineBufferRepeatedPartialFlushes(t *testing.T) {
	lb, _, raw, coloredWriter, rawWriter := testBuffer(t)
	lb.Append([]byte("par"))
	lb.FlushPending(coloredWriter, rawWriter)
	lb.Append([]byte("tial"))
	lb.FlushPending(coloredWriter, rawWriter)
	lb.Append([]byte(" line"))
	lb.FlushPending(coloredWriter, rawWriter)
	lb.Finalize(coloredWriter, rawWriter)
	_ = rawWriter.Flush()
	if raw.String() != "partial line" {
		t.Fatalf("raw output = %q, want %q", raw.String(), "partial line")
	}
}

func TestLineBufferCRLFWithPartialFlush(t *testing.T) {
	lb, _, raw, coloredWriter, rawWriter := testBuffer(t)
	lb.Append([]byte("line\r"))
	lb.FlushPending(coloredWriter, rawWriter)
	lb.Append([]byte("\n"))
	lb.ProcessCompleteLines(coloredWriter, rawWriter)
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
	lb.ProcessCompleteLines(coloredWriter, rawWriter)
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
	lb.FlushPending(coloredWriter, rawWriter)
	lb.Append([]byte("\n"))
	lb.ProcessCompleteLines(coloredWriter, rawWriter)
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
	lb.FlushPending(coloredWriter, rawWriter)
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
	lb.Finalize(coloredWriter, rawWriter)
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
	lb.FlushPending(nil, rawWriter)
	lb.FlushPending(nil, nil)
	_ = rawWriter.Flush()
	if raw.String() != "pending" {
		t.Fatalf("raw output = %q, want %q", raw.String(), "pending")
	}
}
