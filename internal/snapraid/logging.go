package snapraid

import (
	"bytes"
	"log/slog"
	"sync"
)

// loggerWriter is an io.Writer that splits on newlines and sends each line
// into a structured slog.Logger with a given component name.
type loggerWriter struct {
	logger    *slog.Logger
	component string

	// internal buffer for partial lines
	buf bytes.Buffer

	// mutex to protect buf if Write is ever called concurrently
	mu sync.Mutex
}

// newLoggerWriter constructs a loggerWriter that tags every line with component.
func newLoggerWriter(logger *slog.Logger, component string) *loggerWriter {
	return &loggerWriter{
		logger:    logger,
		component: component,
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

		// Remove the trailing '\n' so the logged message doesn’t include it.
		line := string(bytes.TrimRight(lineBytes, "\n"))

		// Now emit it via slog.Logger.Info. The first argument is the
		// “message,” and subsequent args must be key/value pairs:
		//
		//    logger.Info( <message string>, <key1>, <value1>, <key2>, <value2>, … )
		//
		// We choose to log the actual file‐output line as the message, and
		// attach a “component” attribute so we know which step produced it.
		lw.logger.Info(
			line,
			"component", lw.component,
		)

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
	lw.logger.Info(
		line,
		"component", lw.component,
		"partial_noNL", true,
	)

	// Clear the buffer
	lw.buf.Reset()
}
