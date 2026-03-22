package parser

import (
	"fmt"
	"testing"

	"github.com/albertocavalcante/mdxgo/syntax"
)

// --- Code span tests ---

func TestInlineCodeSpan(t *testing.T) {
	src := "`code`\n"
	green := Parse([]byte(src), Options{})
	assertRoundTrip(t, src, Options{})

	para := green.Children[0].Node
	assertHasInlineNode(t, para, syntax.CodeSpan)
}

func TestInlineCodeSpanDouble(t *testing.T) {
	src := "``code span``\n"
	green := Parse([]byte(src), Options{})
	assertRoundTrip(t, src, Options{})

	para := green.Children[0].Node
	assertHasInlineNode(t, para, syntax.CodeSpan)
}

func TestInlineCodeSpanWithBacktick(t *testing.T) {
	src := "`` `foo` ``\n"
	assertRoundTrip(t, src, Options{})

	green := Parse([]byte(src), Options{})
	para := green.Children[0].Node
	assertHasInlineNode(t, para, syntax.CodeSpan)
}

func TestInlineCodeSpanUnclosed(t *testing.T) {
	// Unclosed code spans should be literal backticks.
	src := "`unclosed\n"
	assertRoundTrip(t, src, Options{})
}

func TestInlineCodeSpanMultiLine(t *testing.T) {
	src := "`line1\nline2`\n"
	assertRoundTrip(t, src, Options{})

	green := Parse([]byte(src), Options{})
	para := green.Children[0].Node
	assertHasInlineNode(t, para, syntax.CodeSpan)
}

// --- Backslash escape tests ---

func TestInlineBackslashEscape(t *testing.T) {
	src := "\\*not emphasis\\*\n"
	assertRoundTrip(t, src, Options{})

	green := Parse([]byte(src), Options{})
	para := green.Children[0].Node
	assertHasInlineNode(t, para, syntax.BackslashEscape)
}

func TestInlineBackslashEscapeAllPunctuation(t *testing.T) {
	puncts := []byte("!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~")
	for _, p := range puncts {
		src := fmt.Sprintf("\\%c\n", p)
		assertRoundTrip(t, src, Options{})

		green := Parse([]byte(src), Options{})
		para := green.Children[0].Node
		assertHasInlineNode(t, para, syntax.BackslashEscape)
	}
}

func TestInlineBackslashNonPunctuation(t *testing.T) {
	// Backslash before a non-punctuation character is literal.
	src := "\\a\n"
	assertRoundTrip(t, src, Options{})
}

func TestInlineBackslashBeforeNewline(t *testing.T) {
	src := "foo\\\nbar\n"
	assertRoundTrip(t, src, Options{})

	green := Parse([]byte(src), Options{})
	para := green.Children[0].Node
	assertHasInlineNode(t, para, syntax.HardLineBreak)
}

// --- Entity reference tests ---

func TestInlineEntityRef(t *testing.T) {
	src := "&amp;\n"
	assertRoundTrip(t, src, Options{})

	green := Parse([]byte(src), Options{})
	para := green.Children[0].Node
	assertHasInlineNode(t, para, syntax.EntityRef)
}

func TestInlineEntityRefNumeric(t *testing.T) {
	src := "&#35;\n"
	assertRoundTrip(t, src, Options{})

	green := Parse([]byte(src), Options{})
	para := green.Children[0].Node
	assertHasInlineNode(t, para, syntax.EntityRef)
}

func TestInlineEntityRefHex(t *testing.T) {
	src := "&#x23;\n"
	assertRoundTrip(t, src, Options{})

	green := Parse([]byte(src), Options{})
	para := green.Children[0].Node
	assertHasInlineNode(t, para, syntax.EntityRef)
}

func TestInlineEntityRefInvalid(t *testing.T) {
	// Not a valid entity — should be literal.
	src := "&invalid entity;\n"
	assertRoundTrip(t, src, Options{})
}

// --- Autolink tests ---

func TestInlineAutolink(t *testing.T) {
	src := "<http://example.com>\n"
	assertRoundTrip(t, src, Options{})

	green := Parse([]byte(src), Options{})
	para := green.Children[0].Node
	assertHasInlineNode(t, para, syntax.AutolinkSpan)
}

