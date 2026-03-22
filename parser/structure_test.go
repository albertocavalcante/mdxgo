package parser

import (
	"testing"

	"github.com/albertocavalcante/mdxgo/syntax"
)

// Structural tests verify that the parser produces the correct tree shape,
// not just that it round-trips. These complement the round-trip invariant
// by asserting semantic correctness.

func TestStructureEmptyDocument(t *testing.T) {
	green := Parse([]byte(""), Options{})
	if green.Kind != syntax.Document {
		t.Fatalf("kind = %v, want Document", green.Kind)
	}
	if len(green.Children) != 0 {
		t.Errorf("empty input should produce 0 children, got %d", len(green.Children))
	}
	if green.Width != 0 {
		t.Errorf("width = %d, want 0", green.Width)
	}
}

func TestStructureSingleParagraph(t *testing.T) {
	green := Parse([]byte("hello world\n"), Options{})
	assertChildKinds(t, green, syntax.Paragraph)
}

func TestStructureMultipleParagraphs(t *testing.T) {
	green := Parse([]byte("para one\n\npara two\n"), Options{})
	assertChildKinds(t, green, syntax.Paragraph, syntax.BlankLineNode, syntax.Paragraph)
}

func TestStructureATXHeading(t *testing.T) {
	green := Parse([]byte("# Title\n"), Options{})
	assertChildKinds(t, green, syntax.ATXHeading)

	heading := green.Children[0].Node
	// Should contain: HashToken, HeadingTextToken, NewLineToken
	hasHash := false
	hasText := false
	for _, child := range heading.Children {
		if child.Token != nil {
			if child.Token.Kind == syntax.HashToken {
				hasHash = true
				if child.Token.Text != "#" {
					t.Errorf("hash text = %q, want %q", child.Token.Text, "#")
				}
			}
			if child.Token.Kind == syntax.HeadingTextToken {
				hasText = true
			}
		}
	}
	if !hasHash {
		t.Error("ATXHeading missing HashToken")
	}
	if !hasText {
		t.Error("ATXHeading missing HeadingTextToken")
	}
}

func TestStructureATXHeadingLevels(t *testing.T) {
	for level := 1; level <= 6; level++ {
		hashes := ""
		for i := 0; i < level; i++ {
			hashes += "#"
		}
		src := hashes + " Heading\n"
		green := Parse([]byte(src), Options{})
		assertChildKinds(t, green, syntax.ATXHeading)

		heading := green.Children[0].Node
		for _, child := range heading.Children {
			if child.Token != nil && child.Token.Kind == syntax.HashToken {
				if child.Token.Text != hashes {
					t.Errorf("level %d: hash text = %q, want %q", level, child.Token.Text, hashes)
				}
			}
		}
	}
}

func TestStructureThematicBreak(t *testing.T) {
	green := Parse([]byte("---\n"), Options{})
	assertChildKinds(t, green, syntax.ThematicBreak)
}

func TestStructureFencedCodeBlock(t *testing.T) {
	green := Parse([]byte("```go\ncode\n```\n"), Options{})
	assertChildKinds(t, green, syntax.FencedCodeBlock)

	fenced := green.Children[0].Node
	hasFenceOpen := false
	hasFenceClose := false
	hasInfo := false
	hasCode := false
	for _, child := range fenced.Children {
		if child.Token != nil {
			switch child.Token.Kind {
			case syntax.FenceOpenToken:
				hasFenceOpen = true
				if child.Token.Text != "```" {
					t.Errorf("fence open text = %q, want %q", child.Token.Text, "```")
				}
			case syntax.FenceCloseToken:
				hasFenceClose = true
			case syntax.InfoStringToken:
				hasInfo = true
				if child.Token.Text != "go" {
					t.Errorf("info string = %q, want %q", child.Token.Text, "go")
				}
			case syntax.CodeLineToken:
				hasCode = true
			}
		}
	}
	if !hasFenceOpen {
		t.Error("FencedCodeBlock missing FenceOpenToken")
	}
	if !hasFenceClose {
		t.Error("FencedCodeBlock missing FenceCloseToken")
	}
	if !hasInfo {
		t.Error("FencedCodeBlock missing InfoStringToken")
	}
	if !hasCode {
		t.Error("FencedCodeBlock missing CodeLineToken")
	}
}

