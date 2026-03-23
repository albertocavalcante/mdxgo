package parser

// findMatchingBrace scans text starting at startPos (which must point to a '{')
// and returns the position of the matching '}' and true, or (-1, false) if no
// match is found before end of text.
//
// It performs pure brace counting: '{' increments depth, '}' decrements.
// Braces inside string literals (single-quoted, double-quoted, and template
// literals) are ignored. Backslash escapes within strings are respected.
//
// This is intentionally agnostic — no JS semantics, just syntax recognition.
func findMatchingBrace(text string, startPos int) (int, bool) {
	if startPos >= len(text) || text[startPos] != '{' {
		return -1, false
	}

	depth := 0
	i := startPos

	for i < len(text) {
		switch text[i] {
		case '{':
			depth++
			i++
		case '}':
			depth--
			if depth == 0 {
				return i, true
			}
			i++
		case '\'', '"':
			// Skip string literal.
			quote := text[i]
			i++
			for i < len(text) {
				if text[i] == '\\' && i+1 < len(text) {
					i += 2 // skip escaped char
					continue
				}
				if text[i] == quote {
					i++
					break
				}
				i++
			}
		case '`':
			// Skip template literal (may contain ${expr} but we treat
			// the whole thing as opaque — nested braces inside template
			// expressions are rare in MDX attributes and the CST parser's
			// job is just brace counting, not full JS parsing).
			i++
			for i < len(text) {
				if text[i] == '\\' && i+1 < len(text) {
					i += 2
					continue
				}
				if text[i] == '`' {
					i++
					break
				}
				i++
			}
		case '/':
			// Skip line comments (//) and block comments (/* ... */).
			if i+1 < len(text) {
				if text[i+1] == '/' {
					// Line comment: skip to end of line.
					i += 2
					for i < len(text) && text[i] != '\n' && text[i] != '\r' {
						i++
					}
					continue
				}
				if text[i+1] == '*' {
					// Block comment: skip to */.
					i += 2
					for i+1 < len(text) {
						if text[i] == '*' && text[i+1] == '/' {
							i += 2
							break
						}
						i++
					}
					// If we ran out of text without finding */, i is at end.
					if i+1 >= len(text) && !(i >= 2 && text[i-2] == '*' && text[i-1] == '/') {
						i = len(text)
					}
					continue
				}
			}
			i++
		default:
			i++
		}
	}

	return -1, false
}

// findMatchingBraceInLines works like findMatchingBrace but operates across
// multiple source lines. It takes the first line (starting from braceOffset
// within that line), and a function to get subsequent lines.
// Returns the total number of bytes consumed (across all lines) from the
// start of the first line, and whether a match was found.
func findMatchingBraceInLines(firstLine string, braceOffset int, nextLine func() (string, bool)) (int, bool) {
	// Build up the text to scan using a byte buffer to avoid repeated
	// string concatenation (O(n^2) in the worst case).
	buf := make([]byte, 0, len(firstLine)*2) //nolint:mnd // pre-allocate ~2x first line
	buf = append(buf, firstLine...)

	// Try the first line alone.
	text := string(buf)
	if end, ok := findMatchingBrace(text, braceOffset); ok {
		return end, true
	}

	// Keep adding lines.
	for {
		line, ok := nextLine()
		if !ok {
			break
		}
		buf = append(buf, line...)
		text = string(buf)
		if end, ok := findMatchingBrace(text, braceOffset); ok {
			return end, true
		}
	}

	return -1, false
}