func TestInlineAutolinkEmail(t *testing.T) {
	src := "<user@example.com>\n"
	assertRoundTrip(t, src, Options{})

	green := Parse([]byte(src), Options{})
	para := green.Children[0].Node
	assertHasInlineNode(t, para, syntax.AutolinkSpan)
}

// --- Raw HTML tests ---

func TestInlineRawHTML(t *testing.T) {
	src := "<em>text</em>\n"
	assertRoundTrip(t, src, Options{})

	green := Parse([]byte(src), Options{})
	para := green.Children[0].Node
	assertHasInlineNode(t, para, syntax.RawHTMLSpan)
}

func TestInlineRawHTMLSelfClosing(t *testing.T) {
	src := "<br />\n"
	assertRoundTrip(t, src, Options{})

	green := Parse([]byte(src), Options{})
	para := green.Children[0].Node
	assertHasInlineNode(t, para, syntax.RawHTMLSpan)
}

func TestInlineRawHTMLComment(t *testing.T) {
	// An HTML comment on its own line is an HTMLBlock, not inline.
	// To test inline HTML comment, embed it within a paragraph.
	src := "foo <!-- comment --> bar\n"
	assertRoundTrip(t, src, Options{})

	green := Parse([]byte(src), Options{})
	para := green.Children[0].Node
	assertHasInlineNode(t, para, syntax.RawHTMLSpan)
}

// --- Line break tests ---

func TestInlineSoftLineBreak(t *testing.T) {
	src := "foo\nbar\n"
	assertRoundTrip(t, src, Options{})

	green := Parse([]byte(src), Options{})
	para := green.Children[0].Node
	assertHasInlineNode(t, para, syntax.SoftLineBreak)
}

func TestInlineHardLineBreakSpaces(t *testing.T) {
	src := "foo  \nbar\n"
	assertRoundTrip(t, src, Options{})

	green := Parse([]byte(src), Options{})
	para := green.Children[0].Node
	assertHasInlineNode(t, para, syntax.HardLineBreak)
}

func TestInlineHardLineBreakBackslash(t *testing.T) {
	src := "foo\\\nbar\n"
	assertRoundTrip(t, src, Options{})

	green := Parse([]byte(src), Options{})
	para := green.Children[0].Node
	assertHasInlineNode(t, para, syntax.HardLineBreak)
}

// --- Emphasis tests ---

func TestInlineEmphasisStar(t *testing.T) {
	src := "*emphasis*\n"
	assertRoundTrip(t, src, Options{})

	green := Parse([]byte(src), Options{})
	para := green.Children[0].Node
	assertHasInlineNode(t, para, syntax.EmphasisSpan)
}

func TestInlineEmphasisUnderscore(t *testing.T) {
	src := "_emphasis_\n"
	assertRoundTrip(t, src, Options{})

	green := Parse([]byte(src), Options{})
	para := green.Children[0].Node
	assertHasInlineNode(t, para, syntax.EmphasisSpan)
}

func TestInlineStrongStar(t *testing.T) {
	src := "**strong**\n"
	assertRoundTrip(t, src, Options{})

	green := Parse([]byte(src), Options{})
	para := green.Children[0].Node
	assertHasInlineNode(t, para, syntax.StrongSpan)
}

func TestInlineStrongUnderscore(t *testing.T) {
	src := "__strong__\n"
	assertRoundTrip(t, src, Options{})

	green := Parse([]byte(src), Options{})
	para := green.Children[0].Node
	assertHasInlineNode(t, para, syntax.StrongSpan)
}

func TestInlineEmphasisNested(t *testing.T) {
	src := "*foo **bar** baz*\n"
	assertRoundTrip(t, src, Options{})

	green := Parse([]byte(src), Options{})
	para := green.Children[0].Node
	assertHasInlineNode(t, para, syntax.EmphasisSpan)
	assertHasInlineNode(t, para, syntax.StrongSpan)
}

func TestInlineEmphasisWithCodeSpan(t *testing.T) {
	src := "*foo `bar` baz*\n"
	assertRoundTrip(t, src, Options{})

	green := Parse([]byte(src), Options{})
	para := green.Children[0].Node
	assertHasInlineNode(t, para, syntax.EmphasisSpan)
	assertHasInlineNode(t, para, syntax.CodeSpan)
}

