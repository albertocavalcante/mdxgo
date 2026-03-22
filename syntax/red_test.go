package syntax

import "testing"

func TestSyntaxNodeChildren(t *testing.T) {
	t1 := NewGreenToken(HashToken, "##")
	t2 := NewGreenTokenTrivia(HeadingTextToken,
		NewTriviaList(Trivia{Kind: WhitespaceTrivia, Text: " "}),
		"Title",
		TriviaList{},
	)
	t3 := NewGreenToken(NewLineToken, "\n")
	heading := NewGreenNode(ATXHeading, []GreenElement{
		TokenElement(t1),
		TokenElement(t2),
		TokenElement(t3),
	})
	doc := NewGreenNode(Document, []GreenElement{NodeElement(heading)})
	root := NewSyntaxRoot(doc)

	// Root should have 1 child (the heading node).
	count := 0
	for range root.Children() {
		count++
	}
	if count != 1 {
		t.Fatalf("root has %d children, want 1", count)
	}

	// Heading child should be a node.
	elem := root.ChildAt(0)
	if !elem.IsNode() {
		t.Fatal("expected node")
	}
	headingNode := elem.Node
	if headingNode.Kind() != ATXHeading {
		t.Errorf("kind = %v, want ATXHeading", headingNode.Kind())
	}
	if headingNode.Offset() != 0 {
		t.Errorf("offset = %d, want 0", headingNode.Offset())
	}

	// Heading should have 3 children.
	hCount := 0
	for range headingNode.Children() {
		hCount++
	}
	if hCount != 3 {
		t.Fatalf("heading has %d children, want 3", hCount)
	}

	// Check offsets of heading children.
	hashElem := headingNode.ChildAt(0)
	if !hashElem.IsToken() {
		t.Fatal("expected token for hash")
	}
	if hashElem.Token.Offset() != 0 {
		t.Errorf("hash offset = %d, want 0", hashElem.Token.Offset())
	}
	if hashElem.Token.Text() != "##" {
		t.Errorf("hash text = %q, want %q", hashElem.Token.Text(), "##")
	}

	titleElem := headingNode.ChildAt(1)
	if !titleElem.IsToken() {
		t.Fatal("expected token for title")
	}
	// Offset includes leading trivia of the title token.
	if titleElem.Token.Offset() != 2 {
		t.Errorf("title offset = %d, want 2", titleElem.Token.Offset())
	}
	// TextOffset excludes leading trivia.
	if titleElem.Token.TextOffset() != 3 {
		t.Errorf("title text offset = %d, want 3", titleElem.Token.TextOffset())
	}
}

func TestTokenAt(t *testing.T) {
	t1 := NewGreenToken(TextToken, "ab")
	t2 := NewGreenToken(TextToken, "cd")
	doc := NewGreenNode(Document, []GreenElement{
		TokenElement(t1),
		TokenElement(t2),
	})
	root := NewSyntaxRoot(doc)

	tok := root.TokenAt(0)
	if tok == nil || tok.Text() != "ab" {
		t.Errorf("TokenAt(0) = %v, want 'ab'", tok)
	}
	tok = root.TokenAt(2)
	if tok == nil || tok.Text() != "cd" {
		t.Errorf("TokenAt(2) = %v, want 'cd'", tok)
	}
	tok = root.TokenAt(4)
	if tok != nil {
		t.Errorf("TokenAt(4) = %v, want nil", tok)
	}
}
