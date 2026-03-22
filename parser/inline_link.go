package parser

import (
	"github.com/albertocavalcante/mdxgo/syntax"
)

// tryOpenBracket handles '[' which may start a link or image.
// If the preceding character in textBuf is '!', this starts an image.
func (ip *inlineProcessor) tryOpenBracket() {
	s := ip.scanner

	isImage := false
	if len(ip.textBuf) > 0 && ip.textBuf[len(ip.textBuf)-1] == '!' {
		isImage = true
		ip.textBuf = ip.textBuf[:len(ip.textBuf)-1]
	}

	ip.flushText()

	bracketText := "["
	if isImage {
		bracketText = "!["
	}

	// Emit placeholder InlineText.
	inlineText := syntax.NewGreenNode(syntax.InlineText, []syntax.GreenElement{
		syntax.TokenElement(syntax.NewGreenToken(syntax.TextToken, bracketText)),
	})
	outputIdx := len(ip.output)
	ip.output = append(ip.output, syntax.NodeElement(inlineText))

	// Push bracket delimiter.
	ip.brackets = append(ip.brackets, &bracketDelim{
		isImage:   isImage,
		active:    true,
		outputIdx: outputIdx,
	})

	s.advance(1) // skip '['
}

// tryCloseBracket handles ']' which may close a link or image.
func (ip *inlineProcessor) tryCloseBracket() {
	s := ip.scanner

	// Find the most recent active bracket opener.
	openerIdx := -1
	for i := len(ip.brackets) - 1; i >= 0; i-- {
		if ip.brackets[i].active {
			openerIdx = i
			break
		}
	}

	if openerIdx < 0 {
		// No active opener: literal ']'.
		ip.textBuf = append(ip.textBuf, ']')
		s.advance(1)
		return
	}

	opener := ip.brackets[openerIdx]

	// Flush any pending text before we look ahead.
	ip.flushText()

	// Look ahead past ']' for a link destination: (url "title")
	closeBracketPos := s.pos
	s.advance(1) // skip ']'

	if s.pos < len(s.text) && s.text[s.pos] == '(' {
		parenStart := s.pos
		if endPos, ok := tryParseLinkDest(s.text, parenStart); ok {
			// Found valid inline link.
			destRawText := s.text[parenStart:endPos] // includes ( and )

			ip.buildLinkNode(opener, openerIdx, destRawText, closeBracketPos)
			s.pos = endPos
			return
		}
	}

	// No valid link destination found.
	opener.active = false

	// Emit ']' as text.
	ip.output = append(ip.output, syntax.NodeElement(
		syntax.NewGreenNode(syntax.InlineText, []syntax.GreenElement{
			syntax.TokenElement(syntax.NewGreenToken(syntax.TextToken, "]")),
		}),
	))

	// Remove this bracket from the stack.
	ip.brackets = append(ip.brackets[:openerIdx], ip.brackets[openerIdx+1:]...)
}

// tryParseLinkDest parses an inline link destination+title starting at '('.
// Returns the end position (after ')') and success.
func tryParseLinkDest(text string, pos int) (int, bool) {
	if pos >= len(text) || text[pos] != '(' {
		return 0, false
	}
	pos++ // skip '('

	// Skip optional whitespace (including newlines).
	pos = skipLinkWhitespace(text, pos)
	if pos >= len(text) {
		return 0, false
	}

	// Empty link: ()
	if text[pos] == ')' {
		return pos + 1, true
	}

	// Parse destination.
	if text[pos] == '<' {
		// Angle-bracket destination.
		pos++ // skip '<'
		for pos < len(text) {
			if text[pos] == '>' {
				pos++ // skip '>'
				break
			}
			if text[pos] == '\n' || text[pos] == '\r' || text[pos] == '<' {
				return 0, false
			}
			if text[pos] == '\\' && pos+1 < len(text) {
				pos++ // skip escaped char
			}
			pos++
		}
	} else {
		// Plain destination: balanced parentheses, no spaces.
		parenDepth := 0
		for pos < len(text) {
			b := text[pos]
			if b == '(' {
				parenDepth++
				if parenDepth > 32 {
					return 0, false
				}
			} else if b == ')' {
				if parenDepth == 0 {
					break
				}
				parenDepth--
			} else if b <= ' ' {
				break // space, tab, or control characters end the destination
			}
			if b == '\\' && pos+1 < len(text) {
				pos++ // skip escaped char
			}
			pos++
		}
	}

	// Skip optional whitespace.
	pos = skipLinkWhitespace(text, pos)
	if pos >= len(text) {
		return 0, false
	}

	// Check for closing paren (no title).
	if text[pos] == ')' {
		return pos + 1, true
	}

	// Try to parse optional title.
	if text[pos] == '"' || text[pos] == '\'' || text[pos] == '(' {
		titleEnd, ok := parseLinkTitle(text, pos)
		if !ok {
			return 0, false
		}
		pos = titleEnd

		// Skip optional whitespace after title.
		pos = skipLinkWhitespace(text, pos)
		if pos >= len(text) {
			return 0, false
		}
	}

	// Must end with ')'.
	if text[pos] != ')' {
		return 0, false
	}
	return pos + 1, true
}

