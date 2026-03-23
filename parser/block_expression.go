package parser

import "github.com/albertocavalcante/mdxgo/syntax"

// isExprBlockStart checks if a line starts an expression block.
// An expression block starts with '{' (with at most 3 spaces indent) in MDX mode.
func isExprBlockStart(li lineInfo) bool {
	if li.blank || li.indent > maxIndent {
		return false
	}
	return len(li.trimmed) > 0 && li.trimmed[0] == '{'
}

// parseExpressionBlock parses a brace-balanced expression block.
// Structure: ExpressionBlock { ExprOpenBrace, ExprContentToken..., ExprCloseBrace }
//
// The parser scans forward counting braces until depth returns to zero.
// If EOF is reached without balance, the block is emitted as an ErrorNode.
func (p *blockParser) parseExpressionBlock(_ lineInfo) {
	p.builder.startNode()

	// Collect all lines that form this expression block.
	// We need to find the matching closing brace.
	firstLine := p.currentLine().Content
	p.advance()

	// Find the position of '{' in the first line.
	braceOffset := -1
	for i := 0; i < len(firstLine); i++ {
		if firstLine[i] == '{' {
			braceOffset = i
			break
		}
	}
	if braceOffset < 0 {
		// Shouldn't happen since isExprBlockStart checked, but be safe.
		p.builder.token(syntax.ErrorToken, firstLine)
		p.builder.finishNode(syntax.ErrorNode)
		return
	}

	// Concatenate lines and look for matching brace.
	allText := firstLine
	for !p.eof() {
		if end, ok := findMatchingBrace(allText, braceOffset); ok {
			// Found the match. The expression ends at position end (inclusive).
			// Everything up to and including the closing brace, plus any trailing
			// content on that line, is part of this expression block.
			_ = end
			p.emitExpressionTokens(allText, braceOffset)
			p.builder.finishNode(syntax.ExpressionBlock)
			return
		}
		// Need more lines.
		allText += p.currentLine().Content
		p.advance()
	}

	// Check one more time with all accumulated text.
	if end, ok := findMatchingBrace(allText, braceOffset); ok {
		_ = end
		p.emitExpressionTokens(allText, braceOffset)
		p.builder.finishNode(syntax.ExpressionBlock)
		return
	}

	// Unclosed expression — emit as ErrorNode.
	p.emitExpressionTokens(allText, braceOffset)
	p.builder.finishNode(syntax.ErrorNode)
}

// emitExpressionTokens emits the tokenized form of an expression block.
// Leading whitespace before '{' becomes IndentToken.
// The '{' becomes ExprOpenBrace.
// Content between braces becomes ExprContentToken (one per line or segment).
// The '}' becomes ExprCloseBrace.
// Any trailing content after '}' on the same line is included in the last ExprContentToken.
func (p *blockParser) emitExpressionTokens(text string, braceOffset int) {
	// Leading indent before the brace.
	if braceOffset > 0 {
		p.builder.token(syntax.IndentToken, text[:braceOffset])
	}

	// Find matching brace to split tokens properly.
	end, ok := findMatchingBrace(text, braceOffset)
	if !ok {
		// Unclosed: emit opening brace + all remaining as content.
		p.builder.token(syntax.ExprOpenBrace, "{")
		remaining := text[braceOffset+1:]
		if len(remaining) > 0 {
			p.builder.token(syntax.ExprContentToken, remaining)
		}
		return
	}

	// Emit opening brace.
	p.builder.token(syntax.ExprOpenBrace, "{")

	// Emit content between braces.
	content := text[braceOffset+1 : end]
	if len(content) > 0 {
		p.builder.token(syntax.ExprContentToken, content)
	}

	// Emit closing brace.
	p.builder.token(syntax.ExprCloseBrace, "}")

	// Emit any trailing content after the closing brace (whitespace, newline).
	trailing := text[end+1:]
	if len(trailing) > 0 {
		p.builder.token(syntax.TextToken, trailing)
	}
}
