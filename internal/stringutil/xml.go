package stringutil

import "strings"

// ExtractXMLTag returns the trimmed content of the first occurrence of
// <tag>…</tag> in s, or "" if not found.
func ExtractXMLTag(s, tag string) string {
	open := "<" + tag + ">"
	close := "</" + tag + ">"
	start := strings.Index(s, open)
	end := strings.Index(s, close)
	if start >= 0 && end > start+len(open) {
		return strings.TrimSpace(s[start+len(open) : end])
	}
	return ""
}
