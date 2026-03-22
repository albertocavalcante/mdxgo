package parser

import (
	"github.com/albertocavalcante/mdxgo/syntax"
)

// listMarker describes a detected list item marker.
type listMarker struct {
	ordered bool   // true for ordered (1., 2), etc.)
	bullet  byte   // '-', '+', '*' for unordered
	text    string // the full marker text including trailing space
	marker  string // just the marker characters (e.g., "- " or "1. ")
	indent  int    // indentation of the marker
	width   int    // total width consumed (indent + marker + space after)
}

// isListItemStart checks if a line starts a list item.
// Returns the marker details and true if it's a list item start.
func isListItemStart(li lineInfo) (listMarker, bool) {
	if li.blank || li.indent > maxIndent {
		return listMarker{}, false
	}

	s := li.trimmed
	if s == "" {
		return listMarker{}, false
	}

	// Bullet list: -, +, *
	if (s[0] == '-' || s[0] == '+' || s[0] == '*') && len(s) > 1 && (s[1] == ' ' || s[1] == '\t') {
		// Make sure it's not a thematic break.
		if s[0] == '-' || s[0] == '*' {
			if isThematicBreak(li) {
				return listMarker{}, false
			}
		}
		return listMarker{
			ordered: false,
			bullet:  s[0],
			text:    string(s[0]) + string(s[1]),
			marker:  string(s[0]),
			indent:  li.indent,
			width:   li.indent + bulletMarkerWidth,
		}, true
	}

	// Ordered list: digits followed by . or )
	if s[0] >= '0' && s[0] <= '9' {
		i := 0
		for i < len(s) && i < 9 && s[i] >= '0' && s[i] <= '9' {
			i++
		}
		if i < len(s) && (s[i] == '.' || s[i] == ')') {
			delim := i + 1
			if delim < len(s) && (s[delim] == ' ' || s[delim] == '\t') {
				markerStr := s[:delim]
				return listMarker{
					ordered: true,
					text:    s[:delim+1],
					marker:  markerStr,
					indent:  li.indent,
					width:   li.indent + delim + 1,
				}, true
			}
			// Empty list item (marker at end of line).
			if delim == len(s) {
				return listMarker{
					ordered: true,
					text:    s[:delim],
					marker:  s[:delim],
					indent:  li.indent,
					width:   li.indent + delim,
				}, true
			}
		}
	}

	return listMarker{}, false
}

// parseList parses a list (bullet or ordered) consisting of consecutive
// list items with the same marker type.
func (p *blockParser) parseList(firstLI lineInfo, firstMarker listMarker) {
	listKind := syntax.BulletList
	if firstMarker.ordered {
		listKind = syntax.OrderedList
	}
	p.builder.startNode()

	p.parseListItem(firstLI, firstMarker)

	// Continue with more list items of the same type.
	for !p.eof() {
		li := analyzeLine(p.currentLine().Content)
		if li.blank {
			// Blank line might separate loose list items.
			// For now, end the list at blank lines.
			break
		}
		marker, ok := isListItemStart(li)
		if !ok || marker.ordered != firstMarker.ordered {
			break
		}
		if !firstMarker.ordered && marker.bullet != firstMarker.bullet {
			break
		}
		p.parseListItem(li, marker)
	}

	p.builder.finishNode(listKind)
}

// parseListItem parses a single list item.
func (p *blockParser) parseListItem(li lineInfo, marker listMarker) {
	p.builder.startNode()

	raw := li.content
	pos := 0

	// Leading indent.
	if li.indent > 0 {
		p.builder.token(syntax.IndentToken, raw[:li.indent])
		pos = li.indent
	}

	// Marker token.
	markerKind := syntax.BulletMarker
	if marker.ordered {
		markerKind = syntax.OrderedMarker
	}

	markerEnd := pos + len(marker.marker)
	markerText := raw[pos:markerEnd]
	pos = markerEnd

	// Space after marker — trailing trivia.
	spaceAfter := ""
	if pos < len(raw) && (raw[pos] == ' ' || raw[pos] == '\t') {
		spaceAfter = string(raw[pos])
		pos++
	}

	if spaceAfter != "" {
		p.builder.tokenTrivia(markerKind,
			syntax.TriviaList{},
			markerText,
			syntax.NewTriviaList(syntax.Trivia{Kind: syntax.WhitespaceTrivia, Text: spaceAfter}),
		)
	} else {
		p.builder.token(markerKind, markerText)
	}

	// Content after marker on this line.
	if pos < len(raw) {
		p.builder.token(syntax.TextToken, raw[pos:])
	}

	if li.newline != "" {
		p.builder.token(syntax.NewLineToken, li.newline)
	}
	p.advance()

	// Continuation lines: lines indented enough to be part of this item.
	contentIndent := marker.width
	if contentIndent == 0 {
		contentIndent = li.indent + bulletMarkerWidth
	}
	for !p.eof() {
		nextLI := analyzeLine(p.currentLine().Content)
		if nextLI.blank {
			break
		}
		// Check if this is a new list item or a new block.
		if _, ok := isListItemStart(nextLI); ok {
			break
		}
		if nextLI.indent < contentIndent {
			break
		}
		// Continuation line — emit as text.
		p.builder.token(syntax.TextToken, p.currentLine().Content)
		p.advance()
	}

	p.builder.finishNode(syntax.ListItem)
}
