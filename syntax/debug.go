package syntax

import (
	"fmt"
	"strings"
)

// DebugDump returns a human-readable tree dump of a green tree,
// useful for test assertions and debugging.
func DebugDump(root *GreenNode) string {
	var b strings.Builder
	dumpGreenNode(&b, root, 0)
	return b.String()
}

func dumpGreenNode(b *strings.Builder, n *GreenNode, depth int) {
	indent := strings.Repeat("  ", depth)
	fmt.Fprintf(b, "%s%s [%d]\n", indent, n.Kind, n.Width)
	for _, child := range n.Children {
		if child.Token != nil {
			dumpGreenToken(b, child.Token, depth+1)
		} else {
			dumpGreenNode(b, child.Node, depth+1)
		}
	}
}

func dumpGreenToken(b *strings.Builder, t *GreenToken, depth int) {
	indent := strings.Repeat("  ", depth)
	text := t.Text
	const maxTextLen = 40
	if len(text) > maxTextLen {
		text = text[:maxTextLen-3] + "..."
	}
	text = strings.ReplaceAll(text, "\n", "\\n")
	text = strings.ReplaceAll(text, "\r", "\\r")
	text = strings.ReplaceAll(text, "\t", "\\t")

	if t.LeadingTrivia.IsEmpty() && t.TrailingTrivia.IsEmpty() {
		fmt.Fprintf(b, "%s%s %q\n", indent, t.Kind, text)
	} else {
		fmt.Fprintf(b, "%s%s %q (lead=%d, trail=%d)\n",
			indent, t.Kind, text,
			t.LeadingTrivia.Width(), t.TrailingTrivia.Width())
	}
}
