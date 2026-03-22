package parser

import (
	"testing"

	"github.com/albertocavalcante/mdxgo/syntax"
)

// assertRoundTrip verifies the fundamental invariant:
// syntax.FullText(Parse(src)) == src
func assertRoundTrip(t *testing.T, src string, opts Options) {
	t.Helper()
	green := Parse([]byte(src), opts)
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed:\ninput:  %q\noutput: %q", src, got)
	}
}

func TestRoundTripEmpty(t *testing.T) {
	assertRoundTrip(t, "", Options{})
}

func TestRoundTripSingleLine(t *testing.T) {
	assertRoundTrip(t, "hello world", Options{})
}

func TestRoundTripSingleLineNewline(t *testing.T) {
	assertRoundTrip(t, "hello world\n", Options{})
}

func TestRoundTripMultiLine(t *testing.T) {
	assertRoundTrip(t, "line one\nline two\nline three\n", Options{})
}

func TestRoundTripBlankLines(t *testing.T) {
	assertRoundTrip(t, "paragraph one\n\nparagraph two\n", Options{})
}

func TestRoundTripATXHeading(t *testing.T) {
	cases := []string{
		"# Heading 1\n",
		"## Heading 2\n",
		"### Heading 3\n",
		"#### Heading 4\n",
		"##### Heading 5\n",
		"###### Heading 6\n",
		"# Heading with closing ##\n",
		"#  Extra space\n",
		"# \n", // empty heading
	}
	for _, src := range cases {
		t.Run(src, func(t *testing.T) {
			assertRoundTrip(t, src, Options{})
		})
	}
}

func TestRoundTripATXHeadingIndented(t *testing.T) {
	assertRoundTrip(t, " # Heading\n", Options{})
	assertRoundTrip(t, "  ## Heading\n", Options{})
	assertRoundTrip(t, "   ### Heading\n", Options{})
}

func TestRoundTripThematicBreak(t *testing.T) {
	cases := []string{
		"---\n",
		"***\n",
		"___\n",
		"- - -\n",
		"* * *\n",
		"  ---\n",
		"   ___\n",
		"----------\n",
	}
	for _, src := range cases {
		t.Run(src, func(t *testing.T) {
			assertRoundTrip(t, src, Options{})
		})
	}
}

func TestRoundTripFencedCode(t *testing.T) {
	cases := []string{
		"```\ncode\n```\n",
		"```go\nfunc main() {}\n```\n",
		"~~~\ncode\n~~~\n",
		"````\ncode with ``` backticks\n````\n",
		"```\n\n```\n",
		"```\nunclosed fence\n",
	}
	for _, src := range cases {
		t.Run(src, func(t *testing.T) {
			assertRoundTrip(t, src, Options{})
		})
	}
}

func TestRoundTripIndentedCode(t *testing.T) {
	cases := []string{
		"    code line\n",
		"    line 1\n    line 2\n",
		"    line 1\n\n    line 2\n",
	}
	for _, src := range cases {
		t.Run(src, func(t *testing.T) {
			assertRoundTrip(t, src, Options{})
		})
	}
}

func TestRoundTripBlockQuote(t *testing.T) {
	cases := []string{
		"> quote\n",
		"> line 1\n> line 2\n",
		"> # Heading\n",
		">  two spaces\n",
	}
	for _, src := range cases {
		t.Run(src, func(t *testing.T) {
			assertRoundTrip(t, src, Options{})
		})
	}
}

func TestRoundTripBulletList(t *testing.T) {
	cases := []string{
		"- item\n",
		"- item 1\n- item 2\n",
		"* item\n",
		"+ item\n",
	}
	for _, src := range cases {
		t.Run(src, func(t *testing.T) {
			assertRoundTrip(t, src, Options{})
		})
	}
}

func TestRoundTripOrderedList(t *testing.T) {
	cases := []string{
		"1. item\n",
		"1. first\n2. second\n",
		"1) item\n",
	}
	for _, src := range cases {
		t.Run(src, func(t *testing.T) {
			assertRoundTrip(t, src, Options{})
		})
	}
}

