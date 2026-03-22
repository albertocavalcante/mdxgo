package parser

import (
	"strings"

	"github.com/albertocavalcante/mdxgo/syntax"
)

// tryAutolinkOrRawHTML attempts to parse an autolink or raw HTML tag
// at the current position (starting with '<').
//
// Per CommonMark spec:
//   - Autolinks (§6.7): <scheme:...> or <email@domain>
//   - Raw HTML (§6.8): <tag>, </tag>, <!-- comment -->, <? PI ?>, <![CDATA[...]]>, <!DECL>
//
// We try autolink first, then raw HTML, then treat '<' as literal text.
func (ip *inlineProcessor) tryAutolinkOrRawHTML() {
	s := ip.scanner
	start := s.pos

	// Find the closing >.
	closeIdx := -1
	for i := start + 1; i < len(s.text); i++ {
		if s.text[i] == '>' {
			closeIdx = i
			break
		}
		// Autolinks and most raw HTML cannot span newlines.
		if s.text[i] == '\n' || s.text[i] == '\r' {
			break
		}
	}

	if closeIdx < 0 {
		// No closing >: literal <.
		ip.textBuf = append(ip.textBuf, '<')
		s.advance(1)
		return
	}

	content := s.text[start+1 : closeIdx]
	fullText := s.text[start : closeIdx+1]

	// Try autolink.
	if isAutolink(content) {
		ip.flushText()
		ip.output = append(ip.output, syntax.NodeElement(
			syntax.NewGreenNode(syntax.AutolinkSpan, []syntax.GreenElement{
				syntax.TokenElement(syntax.NewGreenToken(syntax.AutolinkToken, fullText)),
			}),
		))
		s.pos = closeIdx + 1
		return
	}

	// Try email autolink.
	if isEmailAutolink(content) {
		ip.flushText()
		ip.output = append(ip.output, syntax.NodeElement(
			syntax.NewGreenNode(syntax.AutolinkSpan, []syntax.GreenElement{
				syntax.TokenElement(syntax.NewGreenToken(syntax.AutolinkToken, fullText)),
			}),
		))
		s.pos = closeIdx + 1
		return
	}

	// Try raw HTML (open tag, closing tag, comment, PI, declaration, CDATA).
	if isInlineRawHTML(content) {
		ip.flushText()
		ip.output = append(ip.output, syntax.NodeElement(
			syntax.NewGreenNode(syntax.RawHTMLSpan, []syntax.GreenElement{
				syntax.TokenElement(syntax.NewGreenToken(syntax.RawHTMLToken, fullText)),
			}),
		))
		s.pos = closeIdx + 1
		return
	}

	// For HTML comments and CDATA, we need to handle multi-character closing
	// sequences that may contain >.
	if start+4 <= len(s.text) && s.text[start:start+4] == "<!--" {
		endIdx := ip.findCommentEnd(start + 4)
		if endIdx >= 0 {
			fullText := s.text[start : endIdx+3]
			ip.flushText()
			ip.output = append(ip.output, syntax.NodeElement(
				syntax.NewGreenNode(syntax.RawHTMLSpan, []syntax.GreenElement{
					syntax.TokenElement(syntax.NewGreenToken(syntax.RawHTMLToken, fullText)),
				}),
			))
			s.pos = endIdx + 3
			return
		}
	}

	if start+9 <= len(s.text) && s.text[start:start+9] == "<![CDATA[" {
		endIdx := ip.findCDATAEnd(start + 9)
		if endIdx >= 0 {
			fullText := s.text[start : endIdx+3]
			ip.flushText()
			ip.output = append(ip.output, syntax.NodeElement(
				syntax.NewGreenNode(syntax.RawHTMLSpan, []syntax.GreenElement{
					syntax.TokenElement(syntax.NewGreenToken(syntax.RawHTMLToken, fullText)),
				}),
			))
			s.pos = endIdx + 3
			return
		}
	}

	if start+2 <= len(s.text) && s.text[start:start+2] == "<?" {
		endIdx := ip.findPIEnd(start + 2)
		if endIdx >= 0 {
			fullText := s.text[start : endIdx+2]
			ip.flushText()
			ip.output = append(ip.output, syntax.NodeElement(
				syntax.NewGreenNode(syntax.RawHTMLSpan, []syntax.GreenElement{
					syntax.TokenElement(syntax.NewGreenToken(syntax.RawHTMLToken, fullText)),
				}),
			))
			s.pos = endIdx + 2
			return
		}
	}

	// Not a recognized construct: literal <.
	ip.textBuf = append(ip.textBuf, '<')
	s.advance(1)
}

