package snapraid

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"sync"
)

// loggerWriter is an io.Writer that splits on newlines and sends each line
// into a structured slog.Logger with a given tag.
type loggerWriter struct {
	logger *slog.Logger // logger is the structured slog.Logger instance to send each completed line.
	tag    string       // tag is the component name to use for each line.
	level  slog.Level   // level is the slog.Level to use for each line.
	buf    bytes.Buffer // buf holds partial data until a newline is encountered.
	mu     sync.Mutex   // mu protects buf if Write is called concurrently.
}

// newLoggerWriter constructs a loggerWriter that tags every line with tag.
func newLoggerWriter(logger *slog.Logger, tag string, level slog.Level) *loggerWriter {
	return &loggerWriter{
		logger: logger,
		level:  level,
		tag:    tag,
	}
}

// Write implements io.Writer. It accumulates bytes in an internal buffer until
// it sees a newline. For each complete line, it calls logger.Info with that line
// as the message, plus key/value pairs. Any partial (non-terminated) line is
// kept in buf until the next Write.
func (lw *loggerWriter) Write(p []byte) (n int, err error) {
	lw.mu.Lock()
	defer lw.mu.Unlock()

	total := 0

	for len(p) > 0 {
		// Look for the next newline in p
		idx := bytes.IndexByte(p, '\n')
		if idx < 0 {
			// No newline in this chunk: buffer everything and return
			m, _ := lw.buf.Write(p)
			total += m
			return total, nil
		}

		// There is at least one newline at p[idx].
		// Write up through and including that newline into our buffer.
		m, _ := lw.buf.Write(p[:idx+1])
		total += m

		// Extract the buffered bytes (a complete line, including '\n').
		lineBytes := lw.buf.Bytes()

		line := string(bytes.TrimRight(lineBytes, "\n")) // Remove the trailing '\n'
		line = strings.TrimSpace(line)                   // Trim any surrounding whitespace

		// If the trimmed line is empty, skip logging
		if line != "" {
			lw.logger.Log(context.Background(), lw.level, line, "tag", lw.tag)
		}

		// Reset the buffer, because we’ve consumed everything up to the newline.
		lw.buf.Reset()

		// Advance p past the bytes we just consumed (including that '\n').
		p = p[idx+1:]
	}

	return total, nil
}

// Flush any remaining partial line (i.e. if the process wrote bytes
// without a trailing newline). We log that final fragment so it isn’t lost.
func (lw *loggerWriter) Flush() {
	lw.mu.Lock()
	defer lw.mu.Unlock()

	if lw.buf.Len() == 0 {
		return
	}

	// Whatever is left is a partial line (no '\n').
	line := lw.buf.String()

	// We log it as a normal message, but add "partial_noNL": true so that
	// if you’re looking at the structured logs, you know this line lacked a newline.
	lw.logger.Log(context.Background(), lw.level, line, "tag", lw.tag, "partial_noNL", true)

	// Clear the buffer
	lw.buf.Reset()
}
