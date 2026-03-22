package parser

import (
	"strings"

	"github.com/albertocavalcante/mdxgo/syntax"
)

// isESMStart checks if a line starts an ESM declaration (import/export).
func isESMStart(li lineInfo) bool {
	if li.indent > 0 || li.blank {
		return false
	}
	return strings.HasPrefix(li.trimmed, "import ") ||
		strings.HasPrefix(li.trimmed, "import\t") ||
		li.trimmed == "import" ||
		strings.HasPrefix(li.trimmed, "export ") ||
		strings.HasPrefix(li.trimmed, "export\t") ||
		li.trimmed == "export"
}

// parseESM parses an ESM (import/export) block.
// ESM continues until a blank line or a line that doesn't look like
// continuation (doesn't start with import/export and isn't indented).
func (p *blockParser) parseESM(_ lineInfo) {
	p.builder.startNode()

	p.builder.token(syntax.ESMLineToken, p.currentLine().Content)
	p.advance()

	// Continue with indented or import/export lines.
	for !p.eof() {
		li := analyzeLine(p.currentLine().Content)
		if li.blank {
			break
		}
		// ESM continuation: line starts with import/export or is indented
		// (for multi-line imports).
		if isESMStart(li) || li.indent > 0 {
			p.builder.token(syntax.ESMLineToken, p.currentLine().Content)
			p.advance()
		} else {
			break
		}
	}

	p.builder.finishNode(syntax.ESMDeclaration)
}