func TestStructureFencedCodeUnclosed(t *testing.T) {
	green := Parse([]byte("```\nunclosed code\n"), Options{})
	assertChildKinds(t, green, syntax.FencedCodeBlock)

	fenced := green.Children[0].Node
	hasFenceClose := false
	for _, child := range fenced.Children {
		if child.Token != nil && child.Token.Kind == syntax.FenceCloseToken {
			hasFenceClose = true
		}
	}
	if hasFenceClose {
		t.Error("unclosed FencedCodeBlock should not have FenceCloseToken")
	}
}

func TestStructureIndentedCode(t *testing.T) {
	green := Parse([]byte("    code\n"), Options{MDX: false})
	assertChildKinds(t, green, syntax.IndentedCodeBlock)
}

func TestStructureIndentedCodeDisabledInMDX(t *testing.T) {
	green := Parse([]byte("    not code in mdx\n"), Options{MDX: true})
	// In MDX mode, 4-space indent should be a paragraph, not code.
	assertChildKinds(t, green, syntax.Paragraph)
}

func TestStructureBlockQuote(t *testing.T) {
	green := Parse([]byte("> quote\n"), Options{})
	assertChildKinds(t, green, syntax.BlockQuote)

	bq := green.Children[0].Node
	hasMarker := false
	for _, child := range bq.Children {
		if child.Token != nil && child.Token.Kind == syntax.BlockQuoteMarker {
			hasMarker = true
			if child.Token.Text != ">" {
				t.Errorf("marker text = %q, want %q", child.Token.Text, ">")
			}
		}
	}
	if !hasMarker {
		t.Error("BlockQuote missing BlockQuoteMarker")
	}
}

func TestStructureBulletList(t *testing.T) {
	green := Parse([]byte("- a\n- b\n"), Options{})
	assertChildKinds(t, green, syntax.BulletList)

	list := green.Children[0].Node
	itemCount := 0
	for _, child := range list.Children {
		if child.Node != nil && child.Node.Kind == syntax.ListItem {
			itemCount++
		}
	}
	if itemCount != 2 {
		t.Errorf("bullet list has %d items, want 2", itemCount)
	}
}

func TestStructureOrderedList(t *testing.T) {
	green := Parse([]byte("1. a\n2. b\n"), Options{})
	assertChildKinds(t, green, syntax.OrderedList)

	list := green.Children[0].Node
	itemCount := 0
	for _, child := range list.Children {
		if child.Node != nil && child.Node.Kind == syntax.ListItem {
			itemCount++
		}
	}
	if itemCount != 2 {
		t.Errorf("ordered list has %d items, want 2", itemCount)
	}
}

func TestStructureListItemHasMarker(t *testing.T) {
	green := Parse([]byte("- item\n"), Options{})
	list := green.Children[0].Node
	item := list.Children[0].Node

	hasMarker := false
	for _, child := range item.Children {
		if child.Token != nil && child.Token.Kind == syntax.BulletMarker {
			hasMarker = true
		}
	}
	if !hasMarker {
		t.Error("ListItem missing BulletMarker")
	}
}

func TestStructureSetextHeading(t *testing.T) {
	green := Parse([]byte("Title\n=====\n"), Options{})
	assertChildKinds(t, green, syntax.SetextHeading)
}

func TestStructureBlankLine(t *testing.T) {
	green := Parse([]byte("\n"), Options{})
	assertChildKinds(t, green, syntax.BlankLineNode)
}

func TestStructureHTMLBlock(t *testing.T) {
	green := Parse([]byte("<div>\ncontent\n</div>\n"), Options{MDX: false})
	assertChildKinds(t, green, syntax.HTMLBlock)
}

func TestStructureHTMLBlockDisabledInMDX(t *testing.T) {
	// In MDX mode, HTML blocks are disabled — parsed as paragraphs.
	green := Parse([]byte("<div>\ncontent\n</div>\n"), Options{MDX: true})
	// Should not be HTMLBlock.
	for _, child := range green.Children {
		if child.Node != nil && child.Node.Kind == syntax.HTMLBlock {
			t.Error("HTML blocks should be disabled in MDX mode")
		}
	}
}

