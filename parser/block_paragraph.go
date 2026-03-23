package parser

import "github.com/albertocavalcante/mdxgo/syntax"

// parseParagraph parses a paragraph. A paragraph is a sequence of non-blank
// lines that cannot be interpreted as other block-level constructs.
// A setext heading underline on the second line converts this to a SetextHeading.
func (p *blockParser) parseParagraph(_ lineInfo) {
	p.builder.startNode()

	// First line.
	p.builder.token(syntax.TextToken, p.currentLine().Content)
	p.advance()

	// Collect continuation lines.
	for !p.eof() {
		li := analyzeLine(p.currentLine().Content)

		if li.blank {
			break
		}

		// Check for setext underline on this line.
		if _, ok := isSetextUnderline(li); ok {
			// Convert to setext heading.
			p.builder.token(syntax.SetextUnderline, p.currentLine().Content)
			p.advance()
			p.builder.finishNode(syntax.SetextHeading)
			return
		}

		// Check if this line starts a new block.
		if tryATXHeading(li) || isThematicBreak(li) || isFenceOpen(li) || isBlockQuoteStart(li) {
			break
		}
		if _, ok := isListItemStart(li); ok {
			break
		}
		if !p.opts.MDX && isHTMLBlockStart(li) {
			break
		}
		if !p.opts.MDX && li.indent >= indentedCodeIndent {
			break
		}
		if p.opts.MDX && isJSXBlockStart(li) {
			break
		}
		if p.opts.MDX && isExprBlockStart(li) {
			break
		}

		// Continuation line.
		p.builder.token(syntax.TextToken, p.currentLine().Content)
		p.advance()
	}

	p.builder.finishNode(syntax.Paragraph)
}