// parseLinkTitle parses a link title in quotes or parens starting at pos.
// Returns the position after the closing delimiter.
func parseLinkTitle(text string, pos int) (int, bool) {
	if pos >= len(text) {
		return 0, false
	}
	openDelim := text[pos]
	closeDelim := openDelim
	if openDelim == '(' {
		closeDelim = ')'
	}
	pos++ // skip opening delimiter

	for pos < len(text) {
		if text[pos] == closeDelim {
			return pos + 1, true
		}
		if text[pos] == '\\' && pos+1 < len(text) {
			pos++ // skip escaped char
		}
		pos++
	}
	return 0, false
}

func skipLinkWhitespace(text string, pos int) int {
	for pos < len(text) && (text[pos] == ' ' || text[pos] == '\t' || text[pos] == '\n' || text[pos] == '\r') {
		pos++
	}
	return pos
}

// buildLinkNode constructs a Link or Image node from the matched bracket
// opener, link content, and raw destination text.
func (ip *inlineProcessor) buildLinkNode(opener *bracketDelim, openerIdx int, destRawText string, _ int) {
	nodeKind := syntax.Link
	if opener.isImage {
		nodeKind = syntax.Image
	}

	openOutputIdx := opener.outputIdx

	// Build children for the Link/Image node.
	var children []syntax.GreenElement

	// Opening marker: [ or ![
	if opener.isImage {
		children = append(children, syntax.TokenElement(
			syntax.NewGreenToken(syntax.ExclMarkToken, "!"),
		))
	}
	children = append(children, syntax.TokenElement(
		syntax.NewGreenToken(syntax.OpenBracketToken, "["),
	))

	// Content between opener and closer (everything in output after the opener).
	for i := openOutputIdx + 1; i < len(ip.output); i++ {
		children = append(children, ip.output[i])
	}

	// Closing bracket.
	children = append(children, syntax.TokenElement(
		syntax.NewGreenToken(syntax.CloseBracketToken, "]"),
	))

	// Destination text (includes parens): (url "title")
	children = append(children, syntax.TokenElement(
		syntax.NewGreenToken(syntax.TextToken, destRawText),
	))

	linkNode := syntax.NewGreenNode(nodeKind, children)

	// Rebuild output: keep everything before opener, replace with link node.
	newOutput := make([]syntax.GreenElement, openOutputIdx, openOutputIdx+1)
	copy(newOutput, ip.output[:openOutputIdx])
	newOutput = append(newOutput, syntax.NodeElement(linkNode))

	// Mark emphasis delimiters that were consumed by the link as inactive.
	// Their content is now inside the link node.
	for _, d := range ip.delimiters {
		if d.outputIdx >= openOutputIdx {
			d.outputIdx = -1
			d.active = false
		}
	}

	ip.output = newOutput

	// For non-image links, deactivate earlier '[' openers (links can't nest).
	if !opener.isImage {
		for i := 0; i < openerIdx; i++ {
			if !ip.brackets[i].isImage {
				ip.brackets[i].active = false
			}
		}
	}

	// Remove processed and intervening brackets.
	ip.brackets = ip.brackets[:openerIdx]
}

// bracketDelim represents an opening bracket on the bracket stack.
type bracketDelim struct {
	isImage   bool
	active    bool
	outputIdx int
}
