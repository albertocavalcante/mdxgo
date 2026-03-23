package parser

import "github.com/albertocavalcante/mdxgo/syntax"

// isJSXBlockStart checks if a line starts a JSX block element.
// A JSX block starts with '<' followed by '/', '>', or a JSX identifier
// start character (uppercase letter or lowercase for HTML elements).
// Must NOT match HTML comments (<!--), processing instructions (<?),
// CDATA (<![CDATA[), or declarations (<!LETTER for non-JSX).
// Must NOT match URI autolinks like <https://...> or <mailto:...>.
func isJSXBlockStart(li lineInfo) bool {
	if li.blank || li.indent > maxIndent {
		return false
	}
	s := li.trimmed
	if len(s) < 2 || s[0] != '<' {
		return false
	}

	next := s[1]

	// Fragment: <>
	if next == '>' {
		return true
	}

	// Closing tag: </
	if next == '/' {
		// Must be followed by identifier start or >
		if len(s) > 2 && (isJSXIdentStart(s[2]) || s[2] == '>') {
			return true
		}
		return false
	}

	// Exclude HTML comments: <!--
	if next == '!' {
		return false
	}

	// Exclude processing instructions: <?
	if next == '?' {
		return false
	}

	// Opening tag: <Identifier or <identifier
	if !isJSXIdentStart(next) {
		return false
	}

	// Read the full identifier to check what follows.
	// Reject if identifier is followed by "://" (URL scheme).
	i := 2
	for i < len(s) && (isJSXIdentPart(s[i]) || s[i] == ':' || s[i] == '.') {
		// If we see "://" it's a URL, not JSX.
		if s[i] == ':' && i+2 < len(s) && s[i+1] == '/' && s[i+2] == '/' {
			return false
		}
		// If we see ":" followed by a digit or non-identifier (like port number),
		// and the identifier so far looks like a hostname, reject.
		if s[i] == ':' && i+1 < len(s) && !isJSXIdentStart(s[i+1]) && s[i+1] != '>' {
			return false
		}
		i++
	}

	return true
}

// isJSXIdentStart reports whether b can start a JSX identifier.
// JSX identifiers can start with ASCII letters, '_', or '$'.
func isJSXIdentStart(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_' || b == '$'
}

// isJSXIdentPart reports whether b can continue a JSX identifier.
func isJSXIdentPart(b byte) bool {
	return isJSXIdentStart(b) || (b >= '0' && b <= '9') || b == '-'
}

// parseJSXBlock parses a JSX block element.
// A JSX block is a single tag construct (opening, closing, self-closing, or fragment)
// that may span multiple lines. The content between opening and closing tags is
// NOT contained within the JSX block — it's normal block content.
//
// Structure: JSXBlock { JSXOpeningTag | JSXClosingTag | JSXSelfClosingTag | JSXFragment }
func (p *blockParser) parseJSXBlock(_ lineInfo) {
	p.builder.startNode()

	// Collect lines until we find the closing '>' of the tag.
	firstLine := p.currentLine().Content
	p.advance()

	allText := firstLine

	// Check if the tag closes on the first line.
	if tagEnd := findJSXTagEnd(allText); tagEnd >= 0 {
		p.emitJSXBlockTokens(allText, tagEnd)
		p.builder.finishNode(syntax.JSXBlock)
		return
	}

	// Multi-line tag: keep adding lines until we find '>'.
	for !p.eof() {
		allText += p.currentLine().Content
		p.advance()
		if tagEnd := findJSXTagEnd(allText); tagEnd >= 0 {
			p.emitJSXBlockTokens(allText, tagEnd)
			p.builder.finishNode(syntax.JSXBlock)
			return
		}
	}

	// No closing '>' found — emit as ErrorNode.
	p.emitJSXBlockAsError(allText)
	p.builder.finishNode(syntax.ErrorNode)
}

// findJSXTagEnd finds the position of the closing '>' of a JSX tag,
// respecting string literals and expression attributes.
// Returns the index of '>' or -1.
func findJSXTagEnd(text string) int {
	if len(text) == 0 || text[0] != '<' {
		// Account for leading whitespace.
		i := 0
		for i < len(text) && (text[i] == ' ' || text[i] == '\t') {
			i++
		}
		if i >= len(text) || text[i] != '<' {
			return -1
		}
	}

	i := 0
	// Skip leading whitespace.
	for i < len(text) && (text[i] == ' ' || text[i] == '\t') {
		i++
	}
	if i >= len(text) || text[i] != '<' {
		return -1
	}
	i++ // skip '<'

	for i < len(text) {
		switch text[i] {
		case '>':
			return i
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
				i++ // skip closing quote
			}
		case '{':
			// Skip expression attribute.
			if end, ok := findMatchingBrace(text, i); ok {
				i = end + 1
			} else {
				return -1 // unclosed expression
			}
		default:
			i++
		}
	}

	return -1
}

