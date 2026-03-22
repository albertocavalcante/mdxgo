package syntax

// GreenElement is a tagged union: either a *GreenToken or a *GreenNode.
// Exactly one field is non-nil.
type GreenElement struct {
	Token *GreenToken
	Node  *GreenNode
}

// Kind returns the SyntaxKind of the element.
func (e GreenElement) Kind() SyntaxKind {
	if e.Token != nil {
		return e.Token.Kind
	}
	return e.Node.Kind
}

// Width returns the total byte width of the element.
func (e GreenElement) Width() int {
	if e.Token != nil {
		return e.Token.Width()
	}
	return e.Node.Width
}

// IsToken reports whether this element wraps a token.
func (e GreenElement) IsToken() bool { return e.Token != nil }

// IsNode reports whether this element wraps a node.
func (e GreenElement) IsNode() bool { return e.Node != nil }

// TokenElement wraps a GreenToken as a GreenElement.
func TokenElement(t *GreenToken) GreenElement {
	return GreenElement{Token: t}
}

// NodeElement wraps a GreenNode as a GreenElement.
func NodeElement(n *GreenNode) GreenElement {
	return GreenElement{Node: n}
}

// -------------------------------------------------------------------
// GreenToken — immutable leaf node in the green tree
// -------------------------------------------------------------------

// GreenToken is an immutable leaf in the green (data) tree.
// It holds its kind, the literal source text, and any attached trivia.
type GreenToken struct {
	Kind           SyntaxKind
	LeadingTrivia  TriviaList
	Text           string
	TrailingTrivia TriviaList
}

// NewGreenToken creates a GreenToken with optional trivia.
func NewGreenToken(kind SyntaxKind, text string) *GreenToken {
	return &GreenToken{Kind: kind, Text: text}
}

// NewGreenTokenTrivia creates a GreenToken with leading and trailing trivia.
func NewGreenTokenTrivia(kind SyntaxKind, leading TriviaList, text string, trailing TriviaList) *GreenToken {
	return &GreenToken{
		Kind:           kind,
		LeadingTrivia:  leading,
		Text:           text,
		TrailingTrivia: trailing,
	}
}

// Width returns the total byte width: leading trivia + text + trailing trivia.
func (t *GreenToken) Width() int {
	return t.LeadingTrivia.Width() + len(t.Text) + t.TrailingTrivia.Width()
}

// FullText returns the complete source text including trivia.
func (t *GreenToken) FullText() string {
	lt := t.LeadingTrivia.Text()
	tt := t.TrailingTrivia.Text()
	if lt == "" && tt == "" {
		return t.Text
	}
	buf := make([]byte, 0, len(lt)+len(t.Text)+len(tt))
	buf = append(buf, lt...)
	buf = append(buf, t.Text...)
	buf = append(buf, tt...)
	return string(buf)
}

// WithLeadingTrivia returns a copy of this token with different leading trivia.
func (t *GreenToken) WithLeadingTrivia(trivia TriviaList) *GreenToken {
	return &GreenToken{
		Kind:           t.Kind,
		LeadingTrivia:  trivia,
		Text:           t.Text,
		TrailingTrivia: t.TrailingTrivia,
	}
}

// WithTrailingTrivia returns a copy of this token with different trailing trivia.
func (t *GreenToken) WithTrailingTrivia(trivia TriviaList) *GreenToken {
	return &GreenToken{
		Kind:           t.Kind,
		LeadingTrivia:  t.LeadingTrivia,
		Text:           t.Text,
		TrailingTrivia: trivia,
	}
}

// WithText returns a copy of this token with different text content.
func (t *GreenToken) WithText(text string) *GreenToken {
	return &GreenToken{
		Kind:           t.Kind,
		LeadingTrivia:  t.LeadingTrivia,
		Text:           text,
		TrailingTrivia: t.TrailingTrivia,
	}
}

// -------------------------------------------------------------------
// GreenNode — immutable interior node in the green tree
// -------------------------------------------------------------------

// GreenNode is an immutable interior node in the green (data) tree.
// It stores its kind, cached width, and an ordered list of child elements.
type GreenNode struct {
	Kind     SyntaxKind
	Width    int
	Children []GreenElement
}

// NewGreenNode creates a GreenNode with the given kind and children.
// Width is computed automatically from children.
func NewGreenNode(kind SyntaxKind, children []GreenElement) *GreenNode {
	w := 0
	for _, c := range children {
		w += c.Width()
	}
	cp := make([]GreenElement, len(children))
	copy(cp, children)
	return &GreenNode{Kind: kind, Width: w, Children: cp}
}

// ChildCount returns the number of direct children.
func (n *GreenNode) ChildCount() int { return len(n.Children) }

// ReplaceChild returns a new GreenNode with child at index i replaced.
// All other children are structurally shared.
func (n *GreenNode) ReplaceChild(i int, newChild GreenElement) *GreenNode {
	newChildren := make([]GreenElement, len(n.Children))
	copy(newChildren, n.Children)
	newChildren[i] = newChild
	w := 0
	for _, c := range newChildren {
		w += c.Width()
	}
	return &GreenNode{Kind: n.Kind, Width: w, Children: newChildren}
}

// AppendChild returns a new GreenNode with an additional child appended.
func (n *GreenNode) AppendChild(child GreenElement) *GreenNode {
	newChildren := make([]GreenElement, len(n.Children)+1)
	copy(newChildren, n.Children)
	newChildren[len(n.Children)] = child
	return &GreenNode{Kind: n.Kind, Width: n.Width + child.Width(), Children: newChildren}
}
