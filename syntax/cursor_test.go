package syntax

import "testing"

func TestCursorNavigation(t *testing.T) {
	doc := buildTestTree()
	c := NewCursor(doc)

	// At root.
	if c.Kind() != Document {
		t.Errorf("initial kind = %v, want Document", c.Kind())
	}
	if c.Depth() != 0 {
		t.Errorf("initial depth = %d, want 0", c.Depth())
	}

	// Move to first child (ATXHeading).
	if !c.FirstChild() {
		t.Fatal("FirstChild failed")
	}
	if c.Kind() != ATXHeading {
		t.Errorf("kind = %v, want ATXHeading", c.Kind())
	}
	if c.Depth() != 1 {
		t.Errorf("depth = %d, want 1", c.Depth())
	}

	// Move to next sibling (Paragraph).
	if !c.NextSibling() {
		t.Fatal("NextSibling failed")
	}
	if c.Kind() != Paragraph {
		t.Errorf("kind = %v, want Paragraph", c.Kind())
	}

	// No more siblings.
	if c.NextSibling() {
		t.Error("expected NextSibling to return false")
	}

	// Move back to previous sibling.
	if !c.PrevSibling() {
		t.Fatal("PrevSibling failed")
	}
	if c.Kind() != ATXHeading {
		t.Errorf("kind = %v, want ATXHeading", c.Kind())
	}

	// Move back to parent.
	if !c.Parent() {
		t.Fatal("Parent failed")
	}
	if c.Kind() != Document {
		t.Errorf("kind = %v, want Document", c.Kind())
	}

	// Parent at root should fail.
	if c.Parent() {
		t.Error("Parent at root should return false")
	}
}

func TestCursorFirstChildToken(t *testing.T) {
	doc := buildTestTree()
	c := NewCursor(doc)

	// Navigate to ATXHeading -> first child (HashToken).
	c.FirstChild()
	if !c.FirstChild() {
		t.Fatal("FirstChild on ATXHeading failed")
	}
	if c.Kind() != HashToken {
		t.Errorf("kind = %v, want HashToken", c.Kind())
	}
	if !c.IsToken() {
		t.Error("expected IsToken = true")
	}
	if c.Token().Text != "#" {
		t.Errorf("token text = %q, want %q", c.Token().Text, "#")
	}

	// FirstChild on a token should fail.
	if c.FirstChild() {
		t.Error("FirstChild on token should return false")
	}
}

func TestCursorLastChild(t *testing.T) {
	doc := buildTestTree()
	c := NewCursor(doc)

	// Navigate to last child (Paragraph).
	if !c.LastChild() {
		t.Fatal("LastChild failed")
	}
	if c.Kind() != Paragraph {
		t.Errorf("kind = %v, want Paragraph", c.Kind())
	}

	// Navigate to last child of Paragraph (second TextToken).
	if !c.LastChild() {
		t.Fatal("LastChild on Paragraph failed")
	}
	if c.Kind() != TextToken {
		t.Errorf("kind = %v, want TextToken", c.Kind())
	}
	if c.Token().Text != " world" {
		t.Errorf("token text = %q, want %q", c.Token().Text, " world")
	}
}

func TestCursorGotoChild(t *testing.T) {
	doc := buildTestTree()
	c := NewCursor(doc)

	// GotoChild(1) should go to Paragraph.
	if !c.GotoChild(1) {
		t.Fatal("GotoChild(1) failed")
	}
	if c.Kind() != Paragraph {
		t.Errorf("kind = %v, want Paragraph", c.Kind())
	}

	// Invalid index.
	c.Parent()
	if c.GotoChild(5) {
		t.Error("GotoChild(5) should return false")
	}
}

func TestCursorReset(t *testing.T) {
	doc := buildTestTree()
	c := NewCursor(doc)

	c.FirstChild()
	c.FirstChild()

	c.Reset(doc)
	if c.Kind() != Document {
		t.Errorf("kind after reset = %v, want Document", c.Kind())
	}
	if c.Depth() != 0 {
		t.Errorf("depth after reset = %d, want 0", c.Depth())
	}
}

func TestCursorEmptyNode(t *testing.T) {
	doc := NewGreenNode(Document, nil)
	c := NewCursor(doc)

	if c.FirstChild() {
		t.Error("FirstChild on empty node should return false")
	}
	if c.LastChild() {
		t.Error("LastChild on empty node should return false")
	}
}

func TestCursorPrevSiblingAtFirst(t *testing.T) {
	doc := buildTestTree()
	c := NewCursor(doc)
	c.FirstChild() // ATXHeading (first child)

	if c.PrevSibling() {
		t.Error("PrevSibling at first child should return false")
	}
}