// findCommentEnd finds the end of an HTML comment (-->), starting search from pos.
// Returns the index of the '-' in '-->' or -1.
func (ip *inlineProcessor) findCommentEnd(from int) int {
	s := ip.scanner
	for i := from; i+2 < len(s.text); i++ {
		if s.text[i] == '-' && s.text[i+1] == '-' && s.text[i+2] == '>' {
			return i
		}
	}
	return -1
}

// findCDATAEnd finds the end of a CDATA section (]]>), starting search from pos.
func (ip *inlineProcessor) findCDATAEnd(from int) int {
	s := ip.scanner
	for i := from; i+2 < len(s.text); i++ {
		if s.text[i] == ']' && s.text[i+1] == ']' && s.text[i+2] == '>' {
			return i
		}
	}
	return -1
}

// findPIEnd finds the end of a processing instruction (?>), starting search from pos.
func (ip *inlineProcessor) findPIEnd(from int) int {
	s := ip.scanner
	for i := from; i+1 < len(s.text); i++ {
		if s.text[i] == '?' && s.text[i+1] == '>' {
			return i
		}
	}
	return -1
}

// isAutolink checks if content (without < and >) is a URI autolink.
// Per CommonMark: scheme ":" followed by any characters except space, <, >.
// Scheme is 2-32 ASCII letters followed by optional letters/digits/+/-./.
func isAutolink(content string) bool {
	if len(content) == 0 {
		return false
	}
	// Find the colon.
	colonIdx := strings.IndexByte(content, ':')
	if colonIdx < 2 || colonIdx > 32 {
		return false
	}

	scheme := content[:colonIdx]
	// First char must be ASCII letter.
	if !isASCIILetter(scheme[0]) {
		return false
	}
	for i := 1; i < len(scheme); i++ {
		b := scheme[i]
		if !isASCIILetter(b) && !(b >= '0' && b <= '9') && b != '+' && b != '-' && b != '.' {
			return false
		}
	}

	// Rest (after colon) must not contain spaces, <, >.
	rest := content[colonIdx+1:]
	for i := 0; i < len(rest); i++ {
		if rest[i] == ' ' || rest[i] == '<' || rest[i] == '>' {
			return false
		}
	}

	return true
}

// isEmailAutolink checks if content (without < and >) is an email autolink.
func isEmailAutolink(content string) bool {
	if len(content) == 0 {
		return false
	}

	atIdx := strings.IndexByte(content, '@')
	if atIdx < 1 || atIdx >= len(content)-1 {
		return false
	}

	// Local part: atext characters.
	local := content[:atIdx]
	for i := 0; i < len(local); i++ {
		if !isAtextChar(local[i]) {
			if local[i] == '.' && i > 0 && i < len(local)-1 {
				continue
			}
			return false
		}
	}

	// Domain: labels separated by dots.
	domain := content[atIdx+1:]
	if len(domain) == 0 || domain[0] == '-' {
		return false
	}
	labels := strings.Split(domain, ".")
	if len(labels) < 2 {
		return false
	}
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return false
		}
		if label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for i := 0; i < len(label); i++ {
			b := label[i]
			if !isASCIILetter(b) && !(b >= '0' && b <= '9') && b != '-' {
				return false
			}
		}
	}

	return true
}

func isAtextChar(b byte) bool {
	if isASCIILetter(b) || (b >= '0' && b <= '9') {
		return true
	}
	switch b {
	case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '/', '=', '?', '^', '_', '`', '{', '|', '}', '~':
		return true
	}
	return false
}

