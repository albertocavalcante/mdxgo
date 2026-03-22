package syntax

import "testing"

func TestGreenTokenWidth(t *testing.T) {
	tok := NewGreenToken(TextToken, "hello")
	if got := tok.Width(); got != 5 {
		t.Errorf("Width() = %d, want 5", got)
	}
}

func TestGreenTokenWidthWithTrivia(t *testing.T) {
	tok := NewGreenTokenTrivia(TextToken,
		NewTriviaList(Trivia{Kind: WhitespaceTrivia, Text: "  "}),
		"hello",
		NewTriviaList(Trivia{Kind: WhitespaceTrivia, Text: " "}),
	)
	if got := tok.Width(); got != 8 {
		t.Errorf("Width() = %d, want 8 (2+5+1)", got)
	}
}

func TestGreenTokenFullText(t *testing.T) {
	tok := NewGreenTokenTrivia(TextToken,
		NewTriviaList(Trivia{Kind: WhitespaceTrivia, Text: "\t"}),
		"world",
		NewTriviaList(Trivia{Kind: EndOfLineTrivia, Text: "\n"}),
	)
	if got := tok.FullText(); got != "\tworld\n" {
		t.Errorf("FullText() = %q, want %q", got, "\tworld\n")
	}
}

func TestGreenNodeWidth(t *testing.T) {
	t1 := NewGreenToken(TextToken, "hello")
	t2 := NewGreenToken(TextToken, " world")
	node := NewGreenNode(Paragraph, []GreenElement{
		TokenElement(t1),
		TokenElement(t2),
	})
	if got := node.Width; got != 11 {
		t.Errorf("Width = %d, want 11", got)
	}
}

func TestGreenNodeReplaceChild(t *testing.T) {
	t1 := NewGreenToken(TextToken, "old")
	t2 := NewGreenToken(TextToken, "other")
	node := NewGreenNode(Paragraph, []GreenElement{
		TokenElement(t1),
		TokenElement(t2),
	})

	newTok := NewGreenToken(TextToken, "new!")
	node2 := node.ReplaceChild(0, TokenElement(newTok))

	// Original unchanged.
	if node.Children[0].Token.Text != "old" {
		t.Error("original node was mutated")
	}
	// New node has replacement.
	if node2.Children[0].Token.Text != "new!" {
		t.Errorf("got %q, want %q", node2.Children[0].Token.Text, "new!")
	}
	// Width updated.
	if node2.Width != 9 {
		t.Errorf("Width = %d, want 9", node2.Width)
	}
}

func TestFullTextRoundTrip(t *testing.T) {
	src := "hello world\n"
	tok := NewGreenToken(TextToken, src)
	doc := NewGreenNode(Document, []GreenElement{TokenElement(tok)})
	if got := FullText(doc); got != src {
		t.Errorf("FullText = %q, want %q", got, src)
	}
}
