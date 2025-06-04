package snapraid

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// WriteJSON writes the RunResult to a timestamped JSON file in the given directory.
func (r RunResult) WriteJSON(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output dir: %w", err)
	}

	filename := r.Timestamp + ".json"
	path := filepath.Join(dir, filename)

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create result file: %w", err)
	}
	defer file.Close() // nolint:errcheck

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	if err := enc.Encode(r); err != nil {
		return fmt.Errorf("failed to encode result JSON: %w", err)
	}
	return nil
}
