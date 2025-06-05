package testutils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// WriteScriptFile writes a shell script with the given content to a temp file.
// Returns the full path to the script file.
func WriteScriptFile(t *testing.T, content string, code int) string {
	t.Helper()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "script.sh")

	var sb strings.Builder
	sb.WriteString("#!/bin/sh\n")
	sb.WriteString(content)
	if !strings.HasSuffix(content, "\n") {
		sb.WriteString("\n")
	}
	sb.WriteString(fmt.Sprintf("exit %d\n", code))

	err := os.WriteFile(path, []byte(sb.String()), 0o700)
	assert.NoError(t, err)

	return path
}

func WriteFile(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "file.txt")
	err := os.WriteFile(path, []byte(content), 0o600)
	assert.NoError(t, err)
	return path
}