func TestRoundTripSetextHeading(t *testing.T) {
	cases := []string{
		"Heading\n=======\n",
		"Heading\n-------\n",
		"Multi line\nparagraph\n=========\n",
	}
	for _, src := range cases {
		t.Run(src, func(t *testing.T) {
			assertRoundTrip(t, src, Options{})
		})
	}
}

func TestRoundTripHTMLBlock(t *testing.T) {
	cases := []string{
		"<div>\nfoo\n</div>\n",
		"<!-- comment -->\n",
	}
	for _, src := range cases {
		t.Run(src, func(t *testing.T) {
			assertRoundTrip(t, src, Options{})
		})
	}
}

func TestRoundTripFrontmatter(t *testing.T) {
	cases := []string{
		"---\ntitle: test\n---\n",
		"---\ntitle: test\ndescription: foo\n---\n",
		"---\n---\n",
	}
	for _, src := range cases {
		t.Run(src, func(t *testing.T) {
			assertRoundTrip(t, src, Options{MDX: true})
		})
	}
}

func TestRoundTripESM(t *testing.T) {
	cases := []string{
		"import Foo from 'foo'\n",
		"export const x = 1\n",
	}
	for _, src := range cases {
		t.Run(src, func(t *testing.T) {
			assertRoundTrip(t, src, Options{MDX: true})
		})
	}
}

func TestRoundTripComplex(t *testing.T) {
	src := `# Title

Some paragraph text here
with multiple lines.

## Section

- Item 1
- Item 2

` + "```go" + `
func main() {
    fmt.Println("hello")
}
` + "```" + `

> A blockquote
> with two lines

---

Another paragraph.
`
	assertRoundTrip(t, src, Options{})
}

func TestRoundTripMDXComplex(t *testing.T) {
	src := `---
title: My Page
---

import { Card } from '@components/Card'

# Hello World

Some text here.

export const metadata = {
  author: "test"
}
`
	assertRoundTrip(t, src, Options{MDX: true})
}

func TestRoundTripCRLF(t *testing.T) {
	assertRoundTrip(t, "hello\r\nworld\r\n", Options{})
}

func TestRoundTripMixedEndings(t *testing.T) {
	assertRoundTrip(t, "line1\nline2\r\nline3\r", Options{})
}

func TestRoundTripTabsPreserved(t *testing.T) {
	assertRoundTrip(t, "\tindented\n", Options{})
	assertRoundTrip(t, "\t\tdeep\n", Options{})
}

func TestRoundTripTrailingSpaces(t *testing.T) {
	assertRoundTrip(t, "text with trailing   \n", Options{})
}

func TestDocumentStructure(t *testing.T) {
	src := "# Title\n\nParagraph\n"
	green := Parse([]byte(src), Options{})

	if green.Kind != syntax.Document {
		t.Fatalf("root kind = %v, want Document", green.Kind)
	}

	// Should have: ATXHeading, BlankLineNode, Paragraph
	kinds := childKinds(green)
	if len(kinds) != 3 {
		t.Fatalf("got %d children, want 3: %v", len(kinds), kinds)
	}
	if kinds[0] != syntax.ATXHeading {
		t.Errorf("child[0] = %v, want ATXHeading", kinds[0])
	}
	if kinds[1] != syntax.BlankLineNode {
		t.Errorf("child[1] = %v, want BlankLineNode", kinds[1])
	}
	if kinds[2] != syntax.Paragraph {
		t.Errorf("child[2] = %v, want Paragraph", kinds[2])
	}
}

func TestFencedCodeStructure(t *testing.T) {
	src := "```go\nfmt.Println()\n```\n"
	green := Parse([]byte(src), Options{})

	if len(green.Children) != 1 {
		t.Fatalf("got %d children, want 1", len(green.Children))
	}
	fenced := green.Children[0].Node
	if fenced == nil {
		t.Fatal("expected a node")
	}
	if fenced.Kind != syntax.FencedCodeBlock {
		t.Fatalf("kind = %v, want FencedCodeBlock", fenced.Kind)
	}
}

// childKinds returns the SyntaxKinds of a node's direct children.
func childKinds(n *syntax.GreenNode) []syntax.SyntaxKind {
	var kinds []syntax.SyntaxKind
	for _, c := range n.Children {
		kinds = append(kinds, c.Kind())
	}
	return kinds
}