// emitJSXBlockTokens tokenizes a complete JSX tag and emits tokens.
func (p *blockParser) emitJSXBlockTokens(text string, tagEnd int) {
	// Find leading whitespace (indent).
	indent := 0
	for indent < len(text) && (text[indent] == ' ' || text[indent] == '\t') {
		indent++
	}

	if indent > 0 {
		p.builder.token(syntax.IndentToken, text[:indent])
	}

	// Use the JSX scanner to tokenize the tag.
	tagText := text[indent : tagEnd+1]
	scanJSXTag(tagText, p.builder)

	// Emit any trailing content after the tag (whitespace, newline).
	trailing := text[tagEnd+1:]
	if len(trailing) > 0 {
		p.builder.token(syntax.JSXTextToken, trailing)
	}
}

// emitJSXBlockAsError emits the entire text as an error token.
func (p *blockParser) emitJSXBlockAsError(text string) {
	p.builder.token(syntax.ErrorToken, text)
}

// scanJSXTag tokenizes a JSX tag string (from '<' to '>') and emits tokens
// to the builder. Determines the tag type (opening, closing, self-closing, fragment)
// and wraps the tokens in the appropriate sub-node.
func scanJSXTag(tag string, b *builder) {
	if len(tag) < 2 {
		b.token(syntax.JSXTextToken, tag)
		return
	}

	b.startNode()

	i := 0

	// Opening '<'
	b.token(syntax.JSXOpenAngle, "<")
	i++

	// Determine tag type.
	isClosing := false

	if i < len(tag) && tag[i] == '/' {
		isClosing = true
		b.token(syntax.JSXSlash, "/")
		i++
	}

	if i < len(tag) && tag[i] == '>' {
		// Fragment: <> or </>
		b.token(syntax.JSXCloseAngle, ">")
		b.finishNode(syntax.JSXFragment)
		return
	}

	// Tag name: may include dots (member expression) or colons (namespace).
	nameStart := i
	if i < len(tag) && isJSXIdentStart(tag[i]) {
		// Read first identifier segment.
		segStart := i
		for i < len(tag) && isJSXIdentPart(tag[i]) {
			i++
		}
		b.token(syntax.JSXIdentifier, tag[segStart:i])

		// Handle member expressions (Foo.Bar.Baz) and namespaces (xml:tag).
		for i < len(tag) {
			if tag[i] == '.' {
				b.token(syntax.JSXDot, ".")
				i++
				segStart = i
				for i < len(tag) && isJSXIdentPart(tag[i]) {
					i++
				}
				if i > segStart {
					b.token(syntax.JSXIdentifier, tag[segStart:i])
				}
			} else if tag[i] == ':' {
				// Namespace — emit colon as part of identifier.
				b.token(syntax.JSXDot, ":")
				i++
				segStart = i
				for i < len(tag) && isJSXIdentPart(tag[i]) {
					i++
				}
				if i > segStart {
					b.token(syntax.JSXIdentifier, tag[segStart:i])
				}
			} else {
				break
			}
		}
	}

	if i == nameStart && !isClosing {
		// No valid tag name found — emit rest as text.
		b.token(syntax.JSXTextToken, tag[i:])
		b.finishNode(syntax.JSXOpeningTag)
		return
	}

	// Skip whitespace.
	wsStart := i
	for i < len(tag) && (tag[i] == ' ' || tag[i] == '\t' || tag[i] == '\n' || tag[i] == '\r') {
		i++
	}
	if i > wsStart {
		b.token(syntax.JSXTextToken, tag[wsStart:i])
	}

	if isClosing {
		// Closing tag: </Name> — may have unexpected content before >.
		// Emit any remaining content before '>' as JSXTextToken.
		if i < len(tag) && tag[i] != '>' {
			extraStart := i
			for i < len(tag) && tag[i] != '>' {
				i++
			}
			b.token(syntax.JSXTextToken, tag[extraStart:i])
		}
		if i < len(tag) && tag[i] == '>' {
			b.token(syntax.JSXCloseAngle, ">")
			i++
		}
		// Emit any remaining text after '>'.
		if i < len(tag) {
			b.token(syntax.JSXTextToken, tag[i:])
		}
		b.finishNode(syntax.JSXClosingTag)
		return
	}

	// Parse attributes.
	for i < len(tag) && tag[i] != '>' && tag[i] != '/' {
		attrStart := i

		if tag[i] == '{' {
			// Expression attribute: {...spread} or {expr}
			b.startNode()
			if end, ok := findMatchingBrace(tag, i); ok {
				b.token(syntax.ExprOpenBrace, "{")
				content := tag[i+1 : end]
				if len(content) > 0 {
					b.token(syntax.ExprContentToken, content)
				}
				b.token(syntax.ExprCloseBrace, "}")
				i = end + 1
			} else {
				// Unclosed — emit as text.
				b.token(syntax.JSXTextToken, tag[i:])
				i = len(tag)
				b.finishNode(syntax.JSXExprAttribute)
				break
			}
			b.finishNode(syntax.JSXExprAttribute)
		} else if isJSXIdentStart(tag[i]) || tag[i] == '_' {
			// Named attribute.
			b.startNode()

			// Attribute name.
			nameS := i
			for i < len(tag) && (isJSXIdentPart(tag[i]) || tag[i] == ':') {
				i++
			}
			b.token(syntax.JSXIdentifier, tag[nameS:i])

			// Skip whitespace around '='.
			ws1Start := i
			for i < len(tag) && (tag[i] == ' ' || tag[i] == '\t' || tag[i] == '\n' || tag[i] == '\r') {
				i++
			}
			if i > ws1Start {
				b.token(syntax.JSXTextToken, tag[ws1Start:i])
			}

			if i < len(tag) && tag[i] == '=' {
				b.token(syntax.JSXEquals, "=")
				i++

				// Skip whitespace after '='.
				ws2Start := i
				for i < len(tag) && (tag[i] == ' ' || tag[i] == '\t' || tag[i] == '\n' || tag[i] == '\r') {
					i++
				}
				if i > ws2Start {
					b.token(syntax.JSXTextToken, tag[ws2Start:i])
				}

				// Attribute value.
				if i < len(tag) {
					if tag[i] == '"' || tag[i] == '\'' {
						// String literal value.
						quote := tag[i]
						valStart := i
						i++
						for i < len(tag) && tag[i] != quote {
							if tag[i] == '\\' && i+1 < len(tag) {
								i++
							}
							i++
						}
						if i < len(tag) {
							i++ // closing quote
						}
						b.token(syntax.JSXStringLiteral, tag[valStart:i])
					} else if tag[i] == '{' {
						// Expression value.
						if end, ok := findMatchingBrace(tag, i); ok {
							b.token(syntax.ExprOpenBrace, "{")
							content := tag[i+1 : end]
							if len(content) > 0 {
								b.token(syntax.ExprContentToken, content)
							}
							b.token(syntax.ExprCloseBrace, "}")
							i = end + 1
						} else {
							b.token(syntax.JSXTextToken, tag[i:])
							i = len(tag)
							b.finishNode(syntax.JSXAttribute)
							break
						}
					}
				}
				b.finishNode(syntax.JSXAttribute)
			} else {
				// Boolean attribute (no value).
				b.finishNode(syntax.JSXAttribute)
			}
		} else {
			// Skip whitespace or unexpected characters.
			i++
			if i > attrStart {
				b.token(syntax.JSXTextToken, tag[attrStart:i])
			}
		}

		// Skip inter-attribute whitespace.
		wsS := i
		for i < len(tag) && (tag[i] == ' ' || tag[i] == '\t' || tag[i] == '\n' || tag[i] == '\r') {
			i++
		}
		if i > wsS {
			b.token(syntax.JSXTextToken, tag[wsS:i])
		}

		if i == attrStart {
			// No progress — break to avoid infinite loop.
			break
		}
	}

	// Self-closing or closing angle.
	isSelfClosing := false
	if i < len(tag) && tag[i] == '/' {
		isSelfClosing = true
		b.token(syntax.JSXSlash, "/")
		i++
	}
	if i < len(tag) && tag[i] == '>' {
		b.token(syntax.JSXCloseAngle, ">")
		i++
	}

	// Emit any remaining unconsumed bytes.
	if i < len(tag) {
		b.token(syntax.JSXTextToken, tag[i:])
	}

	if isSelfClosing {
		b.finishNode(syntax.JSXSelfClosingTag)
	} else {
		b.finishNode(syntax.JSXOpeningTag)
	}
}
