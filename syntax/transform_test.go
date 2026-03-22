package syntax

import "testing"

func TestInsertBefore(t *testing.T) {
	doc := NewGreenNode(Document, []GreenElement{
		TokenElement(NewGreenToken(TextToken, "aaa")),
		TokenElement(NewGreenToken(TextToken, "ccc")),
	})

	result := InsertBefore(doc, 1, TokenElement(NewGreenToken(TextToken, "bbb")))
	if result == nil {
		t.Fatal("InsertBefore returned nil")
	}
	if len(result.Children) != 3 {
		t.Fatalf("got %d children, want 3", len(result.Children))
	}
	if FullText(result) != "aaabbbccc" {
		t.Errorf("got %q, want %q", FullText(result), "aaabbbccc")
	}
	// Original unchanged.
	if FullText(doc) != "aaaccc" {
		t.Errorf("original changed: %q", FullText(doc))
	}
}

func TestInsertBeforeAtStart(t *testing.T) {
	doc := NewGreenNode(Document, []GreenElement{
		TokenElement(NewGreenToken(TextToken, "bbb")),
	})
	result := InsertBefore(doc, 0, TokenElement(NewGreenToken(TextToken, "aaa")))
	if FullText(result) != "aaabbb" {
		t.Errorf("got %q, want %q", FullText(result), "aaabbb")
	}
}

func TestInsertAfter(t *testing.T) {
	doc := NewGreenNode(Document, []GreenElement{
		TokenElement(NewGreenToken(TextToken, "aaa")),
		TokenElement(NewGreenToken(TextToken, "ccc")),
	})

	result := InsertAfter(doc, 0, TokenElement(NewGreenToken(TextToken, "bbb")))
	if result == nil {
		t.Fatal("InsertAfter returned nil")
	}
	if FullText(result) != "aaabbbccc" {
		t.Errorf("got %q, want %q", FullText(result), "aaabbbccc")
	}
}

func TestRemoveChild(t *testing.T) {
	doc := NewGreenNode(Document, []GreenElement{
		TokenElement(NewGreenToken(TextToken, "aaa")),
		TokenElement(NewGreenToken(TextToken, "bbb")),
		TokenElement(NewGreenToken(TextToken, "ccc")),
	})

	result := RemoveChild(doc, 1)
	if result == nil {
		t.Fatal("RemoveChild returned nil")
	}
	if len(result.Children) != 2 {
		t.Fatalf("got %d children, want 2", len(result.Children))
	}
	if FullText(result) != "aaaccc" {
		t.Errorf("got %q, want %q", FullText(result), "aaaccc")
	}
	// Original unchanged.
	if FullText(doc) != "aaabbbccc" {
		t.Errorf("original changed: %q", FullText(doc))
	}
}

func TestRemoveChildInvalid(t *testing.T) {
	doc := NewGreenNode(Document, []GreenElement{
		TokenElement(NewGreenToken(TextToken, "x")),
	})
	if RemoveChild(doc, 5) != nil {
		t.Error("expected nil for invalid index")
	}
	if RemoveChild(doc, -1) != nil {
		t.Error("expected nil for negative index")
	}
}

func TestTransformChildren(t *testing.T) {
	doc := NewGreenNode(Document, []GreenElement{
		TokenElement(NewGreenToken(TextToken, "aaa")),
		TokenElement(NewGreenToken(TextToken, "bbb")),
		TokenElement(NewGreenToken(TextToken, "ccc")),
	})

	// Upper-case all token text by replacing with new tokens.
	result := TransformChildren(doc, func(e GreenElement) *GreenElement {
		if e.Token != nil && e.Token.Text == "bbb" {
			newElem := TokenElement(NewGreenToken(TextToken, "BBB"))
			return &newElem
		}
		return &e
	})

	if FullText(result) != "aaaBBBccc" {
		t.Errorf("got %q, want %q", FullText(result), "aaaBBBccc")
	}
}

func TestTransformChildrenRemove(t *testing.T) {
	doc := NewGreenNode(Document, []GreenElement{
		TokenElement(NewGreenToken(TextToken, "aaa")),
		TokenElement(NewGreenToken(TextToken, "bbb")),
		TokenElement(NewGreenToken(TextToken, "ccc")),
	})

	// Remove "bbb".
	result := TransformChildren(doc, func(e GreenElement) *GreenElement {
		if e.Token != nil && e.Token.Text == "bbb" {
			return nil
		}
		return &e
	})

	if FullText(result) != "aaaccc" {
		t.Errorf("got %q, want %q", FullText(result), "aaaccc")
	}
}

func TestTransformChildrenUnchanged(t *testing.T) {
	doc := NewGreenNode(Document, []GreenElement{
		TokenElement(NewGreenToken(TextToken, "aaa")),
	})

	// No changes: should return same pointer.
	result := TransformChildren(doc, func(e GreenElement) *GreenElement {
		return &e
	})

	if result != doc {
		t.Error("expected same pointer when nothing changed")
	}
}

func TestMapNodes(t *testing.T) {
	// Build: Document { Paragraph { TextToken "hello" } }
	para := NewGreenNode(Paragraph, []GreenElement{
		TokenElement(NewGreenToken(TextToken, "hello")),
	})
	doc := NewGreenNode(Document, []GreenElement{
		NodeElement(para),
	})

	// Replace all Paragraph nodes with ATXHeading nodes (same children).
	result := MapNodes(doc, func(n *GreenNode) *GreenNode {
		if n.Kind == Paragraph {
			return NewGreenNode(ATXHeading, n.Children)
		}
		return n
	})

	// The document should now contain an ATXHeading instead of Paragraph.
	if result.Children[0].Node.Kind != ATXHeading {
		t.Errorf("kind = %v, want ATXHeading", result.Children[0].Node.Kind)
	}
	if FullText(result) != "hello" {
		t.Errorf("got %q, want %q", FullText(result), "hello")
	}
}

func TestInsertBeforeInvalidIndex(t *testing.T) {
	doc := NewGreenNode(Document, []GreenElement{
		TokenElement(NewGreenToken(TextToken, "x")),
	})
	if InsertBefore(doc, -1, TokenElement(NewGreenToken(TextToken, "y"))) != nil {
		t.Error("expected nil for negative index")
	}
	if InsertBefore(doc, 5, TokenElement(NewGreenToken(TextToken, "y"))) != nil {
		t.Error("expected nil for out-of-range index")
	}
}

func TestInsertAfterInvalidIndex(t *testing.T) {
	doc := NewGreenNode(Document, []GreenElement{
		TokenElement(NewGreenToken(TextToken, "x")),
	})
	if InsertAfter(doc, -2, TokenElement(NewGreenToken(TextToken, "y"))) != nil {
		t.Error("expected nil for invalid index")
	}
	if InsertAfter(doc, 5, TokenElement(NewGreenToken(TextToken, "y"))) != nil {
		t.Error("expected nil for out-of-range index")
	}
}
