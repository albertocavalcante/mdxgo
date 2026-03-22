package parser

import (
	"strings"

	"github.com/albertocavalcante/mdxgo/syntax"
)

// fenceInfo captures the opening fence details.
type fenceInfo struct {
	char   byte // '`' or '~'
	count  int  // number of fence chars
	indent int  // indentation of the fence
}

// isFenceOpen checks if a line starts a fenced code block.
func isFenceOpen(li lineInfo) bool {
	if li.indent > maxIndent {
		return false
	}
	s := li.trimmed
	if len(s) < minFenceLength {
		return false
	}
	ch := s[0]
	if ch != '`' && ch != '~' {
		return false
	}
	count := 0
	for count < len(s) && s[count] == ch {
		count++
	}
	if count < minFenceLength {
		return false
	}
	// Backtick fences cannot have backticks in the info string.
	if ch == '`' && strings.ContainsRune(s[count:], '`') {
		return false
	}
	return true
}

// parseFenceOpen extracts fence details from a line known to be a fence open.
func parseFenceOpen(li lineInfo) fenceInfo {
	s := li.trimmed
	ch := s[0]
	count := 0
	for count < len(s) && s[count] == ch {
		count++
	}
	return fenceInfo{char: ch, count: count, indent: li.indent}
}

// isFenceClose checks if a line is a closing fence matching the opener.
func isFenceClose(li lineInfo, fi fenceInfo) bool {
	if li.indent > maxIndent {
		return false
	}
	s := li.trimmed
	if s == "" {
		return false
	}
	if s[0] != fi.char {
		return false
	}
	count := 0
	for count < len(s) && s[count] == fi.char {
		count++
	}
	if count < fi.count {
		return false
	}
	// Only whitespace may follow the closing fence.
	rest := strings.TrimRight(s[count:], " \t")
	return rest == ""
}

// parseFencedCode parses a fenced code block: opening fence, content lines,
// optional closing fence.
func (p *blockParser) parseFencedCode(openLI lineInfo) {
	fi := parseFenceOpen(openLI)
	p.builder.startNode()

	// Opening fence line.
	raw := openLI.content
	pos := 0

	if openLI.indent > 0 {
		p.builder.token(syntax.IndentToken, raw[:openLI.indent])
		pos = openLI.indent
	}

	// Fence characters.
	fenceEnd := pos + fi.count
	fenceText := raw[pos:fenceEnd]
	pos = fenceEnd

	// Info string (rest of line after fence).
	if pos < len(raw) {
		infoText := raw[pos:]
		// Trailing trivia on the fence token: nothing. Info is a separate token.
		p.builder.token(syntax.FenceOpenToken, fenceText)
		p.builder.token(syntax.InfoStringToken, infoText)
	} else {
		p.builder.token(syntax.FenceOpenToken, fenceText)
	}

	if openLI.newline != "" {
		p.builder.token(syntax.NewLineToken, openLI.newline)
	}
	p.advance()

	// Content lines until closing fence or EOF.
	for !p.eof() {
		cl := p.currentLine()
		cli := analyzeLine(cl.Content)

		if isFenceClose(cli, fi) {
			// Closing fence line.
			closeRaw := cli.content
			cpos := 0
			if cli.indent > 0 {
				p.builder.token(syntax.IndentToken, closeRaw[:cli.indent])
				cpos = cli.indent
			}
			p.builder.token(syntax.FenceCloseToken, closeRaw[cpos:])
			if cli.newline != "" {
				p.builder.token(syntax.NewLineToken, cli.newline)
			}
			p.advance()
			p.builder.finishNode(syntax.FencedCodeBlock)
			return
		}

		// Code content line — emit entire line as a single CodeLineToken.
		p.builder.token(syntax.CodeLineToken, cl.Content)
		p.advance()
	}

	// Unclosed fence — still valid, just no closing fence token.
	p.builder.finishNode(syntax.FencedCodeBlock)
}
