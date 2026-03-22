package syntax

import "testing"

func TestFindAll(t *testing.T) {
	doc := buildTestTree()
	results := FindAll(doc, func(n *GreenNode) bool {
		return n.Kind == Paragraph || n.Kind == ATXHeading
	})
	if len(results) != 2 {
		t.Errorf("FindAll returned %d results, want 2", len(results))
	}
}

func TestFindAllNoMatch(t *testing.T) {
	doc := buildTestTree()
	results := FindAll(doc, func(n *GreenNode) bool {
		return n.Kind == FencedCodeBlock
	})
	if len(results) != 0 {
		t.Errorf("FindAll returned %d results, want 0", len(results))
	}
}

func TestFindFirst(t *testing.T) {
	doc := buildTestTree()
	result := FindFirst(doc, func(n *GreenNode) bool {
		return n.Kind == Paragraph
	})
	if result == nil {
		t.Fatal("FindFirst returned nil")
	}
	if result.Kind != Paragraph {
		t.Errorf("kind = %v, want Paragraph", result.Kind)
	}
}

func TestFindFirstNoMatch(t *testing.T) {
	doc := buildTestTree()
	result := FindFirst(doc, func(n *GreenNode) bool {
		return n.Kind == FencedCodeBlock
	})
	if result != nil {
		t.Error("FindFirst should return nil when no match")
	}
}

func TestFindToken(t *testing.T) {
	doc := buildTestTree()
	tok := FindToken(doc, func(t *GreenToken) bool {
		return t.Kind == HashToken
	})
	if tok == nil {
		t.Fatal("FindToken returned nil")
	}
	if tok.Text != "#" {
		t.Errorf("token text = %q, want %q", tok.Text, "#")
	}
}

func TestFindAllTokens(t *testing.T) {
	doc := buildTestTree()
	tokens := FindAllTokens(doc, func(t *GreenToken) bool {
		return t.Kind == TextToken
	})
	if len(tokens) != 2 {
		t.Errorf("FindAllTokens returned %d tokens, want 2", len(tokens))
	}
}

func TestCountNodes(t *testing.T) {
	doc := buildTestTree()
	count := CountNodes(doc, func(n *GreenNode) bool {
		return true
	})
	// Document + ATXHeading + Paragraph = 3
	if count != 3 {
		t.Errorf("CountNodes = %d, want 3", count)
	}
}
