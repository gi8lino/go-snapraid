package snapraid

import (
	"context"
	"log/slog"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// entry captures a logged message and its attributes.
type entry struct {
	msg   string
	attrs map[string]any
}

// testHandler implements slog.Handler to collect log entries.
type testHandler struct {
	mu      sync.Mutex
	entries *[]entry
}

func (h *testHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (h *testHandler) Handle(ctx context.Context, r slog.Record) error {
	attrs := make(map[string]any)
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})
	msg := r.Clone().Message
	h.mu.Lock()
	*h.entries = append(*h.entries, entry{msg: msg, attrs: attrs})
	h.mu.Unlock()
	return nil
}

func (h *testHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *testHandler) WithGroup(name string) slog.Handler {
	return h
}

// newTestLogger creates a slog.Logger whose testHandler collects entries into a slice.
func newTestLogger(collected *[]entry) *slog.Logger {
	h := &testHandler{entries: collected}
	return slog.New(h)
}

func TestLoggerWriter_Write(t *testing.T) {
	t.Parallel()

	t.Run("Single complete line", func(t *testing.T) {
		t.Parallel()

		var collected []entry
		logger := newTestLogger(&collected)
		lw := newLoggerWriter(logger, "testcomp", slog.LevelInfo)
		input := []byte("hello world\n")
		n, err := lw.Write(input)
		assert.NoError(t, err)
		assert.Equal(t, len(input), n)

		assert.Len(t, collected, 1)

		e := collected[0]
		assert.Equal(t, "hello world", e.msg)
		assert.Equal(t, "testcomp", e.attrs["tag"])
		_, hasPartial := e.attrs["partial_noNL"]
		assert.False(t, hasPartial)
	})

	t.Run("Multiple lines in one Write", func(t *testing.T) {
		t.Parallel()

		var collected []entry
		logger := newTestLogger(&collected)
		lw := newLoggerWriter(logger, "multi", slog.LevelDebug)
		input := []byte("line1\nline2\n")
		n, err := lw.Write(input)
		assert.NoError(t, err)
		assert.Equal(t, len(input), n)

		assert.Len(t, collected, 2)
		assert.Equal(t, "line1", collected[0].msg)
		assert.Equal(t, "multi", collected[0].attrs["tag"])
		assert.Equal(t, "line2", collected[1].msg)
		assert.Equal(t, "multi", collected[1].attrs["tag"])
	})

	t.Run("Partial line buffered, then completed", func(t *testing.T) {
		t.Parallel()

		var collected []entry
		logger := newTestLogger(&collected)
		lw := newLoggerWriter(logger, "partial", slog.LevelDebug)

		// First write: no newline, should buffer
		part := []byte("incomplete")
		n1, err1 := lw.Write(part)
		assert.NoError(t, err1)
		assert.Equal(t, len(part), n1)
		assert.Len(t, collected, 0)

		// Second write: completes the line
		rest := []byte(" line\n")
		n2, err2 := lw.Write(rest)
		assert.NoError(t, err2)
		assert.Equal(t, len(rest), n2)

		assert.Len(t, collected, 1)
		e := collected[0]
		assert.Equal(t, "incomplete line", e.msg)
		assert.Equal(t, "partial", e.attrs["tag"])
	})

	t.Run("Skip empty lines", func(t *testing.T) {
		t.Parallel()

		var collected []entry
		logger := newTestLogger(&collected)
		lw := newLoggerWriter(logger, "skip", slog.LevelDebug)
		input := []byte("\n   \nvalid\n\n")
		n, err := lw.Write(input)
		assert.NoError(t, err)
		assert.Equal(t, len(input), n)

		assert.Len(t, collected, 1)
		e := collected[0]
		assert.Equal(t, "valid", e.msg)
		assert.Equal(t, "skip", e.attrs["tag"])
	})
}

func TestLoggerWriter_Flush(t *testing.T) {
	t.Parallel()

	t.Run("Flush without partial does nothing", func(t *testing.T) {
		t.Parallel()

		var collected []entry
		logger := newTestLogger(&collected)
		lw := newLoggerWriter(logger, "no_partial", slog.LevelDebug)

		// No writes at all; Flush should not produce entries
		lw.Flush()
		assert.Len(t, collected, 0)

		// Write a complete line; Flush should still not produce extra
		_, _ = lw.Write([]byte("complete\n"))
		lw.Flush()

		assert.Len(t, collected, 1)
		assert.Equal(t, "complete", collected[0].msg)
	})

	t.Run("Flush logs remaining partial line", func(t *testing.T) {
		t.Parallel()

		var collected []entry
		logger := newTestLogger(&collected)
		lw := newLoggerWriter(logger, "flush", slog.LevelDebug)

		// Write a partial line without newline
		_, _ = lw.Write([]byte("leftover"))
		lw.Flush()

		assert.Len(t, collected, 1)
		e := collected[0]
		assert.Equal(t, "leftover", e.msg)
		assert.Equal(t, "flush", e.attrs["tag"])
		assert.Equal(t, true, e.attrs["partial_noNL"])

		// After flush, buffer should be cleared; another Flush does nothing
		collected = collected[:0]
		lw.Flush()
		assert.Len(t, collected, 0)
	})
}
