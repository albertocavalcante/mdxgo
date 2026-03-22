package syntax

import "testing"

func buildTestTree() *GreenNode {
	// Document {
	//   ATXHeading {
	//     HashToken "#"
	//     HeadingTextToken "title"
	//   }
	//   Paragraph {
	//     TextToken "hello"
	//     TextToken " world"
	//   }
	// }
	heading := NewGreenNode(ATXHeading, []GreenElement{
		TokenElement(NewGreenToken(HashToken, "#")),
		TokenElement(NewGreenToken(HeadingTextToken, "title")),
	})
	para := NewGreenNode(Paragraph, []GreenElement{
		TokenElement(NewGreenToken(TextToken, "hello")),
		TokenElement(NewGreenToken(TextToken, " world")),
	})
	return NewGreenNode(Document, []GreenElement{
		NodeElement(heading),
		NodeElement(para),
	})
}

func TestWalkVisitsAllNodes(t *testing.T) {
	doc := buildTestTree()
	var visited []SyntaxKind
	Walk(doc, func(n *GreenNode, depth int) WalkAction {
		visited = append(visited, n.Kind)
		return Continue
	})

	// Should visit: Document, ATXHeading, Paragraph
	expected := []SyntaxKind{Document, ATXHeading, Paragraph}
	if len(visited) != len(expected) {
		t.Fatalf("visited %d nodes, want %d: %v", len(visited), len(expected), visited)
	}
	for i, k := range expected {
		if visited[i] != k {
			t.Errorf("visited[%d] = %v, want %v", i, visited[i], k)
		}
	}
}

func TestWalkSkipChildren(t *testing.T) {
	doc := buildTestTree()
	var visited []SyntaxKind
	Walk(doc, func(n *GreenNode, depth int) WalkAction {
		visited = append(visited, n.Kind)
		if n.Kind == ATXHeading {
			return SkipChildren
		}
		return Continue
	})

	// Should visit: Document, ATXHeading (children skipped), Paragraph
	expected := []SyntaxKind{Document, ATXHeading, Paragraph}
	if len(visited) != len(expected) {
		t.Fatalf("visited %d nodes, want %d: %v", len(visited), len(expected), visited)
	}
}

func TestWalkStop(t *testing.T) {
	doc := buildTestTree()
	var visited []SyntaxKind
	Walk(doc, func(n *GreenNode, depth int) WalkAction {
		visited = append(visited, n.Kind)
		if n.Kind == ATXHeading {
			return Stop
		}
		return Continue
	})

	// Should stop after ATXHeading
	expected := []SyntaxKind{Document, ATXHeading}
	if len(visited) != len(expected) {
		t.Fatalf("visited %d nodes, want %d: %v", len(visited), len(expected), visited)
	}
}

func TestWalkDepth(t *testing.T) {
	doc := buildTestTree()
	depths := make(map[SyntaxKind]int)
	Walk(doc, func(n *GreenNode, depth int) WalkAction {
		depths[n.Kind] = depth
		return Continue
	})

	if depths[Document] != 0 {
		t.Errorf("Document depth = %d, want 0", depths[Document])
	}
	if depths[ATXHeading] != 1 {
		t.Errorf("ATXHeading depth = %d, want 1", depths[ATXHeading])
	}
	if depths[Paragraph] != 1 {
		t.Errorf("Paragraph depth = %d, want 1", depths[Paragraph])
	}
}

func TestWalkAllVisitsNodesAndTokens(t *testing.T) {
	doc := buildTestTree()
	var kinds []SyntaxKind
	WalkAll(doc, func(elem WalkElement, depth int) WalkAllAction {
		kinds = append(kinds, elem.Kind())
		return ContinueAll
	})

	// Document, ATXHeading, HashToken, HeadingTextToken, Paragraph, TextToken, TextToken
	expected := []SyntaxKind{Document, ATXHeading, HashToken, HeadingTextToken, Paragraph, TextToken, TextToken}
	if len(kinds) != len(expected) {
		t.Fatalf("visited %d elements, want %d: %v", len(kinds), len(expected), kinds)
	}
	for i, k := range expected {
		if kinds[i] != k {
			t.Errorf("kinds[%d] = %v, want %v", i, kinds[i], k)
		}
	}
}

func TestNodesIterator(t *testing.T) {
	doc := buildTestTree()
	var kinds []SyntaxKind
	for n := range Nodes(doc) {
		kinds = append(kinds, n.Kind)
	}

	expected := []SyntaxKind{Document, ATXHeading, Paragraph}
	if len(kinds) != len(expected) {
		t.Fatalf("got %d nodes, want %d: %v", len(kinds), len(expected), kinds)
	}
	for i, k := range expected {
		if kinds[i] != k {
			t.Errorf("kinds[%d] = %v, want %v", i, kinds[i], k)
		}
	}
}

func TestTokensIterator(t *testing.T) {
	doc := buildTestTree()
	var texts []string
	for tok := range Tokens(doc) {
		texts = append(texts, tok.Text)
	}

	expected := []string{"#", "title", "hello", " world"}
	if len(texts) != len(expected) {
		t.Fatalf("got %d tokens, want %d: %v", len(texts), len(expected), texts)
	}
	for i, txt := range expected {
		if texts[i] != txt {
			t.Errorf("texts[%d] = %q, want %q", i, texts[i], txt)
		}
	}
}

func TestNodesOfKindIterator(t *testing.T) {
	doc := buildTestTree()
	count := 0
	for n := range NodesOfKind(doc, Paragraph) {
		count++
		if n.Kind != Paragraph {
			t.Errorf("expected Paragraph, got %v", n.Kind)
		}
	}
	if count != 1 {
		t.Errorf("got %d Paragraphs, want 1", count)
	}
}

func TestTokensOfKindIterator(t *testing.T) {
	doc := buildTestTree()
	count := 0
	for tok := range TokensOfKind(doc, TextToken) {
		count++
		if tok.Kind != TextToken {
			t.Errorf("expected TextToken, got %v", tok.Kind)
		}
	}
	if count != 2 {
		t.Errorf("got %d TextTokens, want 2", count)
	}
}

func TestNodesIteratorBreak(t *testing.T) {
	doc := buildTestTree()
	count := 0
	for range Nodes(doc) {
		count++
		if count == 2 {
			break
		}
	}
	if count != 2 {
		t.Errorf("got %d iterations, want 2", count)
	}
}
