package syntax

import "iter"

// -------------------------------------------------------------------
// SyntaxElement — tagged union of red-tree wrappers
// -------------------------------------------------------------------

// SyntaxElement is a tagged union: either a *SyntaxNode or a *SyntaxToken.
type SyntaxElement struct {
	Node  *SyntaxNode
	Token *SyntaxToken
}

// Kind returns the SyntaxKind of the element.
func (e SyntaxElement) Kind() SyntaxKind {
	if e.Token != nil {
		return e.Token.Kind()
	}
	return e.Node.Kind()
}

// Offset returns the absolute byte offset of the element.
func (e SyntaxElement) Offset() int {
	if e.Token != nil {
		return e.Token.Offset()
	}
	return e.Node.Offset()
}

// Width returns the total width of the element.
func (e SyntaxElement) Width() int {
	if e.Token != nil {
		return e.Token.green.Width()
	}
	return e.Node.green.Width
}

// IsToken reports whether this is a token.
func (e SyntaxElement) IsToken() bool { return e.Token != nil }

// IsNode reports whether this is a node.
func (e SyntaxElement) IsNode() bool { return e.Node != nil }

// -------------------------------------------------------------------
// SyntaxNode — ephemeral red wrapper for GreenNode
// -------------------------------------------------------------------

// SyntaxNode is an ephemeral wrapper manufactured during tree traversal.
// It provides parent pointers and absolute offsets computed from green widths.
type SyntaxNode struct {
	green  *GreenNode
	parent *SyntaxNode
	offset int // absolute byte offset in source
	index  int // child index within parent
}

// NewSyntaxRoot creates the root red node for a green tree.
func NewSyntaxRoot(green *GreenNode) *SyntaxNode {
	return &SyntaxNode{green: green, parent: nil, offset: 0, index: 0}
}

// Green returns the underlying green node.
func (n *SyntaxNode) Green() *GreenNode { return n.green }

// Kind returns the syntax kind.
func (n *SyntaxNode) Kind() SyntaxKind { return n.green.Kind }

// Offset returns the absolute byte offset in source.
func (n *SyntaxNode) Offset() int { return n.offset }

// Width returns the total byte width.
func (n *SyntaxNode) Width() int { return n.green.Width }

// End returns the exclusive end offset.
func (n *SyntaxNode) End() int { return n.offset + n.green.Width }

// Parent returns the parent red node, or nil for the root.
func (n *SyntaxNode) Parent() *SyntaxNode { return n.parent }

// Index returns this node's child index within its parent.
func (n *SyntaxNode) Index() int { return n.index }

// ChildCount returns the number of direct children.
func (n *SyntaxNode) ChildCount() int { return len(n.green.Children) }

// Children returns an iterator over the children as SyntaxElements.
// Red wrappers are manufactured on-the-fly from green children.
func (n *SyntaxNode) Children() iter.Seq2[int, SyntaxElement] {
	return func(yield func(int, SyntaxElement) bool) {
		childOffset := n.offset
		for i, gc := range n.green.Children {
			var elem SyntaxElement
			if gc.Token != nil {
				elem = SyntaxElement{Token: &SyntaxToken{
					green:  gc.Token,
					parent: n,
					offset: childOffset,
					index:  i,
				}}
			} else {
				elem = SyntaxElement{Node: &SyntaxNode{
					green:  gc.Node,
					parent: n,
					offset: childOffset,
					index:  i,
				}}
			}
			if !yield(i, elem) {
				return
			}
			childOffset += gc.Width()
		}
	}
}

// ChildAt returns the i-th child as a SyntaxElement.
func (n *SyntaxNode) ChildAt(i int) SyntaxElement {
	childOffset := n.offset
	for j := range i {
		childOffset += n.green.Children[j].Width()
	}
	gc := n.green.Children[i]
	if gc.Token != nil {
		return SyntaxElement{Token: &SyntaxToken{
			green:  gc.Token,
			parent: n,
			offset: childOffset,
			index:  i,
		}}
	}
	return SyntaxElement{Node: &SyntaxNode{
		green:  gc.Node,
		parent: n,
		offset: childOffset,
		index:  i,
	}}
}

// TokenAt walks the tree to find the leaf token containing the given offset.
func (n *SyntaxNode) TokenAt(offset int) *SyntaxToken {
	if offset < n.offset || offset >= n.End() {
		return nil
	}
	childOffset := n.offset
	for i, gc := range n.green.Children {
		w := gc.Width()
		if offset < childOffset+w {
			if gc.Token != nil {
				return &SyntaxToken{
					green:  gc.Token,
					parent: n,
					offset: childOffset,
					index:  i,
				}
			}
			child := &SyntaxNode{
				green:  gc.Node,
				parent: n,
				offset: childOffset,
				index:  i,
			}
			return child.TokenAt(offset)
		}
		childOffset += w
	}
	return nil
}

// -------------------------------------------------------------------
// SyntaxToken — ephemeral red wrapper for GreenToken
// -------------------------------------------------------------------

// SyntaxToken is an ephemeral wrapper for a green token, providing
// parent pointers and absolute positions.
type SyntaxToken struct {
	green  *GreenToken
	parent *SyntaxNode
	offset int // absolute byte offset (including leading trivia)
	index  int // child index within parent
}

// Green returns the underlying green token.
func (t *SyntaxToken) Green() *GreenToken { return t.green }

// Kind returns the syntax kind.
func (t *SyntaxToken) Kind() SyntaxKind { return t.green.Kind }

// Offset returns the absolute byte offset (including leading trivia).
func (t *SyntaxToken) Offset() int { return t.offset }

// TextOffset returns the absolute byte offset of the token text,
// excluding leading trivia.
func (t *SyntaxToken) TextOffset() int { return t.offset + t.green.LeadingTrivia.Width() }

// TextEnd returns the exclusive end offset of the token text,
// excluding trailing trivia.
func (t *SyntaxToken) TextEnd() int { return t.TextOffset() + len(t.green.Text) }

// End returns the exclusive end offset including trailing trivia.
func (t *SyntaxToken) End() int { return t.offset + t.green.Width() }

// Text returns the token text without trivia.
func (t *SyntaxToken) Text() string { return t.green.Text }

// FullText returns the complete text including trivia.
func (t *SyntaxToken) FullText() string { return t.green.FullText() }

// Parent returns the parent red node.
func (t *SyntaxToken) Parent() *SyntaxNode { return t.parent }

// Index returns this token's child index within its parent.
func (t *SyntaxToken) Index() int { return t.index }

// LeadingTrivia returns the leading trivia list.
func (t *SyntaxToken) LeadingTrivia() TriviaList { return t.green.LeadingTrivia }

// TrailingTrivia returns the trailing trivia list.
func (t *SyntaxToken) TrailingTrivia() TriviaList { return t.green.TrailingTrivia }