func TestStructureFrontmatter(t *testing.T) {
	green := Parse([]byte("---\ntitle: test\n---\n"), Options{MDX: true})
	assertChildKinds(t, green, syntax.Frontmatter)

	fm := green.Children[0].Node
	fenceCount := 0
	contentCount := 0
	for _, child := range fm.Children {
		if child.Token != nil {
			switch child.Token.Kind {
			case syntax.FrontmatterFence:
				fenceCount++
			case syntax.FrontmatterLine:
				contentCount++
			}
		}
	}
	if fenceCount != 2 {
		t.Errorf("frontmatter has %d fences, want 2", fenceCount)
	}
	if contentCount != 1 {
		t.Errorf("frontmatter has %d content lines, want 1", contentCount)
	}
}

func TestStructureFrontmatterNotInCommonMark(t *testing.T) {
	green := Parse([]byte("---\ntitle: test\n---\n"), Options{MDX: false})
	// In CommonMark mode, --- is a thematic break, not frontmatter.
	for _, child := range green.Children {
		if child.Node != nil && child.Node.Kind == syntax.Frontmatter {
			t.Error("frontmatter should not be parsed in CommonMark mode")
		}
	}
}

func TestStructureESM(t *testing.T) {
	green := Parse([]byte("import Foo from 'foo'\n"), Options{MDX: true})
	assertChildKinds(t, green, syntax.ESMDeclaration)
}

func TestStructureESMNotInCommonMark(t *testing.T) {
	green := Parse([]byte("import Foo from 'foo'\n"), Options{MDX: false})
	// In CommonMark mode, import is a paragraph.
	assertChildKinds(t, green, syntax.Paragraph)
}

func TestStructureRedTreeOffsets(t *testing.T) {
	src := "# A\n\n## B\n"
	green := Parse([]byte(src), Options{})
	root := syntax.NewSyntaxRoot(green)

	if root.Offset() != 0 {
		t.Errorf("root offset = %d, want 0", root.Offset())
	}
	if root.Width() != len(src) {
		t.Errorf("root width = %d, want %d", root.Width(), len(src))
	}
	if root.End() != len(src) {
		t.Errorf("root end = %d, want %d", root.End(), len(src))
	}

	// First child (heading "# A\n") starts at offset 0.
	first := root.ChildAt(0)
	if first.Offset() != 0 {
		t.Errorf("first child offset = %d, want 0", first.Offset())
	}

	// Second child (blank line "\n") starts at offset 4.
	second := root.ChildAt(1)
	if second.Offset() != 4 {
		t.Errorf("second child offset = %d, want 4", second.Offset())
	}

	// Third child (heading "## B\n") starts at offset 5.
	third := root.ChildAt(2)
	if third.Offset() != 5 {
		t.Errorf("third child offset = %d, want 5", third.Offset())
	}
}

func TestStructureRedTreeParentPointers(t *testing.T) {
	green := Parse([]byte("# A\n"), Options{})
	root := syntax.NewSyntaxRoot(green)

	if root.Parent() != nil {
		t.Error("root should have nil parent")
	}

	heading := root.ChildAt(0)
	if heading.IsNode() {
		if heading.Node.Parent() != root {
			t.Error("heading parent should be root")
		}
	}
}

func TestStructureWidthEqualsInputLength(t *testing.T) {
	cases := []string{
		"",
		"a",
		"hello\n",
		"# Title\n\nParagraph\n",
		"```\ncode\n```\n",
		"- a\n- b\n",
		"> q\n",
		"---\n",
		"line\r\n",
		"mixed\r\nlines\nhere\r",
	}
	for _, src := range cases {
		green := Parse([]byte(src), Options{})
		if green.Width != len(src) {
			t.Errorf("width %d != len %d for %q", green.Width, len(src), src)
		}
	}
}

// assertChildKinds checks that a node has exactly the specified child kinds.
func assertChildKinds(t *testing.T, n *syntax.GreenNode, expected ...syntax.SyntaxKind) {
	t.Helper()
	got := childKinds(n)
	if len(got) != len(expected) {
		t.Fatalf("got %d children %v, want %d %v", len(got), got, len(expected), expected)
	}
	for i, k := range expected {
		if got[i] != k {
			t.Errorf("child[%d] = %v, want %v", i, got[i], k)
		}
	}
}
