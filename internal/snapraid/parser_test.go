package snapraid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDiff(t *testing.T) {
	lines := []string{
		"Comparing...",
		"add filme/XOXO\\ \\(2016\\)/XOXO.2016.German.DL.1080p.WEB.x264.iNTERNAL-BiGiNT.mkv",
		"add filme/XOXO\\ \\(2016\\)/XOXO.2016.German.DL.1080p.WEB.x264.iNTERNAL-BiGiNT.sfv",
		"   21146 equal",
		"       2 added",
		"       0 removed",
		"       0 updated",
		"       0 moved",
		"       0 copied",
		"       0 restored",
	}

	result := parseDiff(lines)

	assert.Equal(t, 21146, result.Equal)
	assert.Equal(t, 2, result.Added)
	assert.Equal(t, []string{
		"filme/XOXO\\ \\(2016\\)/XOXO.2016.German.DL.1080p.WEB.x264.iNTERNAL-BiGiNT.mkv",
		"filme/XOXO\\ \\(2016\\)/XOXO.2016.German.DL.1080p.WEB.x264.iNTERNAL-BiGiNT.sfv",
	}, result.AddedFiles)
}
