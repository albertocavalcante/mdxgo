package parser

import (
	"strings"

	"github.com/albertocavalcante/mdxgo/syntax"
)

// tryATXHeading checks whether the line is an ATX heading (# … ######).
func tryATXHeading(li lineInfo) bool {
	s := li.trimmed
	if s == "" || s[0] != '#' {
		return false
	}
	i := 0
	for i < len(s) && i < maxHeadingLevel && s[i] == '#' {
		i++
	}
	if i > maxHeadingLevel {
		return false
	}
	// Must be followed by space/tab or end of line.
	if i < len(s) && s[i] != ' ' && s[i] != '\t' {
		return false
	}
	return true
}

// parseATXHeading parses an ATX heading line into an ATXHeading node.
// Structure: [IndentToken] HashToken [HeadingTextToken] [NewLineToken]
func (p *blockParser) parseATXHeading(li lineInfo) {
	p.builder.startNode()

	raw := li.content
	pos := 0

	// Leading indent (spaces before #).
	if li.indent > 0 && li.indent <= maxIndent {
		indentText := raw[:li.indent]
		p.builder.token(syntax.IndentToken, indentText)
		pos = li.indent
	}

	// Hash prefix.
	hashStart := pos
	for pos < len(raw) && raw[pos] == '#' {
		pos++
	}
	hashText := raw[hashStart:pos]

	// Space after hashes — trailing trivia on the hash token.
	spaceAfterHash := ""
	if pos < len(raw) && (raw[pos] == ' ' || raw[pos] == '\t') {
		spaceStart := pos
		pos++
		spaceAfterHash = raw[spaceStart:pos]
	}

	if spaceAfterHash != "" {
		p.builder.tokenTrivia(syntax.HashToken,
			syntax.TriviaList{},
			hashText,
			syntax.NewTriviaList(syntax.Trivia{Kind: syntax.WhitespaceTrivia, Text: spaceAfterHash}),
		)
	} else {
		p.builder.token(syntax.HashToken, hashText)
	}

	// Content: everything up to optional closing hashes and trailing whitespace.
	if pos < len(raw) {
		remaining := raw[pos:]

		// Strip optional closing sequence: trailing spaces + hashes + spaces.
		content := remaining
		content = strings.TrimRight(content, " \t")
		if content != "" && content[len(content)-1] == '#' {
			// Check for closing hashes.
			closeEnd := len(content)
			for closeEnd > 0 && content[closeEnd-1] == '#' {
				closeEnd--
			}
			// Closing hashes must be preceded by space or be the entire content.
			if closeEnd == 0 || content[closeEnd-1] == ' ' || content[closeEnd-1] == '\t' {
				closingHashes := content[closeEnd:]
				_ = closingHashes
				// For round-trip: emit the entire remaining text as one token.
				// We don't split closing hashes in phase 2 — they're part of content.
			}
		}

		// Emit the full remaining line content as a single HeadingTextToken.
		p.builder.token(syntax.HeadingTextToken, remaining)
	}

	// Line ending.
	if li.newline != "" {
		p.builder.token(syntax.NewLineToken, li.newline)
	}

	p.builder.finishNode(syntax.ATXHeading)
	p.advance()
}

// isSetextUnderline checks if a line is a setext heading underline (=== or ---).
func isSetextUnderline(li lineInfo) (byte, bool) {
	if li.blank || li.indent > maxIndent {
		return 0, false
	}
	s := li.trimmed
	if s == "" {
		return 0, false
	}
	ch := s[0]
	if ch != '=' && ch != '-' {
		return 0, false
	}
	trimmed := strings.TrimRight(s, string(ch))
	trimmed = strings.TrimRight(trimmed, " \t")
	if trimmed != "" {
		return 0, false
	}
	return ch, true
}
