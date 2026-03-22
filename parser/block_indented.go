package parser

import "github.com/albertocavalcante/mdxgo/syntax"

// parseIndentedCode parses an indented code block (4+ spaces of indentation).
// Consecutive indented lines form the block. Blank lines between them
// are included. The block ends at the first non-blank, non-indented line.
func (p *blockParser) parseIndentedCode(_ lineInfo) {
	p.builder.startNode()

	// Emit the first line.
	p.builder.token(syntax.IndentedCodeToken, p.currentLine().Content)
	p.advance()

	// Continue consuming lines that are indented 4+ or blank.
	for !p.eof() {
		li := analyzeLine(p.currentLine().Content)

		switch {
		case li.blank:
			// Blank lines are tentatively included — they'll be part of
			// the code block only if followed by another indented line.
			// For simplicity and round-trip correctness, we buffer them.
			blankStart := p.pos
			for !p.eof() {
				bli := analyzeLine(p.currentLine().Content)
				if !bli.blank {
					break
				}
				p.advance()
			}
			// Check if the next non-blank line is also indented.
			if !p.eof() {
				nextLI := analyzeLine(p.currentLine().Content)
				if nextLI.indent >= indentedCodeIndent {
					// Include the blank lines in the code block.
					for i := blankStart; i < p.pos; i++ {
						p.builder.token(syntax.IndentedCodeToken, p.lines[i].Content)
					}
					continue
				}
			}
			// Blank lines are NOT part of the code block — back up.
			p.pos = blankStart
			return

		case li.indent >= indentedCodeIndent:
			p.builder.token(syntax.IndentedCodeToken, p.currentLine().Content)
			p.advance()

		default:
			return
		}
	}

	p.builder.finishNode(syntax.IndentedCodeBlock)
}
