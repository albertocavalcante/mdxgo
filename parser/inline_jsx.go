package parser

import "github.com/albertocavalcante/mdxgo/syntax"

// tryJSXInline attempts to parse inline JSX at the current position (starting
// with '<'). In MDX mode, '<' triggers JSX parsing instead of autolink/raw HTML.
//
// If the content after '<' doesn't form a valid JSX tag, '<' is treated as
// literal text.
//
// Structure: JSXInline { JSXOpeningTag | JSXClosingTag | JSXSelfClosingTag | JSXFragment }
func (ip *inlineProcessor) tryJSXInline() {
	s := ip.scanner
	start := s.pos

	// Quick validation: what follows '<'?
	if start+1 >= len(s.text) {
		ip.textBuf = append(ip.textBuf, '<')
		s.advance(1)
		return
	}

	next := s.text[start+1]

	// Fragment: <>
	if next == '>' {
		ip.flushText()
		ip.output = append(ip.output, syntax.NodeElement(
			syntax.NewGreenNode(syntax.JSXInline, []syntax.GreenElement{
				syntax.NodeElement(syntax.NewGreenNode(syntax.JSXFragment, []syntax.GreenElement{
					syntax.TokenElement(syntax.NewGreenToken(syntax.JSXOpenAngle, "<")),
					syntax.TokenElement(syntax.NewGreenToken(syntax.JSXCloseAngle, ">")),
				})),
			}),
		))
		s.pos = start + 2
		return
	}

	// Closing tag: </
	if next == '/' {
		if start+2 < len(s.text) && (isJSXIdentStart(s.text[start+2]) || s.text[start+2] == '>') {
			// Looks like closing tag or closing fragment.
		} else {
			ip.textBuf = append(ip.textBuf, '<')
			s.advance(1)
			return
		}
	} else if next == '!' || next == '?' {
		// Not JSX — treat '<' as literal text (comments, PIs not JSX).
		ip.textBuf = append(ip.textBuf, '<')
		s.advance(1)
		return
	} else if !isJSXIdentStart(next) {
		// Not a valid tag start.
		ip.textBuf = append(ip.textBuf, '<')
		s.advance(1)
		return
	}

	// Check for URL-like patterns: <scheme://
	if isJSXIdentStart(next) {
		// Scan ahead to see if this looks like a URL.
		j := start + 1
		for j < len(s.text) && (isJSXIdentPart(s.text[j]) || s.text[j] == '.' || s.text[j] == ':') {
			if s.text[j] == ':' && j+2 < len(s.text) && s.text[j+1] == '/' && s.text[j+2] == '/' {
				// Looks like a URL scheme — not JSX.
				ip.textBuf = append(ip.textBuf, '<')
				s.advance(1)
				return
			}
			if s.text[j] == ':' && j+1 < len(s.text) && !isJSXIdentStart(s.text[j+1]) && s.text[j+1] != '>' {
				// Colon followed by non-identifier (like port number).
				ip.textBuf = append(ip.textBuf, '<')
				s.advance(1)
				return
			}
			j++
		}
	}

	// Find the closing '>' of the tag.
	tagEnd := findInlineJSXTagEnd(s.text, start)
	if tagEnd < 0 {
		// No valid closing — treat '<' as literal.
		ip.textBuf = append(ip.textBuf, '<')
		s.advance(1)
		return
	}

	ip.flushText()

	// Tokenize the tag using the shared scanner.
	tagText := s.text[start : tagEnd+1]
	b := newBuilder()
	b.startNode()
	scanJSXTag(tagText, b)
	b.finishNode(syntax.JSXInline)
	tree := b.finish()

	// The builder produces a Document wrapping our JSXInline node.
	// Extract the JSXInline.
	if len(tree.Children) > 0 && tree.Children[0].Node != nil {
		ip.output = append(ip.output, tree.Children[0])
	}

	s.pos = tagEnd + 1
}

// findInlineJSXTagEnd finds the closing '>' of an inline JSX tag starting
// at position start. It respects string literals and expression attributes.
// Inline tags cannot span across newlines (except within expressions/strings).
// Returns the index of '>' or -1.
func findInlineJSXTagEnd(text string, start int) int {
	if start >= len(text) || text[start] != '<' {
		return -1
	}

	i := start + 1

	for i < len(text) {
		switch text[i] {
		case '>':
			return i
		case '\n', '\r':
			// Inline JSX tags generally don't span newlines,
			// but multi-line tags can exist. We allow it for
			// compatibility with the block-level behavior.
			i++
		case '"', '\'':
			// Skip string literal.
			quote := text[i]
			i++
			for i < len(text) && text[i] != quote {
				if text[i] == '\\' && i+1 < len(text) {
					i++
				}
				i++
			}
			if i < len(text) {
				i++ // closing quote
			}
		case '{':
			// Skip expression attribute.
			if end, ok := findMatchingBrace(text, i); ok {
				i = end + 1
			} else {
				return -1 // unclosed expression — not a valid tag
			}
		default:
			i++
		}
	}

	return -1
}
