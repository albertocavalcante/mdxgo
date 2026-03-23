package parser

import "github.com/albertocavalcante/mdxgo/syntax"

// tryExpressionInline attempts to parse an inline MDX expression at the
// current position (starting with '{').
//
// In MDX mode, '{' starts an expression. The parser scans forward for the
// matching '}' using brace counting. If no match is found, '{' is treated
// as literal text.
//
// Structure: ExpressionInline { ExprOpenBrace, ExprContentToken, ExprCloseBrace }
func (ip *inlineProcessor) tryExpressionInline() {
	s := ip.scanner
	start := s.pos

	// Use findMatchingBrace to locate the closing '}'.
	end, ok := findMatchingBrace(s.text, start)
	if !ok {
		// No matching brace — treat '{' as literal text.
		ip.textBuf = append(ip.textBuf, '{')
		s.advance(1)
		return
	}

	ip.flushText()

	// Build ExpressionInline node.
	children := []syntax.GreenElement{
		syntax.TokenElement(syntax.NewGreenToken(syntax.ExprOpenBrace, "{")),
	}

	content := s.text[start+1 : end]
	if len(content) > 0 {
		children = append(children,
			syntax.TokenElement(syntax.NewGreenToken(syntax.ExprContentToken, content)),
		)
	}

	children = append(children,
		syntax.TokenElement(syntax.NewGreenToken(syntax.ExprCloseBrace, "}")),
	)

	ip.output = append(ip.output, syntax.NodeElement(
		syntax.NewGreenNode(syntax.ExpressionInline, children),
	))

	s.pos = end + 1
}
