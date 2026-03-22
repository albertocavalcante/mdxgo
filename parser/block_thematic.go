package parser

import "github.com/albertocavalcante/mdxgo/syntax"

// isThematicBreak checks if a line is a thematic break: three or more
// -, *, or _ characters with optional spaces between them.
func isThematicBreak(li lineInfo) bool {
	if li.blank || li.indent > maxIndent {
		return false
	}
	s := li.trimmed
	if s == "" {
		return false
	}
	ch := s[0]
	if ch != '-' && ch != '*' && ch != '_' {
		return false
	}
	count := 0
	for _, c := range s {
		if byte(c) == ch {
			count++
		} else if c != ' ' && c != '\t' {
			return false
		}
	}
	return count >= minThematicBreakChars
}

// parseThematicBreak parses a thematic break line.
func (p *blockParser) parseThematicBreak(li lineInfo) {
	p.builder.startNode()

	if li.indent > 0 {
		p.builder.token(syntax.IndentToken, li.content[:li.indent])
	}

	p.builder.token(syntax.ThematicBreakToken, li.trimmed)

	if li.newline != "" {
		p.builder.token(syntax.NewLineToken, li.newline)
	}

	p.builder.finishNode(syntax.ThematicBreak)
	p.advance()
}
