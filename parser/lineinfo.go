package parser

import "strings"

// lineInfo holds pre-computed properties of a source line to avoid
// re-scanning in each block parser.
type lineInfo struct {
	raw     string // full line content including newline
	trimmed string // leading whitespace removed
	indent  int    // number of leading spaces (tabs count as 1 for offset, but we preserve them)
	blank   bool   // line is entirely whitespace
	newline string // the line ending: "\n", "\r\n", "\r", or ""
	content string // line content without the trailing newline
}

// analyeLine pre-computes properties for a source line.
func analyzeLine(raw string) lineInfo {
	li := lineInfo{raw: raw}

	// Find and strip trailing newline.
	switch {
	case strings.HasSuffix(raw, "\r\n"):
		li.newline = "\r\n"
		li.content = raw[:len(raw)-2]
	case strings.HasSuffix(raw, "\n"):
		li.newline = "\n"
		li.content = raw[:len(raw)-1]
	case strings.HasSuffix(raw, "\r"):
		li.newline = "\r"
		li.content = raw[:len(raw)-1]
	default:
		li.content = raw
	}

	// Count leading whitespace and produce trimmed version.
	i := countLeadingWhitespace(li.content)
	li.indent = i
	li.trimmed = li.content[i:]
	li.blank = li.trimmed == ""

	return li
}

// countLeadingWhitespace returns the number of leading space/tab characters.
// Tabs count as 1 for indent tracking (not expanded).
func countLeadingWhitespace(s string) int {
	for i := range len(s) {
		if s[i] != ' ' && s[i] != '\t' {
			return i
		}
	}
	return len(s)
}
