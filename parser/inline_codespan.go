package parser

import "github.com/albertocavalcante/mdxgo/syntax"

// tryCodeSpan attempts to parse a code span starting at the current position.
// A code span begins with one or more backticks and is closed by a matching
// run of backticks. Per CommonMark spec (§6.1):
//   - The opening and closing runs must have the same number of backticks.
//   - If a single space is at both the start and end of the content, and the
//     content is not all spaces, those spaces are stripped. (We don't strip
//     for the CST; we preserve all bytes for round-trip.)
//   - Line endings within code spans are converted to spaces in HTML, but
//     the CST preserves original text for round-trip.
//   - A backtick string that is not closed is treated as literal text.
func (ip *inlineProcessor) tryCodeSpan() {
	s := ip.scanner
	start := s.pos

	// Count opening backticks.
	openLen := 0
	for s.pos+openLen < len(s.text) && s.text[s.pos+openLen] == '`' {
		openLen++
	}

	openTicks := s.text[start : start+openLen]

	// Search for matching closing backtick run.
	searchPos := start + openLen
	for searchPos < len(s.text) {
		// Find next backtick.
		idx := -1
		for i := searchPos; i < len(s.text); i++ {
			if s.text[i] == '`' {
				idx = i
				break
			}
		}
		if idx < 0 {
			break // no more backticks
		}

		// Count consecutive backticks at idx.
		closeLen := 0
		for idx+closeLen < len(s.text) && s.text[idx+closeLen] == '`' {
			closeLen++
		}

		if closeLen == openLen {
			// Found matching close. Emit code span.
			content := s.text[start+openLen : idx]
			closeTicks := s.text[idx : idx+closeLen]

			ip.flushText()
			ip.output = append(ip.output, syntax.NodeElement(
				syntax.NewGreenNode(syntax.CodeSpan, []syntax.GreenElement{
					syntax.TokenElement(syntax.NewGreenToken(syntax.BacktickToken, openTicks)),
					syntax.TokenElement(syntax.NewGreenToken(syntax.TextToken, content)),
					syntax.TokenElement(syntax.NewGreenToken(syntax.BacktickToken, closeTicks)),
				}),
			))
			s.pos = idx + closeLen
			return
		}

		searchPos = idx + closeLen
	}

	// No matching close found. Treat opening backticks as literal text.
	ip.textBuf = append(ip.textBuf, openTicks...)
	s.pos = start + openLen
}