// isInlineRawHTML checks if content (without < and >) looks like inline raw HTML.
func isInlineRawHTML(content string) bool {
	if len(content) == 0 {
		return false
	}

	// Closing tag: /tagname with optional whitespace
	if content[0] == '/' {
		return isClosingTag(content[1:])
	}

	// Declaration: !LETTER...
	if content[0] == '!' && len(content) > 1 && isASCIILetter(content[1]) {
		return true
	}

	// Open tag: tagname [attrs] [/]
	return isOpenTag(content)
}

// isOpenTag checks if content matches an HTML open or self-closing tag.
func isOpenTag(content string) bool {
	if len(content) == 0 || !isASCIILetter(content[0]) {
		return false
	}

	// Read tag name.
	i := 1
	for i < len(content) && isTagNameChar(content[i]) {
		i++
	}

	// Skip attributes.
	for i < len(content) {
		// Skip whitespace.
		if content[i] != ' ' && content[i] != '\t' && content[i] != '\n' && content[i] != '\r' {
			break
		}
		for i < len(content) && (content[i] == ' ' || content[i] == '\t' || content[i] == '\n' || content[i] == '\r') {
			i++
		}
		if i >= len(content) {
			break
		}
		// Check for self-closing /.
		if content[i] == '/' && i == len(content)-1 {
			return true
		}
		if content[i] == '/' {
			break
		}
		// Try to read an attribute.
		attrEnd := tryReadAttribute(content, i)
		if attrEnd == i {
			break
		}
		i = attrEnd
	}

	if i == len(content) {
		return true
	}
	if i == len(content)-1 && content[i] == '/' {
		return true
	}

	return false
}

// tryReadAttribute tries to read an HTML attribute from content starting at pos.
// Returns the position after the attribute, or pos if no attribute found.
func tryReadAttribute(content string, pos int) int {
	i := pos
	// Attribute name: [a-zA-Z_:][a-zA-Z0-9_.:-]*
	if i >= len(content) {
		return pos
	}
	if !isASCIILetter(content[i]) && content[i] != '_' && content[i] != ':' {
		return pos
	}
	i++
	for i < len(content) && isAttrNameChar(content[i]) {
		i++
	}

	// Optional value specification.
	j := i
	// Skip whitespace.
	for j < len(content) && (content[j] == ' ' || content[j] == '\t' || content[j] == '\n' || content[j] == '\r') {
		j++
	}
	if j < len(content) && content[j] == '=' {
		j++
		// Skip whitespace.
		for j < len(content) && (content[j] == ' ' || content[j] == '\t' || content[j] == '\n' || content[j] == '\r') {
			j++
		}
		if j < len(content) {
			if content[j] == '\'' {
				// Single-quoted value.
				end := strings.IndexByte(content[j+1:], '\'')
				if end >= 0 {
					return j + 1 + end + 1
				}
				return pos
			}
			if content[j] == '"' {
				// Double-quoted value.
				end := strings.IndexByte(content[j+1:], '"')
				if end >= 0 {
					return j + 1 + end + 1
				}
				return pos
			}
			// Unquoted value: no spaces, quotes, =, <, >, `.
			k := j
			for k < len(content) && !isUnquotedAttrValEnd(content[k]) {
				k++
			}
			if k > j {
				return k
			}
			return pos
		}
		return pos
	}

	return i
}

func isClosingTag(content string) bool {
	if len(content) == 0 || !isASCIILetter(content[0]) {
		return false
	}
	i := 1
	for i < len(content) && isTagNameChar(content[i]) {
		i++
	}
	// Optional trailing whitespace.
	for i < len(content) && (content[i] == ' ' || content[i] == '\t') {
		i++
	}
	return i == len(content)
}

func isTagNameChar(b byte) bool {
	return isASCIILetter(b) || (b >= '0' && b <= '9') || b == '-'
}

func isAttrNameChar(b byte) bool {
	return isASCIILetter(b) || (b >= '0' && b <= '9') || b == '_' || b == '.' || b == ':' || b == '-'
}

func isUnquotedAttrValEnd(b byte) bool {
	switch b {
	case ' ', '\t', '\n', '\r', '"', '\'', '=', '<', '>', '`':
		return true
	}
	return false
}
