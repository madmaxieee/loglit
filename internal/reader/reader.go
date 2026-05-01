package reader

import (
	"bufio"
	"bytes"
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
// partial output for the colored writer using ANSI escape sequences.
func (lb *LineBuffer) ProcessCompleteLines(coloredWriter, rawWriter *bufio.Writer) {
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

		if lb.coloredFlushed > 0 {
			coloredWriter.WriteString("\033[2K\r")
		}
		coloredLine, _ := lb.renderer.Render(line)
		coloredWriter.WriteString(coloredLine)
		coloredWriter.WriteByte('\n')

		if lb.rawFlushed > 0 {
			rawWriter.WriteString(line[lb.rawFlushed:])
		} else {
			rawWriter.WriteString(line)
		}
		rawWriter.WriteByte('\n')

		lb.buf = lb.buf[idx+1:]
		lb.coloredFlushed = 0
		lb.rawFlushed = 0
	}
}

// FlushPending writes any buffered but not-yet-completed line data to the
// writers, tracking how much has been flushed so far. Passing a nil writer
// skips that output.
//
// For the colored writer, the pending line is rendered and the entire line is
// redrawn (after clearing the previous partial output) so that partial lines
// appear colorized in real time.
func (lb *LineBuffer) FlushPending(coloredWriter, rawWriter *bufio.Writer) {
	if len(lb.buf) == 0 {
		return
	}
	pending := string(lb.buf)
	if coloredWriter != nil && len(pending) > lb.coloredFlushed {
		coloredWriter.WriteString("\033[2K\r")
		coloredLine, _ := lb.renderer.Render(pending)
		coloredWriter.WriteString(coloredLine)
		lb.coloredFlushed = len(pending)
	}
	if rawWriter != nil && len(pending) > lb.rawFlushed {
		rawWriter.WriteString(pending[lb.rawFlushed:])
		lb.rawFlushed = len(pending)
	}
}

// Finalize treats any remaining buffered data as a final line and writes it
// to the writers, even if it lacks a trailing newline.
func (lb *LineBuffer) Finalize(coloredWriter, rawWriter *bufio.Writer) {
	if len(lb.buf) == 0 {
		return
	}
	line := string(lb.buf)
	if lb.coloredFlushed > 0 {
		coloredWriter.WriteString("\033[2K\r")
	}
	coloredLine, _ := lb.renderer.Render(line)
	coloredWriter.WriteString(coloredLine)
	coloredWriter.WriteByte('\n')

	if lb.rawFlushed > 0 {
		rawWriter.WriteString(line[lb.rawFlushed:])
	} else {
		rawWriter.WriteString(line)
	}
	rawWriter.WriteByte('\n')

	lb.buf = nil
	lb.coloredFlushed = 0
	lb.rawFlushed = 0
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
