package syntax

import "testing"

// --- ReplaceChild tests ---

func TestReplaceChildDirect(t *testing.T) {
	// Build: Document { Paragraph { TextToken "hello" } }
	para := NewGreenNode(Paragraph, []GreenElement{
		TokenElement(NewGreenToken(TextToken, "hello")),
	})
	doc := NewGreenNode(Document, []GreenElement{
		NodeElement(para),
	})

	// Replace the TextToken in the Paragraph (path=[], childIdx=0 of child at path [0]).
	newToken := TokenElement(NewGreenToken(TextToken, "world"))
	result := ReplaceChild(doc, []int{0}, 0, newToken)
	if result == nil {
		t.Fatal("ReplaceChild returned nil")
	}

	// Root should be different (new copy).
	if result == doc {
		t.Error("expected new root, got same pointer")
	}

	// Original should be unchanged.
	if FullText(doc) != "hello" {
		t.Errorf("original changed: %q", FullText(doc))
	}

	// New tree should have "world".
	if FullText(result) != "world" {
		t.Errorf("new tree: %q, want %q", FullText(result), "world")
	}
}

func TestReplaceChildRootLevel(t *testing.T) {
	tok1 := TokenElement(NewGreenToken(TextToken, "aaa"))
	tok2 := TokenElement(NewGreenToken(TextToken, "bbb"))
	doc := NewGreenNode(Document, []GreenElement{tok1, tok2})

	newTok := TokenElement(NewGreenToken(TextToken, "ccc"))
	result := ReplaceChild(doc, nil, 1, newTok)
	if result == nil {
		t.Fatal("ReplaceChild returned nil")
	}
	if FullText(result) != "aaaccc" {
		t.Errorf("got %q, want %q", FullText(result), "aaaccc")
	}
	// Original unchanged.
	if FullText(doc) != "aaabbb" {
		t.Errorf("original changed: %q", FullText(doc))
	}
}

func TestReplaceChildInvalidPath(t *testing.T) {
	doc := NewGreenNode(Document, []GreenElement{
		TokenElement(NewGreenToken(TextToken, "x")),
	})

	// Path points to index 5, which doesn't exist.
	result := ReplaceChild(doc, []int{5}, 0, TokenElement(NewGreenToken(TextToken, "y")))
	if result != nil {
		t.Error("expected nil for invalid path")
	}
}

func TestReplaceChildDeepPath(t *testing.T) {
	// Build: Document { BlockQuote { Paragraph { TextToken "deep" } } }
	para := NewGreenNode(Paragraph, []GreenElement{
		TokenElement(NewGreenToken(TextToken, "deep")),
	})
	bq := NewGreenNode(BlockQuote, []GreenElement{
		NodeElement(para),
	})
	doc := NewGreenNode(Document, []GreenElement{
		NodeElement(bq),
	})

	newTok := TokenElement(NewGreenToken(TextToken, "DEEP"))
	result := ReplaceChild(doc, []int{0, 0}, 0, newTok)
	if result == nil {
		t.Fatal("ReplaceChild returned nil")
	}
	if FullText(result) != "DEEP" {
		t.Errorf("got %q, want %q", FullText(result), "DEEP")
	}
	if FullText(doc) != "deep" {
		t.Errorf("original changed: %q", FullText(doc))
	}
}

// --- ReplaceDescendant tests ---

func TestReplaceDescendantNode(t *testing.T) {
	para1 := NewGreenNode(Paragraph, []GreenElement{
		TokenElement(NewGreenToken(TextToken, "para1")),
	})
	para2 := NewGreenNode(Paragraph, []GreenElement{
		TokenElement(NewGreenToken(TextToken, "para2")),
	})
	doc := NewGreenNode(Document, []GreenElement{
		NodeElement(para1),
		NodeElement(para2),
	})

	newPara := NewGreenNode(Paragraph, []GreenElement{
		TokenElement(NewGreenToken(TextToken, "replaced")),
	})

	result := ReplaceDescendant(doc, []int{1}, NodeElement(newPara))
	if result == nil {
		t.Fatal("ReplaceDescendant returned nil")
	}

	if FullText(result) != "para1replaced" {
		t.Errorf("got %q, want %q", FullText(result), "para1replaced")
	}
	if FullText(doc) != "para1para2" {
		t.Errorf("original changed: %q", FullText(doc))
	}
}

func TestReplaceDescendantDeep(t *testing.T) {
	inner := NewGreenNode(Paragraph, []GreenElement{
		TokenElement(NewGreenToken(TextToken, "inner")),
	})
	outer := NewGreenNode(BlockQuote, []GreenElement{
		NodeElement(inner),
	})
	doc := NewGreenNode(Document, []GreenElement{
		NodeElement(outer),
	})

	newInner := NewGreenNode(Paragraph, []GreenElement{
		TokenElement(NewGreenToken(TextToken, "NEW")),
	})

	result := ReplaceDescendant(doc, []int{0, 0}, NodeElement(newInner))
	if result == nil {
		t.Fatal("ReplaceDescendant returned nil")
	}
	if FullText(result) != "NEW" {
		t.Errorf("got %q, want %q", FullText(result), "NEW")
	}
}

