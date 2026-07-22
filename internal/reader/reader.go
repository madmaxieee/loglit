package reader

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"github.com/madmaxieee/loglit/internal/renderer"
)

// LineBuffer accumulates incoming chunks and processes complete lines,
// while supporting periodic flushing of incomplete lines.
type LineBuffer struct {
	renderer       *renderer.Renderer
	buf            []byte
	coloredFlushed int
	rawFlushed     int
}

// NewLineBuffer creates a new LineBuffer.
func NewLineBuffer(renderer *renderer.Renderer) *LineBuffer {
	return &LineBuffer{renderer: renderer}
}

// Append adds incoming data to the internal buffer.
func (lb *LineBuffer) Append(data []byte) {
	lb.buf = append(lb.buf, data...)
}

// ProcessCompleteLines finds and renders all complete lines (ending in \n),
// writing them to the provided writers. It handles clearing previously-flushed
// partial output for the colored writer using ANSI escape sequences. It returns
// the first writer error encountered.
func (lb *LineBuffer) ProcessCompleteLines(coloredWriter, rawWriter *bufio.Writer) error {
	for {
		idx := bytes.IndexByte(lb.buf, '\n')
		if idx == -1 {
			break
		}

		lineBytes := lb.buf[:idx]
		if len(lineBytes) > 0 && lineBytes[len(lineBytes)-1] == '\r' {
			lineBytes = lineBytes[:len(lineBytes)-1]
		}
		line := string(lineBytes)

		if coloredWriter != nil && lb.coloredFlushed > 0 {
			if _, err := coloredWriter.WriteString("\033[2K\r"); err != nil {
				return err
			}
		}
		if coloredWriter != nil {
			coloredLine, err := lb.renderer.Render(line)
			if err != nil {
				return fmt.Errorf("render line: %w", err)
			}
			if _, err := coloredWriter.WriteString(coloredLine); err != nil {
				return err
			}
			if err := coloredWriter.WriteByte('\n'); err != nil {
				return err
			}
		}

		if _, err := rawWriter.Write(lb.buf[lb.rawFlushed : idx+1]); err != nil {
			return err
		}

		lb.buf = lb.buf[idx+1:]
		lb.coloredFlushed = 0
		lb.rawFlushed = 0
	}
	return nil
}

// FlushPending writes any buffered but not-yet-completed line data to the
// writers, tracking how much has been flushed so far. Passing a nil writer
// skips that output.
//
// For the colored writer, the pending line is rendered and the entire line is
// redrawn (after clearing the previous partial output) so that partial lines
// appear colorized in real time. It returns the first writer error encountered.
func (lb *LineBuffer) FlushPending(coloredWriter, rawWriter *bufio.Writer) error {
	if len(lb.buf) == 0 {
		return nil
	}
	pending := string(lb.buf)
	if coloredWriter != nil && len(pending) > lb.coloredFlushed {
		if _, err := coloredWriter.WriteString("\033[2K\r"); err != nil {
			return err
		}
		coloredLine, err := lb.renderer.Render(pending)
		if err != nil {
			return fmt.Errorf("render pending line: %w", err)
		}
		if _, err := coloredWriter.WriteString(coloredLine); err != nil {
			return err
		}
		lb.coloredFlushed = len(pending)
	}
	if rawWriter != nil && len(lb.buf) > lb.rawFlushed {
		if _, err := rawWriter.Write(lb.buf[lb.rawFlushed:]); err != nil {
			return err
		}
		lb.rawFlushed = len(lb.buf)
	}
	return nil
}

// Finalize treats any remaining buffered data as a final line and writes it
// to the writers, even if it lacks a trailing newline. It returns the first
// writer error encountered.
func (lb *LineBuffer) Finalize(coloredWriter, rawWriter *bufio.Writer) error {
	if len(lb.buf) == 0 {
		return nil
	}
	line := string(lb.buf)
	if coloredWriter != nil && lb.coloredFlushed > 0 {
		if _, err := coloredWriter.WriteString("\033[2K\r"); err != nil {
			return err
		}
	}
	if coloredWriter != nil {
		coloredLine, err := lb.renderer.Render(line)
		if err != nil {
			return fmt.Errorf("render final line: %w", err)
		}
		if _, err := coloredWriter.WriteString(coloredLine); err != nil {
			return err
		}
		if err := coloredWriter.WriteByte('\n'); err != nil {
			return err
		}
	}

	if _, err := rawWriter.Write(lb.buf[lb.rawFlushed:]); err != nil {
		return err
	}

	lb.buf = nil
	lb.coloredFlushed = 0
	lb.rawFlushed = 0
	return nil
}

// ReadChunks reads data from the provided reader in chunks and sends them
// on the returned channel. The channel is closed when reading is done.
func ReadChunks(r io.Reader) <-chan []byte {
	chunkCh := make(chan []byte)
	go func() {
		defer close(chunkCh)
		buf := make([]byte, 4096)
		for {
			n, err := r.Read(buf)
			if n > 0 {
				chunk := make([]byte, n)
				copy(chunk, buf[:n])
				chunkCh <- chunk
			}
			if err != nil {
				return
			}
		}
	}()
	return chunkCh
}