func TestInlineEmphasisUnmatched(t *testing.T) {
	// Unmatched delimiter should remain as text.
	src := "*unmatched\n"
	assertRoundTrip(t, src, Options{})
}

func TestInlineEmphasisTripleStar(t *testing.T) {
	src := "***strong and em***\n"
	assertRoundTrip(t, src, Options{})
}

func TestSpecEmphasisAndStrongEmphasis(t *testing.T) {
	examples := loadSpecExamples(t)
	for _, ex := range examples {
		if ex.Section != "Emphasis and strong emphasis" {
			continue
		}
		t.Run(fmt.Sprintf("example_%d", ex.Example), func(t *testing.T) {
			assertRoundTrip(t, ex.Markdown, Options{})
		})
	}
}

// --- Link tests ---

func TestInlineLink(t *testing.T) {
	src := "[text](url)\n"
	assertRoundTrip(t, src, Options{})

	green := Parse([]byte(src), Options{})
	para := green.Children[0].Node
	assertHasInlineNode(t, para, syntax.Link)
}

func TestInlineLinkWithTitle(t *testing.T) {
	src := "[text](url \"title\")\n"
	assertRoundTrip(t, src, Options{})

	green := Parse([]byte(src), Options{})
	para := green.Children[0].Node
	assertHasInlineNode(t, para, syntax.Link)
}

func TestInlineLinkEmpty(t *testing.T) {
	src := "[text]()\n"
	assertRoundTrip(t, src, Options{})
}

func TestInlineImage(t *testing.T) {
	src := "![alt](image.png)\n"
	assertRoundTrip(t, src, Options{})

	green := Parse([]byte(src), Options{})
	para := green.Children[0].Node
	assertHasInlineNode(t, para, syntax.Image)
}

func TestInlineImageWithTitle(t *testing.T) {
	src := "![alt](image.png \"title\")\n"
	assertRoundTrip(t, src, Options{})
}

func TestInlineLinkWithEmphasis(t *testing.T) {
	src := "[*emphasis*](url)\n"
	assertRoundTrip(t, src, Options{})

	green := Parse([]byte(src), Options{})
	para := green.Children[0].Node
	assertHasInlineNode(t, para, syntax.Link)
}

func TestInlineLinkNoMatch(t *testing.T) {
	// No matching destination: brackets are literal.
	src := "[text]\n"
	assertRoundTrip(t, src, Options{})
}

func TestSpecLinks(t *testing.T) {
	examples := loadSpecExamples(t)
	for _, ex := range examples {
		if ex.Section != "Links" {
			continue
		}
		t.Run(fmt.Sprintf("example_%d", ex.Example), func(t *testing.T) {
			assertRoundTrip(t, ex.Markdown, Options{})
		})
	}
}

func TestSpecImages(t *testing.T) {
	examples := loadSpecExamples(t)
	for _, ex := range examples {
		if ex.Section != "Images" {
			continue
		}
		t.Run(fmt.Sprintf("example_%d", ex.Example), func(t *testing.T) {
			assertRoundTrip(t, ex.Markdown, Options{})
		})
	}
}

// --- Heading inline tests ---

func TestInlineInHeading(t *testing.T) {
	src := "# Hello `code`\n"
	assertRoundTrip(t, src, Options{})

	green := Parse([]byte(src), Options{})
	heading := green.Children[0].Node
	if heading.Kind != syntax.ATXHeading {
		t.Fatalf("expected ATXHeading, got %v", heading.Kind)
	}
	assertHasInlineNode(t, heading, syntax.CodeSpan)
}

func TestInlineInSetextHeading(t *testing.T) {
	src := "Hello `code`\n======\n"
	assertRoundTrip(t, src, Options{})

	green := Parse([]byte(src), Options{})
	heading := green.Children[0].Node
	if heading.Kind != syntax.SetextHeading {
		t.Fatalf("expected SetextHeading, got %v", heading.Kind)
	}
	assertHasInlineNode(t, heading, syntax.CodeSpan)
}

// --- Round-trip with inline constructs in block contexts ---

func TestInlineInBlockQuote(t *testing.T) {
	src := "> `code` **bold**\n"
	assertRoundTrip(t, src, Options{})
}

func TestInlineInListItem(t *testing.T) {
	src := "- `code` in a list\n"
	assertRoundTrip(t, src, Options{})
}

