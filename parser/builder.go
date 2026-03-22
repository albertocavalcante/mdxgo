package parser

import "github.com/albertocavalcante/mdxgo/syntax"

// builder constructs a green tree bottom-up.
// Blocks push tokens/nodes as children, then call finishNode to wrap them.
type builder struct {
	stack []syntax.GreenElement
	marks []int // stack of start positions for open nodes
}

func newBuilder() *builder {
	return &builder{}
}

// token appends a green token to the current position.
func (b *builder) token(kind syntax.SyntaxKind, text string) {
	b.stack = append(b.stack, syntax.TokenElement(
		syntax.NewGreenToken(kind, text),
	))
}

// tokenTrivia appends a green token with trivia.
func (b *builder) tokenTrivia(kind syntax.SyntaxKind, leading syntax.TriviaList, text string, trailing syntax.TriviaList) {
	b.stack = append(b.stack, syntax.TokenElement(
		syntax.NewGreenTokenTrivia(kind, leading, text, trailing),
	))
}

// startNode marks the current stack position so that finishNode
// can collect all children pushed after this mark.
func (b *builder) startNode() {
	b.marks = append(b.marks, len(b.stack))
}

// finishNode pops children since the last startNode mark and
// wraps them in a GreenNode of the given kind.
func (b *builder) finishNode(kind syntax.SyntaxKind) {
	mark := b.marks[len(b.marks)-1]
	b.marks = b.marks[:len(b.marks)-1]
	children := make([]syntax.GreenElement, len(b.stack)-mark)
	copy(children, b.stack[mark:])
	b.stack = b.stack[:mark]
	b.stack = append(b.stack, syntax.NodeElement(
		syntax.NewGreenNode(kind, children),
	))
}

// finish returns the completed Document root node.
func (b *builder) finish() *syntax.GreenNode {
	return syntax.NewGreenNode(syntax.Document, b.stack)
}
