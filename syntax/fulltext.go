package syntax

import "strings"

// FullText reconstructs the complete source text from a green tree.
// This is the critical round-trip function: FullText(Parse(src)) == src.
func FullText(root *GreenNode) string {
	var b strings.Builder
	b.Grow(root.Width)
	appendGreenNode(&b, root)
	return b.String()
}

func appendGreenNode(b *strings.Builder, n *GreenNode) {
	for _, child := range n.Children {
		if child.Token != nil {
			appendGreenToken(b, child.Token)
		} else {
			appendGreenNode(b, child.Node)
		}
	}
}

func appendGreenToken(b *strings.Builder, t *GreenToken) {
	for i := range t.LeadingTrivia.Len() {
		b.WriteString(t.LeadingTrivia.At(i).Text)
	}
	b.WriteString(t.Text)
	for i := range t.TrailingTrivia.Len() {
		b.WriteString(t.TrailingTrivia.At(i).Text)
	}
}

// FullTextNode reconstructs the complete source text from a red node.
func FullTextNode(n *SyntaxNode) string {
	return FullText(n.Green())
}
