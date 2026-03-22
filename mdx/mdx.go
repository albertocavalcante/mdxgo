// Package mdx provides a convenience entry point for parsing MDX v3 files
// into a lossless Concrete Syntax Tree.
package mdx

import (
	"github.com/albertocavalcante/mdxgo/parser"
	"github.com/albertocavalcante/mdxgo/syntax"
)

// Parse parses MDX v3 source into a green CST rooted at a Document node.
// The round-trip invariant holds: syntax.FullText(Parse(src)) == string(src).
func Parse(src []byte) *syntax.GreenNode {
	return parser.Parse(src, parser.Options{MDX: true})
}

// ParseCommonMark parses CommonMark source (no MDX extensions).
func ParseCommonMark(src []byte) *syntax.GreenNode {
	return parser.Parse(src, parser.Options{MDX: false})
}
