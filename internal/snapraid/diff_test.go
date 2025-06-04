package snapraid

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasChanges(t *testing.T) {
	t.Parallel()

	t.Run("No file changes", func(t *testing.T) {
		t.Parallel()
		dr := DiffResult{
			Equal:    42,
			Added:    nil,
			Removed:  nil,
			Updated:  nil,
			Moved:    nil,
			Copied:   nil,
			Restored: nil,
		}
		assert.False(t, dr.HasChanges())
	})

	t.Run("One added file", func(t *testing.T) {
		t.Parallel()
		dr := DiffResult{
			Equal: 10,
			Added: []string{"/path/to/new.mp4"},
		}
		assert.True(t, dr.HasChanges())
	})

	t.Run("Multiple change types", func(t *testing.T) {
		t.Parallel()
		dr := DiffResult{
			Equal:   5,
			Added:   []string{"/a.mp4"},
			Removed: []string{"/old1.txt", "/old2.txt"},
			Updated: []string{"/config.yaml"},
		}
		assert.True(t, dr.HasChanges())
	})

	t.Run("Real example", func(t *testing.T) {
		t.Parallel()

		diff := `Comparing...
add movies/Gladiator\ \(2000\)/Gladiator.2000.German.AC3.DL.1080p.BluRay.x264.FuN.mk
add movies/Interstellar\ \(2014\)/Interstellar.2014.German.AC3.DL.1080p.BluRay.x264.FuN.mkv
add movies/Judge\ Dredd\ \(1995\)/Judge.Dredd.1995.German.AC3.DL.1080p.BluRay.x264.FuN.mkv
add movies/Predator\ \(1987\)/Predator.1987.German.AC3.DL.1080p.BluRay.x264.FuN.mkv
add movies/Zoolander\ \(2001\)/Zoolander.2001.German.AC3.DL.1080p.BluRay.x265-FuN.mkv
copy movies/Mile\ 22\ \(2002\)/Mile.22.2002.German.AC3.DL.1080p.BluRay.x264.FuN.mkv
move movies/The\ Shawshank\ Redemption\ \(1994\)/The.Shawshank.Redemption.1994.German.AC3.DL.1080p.BluRay.x264.FuN.mkv
remove movies/XOXO\ \(2016\)/XOXO.2016.German.DL.1080p.WEB.x264.iNTERNAL-BiGiNT.mkv
remove movies/Zoolander\ 2\ \(2016\)/Zoolander.2.2016.German.AC3.DL.1080p.BluRay.x265-FuN.mkv
restore movies/Crazy, Stupid, Love\. A\. Piano\. \(2009\)/Crazy.Stupid.Love.A.Piano.2009.German.AC3.DL.1080p.BluRay.x264.FuN.mkv
update movies/The\ Matrix\ \(1999\)/The.Matrix.1999.German.AC3.DL.1080p.BluRay.x264.FuN.mkv

   21156 equal
       5 added
       2 removed
       1 updated
       1 moved
       1 copied
       1 restored
There are differences!
`
		lines := strings.Split(diff, "\n")

		dr := parseDiff(lines)

		// Summary count
		assert.Equal(t, 21156, dr.Equal)

		expectedAdded := []string{
			"movies/Gladiator\\ \\(2000\\)/Gladiator.2000.German.AC3.DL.1080p.BluRay.x264.FuN.mk",
			"movies/Interstellar\\ \\(2014\\)/Interstellar.2014.German.AC3.DL.1080p.BluRay.x264.FuN.mkv",
			"movies/Judge\\ Dredd\\ \\(1995\\)/Judge.Dredd.1995.German.AC3.DL.1080p.BluRay.x264.FuN.mkv",
			"movies/Predator\\ \\(1987\\)/Predator.1987.German.AC3.DL.1080p.BluRay.x264.FuN.mkv",
			"movies/Zoolander\\ \\(2001\\)/Zoolander.2001.German.AC3.DL.1080p.BluRay.x265-FuN.mkv",
		}
		expectedRemoved := []string{
			"movies/XOXO\\ \\(2016\\)/XOXO.2016.German.DL.1080p.WEB.x264.iNTERNAL-BiGiNT.mkv",
			"movies/Zoolander\\ 2\\ \\(2016\\)/Zoolander.2.2016.German.AC3.DL.1080p.BluRay.x265-FuN.mkv",
		}
		expectedUpdated := []string{
			"movies/The\\ Matrix\\ \\(1999\\)/The.Matrix.1999.German.AC3.DL.1080p.BluRay.x264.FuN.mkv",
		}
		expectedMoved := []string{
			"movies/The\\ Shawshank\\ Redemption\\ \\(1994\\)/The.Shawshank.Redemption.1994.German.AC3.DL.1080p.BluRay.x264.FuN.mkv",
		}
		expectedCopied := []string{
			"movies/Mile\\ 22\\ \\(2002\\)/Mile.22.2002.German.AC3.DL.1080p.BluRay.x264.FuN.mkv",
		}
		expectedRestored := []string{
			"movies/Crazy, Stupid, Love\\. A\\. Piano\\. \\(2009\\)/Crazy.Stupid.Love.A.Piano.2009.German.AC3.DL.1080p.BluRay.x264.FuN.mkv",
		}

		assert.Equal(t, 21156, dr.Equal)
		assert.Equal(t, expectedAdded, dr.Added)
		assert.Equal(t, expectedRemoved, dr.Removed)
		assert.Equal(t, expectedUpdated, dr.Updated)
		assert.Equal(t, expectedMoved, dr.Moved)
		assert.Equal(t, expectedCopied, dr.Copied)
		assert.Equal(t, expectedRestored, dr.Restored)
	})

	t.Run("Ignores unmatched lines", func(t *testing.T) {
		t.Parallel()

		lines := []string{
			"some random text",
			"123 unknowncategory",
			"addwithoutspace/file.txt",
			"added five",    // not a number => treated as action "added" but "added" is not recognized as action
			"move",          // incomplete
			"remove ",       // missing path
			"   50notEqual", // no space between number and word
			"",
			"\t\t",
		}

		dr := parseDiff(lines)

		// Everything should be zero/empty
		assert.Zero(t, dr.Equal)
		assert.Empty(t, dr.Added)
		assert.Empty(t, dr.Removed)
		assert.Empty(t, dr.Updated)
		assert.Empty(t, dr.Moved)
		assert.Empty(t, dr.Copied)
		assert.Empty(t, dr.Restored)
	})

	t.Run("Mixed counts and paths", func(t *testing.T) {
		t.Parallel()

		lines := []string{
			"2 added",
			"add /one.mp4",
			"5 removed",
			"remove /old1.txt",
			"remove /old2.txt",
			"1 updated",
			"update /config.yaml",
			"   10 equal",
		}

		dr := parseDiff(lines)

		// Summary counts from "2 added" and "5 removed" are ignored (only "equal" is counted)
		assert.Equal(t, 10, dr.Equal)

		// File lists should be populated
		assert.Equal(t, []string{"/one.mp4"}, dr.Added)
		assert.Equal(t, []string{"/old1.txt", "/old2.txt"}, dr.Removed)
		assert.Equal(t, []string{"/config.yaml"}, dr.Updated)

		// Other categories remain empty
		assert.Empty(t, dr.Moved)
		assert.Empty(t, dr.Copied)
		assert.Empty(t, dr.Restored)
	})
}

