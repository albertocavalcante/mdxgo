package parser

import (
	"strings"

	"github.com/albertocavalcante/mdxgo/syntax"
)

// isFrontmatterOpen checks if a line is a frontmatter opening fence (--- or +++).
func isFrontmatterOpen(li lineInfo) bool {
	if li.indent > 0 {
		return false
	}
	s := li.trimmed
	if len(s) < minFenceLength {
		return false
	}
	if (s[0] == '-' && strings.TrimRight(s, "-") == "") ||
		(s[0] == '+' && strings.TrimRight(s, "+") == "") {
		return len(s) >= minFenceLength
	}
	return false
}

// parseFrontmatter parses a YAML/TOML frontmatter block.
func (p *blockParser) parseFrontmatter(openLI lineInfo) {
	fenceChar := openLI.trimmed[0]
	p.builder.startNode()

	// Opening fence.
	p.builder.token(syntax.FrontmatterFence, openLI.content)
	if openLI.newline != "" {
		p.builder.token(syntax.NewLineToken, openLI.newline)
	}
	p.advance()

	// Content lines until closing fence.
	for !p.eof() {
		li := analyzeLine(p.currentLine().Content)

		// Check for closing fence (same char, same or more count).
		if !li.blank && li.indent == 0 {
			s := li.trimmed
			if len(s) >= minFenceLength && s[0] == fenceChar {
				allSame := true
				for _, c := range s {
					if byte(c) != fenceChar {
						allSame = false
						break
					}
				}
				if allSame {
					// Closing fence.
					p.builder.token(syntax.FrontmatterFence, li.content)
					if li.newline != "" {
						p.builder.token(syntax.NewLineToken, li.newline)
					}
					p.advance()
					p.builder.finishNode(syntax.Frontmatter)
					return
				}
			}
		}

		// Content line.
		p.builder.token(syntax.FrontmatterLine, p.currentLine().Content)
		p.advance()
	}

	// Unclosed frontmatter.
	p.builder.finishNode(syntax.Frontmatter)
}
