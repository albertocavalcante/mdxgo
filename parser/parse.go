package parser

import (
	"github.com/albertocavalcante/mdxgo/syntax"
)

// Options configures the parser.
type Options struct {
	// MDX enables MDX v3 extensions: JSX blocks/inline, expressions,
	// ESM, frontmatter. When false, parse as CommonMark.
	MDX bool
}

// Parse parses source bytes into a green CST rooted at a Document node.
// The round-trip invariant holds: syntax.FullText(Parse(src, opts)) == string(src).
//
// Parsing proceeds in two passes:
//  1. Block pass — splits source into block-level nodes (Paragraph, Heading, etc.)
//  2. Inline pass — decomposes text content within Paragraph/Heading nodes into
//     structured inline elements (CodeSpan, BackslashEscape, EmphasisSpan, etc.)
func Parse(src []byte, opts Options) *syntax.GreenNode {
	sc := newScanner(src)
	b := newBuilder()
	p := &blockParser{
		lines:   sc.lines,
		builder: b,
		opts:    opts,
	}
	p.parseDocument()
	blockTree := b.finish()

	// Phase 2: inline parsing pass.
	return parseInlines(blockTree, opts)
}

// blockParser drives the line-oriented first pass of CommonMark parsing.
type blockParser struct {
	lines   []line
	pos     int // current line index
	builder *builder
	opts    Options
}

func (p *blockParser) parseDocument() {
	for p.pos < len(p.lines) {
		p.parseBlock()
	}
}

// parseBlock attempts to parse one block-level construct starting at p.pos.
func (p *blockParser) parseBlock() {
	if p.pos >= len(p.lines) {
		return
	}

	li := analyzeLine(p.lines[p.pos].Content)

	// Try each block parser in priority order.
	// Blank line
	if li.blank {
		p.parseBlankLine(li)
		return
	}

	// Frontmatter (only at position 0 in MDX mode)
	if p.opts.MDX && p.pos == 0 && isFrontmatterOpen(li) {
		p.parseFrontmatter(li)
		return
	}

	// ATX heading
	if tryATXHeading(li) {
		p.parseATXHeading(li)
		return
	}

	// Thematic break (must check before setext to avoid ambiguity)
	if isThematicBreak(li) {
		p.parseThematicBreak(li)
		return
	}

	// Fenced code block
	if isFenceOpen(li) {
		p.parseFencedCode(li)
		return
	}

	// Indented code block (4+ spaces, not in MDX mode where it's disabled)
	if !p.opts.MDX && li.indent >= indentedCodeIndent {
		p.parseIndentedCode(li)
		return
	}

	// HTML block (not in MDX mode)
	if !p.opts.MDX && isHTMLBlockStart(li) {
		p.parseHTMLBlock(li)
		return
	}

	// Block quote
	if isBlockQuoteStart(li) {
		p.parseBlockQuote(li)
		return
	}

	// List item
	if marker, ok := isListItemStart(li); ok {
		p.parseList(li, marker)
		return
	}

	// ESM (MDX mode: import/export at start of line)
	if p.opts.MDX && isESMStart(li) {
		p.parseESM(li)
		return
	}

	// JSX block (MDX mode: line starting with <)
	if p.opts.MDX && isJSXBlockStart(li) {
		p.parseJSXBlock(li)
		return
	}

	// Expression block (MDX mode: line starting with {)
	if p.opts.MDX && isExprBlockStart(li) {
		p.parseExpressionBlock(li)
		return
	}

	// Fallthrough: paragraph (may become setext heading)
	p.parseParagraph(li)
}

// eof reports whether we've consumed all lines.
func (p *blockParser) eof() bool {
	return p.pos >= len(p.lines)
}

// currentLine returns the current line or empty if at EOF.
func (p *blockParser) currentLine() line {
	if p.pos < len(p.lines) {
		return p.lines[p.pos]
	}
	return line{}
}

// advance moves to the next line.
func (p *blockParser) advance() {
	p.pos++
}
