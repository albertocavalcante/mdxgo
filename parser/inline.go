package parser

import (
	"github.com/albertocavalcante/mdxgo/syntax"
)

// inlineParser decomposes flat TextToken/HeadingTextToken content inside
// Paragraph and Heading nodes into structured inline nodes.
//
// Design: the block parser runs first (unchanged), producing a green tree
// where Paragraph/Heading nodes contain flat TextToken/HeadingTextToken
// children. Then parseInlines walks the green tree and rewrites those
// flat tokens into structured inline elements (CodeSpan, BackslashEscape,
// EntityRef, AutolinkSpan, RawHTMLSpan, HardLineBreak, SoftLineBreak,
// InlineText, EmphasisSpan, StrongSpan, Link, Image).
//
// The round-trip invariant is maintained: the FullText of the rewritten
// tree equals the FullText of the original tree.

// parseInlines rewrites the given green tree, decomposing inline content
// within Paragraph and Heading nodes. Returns a new green tree.
func parseInlines(root *syntax.GreenNode, opts Options) *syntax.GreenNode {
	return rewriteInlines(root, opts)
}

// rewriteInlines recursively walks a green node, rewriting containers and
// inline-bearing nodes.
func rewriteInlines(n *syntax.GreenNode, opts Options) *syntax.GreenNode {
	changed := false
	newChildren := make([]syntax.GreenElement, len(n.Children))

	for i, child := range n.Children {
		if child.Node != nil {
			switch child.Node.Kind {
			case syntax.Paragraph, syntax.ATXHeading, syntax.SetextHeading:
				rewritten := rewriteInlineContent(child.Node, opts)
				if rewritten != child.Node {
					newChildren[i] = syntax.NodeElement(rewritten)
					changed = true
					continue
				}
			default:
				// Recurse into containers (BlockQuote, ListItem, etc.)
				rewritten := rewriteInlines(child.Node, opts)
				if rewritten != child.Node {
					newChildren[i] = syntax.NodeElement(rewritten)
					changed = true
					continue
				}
			}
		}
		newChildren[i] = child
	}

	if !changed {
		return n
	}
	return syntax.NewGreenNode(n.Kind, newChildren)
}

// rewriteInlineContent takes a Paragraph or Heading node and rewrites
// its TextToken/HeadingTextToken children into structured inline elements.
func rewriteInlineContent(n *syntax.GreenNode, opts Options) *syntax.GreenNode {
	// Collect the text tokens that contain inline content.
	// Non-text tokens (HashToken, IndentToken, SetextUnderline, NewLineToken)
	// are preserved as-is.
	type textRun struct {
		startIdx int // index in n.Children
		endIdx   int // exclusive
		text     string
	}

	var runs []textRun
	hasInlineContent := false

	for i, child := range n.Children {
		if child.Token != nil && isInlineTextKind(child.Token.Kind) {
			hasInlineContent = true
			// Extend existing run or start new one.
			if len(runs) > 0 && runs[len(runs)-1].endIdx == i {
				runs[len(runs)-1].endIdx = i + 1
				runs[len(runs)-1].text += child.Token.FullText()
			} else {
				runs = append(runs, textRun{
					startIdx: i,
					endIdx:   i + 1,
					text:     child.Token.FullText(),
				})
			}
		}
	}

	if !hasInlineContent {
		return n
	}

	// Check if any run contains inline markup worth parsing.
	hasMarkup := false
	for _, run := range runs {
		if containsInlineMarkup(run.text) {
			hasMarkup = true
			break
		}
	}
	if !hasMarkup {
		return n
	}

	// Rebuild children, replacing text runs with parsed inline elements.
	var newChildren []syntax.GreenElement
	runIdx := 0

	for i, child := range n.Children {
		if runIdx < len(runs) && i == runs[runIdx].startIdx {
			// Parse this text run into inline elements.
			elements := parseInlineContent(runs[runIdx].text, opts)
			newChildren = append(newChildren, elements...)
			i = runs[runIdx].endIdx - 1 // skip consumed children
			// Fast-forward past the run (the loop will i++ for us).
			for j := runs[runIdx].startIdx + 1; j < runs[runIdx].endIdx; j++ {
				_ = j // consumed by parseInlineContent
			}
			runIdx++
			continue
		}
		// Check if we're inside a run (and should skip).
		skip := false
		if runIdx > 0 && i > runs[runIdx-1].startIdx && i < runs[runIdx-1].endIdx {
			skip = true
		}
		if !skip {
			newChildren = append(newChildren, child)
		}
	}

	return syntax.NewGreenNode(n.Kind, newChildren)
}

