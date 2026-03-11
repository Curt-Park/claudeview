package stringutil

import (
	"os"
	"strings"
)

// MdTitle reads the first `# Heading` line from a markdown file and returns
// the heading text. Returns an empty string if the file cannot be read or
// contains no top-level heading.
func MdTitle(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	for _, line := range strings.SplitAfter(string(data), "\n") {
		line = strings.TrimRight(line, "\r\n")
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}
	return ""
}
