package syntax

import "testing"

// ---------------------------------------------------------------------------
// GreenElement
// ---------------------------------------------------------------------------

func TestGreenElementToken(t *testing.T) {
	tok := NewGreenToken(TextToken, "hello")
	elem := TokenElement(tok)

	if !elem.IsToken() {
		t.Fatal("expected IsToken() == true")
	}
	if elem.IsNode() {
		t.Fatal("expected IsNode() == false")
	}
	if elem.Kind() != TextToken {
		t.Errorf("Kind() = %v, want TextToken", elem.Kind())
	}
	if elem.Width() != 5 {
		t.Errorf("Width() = %d, want 5", elem.Width())
	}
}

func TestGreenElementNode(t *testing.T) {
	node := NewGreenNode(Paragraph, nil)
	elem := NodeElement(node)

	if !elem.IsNode() {
		t.Fatal("expected IsNode() == true")
	}
	if elem.IsToken() {
		t.Fatal("expected IsToken() == false")
	}
	if elem.Kind() != Paragraph {
		t.Errorf("Kind() = %v, want Paragraph", elem.Kind())
	}
	if elem.Width() != 0 {
		t.Errorf("Width() = %d, want 0", elem.Width())
	}
}

// ---------------------------------------------------------------------------
// GreenToken
// ---------------------------------------------------------------------------

func TestGreenTokenWidthEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		token     *GreenToken
		wantWidth int
	}{
		{
			name:      "empty text",
			token:     NewGreenToken(TextToken, ""),
			wantWidth: 0,
		},
		{
			name: "with leading trivia only",
			token: NewGreenTokenTrivia(
				TextToken,
				NewTriviaList(Trivia{Kind: WhitespaceTrivia, Text: "  "}),
				"hi",
				TriviaList{},
			),
			wantWidth: 4,
		},
		{
			name: "with trailing trivia only",
			token: NewGreenTokenTrivia(
				TextToken,
				TriviaList{},
				"hi",
				NewTriviaList(Trivia{Kind: WhitespaceTrivia, Text: " "}),
			),
			wantWidth: 3,
		},
		{
			name: "with both trivia",
			token: NewGreenTokenTrivia(
				HashToken,
				NewTriviaList(Trivia{Kind: EndOfLineTrivia, Text: "\n"}),
				"#",
				NewTriviaList(Trivia{Kind: WhitespaceTrivia, Text: " "}),
			),
			wantWidth: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.token.Width(); got != tt.wantWidth {
				t.Errorf("Width() = %d, want %d", got, tt.wantWidth)
			}
		})
	}
}

func TestGreenTokenFullTextEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		token *GreenToken
		want  string
	}{
		{
			name:  "no trivia",
			token: NewGreenToken(TextToken, "hello"),
			want:  "hello",
		},
		{
			name:  "empty",
			token: NewGreenToken(TextToken, ""),
			want:  "",
		},
		{
			name: "leading trivia only",
			token: NewGreenTokenTrivia(
				TextToken,
				NewTriviaList(Trivia{Kind: WhitespaceTrivia, Text: "  "}),
				"x",
				TriviaList{},
			),
			want: "  x",
		},
		{
			name: "trailing trivia only",
			token: NewGreenTokenTrivia(
				TextToken,
				TriviaList{},
				"x",
				NewTriviaList(Trivia{Kind: WhitespaceTrivia, Text: " "}),
			),
			want: "x ",
		},
		{
			name: "both trivia",
			token: NewGreenTokenTrivia(
				HashToken,
				NewTriviaList(Trivia{Kind: EndOfLineTrivia, Text: "\n"}),
				"#",
				NewTriviaList(Trivia{Kind: WhitespaceTrivia, Text: " "}),
			),
			want: "\n# ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.token.FullText(); got != tt.want {
				t.Errorf("FullText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGreenTokenWithLeadingTrivia(t *testing.T) {
	orig := NewGreenTokenTrivia(
		HashToken,
		NewTriviaList(Trivia{Kind: WhitespaceTrivia, Text: " "}),
		"#",
		NewTriviaList(Trivia{Kind: WhitespaceTrivia, Text: "  "}),
	)

	newTrivia := NewTriviaList(Trivia{Kind: EndOfLineTrivia, Text: "\n\n"})
	updated := orig.WithLeadingTrivia(newTrivia)

	// Original unchanged.
	if orig.LeadingTrivia.Width() != 1 {
		t.Error("original leading trivia was mutated")
	}

	if updated.LeadingTrivia.Width() != 2 {
		t.Errorf("new leading trivia width = %d, want 2", updated.LeadingTrivia.Width())
	}
	if updated.Text != "#" {
		t.Errorf("text changed: %q", updated.Text)
	}
	if updated.TrailingTrivia.Width() != 2 {
		t.Errorf("trailing trivia changed: width = %d", updated.TrailingTrivia.Width())
	}
	if updated.Kind != HashToken {
		t.Errorf("kind changed: %v", updated.Kind)
	}
}

func TestGreenTokenWithTrailingTrivia(t *testing.T) {
	orig := NewGreenToken(TextToken, "hi")
	newTrivia := NewTriviaList(Trivia{Kind: WhitespaceTrivia, Text: "\t"})
	updated := orig.WithTrailingTrivia(newTrivia)

	if orig.TrailingTrivia.Width() != 0 {
		t.Error("original trailing trivia was mutated")
	}
	if updated.TrailingTrivia.Width() != 1 {
		t.Errorf("new trailing trivia width = %d, want 1", updated.TrailingTrivia.Width())
	}
	if updated.Text != "hi" {
		t.Errorf("text changed: %q", updated.Text)
	}
}

func TestGreenTokenWithText(t *testing.T) {
	orig := NewGreenTokenTrivia(
		HashToken,
		NewTriviaList(Trivia{Kind: WhitespaceTrivia, Text: " "}),
		"#",
		NewTriviaList(Trivia{Kind: WhitespaceTrivia, Text: " "}),
	)
	updated := orig.WithText("##")

	if orig.Text != "#" {
		t.Error("original text was mutated")
	}
	if updated.Text != "##" {
		t.Errorf("updated text = %q, want %q", updated.Text, "##")
	}
	if updated.LeadingTrivia.Width() != 1 {
		t.Error("leading trivia not preserved")
	}
	if updated.TrailingTrivia.Width() != 1 {
		t.Error("trailing trivia not preserved")
	}
}

// ---------------------------------------------------------------------------
// GreenNode
// ---------------------------------------------------------------------------

func TestGreenNodeWidthComputed(t *testing.T) {
	tok1 := NewGreenToken(TextToken, "abc")
	tok2 := NewGreenToken(NewLineToken, "\n")
	node := NewGreenNode(Paragraph, []GreenElement{
		TokenElement(tok1),
		TokenElement(tok2),
	})

	if node.Width != 4 {
		t.Errorf("Width = %d, want 4", node.Width)
	}
	if node.ChildCount() != 2 {
		t.Errorf("ChildCount() = %d, want 2", node.ChildCount())
	}
}

func TestGreenNodeEmpty(t *testing.T) {
	node := NewGreenNode(Document, nil)
	if node.Width != 0 {
		t.Errorf("Width = %d, want 0", node.Width)
	}
	if node.ChildCount() != 0 {
		t.Errorf("ChildCount() = %d, want 0", node.ChildCount())
	}
}

func TestGreenNodeChildrenDefensiveCopy(t *testing.T) {
	elems := []GreenElement{
		TokenElement(NewGreenToken(TextToken, "a")),
	}
	node := NewGreenNode(Paragraph, elems)

	// Mutate the original slice — should not affect the node.
	elems[0] = TokenElement(NewGreenToken(TextToken, "z"))

	if node.Children[0].Token.Text != "a" {
		t.Error("GreenNode children were not defensively copied")
	}
}

func TestGreenNodeReplaceChildPreservesOther(t *testing.T) {
	tok1 := NewGreenToken(TextToken, "old")
	tok2 := NewGreenToken(NewLineToken, "\n")
	orig := NewGreenNode(Paragraph, []GreenElement{
		TokenElement(tok1),
		TokenElement(tok2),
	})

	replacement := NewGreenToken(TextToken, "brand-new")
	updated := orig.ReplaceChild(0, TokenElement(replacement))

	// Original unchanged.
	if orig.Children[0].Token.Text != "old" {
		t.Error("original was mutated")
	}
	if orig.Width != 4 {
		t.Errorf("original width changed: %d", orig.Width)
	}

	// Updated has new child and recomputed width.
	if updated.Children[0].Token.Text != "brand-new" {
		t.Errorf("replacement text = %q", updated.Children[0].Token.Text)
	}
	if updated.Width != 10 { // "brand-new" + "\n"
		t.Errorf("updated width = %d, want 10", updated.Width)
	}
	if updated.ChildCount() != 2 {
		t.Errorf("child count changed: %d", updated.ChildCount())
	}
}

func TestGreenNodeAppendChild(t *testing.T) {
	tok := NewGreenToken(TextToken, "a")
	orig := NewGreenNode(Paragraph, []GreenElement{
		TokenElement(tok),
	})

	extra := NewGreenToken(NewLineToken, "\n")
	updated := orig.AppendChild(TokenElement(extra))

	if orig.ChildCount() != 1 {
		t.Error("original child count changed")
	}
	if updated.ChildCount() != 2 {
		t.Errorf("updated child count = %d, want 2", updated.ChildCount())
	}
	if updated.Width != 2 {
		t.Errorf("updated width = %d, want 2", updated.Width)
	}
}

// ---------------------------------------------------------------------------
// Trivia & TriviaList
// ---------------------------------------------------------------------------

func TestTriviaWidth(t *testing.T) {
	tr := Trivia{Kind: WhitespaceTrivia, Text: "   "}
	if tr.Width() != 3 {
		t.Errorf("Width() = %d, want 3", tr.Width())
	}
}

func TestTriviaListEmpty(t *testing.T) {
	tl := TriviaList{}
	if !tl.IsEmpty() {
		t.Error("expected empty")
	}
	if tl.Len() != 0 {
		t.Errorf("Len() = %d", tl.Len())
	}
	if tl.Width() != 0 {
		t.Errorf("Width() = %d", tl.Width())
	}
	if tl.Text() != "" {
		t.Errorf("Text() = %q", tl.Text())
	}
}

func TestTriviaListSingle(t *testing.T) {
	tl := NewTriviaList(Trivia{Kind: WhitespaceTrivia, Text: " "})
	if tl.IsEmpty() {
		t.Error("expected non-empty")
	}
	if tl.Len() != 1 {
		t.Errorf("Len() = %d", tl.Len())
	}
	if tl.Width() != 1 {
		t.Errorf("Width() = %d", tl.Width())
	}
	if tl.Text() != " " {
		t.Errorf("Text() = %q", tl.Text())
	}
	if tl.At(0).Kind != WhitespaceTrivia {
		t.Errorf("At(0).Kind = %v", tl.At(0).Kind)
	}
}

func TestTriviaListMultiple(t *testing.T) {
	tl := NewTriviaList(
		Trivia{Kind: WhitespaceTrivia, Text: "  "},
		Trivia{Kind: EndOfLineTrivia, Text: "\n"},
		Trivia{Kind: WhitespaceTrivia, Text: "\t"},
	)
	if tl.Len() != 3 {
		t.Errorf("Len() = %d", tl.Len())
	}
	if tl.Width() != 4 {
		t.Errorf("Width() = %d, want 4", tl.Width())
	}
	if tl.Text() != "  \n\t" {
		t.Errorf("Text() = %q", tl.Text())
	}
}

func TestTriviaListAppend(t *testing.T) {
	tl := NewTriviaList(Trivia{Kind: WhitespaceTrivia, Text: " "})
	tl2 := tl.Append(Trivia{Kind: EndOfLineTrivia, Text: "\n"})

	// Original unchanged.
	if tl.Len() != 1 {
		t.Error("original was mutated")
	}
	if tl2.Len() != 2 {
		t.Errorf("appended Len() = %d, want 2", tl2.Len())
	}
	if tl2.Text() != " \n" {
		t.Errorf("appended Text() = %q", tl2.Text())
	}
}

func TestTriviaListDefensiveCopy(t *testing.T) {
	pieces := []Trivia{{Kind: WhitespaceTrivia, Text: "a"}}
	tl := NewTriviaList(pieces...)

	// Mutate original slice.
	pieces[0] = Trivia{Kind: WhitespaceTrivia, Text: "z"}

	if tl.At(0).Text != "a" {
		t.Error("TriviaList was not defensively copied")
	}
}

// ---------------------------------------------------------------------------
// SyntaxKind
// ---------------------------------------------------------------------------

func TestSyntaxKindClassification(t *testing.T) {
	tests := []struct {
		kind     SyntaxKind
		isTrivia bool
		isToken  bool
		isNode   bool
	}{
		{WhitespaceTrivia, true, false, false},
		{EndOfLineTrivia, true, false, false},
		{LineCommentTrivia, true, false, false},
		{TextToken, false, true, false},
		{HashToken, false, true, false},
		{NewLineToken, false, true, false},
		{Document, false, false, true},
		{Paragraph, false, false, true},
		{ATXHeading, false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.kind.String(), func(t *testing.T) {
			if got := tt.kind.IsTrivia(); got != tt.isTrivia {
				t.Errorf("IsTrivia() = %v, want %v", got, tt.isTrivia)
			}
			if got := tt.kind.IsToken(); got != tt.isToken {
				t.Errorf("IsToken() = %v, want %v", got, tt.isToken)
			}
			if got := tt.kind.IsNode(); got != tt.isNode {
				t.Errorf("IsNode() = %v, want %v", got, tt.isNode)
			}
		})
	}
}

func TestSyntaxKindStringKnown(t *testing.T) {
	if s := Document.String(); s != "Document" {
		t.Errorf("Document.String() = %q", s)
	}
	if s := TextToken.String(); s != "TextToken" {
		t.Errorf("TextToken.String() = %q", s)
	}
}

func TestSyntaxKindStringUnknown(t *testing.T) {
	unknown := SyntaxKind(9999)
	if s := unknown.String(); s != "Unknown" {
		t.Errorf("unknown kind String() = %q, want %q", s, "Unknown")
	}
}

// ---------------------------------------------------------------------------
// FullText / FullTextNode
// ---------------------------------------------------------------------------

func TestFullTextEmpty(t *testing.T) {
	doc := NewGreenNode(Document, nil)
	if got := FullText(doc); got != "" {
		t.Errorf("FullText(empty) = %q", got)
	}
}

func TestFullTextSingleToken(t *testing.T) {
	doc := NewGreenNode(Document, []GreenElement{
		TokenElement(NewGreenToken(TextToken, "hello")),
	})
	if got := FullText(doc); got != "hello" {
		t.Errorf("FullText = %q, want %q", got, "hello")
	}
}

func TestFullTextWithTrivia(t *testing.T) {
	tok := NewGreenTokenTrivia(
		HashToken,
		NewTriviaList(Trivia{Kind: EndOfLineTrivia, Text: "\n"}),
		"#",
		NewTriviaList(Trivia{Kind: WhitespaceTrivia, Text: " "}),
	)
	doc := NewGreenNode(Document, []GreenElement{
		TokenElement(tok),
		TokenElement(NewGreenToken(HeadingTextToken, "Hi")),
	})
	want := "\n# Hi"
	if got := FullText(doc); got != want {
		t.Errorf("FullText = %q, want %q", got, want)
	}
}

func TestFullTextNested(t *testing.T) {
	inner := NewGreenNode(Paragraph, []GreenElement{
		TokenElement(NewGreenToken(TextToken, "text\n")),
	})
	doc := NewGreenNode(Document, []GreenElement{
		NodeElement(inner),
	})
	if got := FullText(doc); got != "text\n" {
		t.Errorf("FullText = %q", got)
	}
}

func TestFullTextNode(t *testing.T) {
	tok := NewGreenToken(TextToken, "abc")
	doc := NewGreenNode(Document, []GreenElement{TokenElement(tok)})
	root := NewSyntaxRoot(doc)
	if got := FullTextNode(root); got != "abc" {
		t.Errorf("FullTextNode = %q", got)
	}
}

// ---------------------------------------------------------------------------
// DebugDump
// ---------------------------------------------------------------------------

func TestDebugDumpEmpty(t *testing.T) {
	doc := NewGreenNode(Document, nil)
	got := DebugDump(doc)
	want := "Document [0]\n"
	if got != want {
		t.Errorf("DebugDump(empty) = %q, want %q", got, want)
	}
}

func TestDebugDumpWithTrivia(t *testing.T) {
	tok := NewGreenTokenTrivia(
		HashToken,
		NewTriviaList(Trivia{Kind: WhitespaceTrivia, Text: " "}),
		"#",
		NewTriviaList(Trivia{Kind: WhitespaceTrivia, Text: "  "}),
	)
	heading := NewGreenNode(ATXHeading, []GreenElement{TokenElement(tok)})
	doc := NewGreenNode(Document, []GreenElement{NodeElement(heading)})
	got := DebugDump(doc)

	// Should contain "(lead=1, trail=2)" for the trivia-bearing token.
	if got == "" {
		t.Fatal("empty dump")
	}
	// Verify structure is present.
	wantSubstrings := []string{"Document", "ATXHeading", "HashToken", "lead=1", "trail=2"}
	for _, sub := range wantSubstrings {
		if !contains(got, sub) {
			t.Errorf("DebugDump missing %q in:\n%s", sub, got)
		}
	}
}

func TestDebugDumpLongText(t *testing.T) {
	long := "abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz"
	tok := NewGreenToken(TextToken, long)
	doc := NewGreenNode(Document, []GreenElement{TokenElement(tok)})
	got := DebugDump(doc)

	// Long text should be truncated with "...".
	if !contains(got, "...") {
		t.Errorf("expected truncation in dump:\n%s", got)
	}
}

// ---------------------------------------------------------------------------
// SyntaxNode (red tree)
// ---------------------------------------------------------------------------

func TestSyntaxNodeRoot(t *testing.T) {
	doc := NewGreenNode(Document, nil)
	root := NewSyntaxRoot(doc)

	if root.Kind() != Document {
		t.Errorf("Kind() = %v", root.Kind())
	}
	if root.Offset() != 0 {
		t.Errorf("Offset() = %d", root.Offset())
	}
	if root.Width() != 0 {
		t.Errorf("Width() = %d", root.Width())
	}
	if root.End() != 0 {
		t.Errorf("End() = %d", root.End())
	}
	if root.Parent() != nil {
		t.Error("root Parent() should be nil")
	}
	if root.Index() != 0 {
		t.Errorf("root Index() = %d", root.Index())
	}
	if root.ChildCount() != 0 {
		t.Errorf("ChildCount() = %d", root.ChildCount())
	}
	if root.Green() != doc {
		t.Error("Green() doesn't match")
	}
}

func TestSyntaxNodeChildrenIteration(t *testing.T) {
	tok1 := NewGreenToken(TextToken, "aaa")
	tok2 := NewGreenToken(NewLineToken, "\n")
	para := NewGreenNode(Paragraph, []GreenElement{
		TokenElement(tok1),
		TokenElement(tok2),
	})
	doc := NewGreenNode(Document, []GreenElement{
		NodeElement(para),
	})

	root := NewSyntaxRoot(doc)
	count := 0
	for i, child := range root.Children() {
		if i == 0 {
			if child.Kind() != Paragraph {
				t.Errorf("child 0 kind = %v", child.Kind())
			}
			if child.Offset() != 0 {
				t.Errorf("child 0 offset = %d", child.Offset())
			}
			if child.Width() != 4 {
				t.Errorf("child 0 width = %d", child.Width())
			}
			if !child.IsNode() {
				t.Error("child 0 should be a node")
			}
		}
		count++
	}
	if count != 1 {
		t.Errorf("iterated %d children, want 1", count)
	}
}

func TestSyntaxNodeChildAt(t *testing.T) {
	tok1 := NewGreenToken(TextToken, "abc")
	tok2 := NewGreenToken(TextToken, "de")
	doc := NewGreenNode(Document, []GreenElement{
		TokenElement(tok1),
		TokenElement(tok2),
	})
	root := NewSyntaxRoot(doc)

	child0 := root.ChildAt(0)
	if !child0.IsToken() {
		t.Fatal("child 0 should be token")
	}
	if child0.Token.Text() != "abc" {
		t.Errorf("child 0 text = %q", child0.Token.Text())
	}
	if child0.Token.Offset() != 0 {
		t.Errorf("child 0 offset = %d", child0.Token.Offset())
	}

	child1 := root.ChildAt(1)
	if child1.Token.Offset() != 3 {
		t.Errorf("child 1 offset = %d, want 3", child1.Token.Offset())
	}
}

func TestSyntaxNodeTokenAt(t *testing.T) {
	tok1 := NewGreenToken(HashToken, "# ")
	tok2 := NewGreenToken(HeadingTextToken, "Hi")
	tok3 := NewGreenToken(NewLineToken, "\n")
	heading := NewGreenNode(ATXHeading, []GreenElement{
		TokenElement(tok1),
		TokenElement(tok2),
		TokenElement(tok3),
	})
	doc := NewGreenNode(Document, []GreenElement{
		NodeElement(heading),
	})
	root := NewSyntaxRoot(doc)

	tests := []struct {
		offset   int
		wantKind SyntaxKind
		wantText string
	}{
		{0, HashToken, "# "},
		{1, HashToken, "# "},
		{2, HeadingTextToken, "Hi"},
		{3, HeadingTextToken, "Hi"},
		{4, NewLineToken, "\n"},
	}

	for _, tt := range tests {
		found := root.TokenAt(tt.offset)
		if found == nil {
			t.Errorf("TokenAt(%d) = nil", tt.offset)
			continue
		}
		if found.Kind() != tt.wantKind {
			t.Errorf("TokenAt(%d).Kind() = %v, want %v", tt.offset, found.Kind(), tt.wantKind)
		}
		if found.Text() != tt.wantText {
			t.Errorf("TokenAt(%d).Text() = %q, want %q", tt.offset, found.Text(), tt.wantText)
		}
	}
}

func TestSyntaxNodeTokenAtOutOfRange(t *testing.T) {
	tok := NewGreenToken(TextToken, "abc")
	doc := NewGreenNode(Document, []GreenElement{TokenElement(tok)})
	root := NewSyntaxRoot(doc)

	if root.TokenAt(-1) != nil {
		t.Error("TokenAt(-1) should be nil")
	}
	if root.TokenAt(3) != nil {
		t.Error("TokenAt(3) should be nil for width=3")
	}
	if root.TokenAt(100) != nil {
		t.Error("TokenAt(100) should be nil")
	}
}

func TestSyntaxNodeParentPointers(t *testing.T) {
	tok := NewGreenToken(TextToken, "x")
	para := NewGreenNode(Paragraph, []GreenElement{TokenElement(tok)})
	doc := NewGreenNode(Document, []GreenElement{NodeElement(para)})
	root := NewSyntaxRoot(doc)

	// Navigate to paragraph.
	paraElem := root.ChildAt(0)
	paraNode := paraElem.Node
	if paraNode.Parent() != root {
		t.Error("paragraph parent should be root")
	}
	if paraNode.Index() != 0 {
		t.Errorf("paragraph index = %d", paraNode.Index())
	}

	// Navigate to token inside paragraph.
	tokElem := paraNode.ChildAt(0)
	if tokElem.Token.Parent() != paraNode {
		t.Error("token parent should be paragraph node")
	}
	if tokElem.Token.Index() != 0 {
		t.Errorf("token index = %d", tokElem.Token.Index())
	}
}

// ---------------------------------------------------------------------------
// SyntaxToken (red tree)
// ---------------------------------------------------------------------------

func TestSyntaxTokenOffsets(t *testing.T) {
	lead := NewTriviaList(Trivia{Kind: WhitespaceTrivia, Text: "  "})
	trail := NewTriviaList(Trivia{Kind: WhitespaceTrivia, Text: " "})
	tok := NewGreenTokenTrivia(HashToken, lead, "#", trail)

	// Place token at offset 10 inside a document.
	prefix := NewGreenToken(TextToken, "0123456789") // 10 bytes
	doc := NewGreenNode(Document, []GreenElement{
		TokenElement(prefix),
		TokenElement(tok),
	})
	root := NewSyntaxRoot(doc)
	st := root.ChildAt(1).Token

	if st.Offset() != 10 {
		t.Errorf("Offset() = %d, want 10", st.Offset())
	}
	if st.TextOffset() != 12 { // 10 + 2 bytes leading trivia
		t.Errorf("TextOffset() = %d, want 12", st.TextOffset())
	}
	if st.TextEnd() != 13 { // 12 + 1 byte text "#"
		t.Errorf("TextEnd() = %d, want 13", st.TextEnd())
	}
	if st.End() != 14 { // 13 + 1 byte trailing trivia
		t.Errorf("End() = %d, want 14", st.End())
	}
}

func TestSyntaxTokenFullText(t *testing.T) {
	lead := NewTriviaList(Trivia{Kind: WhitespaceTrivia, Text: " "})
	trail := NewTriviaList(Trivia{Kind: EndOfLineTrivia, Text: "\n"})
	tok := NewGreenTokenTrivia(TextToken, lead, "hi", trail)
	doc := NewGreenNode(Document, []GreenElement{TokenElement(tok)})
	root := NewSyntaxRoot(doc)

	st := root.ChildAt(0).Token
	if st.FullText() != " hi\n" {
		t.Errorf("FullText() = %q", st.FullText())
	}
	if st.Text() != "hi" {
		t.Errorf("Text() = %q", st.Text())
	}
}

func TestSyntaxTokenTrivia(t *testing.T) {
	lead := NewTriviaList(Trivia{Kind: WhitespaceTrivia, Text: "  "})
	trail := NewTriviaList(Trivia{Kind: EndOfLineTrivia, Text: "\n"})
	tok := NewGreenTokenTrivia(TextToken, lead, "x", trail)
	doc := NewGreenNode(Document, []GreenElement{TokenElement(tok)})
	root := NewSyntaxRoot(doc)

	st := root.ChildAt(0).Token
	if st.LeadingTrivia().Width() != 2 {
		t.Errorf("LeadingTrivia().Width() = %d", st.LeadingTrivia().Width())
	}
	if st.TrailingTrivia().Width() != 1 {
		t.Errorf("TrailingTrivia().Width() = %d", st.TrailingTrivia().Width())
	}
}

func TestSyntaxTokenGreen(t *testing.T) {
	gt := NewGreenToken(TextToken, "x")
	doc := NewGreenNode(Document, []GreenElement{TokenElement(gt)})
	root := NewSyntaxRoot(doc)
	st := root.ChildAt(0).Token

	if st.Green() != gt {
		t.Error("Green() doesn't return the underlying green token")
	}
}

// ---------------------------------------------------------------------------
// SyntaxElement
// ---------------------------------------------------------------------------

func TestSyntaxElementNode(t *testing.T) {
	doc := NewGreenNode(Document, nil)
	root := NewSyntaxRoot(doc)
	elem := SyntaxElement{Node: root}

	if !elem.IsNode() {
		t.Error("expected IsNode()")
	}
	if elem.IsToken() {
		t.Error("expected !IsToken()")
	}
	if elem.Kind() != Document {
		t.Errorf("Kind() = %v", elem.Kind())
	}
	if elem.Offset() != 0 {
		t.Errorf("Offset() = %d", elem.Offset())
	}
	if elem.Width() != 0 {
		t.Errorf("Width() = %d", elem.Width())
	}
}

func TestSyntaxElementToken(t *testing.T) {
	gt := NewGreenToken(TextToken, "abc")
	doc := NewGreenNode(Document, []GreenElement{TokenElement(gt)})
	root := NewSyntaxRoot(doc)
	elem := root.ChildAt(0)

	if !elem.IsToken() {
		t.Error("expected IsToken()")
	}
	if elem.IsNode() {
		t.Error("expected !IsNode()")
	}
	if elem.Kind() != TextToken {
		t.Errorf("Kind() = %v", elem.Kind())
	}
	if elem.Width() != 3 {
		t.Errorf("Width() = %d", elem.Width())
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