func TestReplaceDescendantEmptyPath(t *testing.T) {
	doc := NewGreenNode(Document, nil)
	result := ReplaceDescendant(doc, nil, TokenElement(NewGreenToken(TextToken, "x")))
	if result != nil {
		t.Error("expected nil for empty path")
	}
}

func TestReplaceDescendantInvalidPath(t *testing.T) {
	doc := NewGreenNode(Document, []GreenElement{
		TokenElement(NewGreenToken(TextToken, "x")),
	})
	result := ReplaceDescendant(doc, []int{5}, TokenElement(NewGreenToken(TextToken, "y")))
	if result != nil {
		t.Error("expected nil for invalid path")
	}
}

// --- Structural sharing tests ---

func TestReplaceChildStructuralSharing(t *testing.T) {
	// Build a tree with multiple children. Replacing one child should
	// share all other children with the original.
	para1 := NewGreenNode(Paragraph, []GreenElement{
		TokenElement(NewGreenToken(TextToken, "aaa")),
	})
	para2 := NewGreenNode(Paragraph, []GreenElement{
		TokenElement(NewGreenToken(TextToken, "bbb")),
	})
	para3 := NewGreenNode(Paragraph, []GreenElement{
		TokenElement(NewGreenToken(TextToken, "ccc")),
	})
	doc := NewGreenNode(Document, []GreenElement{
		NodeElement(para1),
		NodeElement(para2),
		NodeElement(para3),
	})

	newPara := NewGreenNode(Paragraph, []GreenElement{
		TokenElement(NewGreenToken(TextToken, "XXX")),
	})
	result := ReplaceDescendant(doc, []int{1}, NodeElement(newPara))
	if result == nil {
		t.Fatal("nil result")
	}

	// Children 0 and 2 should be structurally shared (same GreenNode pointer).
	if result.Children[0].Node != para1 {
		t.Error("child 0 not shared")
	}
	if result.Children[2].Node != para3 {
		t.Error("child 2 not shared")
	}
	// Child 1 should be different.
	if result.Children[1].Node == para2 {
		t.Error("child 1 should be different")
	}
}

// --- FindPath tests ---

func TestFindPath(t *testing.T) {
	heading := NewGreenNode(ATXHeading, []GreenElement{
		TokenElement(NewGreenToken(HashToken, "#")),
		TokenElement(NewGreenToken(HeadingTextToken, "title")),
	})
	para := NewGreenNode(Paragraph, []GreenElement{
		TokenElement(NewGreenToken(TextToken, "text")),
	})
	doc := NewGreenNode(Document, []GreenElement{
		NodeElement(heading),
		NodeElement(para),
	})

	path := FindPath(doc, func(n *GreenNode) bool {
		return n.Kind == Paragraph
	})
	if len(path) != 1 || path[0] != 1 {
		t.Errorf("FindPath = %v, want [1]", path)
	}
}

func TestFindPathDeep(t *testing.T) {
	inner := NewGreenNode(Paragraph, []GreenElement{
		TokenElement(NewGreenToken(TextToken, "deep")),
	})
	bq := NewGreenNode(BlockQuote, []GreenElement{
		NodeElement(inner),
	})
	doc := NewGreenNode(Document, []GreenElement{
		NodeElement(bq),
	})

	path := FindPath(doc, func(n *GreenNode) bool {
		return n.Kind == Paragraph
	})
	if len(path) != 2 || path[0] != 0 || path[1] != 0 {
		t.Errorf("FindPath = %v, want [0, 0]", path)
	}
}

func TestFindPathNotFound(t *testing.T) {
	doc := NewGreenNode(Document, []GreenElement{
		TokenElement(NewGreenToken(TextToken, "x")),
	})

	path := FindPath(doc, func(n *GreenNode) bool {
		return n.Kind == Paragraph
	})
	if path != nil {
		t.Errorf("FindPath = %v, want nil", path)
	}
}

// --- Width consistency tests ---

func TestReplaceChildWidthConsistency(t *testing.T) {
	para := NewGreenNode(Paragraph, []GreenElement{
		TokenElement(NewGreenToken(TextToken, "hello")),
	})
	doc := NewGreenNode(Document, []GreenElement{
		NodeElement(para),
	})

	newTok := TokenElement(NewGreenToken(TextToken, "hi"))
	result := ReplaceChild(doc, []int{0}, 0, newTok)
	if result == nil {
		t.Fatal("nil result")
	}

	// Width should equal sum of children.
	if result.Width != 2 {
		t.Errorf("root width = %d, want 2", result.Width)
	}
	if result.Children[0].Node.Width != 2 {
		t.Errorf("para width = %d, want 2", result.Children[0].Node.Width)
	}
}