func TestValidateThresholds(t *testing.T) {
	t.Parallel()

	t.Run("No violation when all counts within thresholds", func(t *testing.T) {
		t.Parallel()

		result := DiffResult{
			Added:    []string{"a", "b"},
			Removed:  []string{"x"},
			Updated:  []string{"u1", "u2", "u3"},
			Moved:    []string{"m1"},
			Copied:   []string{"c1", "c2"},
			Restored: []string{},
		}
		thresholds := Thresholds{
			Add:     5,
			Remove:  2,
			Update:  5,
			Move:    2,
			Copy:    3,
			Restore: 1,
		}

		err := validateThresholds(result, thresholds)
		assert.NoError(t, err)
	})

	t.Run("Added violation", func(t *testing.T) {
		t.Parallel()

		result := DiffResult{
			Added: []string{"f1", "f2", "f3", "f4"},
		}
		thresholds := Thresholds{
			Add:     3,
			Remove:  -1,
			Update:  -1,
			Move:    -1,
			Copy:    -1,
			Restore: -1,
		}

		err := validateThresholds(result, thresholds)
		assert.Error(t, err)
		assert.EqualError(t, err, "added files exceed threshold (4 > 3)")
	})

	t.Run("Removed violation", func(t *testing.T) {
		t.Parallel()

		result := DiffResult{
			Removed: []string{"o1", "o2", "o3"},
		}
		thresholds := Thresholds{
			Add:     -1,
			Remove:  2,
			Update:  -1,
			Move:    -1,
			Copy:    -1,
			Restore: -1,
		}

		err := validateThresholds(result, thresholds)
		assert.Error(t, err)
		assert.EqualError(t, err, "removed files exceed threshold (3 > 2)")
	})

	t.Run("Updated violation", func(t *testing.T) {
		t.Parallel()

		result := DiffResult{
			Updated: []string{"u1", "u2", "u3", "u4", "u5"},
		}
		thresholds := Thresholds{
			Add:     -1,
			Remove:  -1,
			Update:  4,
			Move:    -1,
			Copy:    -1,
			Restore: -1,
		}

		err := validateThresholds(result, thresholds)
		assert.Error(t, err)
		assert.EqualError(t, err, "updated files exceed threshold (5 > 4)")
	})

	t.Run("Moved violation", func(t *testing.T) {
		t.Parallel()

		result := DiffResult{
			Moved: []string{"m1", "m2", "m3"},
		}
		thresholds := Thresholds{
			Add:     -1,
			Remove:  -1,
			Update:  -1,
			Move:    2,
			Copy:    -1,
			Restore: -1,
		}

		err := validateThresholds(result, thresholds)
		assert.Error(t, err)
		assert.EqualError(t, err, "moved files exceed threshold (3 > 2)")
	})

	t.Run("Copied violation", func(t *testing.T) {
		t.Parallel()

		result := DiffResult{
			Copied: []string{"c1", "c2", "c3", "c4"},
		}
		thresholds := Thresholds{
			Add:     -1,
			Remove:  -1,
			Update:  -1,
			Move:    -1,
			Copy:    3,
			Restore: -1,
		}

		err := validateThresholds(result, thresholds)
		assert.Error(t, err)
		assert.EqualError(t, err, "copied files exceed threshold (4 > 3)")
	})

	t.Run("Restored violation", func(t *testing.T) {
		t.Parallel()

		result := DiffResult{
			Restored: []string{"r1", "r2"},
		}
		thresholds := Thresholds{
			Add:     -1,
			Remove:  -1,
			Update:  -1,
			Move:    -1,
			Copy:    -1,
			Restore: 1,
		}

		err := validateThresholds(result, thresholds)
		assert.Error(t, err)
		assert.EqualError(t, err, "restored files exceed threshold (2 > 1)")
	})

	t.Run("Negative thresholds disable checks", func(t *testing.T) {
		t.Parallel()

		result := DiffResult{
			Added:    []string{"a1", "a2", "a3", "a4"},
			Removed:  []string{"r1", "r2", "r3"},
			Updated:  []string{"u1", "u2"},
			Moved:    []string{"m1", "m2", "m3"},
			Copied:   []string{"c1", "c2", "c3", "c4", "c5"},
			Restored: []string{"t1"},
		}
		thresholds := Thresholds{
			Add:     -1,
			Remove:  -1,
			Update:  -1,
			Move:    -1,
			Copy:    -1,
			Restore: -1,
		}

		err := validateThresholds(result, thresholds)
		assert.NoError(t, err)
	})
}
