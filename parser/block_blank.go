package parser

import "github.com/albertocavalcante/mdxgo/syntax"

// parseBlankLine handles a blank line. Blank lines are represented as
// BlankLineNode containing a single BlankLineToken.
func (p *blockParser) parseBlankLine(li lineInfo) {
	p.builder.startNode()
	p.builder.token(syntax.BlankLineToken, li.raw)
	p.builder.finishNode(syntax.BlankLineNode)
	p.advance()
}