func TestInlinePlainText(t *testing.T) {
	// No inline markup — should still round-trip and produce InlineText.
	src := "Just plain text\n"
	assertRoundTrip(t, src, Options{})
}

func TestInlineNoModificationInCodeBlock(t *testing.T) {
	// Code blocks should not have inline parsing.
	src := "```\n`code` in a fenced block\n```\n"
	assertRoundTrip(t, src, Options{})

	green := Parse([]byte(src), Options{})
	fenced := green.Children[0].Node
	if fenced.Kind != syntax.FencedCodeBlock {
		t.Fatalf("expected FencedCodeBlock, got %v", fenced.Kind)
	}
	// Should not contain any CodeSpan nodes.
	for _, child := range fenced.Children {
		if child.Node != nil && child.Node.Kind == syntax.CodeSpan {
			t.Error("fenced code block should not contain inline CodeSpan nodes")
		}
	}
}

// --- Spec example tests for inline constructs ---

func TestSpecCodeSpans(t *testing.T) {
	examples := loadSpecExamples(t)
	for _, ex := range examples {
		if ex.Section != "Code spans" {
			continue
		}
		t.Run(fmt.Sprintf("example_%d", ex.Example), func(t *testing.T) {
			assertRoundTrip(t, ex.Markdown, Options{})
		})
	}
}

func TestSpecBackslashEscapes(t *testing.T) {
	examples := loadSpecExamples(t)
	for _, ex := range examples {
		if ex.Section != "Backslash escapes" {
			continue
		}
		t.Run(fmt.Sprintf("example_%d", ex.Example), func(t *testing.T) {
			assertRoundTrip(t, ex.Markdown, Options{})
		})
	}
}

func TestSpecEntityReferences(t *testing.T) {
	examples := loadSpecExamples(t)
	for _, ex := range examples {
		if ex.Section != "Entity and numeric character references" {
			continue
		}
		t.Run(fmt.Sprintf("example_%d", ex.Example), func(t *testing.T) {
			assertRoundTrip(t, ex.Markdown, Options{})
		})
	}
}

func TestSpecAutolinks(t *testing.T) {
	examples := loadSpecExamples(t)
	for _, ex := range examples {
		if ex.Section != "Autolinks" {
			continue
		}
		t.Run(fmt.Sprintf("example_%d", ex.Example), func(t *testing.T) {
			assertRoundTrip(t, ex.Markdown, Options{})
		})
	}
}

func TestSpecRawHTML(t *testing.T) {
	examples := loadSpecExamples(t)
	for _, ex := range examples {
		if ex.Section != "Raw HTML" {
			continue
		}
		t.Run(fmt.Sprintf("example_%d", ex.Example), func(t *testing.T) {
			assertRoundTrip(t, ex.Markdown, Options{})
		})
	}
}

func TestSpecHardLineBreaks(t *testing.T) {
	examples := loadSpecExamples(t)
	for _, ex := range examples {
		if ex.Section != "Hard line breaks" {
			continue
		}
		t.Run(fmt.Sprintf("example_%d", ex.Example), func(t *testing.T) {
			assertRoundTrip(t, ex.Markdown, Options{})
		})
	}
}

func TestSpecSoftLineBreaks(t *testing.T) {
	examples := loadSpecExamples(t)
	for _, ex := range examples {
		if ex.Section != "Soft line breaks" {
			continue
		}
		t.Run(fmt.Sprintf("example_%d", ex.Example), func(t *testing.T) {
			assertRoundTrip(t, ex.Markdown, Options{})
		})
	}
}

// --- Helper functions ---

// assertHasInlineNode checks that a node contains at least one child node
// of the given kind (recursively).
func assertHasInlineNode(t *testing.T, n *syntax.GreenNode, kind syntax.SyntaxKind) {
	t.Helper()
	if !hasNodeKind(n, kind) {
		t.Errorf("expected to find %v in %v, tree:\n%s", kind, n.Kind, syntax.DebugDump(n))
	}
}

func hasNodeKind(n *syntax.GreenNode, kind syntax.SyntaxKind) bool {
	for _, child := range n.Children {
		if child.Node != nil {
			if child.Node.Kind == kind {
				return true
			}
			if hasNodeKind(child.Node, kind) {
				return true
			}
		}
	}
	return false
}
