package parser

import "github.com/albertocavalcante/mdxgo/syntax"

// isBlockQuoteStart checks if a line starts a blockquote (optional indent + >).
func isBlockQuoteStart(li lineInfo) bool {
	if li.indent > maxIndent {
		return false
	}
	return li.trimmed != "" && li.trimmed[0] == '>'
}

// parseBlockQuote parses a block quote. Each line's > marker is emitted as a
// BlockQuoteMarker token, the optional space after it as trailing trivia, and
// the rest of the line as a TextToken. Inner block structure is not recursively
// parsed in this phase — the content is flat tokens that preserve round-trip.
func (p *blockParser) parseBlockQuote(_ lineInfo) {
	p.builder.startNode()

	for !p.eof() {
		li := analyzeLine(p.currentLine().Content)

		if !isBlockQuoteStart(li) {
			break
		}

		raw := li.content
		pos := 0

		// Indent before >.
		if li.indent > 0 && li.indent <= maxIndent {
			p.builder.token(syntax.IndentToken, raw[:li.indent])
			pos = li.indent
		}

		// > marker.
		pos++ // skip '>'
		markerText := ">"

		// Optional space after >.
		spaceAfter := ""
		if pos < len(raw) && raw[pos] == ' ' {
			spaceAfter = " "
			pos++
		}

		if spaceAfter != "" {
			p.builder.tokenTrivia(syntax.BlockQuoteMarker,
				syntax.TriviaList{},
				markerText,
				syntax.NewTriviaList(syntax.Trivia{Kind: syntax.WhitespaceTrivia, Text: spaceAfter}),
			)
		} else {
			p.builder.token(syntax.BlockQuoteMarker, markerText)
		}

		// Rest of line content.
		if pos < len(raw) {
			p.builder.token(syntax.TextToken, raw[pos:])
		}

		// Line ending.
		if li.newline != "" {
			p.builder.token(syntax.NewLineToken, li.newline)
		}

		p.advance()
	}

	p.builder.finishNode(syntax.BlockQuote)
}
