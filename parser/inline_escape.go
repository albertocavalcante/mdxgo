package parser

import (
	"github.com/albertocavalcante/mdxgo/syntax"
)

// tryBackslashEscape attempts to parse a backslash escape at the current position.
// Per CommonMark spec (§2.4):
//   - Any ASCII punctuation character can be escaped with a backslash.
//   - A backslash before a newline creates a hard line break.
//   - A backslash before any other character is not an escape; the backslash
//     is treated as a literal backslash.
func (ip *inlineProcessor) tryBackslashEscape() {
	s := ip.scanner
	// s.text[s.pos] == '\\'
	if s.pos+1 >= len(s.text) {
		// Backslash at end of input: literal.
		ip.textBuf = append(ip.textBuf, '\\')
		s.advance(1)
		return
	}

	next := s.text[s.pos+1]

	// Backslash before newline: hard line break.
	if next == '\n' || next == '\r' {
		ip.flushText()
		// Determine line ending length.
		nlLen := 1
		if next == '\r' && s.pos+2 < len(s.text) && s.text[s.pos+2] == '\n' {
			nlLen = 2
		}
		nlText := s.text[s.pos+1 : s.pos+1+nlLen]
		ip.output = append(ip.output, syntax.NodeElement(
			syntax.NewGreenNode(syntax.HardLineBreak, []syntax.GreenElement{
				syntax.TokenElement(syntax.NewGreenToken(syntax.BackslashToken, `\`)),
				syntax.TokenElement(syntax.NewGreenToken(syntax.HardBreakToken, nlText)),
			}),
		))
		s.advance(1 + nlLen)
		return
	}

	// Backslash before ASCII punctuation: escape.
	if isASCIIPunctuation(next) {
		ip.flushText()
		ip.output = append(ip.output, syntax.NodeElement(
			syntax.NewGreenNode(syntax.BackslashEscape, []syntax.GreenElement{
				syntax.TokenElement(syntax.NewGreenToken(syntax.BackslashToken, `\`)),
				syntax.TokenElement(syntax.NewGreenToken(syntax.TextToken, string(next))),
			}),
		))
		s.advance(2)
		return
	}

	// Not a valid escape: literal backslash.
	ip.textBuf = append(ip.textBuf, '\\')
	s.advance(1)
}

// tryEntityRef attempts to parse an entity reference at the current position.
// Per CommonMark spec (§2.5):
//   - HTML entity references: &name;
//   - Decimal numeric character references: &#digits;
//   - Hexadecimal numeric character references: &#x/&#X hexdigits;
// For the CST we preserve the exact source text; we just recognize entity patterns.
func (ip *inlineProcessor) tryEntityRef() {
	s := ip.scanner
	// s.text[s.pos] == '&'
	start := s.pos

	// Look for semicolon to close the entity.
	semiIdx := -1
	for i := start + 1; i < len(s.text) && i < start+32; i++ {
		if s.text[i] == ';' {
			semiIdx = i
			break
		}
		if s.text[i] == '\n' || s.text[i] == '\r' || s.text[i] == '&' {
			break
		}
	}

	if semiIdx < 0 {
		// Not an entity reference: literal &.
		ip.textBuf = append(ip.textBuf, '&')
		s.advance(1)
		return
	}

	body := s.text[start+1 : semiIdx]
	if len(body) == 0 {
		ip.textBuf = append(ip.textBuf, '&')
		s.advance(1)
		return
	}

	valid := false
	if body[0] == '#' {
		valid = isNumericCharRef(body)
	} else {
		valid = isHTMLEntityName(body)
	}

	if !valid {
		ip.textBuf = append(ip.textBuf, '&')
		s.advance(1)
		return
	}

	// Valid entity reference.
	entityText := s.text[start : semiIdx+1]
	ip.flushText()
	ip.output = append(ip.output, syntax.NodeElement(
		syntax.NewGreenNode(syntax.EntityRef, []syntax.GreenElement{
			syntax.TokenElement(syntax.NewGreenToken(syntax.EntityToken, entityText)),
		}),
	))
	s.pos = semiIdx + 1
}

// isNumericCharRef checks if body (without & and ;) is a valid numeric character reference.
// body starts with '#'.
func isNumericCharRef(body string) bool {
	if len(body) < 2 {
		return false
	}
	if body[1] == 'x' || body[1] == 'X' {
		// Hexadecimal: &#xHH...
		if len(body) < 3 || len(body) > 9 {
			return false
		}
		for i := 2; i < len(body); i++ {
			if !isHexDigit(body[i]) {
				return false
			}
		}
		return true
	}
	// Decimal: &#DDD...
	if len(body) > 8 {
		return false
	}
	for i := 1; i < len(body); i++ {
		if body[i] < '0' || body[i] > '9' {
			return false
		}
	}
	return true
}

func isHexDigit(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
}

// isHTMLEntityName checks if name is a valid HTML entity name.
// Per CommonMark, we recognize all HTML5 named character references.
// For simplicity and correctness, we check that it looks like an entity
// name (1-31 alphanumeric characters). A full check would use the
// official HTML entity list, but for a CST parser we're lenient:
// the parser's job is to recognize the syntactic form, not validate semantics.
func isHTMLEntityName(name string) bool {
	if len(name) == 0 || len(name) > 31 {
		return false
	}
	for i := 0; i < len(name); i++ {
		b := name[i]
		if !((b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')) {
			return false
		}
	}
	return true
}

// tryLineBreak handles newlines in inline content.
// A newline in inline content is either a soft line break or part of a
// hard line break (when preceded by 2+ spaces or a backslash).
func (ip *inlineProcessor) tryLineBreak() {
	s := ip.scanner
	b := s.text[s.pos]

	// Determine the newline length.
	nlLen := 1
	if b == '\r' && s.pos+1 < len(s.text) && s.text[s.pos+1] == '\n' {
		nlLen = 2
	}

	nlText := s.text[s.pos : s.pos+nlLen]

	// A newline at the very end of inline content (last line of paragraph)
	// is not a line break — it's just part of the block structure.
	// Check if there's any non-whitespace content after this newline.
	afterNL := s.pos + nlLen
	if afterNL >= len(s.text) {
		// Newline at end: include in text buffer.
		ip.textBuf = append(ip.textBuf, nlText...)
		s.advance(nlLen)
		return
	}

	// Check if the text before this newline ends with 2+ spaces (hard break).
	trailingSpaces := 0
	for i := len(ip.textBuf) - 1; i >= 0; i-- {
		if ip.textBuf[i] == ' ' {
			trailingSpaces++
		} else {
			break
		}
	}

	if trailingSpaces >= 2 {
		// Hard line break: spaces + newline.
		// Remove the trailing spaces from textBuf and emit them as part of the hard break.
		spacesText := string(ip.textBuf[len(ip.textBuf)-trailingSpaces:])
		ip.textBuf = ip.textBuf[:len(ip.textBuf)-trailingSpaces]
		ip.flushText()
		ip.output = append(ip.output, syntax.NodeElement(
			syntax.NewGreenNode(syntax.HardLineBreak, []syntax.GreenElement{
				syntax.TokenElement(syntax.NewGreenToken(syntax.HardBreakToken, spacesText)),
				syntax.TokenElement(syntax.NewGreenToken(syntax.SoftBreakToken, nlText)),
			}),
		))
		s.advance(nlLen)
		return
	}

	// Soft line break.
	ip.flushText()
	ip.output = append(ip.output, syntax.NodeElement(
		syntax.NewGreenNode(syntax.SoftLineBreak, []syntax.GreenElement{
			syntax.TokenElement(syntax.NewGreenToken(syntax.SoftBreakToken, nlText)),
		}),
	))
	s.advance(nlLen)
}

// tryHardBreakSpaces handles spaces that might be part of a hard line break.
// We only need to handle the case where we see a space — the actual hard break
// logic is in tryLineBreak (when the newline comes). So here we just
// accumulate the space into the text buffer.
func (ip *inlineProcessor) tryHardBreakSpaces() {
	ip.textBuf = append(ip.textBuf, ip.scanner.text[ip.scanner.pos])
	ip.scanner.advance(1)
}