// isInlineTextKind reports whether a token kind contains inline content
// that should be decomposed.
func isInlineTextKind(k syntax.SyntaxKind) bool {
	return k == syntax.TextToken || k == syntax.HeadingTextToken
}

// containsInlineMarkup reports whether text contains any character that
// might start an inline construct. Uses a pre-computed lookup table
// (inlineMarkupChars) for the fast path.
func containsInlineMarkup(text string) bool {
	for i := 0; i < len(text); i++ {
		if inlineMarkupChars[text[i]] {
			return true
		}
		// Check for hard line break (2+ spaces before newline).
		if text[i] == ' ' && i+1 < len(text) && (text[i+1] == '\n' || text[i+1] == '\r') {
			return true
		}
	}
	return false
}

// parseInlineContent parses a string of inline content into green tree elements.
func parseInlineContent(text string, opts Options) []syntax.GreenElement {
	ip := &inlineProcessor{
		scanner: newInlineScanner(text),
		opts:    opts,
	}
	ip.parse()
	return ip.output
}

// inlineProcessor processes inline content into structured green elements.
type inlineProcessor struct {
	scanner *inlineScanner
	opts    Options
	output  []syntax.GreenElement
	// textBuf accumulates plain text that will become an InlineText node.
	textBuf []byte
	// delimiters is the delimiter stack for emphasis processing.
	delimiters []*delimiter
	// brackets is the bracket stack for link/image parsing.
	brackets []*bracketDelim
}

// parse drives the inline parsing loop.
func (ip *inlineProcessor) parse() {
	for !ip.scanner.eof() {
		b := ip.scanner.peek()
		switch b {
		case '`':
			ip.tryCodeSpan()
		case '\\':
			ip.tryBackslashEscape()
		case '&':
			ip.tryEntityRef()
		case '<':
			if ip.opts.MDX {
				ip.tryJSXInline()
			} else {
				ip.tryAutolinkOrRawHTML()
			}
		case '{':
			if ip.opts.MDX {
				ip.tryExpressionInline()
			} else {
				ip.textBuf = append(ip.textBuf, '{')
				ip.scanner.advance(1)
			}
		case '*', '_':
			ip.tryEmphasis()
		case '[':
			ip.tryOpenBracket()
		case ']':
			ip.tryCloseBracket()
		case '!':
			// '!' might start an image (![). Accumulate it for now;
			// tryOpenBracket will check for it when '[' is encountered.
			ip.textBuf = append(ip.textBuf, '!')
			ip.scanner.advance(1)
		case '\n', '\r':
			ip.tryLineBreak()
		case ' ':
			ip.tryHardBreakSpaces()
		default:
			ip.textBuf = append(ip.textBuf, b)
			ip.scanner.advance(1)
		}
	}
	ip.flushText()
	ip.processEmphasis()
}

// flushText emits any accumulated plain text as an InlineText node.
func (ip *inlineProcessor) flushText() {
	if len(ip.textBuf) == 0 {
		return
	}
	text := string(ip.textBuf)
	ip.textBuf = ip.textBuf[:0]
	ip.output = append(ip.output, syntax.NodeElement(
		syntax.NewGreenNode(syntax.InlineText, []syntax.GreenElement{
			syntax.TokenElement(syntax.NewGreenToken(syntax.TextToken, text)),
		}),
	))
}

// emitToken is a convenience to flush text and emit a single token.
func (ip *inlineProcessor) emitToken(kind syntax.SyntaxKind, text string) {
	ip.flushText()
	ip.output = append(ip.output, syntax.TokenElement(
		syntax.NewGreenToken(kind, text),
	))
}

// emitNode flushes text and emits a node.
func (ip *inlineProcessor) emitNode(n *syntax.GreenNode) {
	ip.flushText()
	ip.output = append(ip.output, syntax.NodeElement(n))
}
