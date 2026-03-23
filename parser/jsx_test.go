package parser

import (
	"testing"

	"github.com/albertocavalcante/mdxgo/syntax"
)

// --- Block JSX tests ---

func TestJSXBlockOpeningTag(t *testing.T) {
	src := "<Component>\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}
	assertChildKinds(t, green, syntax.JSXBlock)
}

func TestJSXBlockClosingTag(t *testing.T) {
	src := "</Component>\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}
	assertChildKinds(t, green, syntax.JSXBlock)
}

func TestJSXBlockSelfClosing(t *testing.T) {
	src := "<Component />\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}
	assertChildKinds(t, green, syntax.JSXBlock)

	// Should contain a JSXSelfClosingTag sub-node.
	jsx := green.Children[0].Node
	hasSelfClosing := false
	for _, child := range jsx.Children {
		if child.Node != nil && child.Node.Kind == syntax.JSXSelfClosingTag {
			hasSelfClosing = true
		}
	}
	if !hasSelfClosing {
		t.Error("JSXBlock should contain JSXSelfClosingTag")
	}
}

func TestJSXBlockFragment(t *testing.T) {
	src := "<>\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}
	assertChildKinds(t, green, syntax.JSXBlock)
}

func TestJSXBlockClosingFragment(t *testing.T) {
	src := "</>\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}
	assertChildKinds(t, green, syntax.JSXBlock)
}

func TestJSXBlockWithStringAttribute(t *testing.T) {
	src := `<Component name="value" />` + "\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}
	assertChildKinds(t, green, syntax.JSXBlock)
}

func TestJSXBlockWithBooleanAttribute(t *testing.T) {
	src := "<Component disabled />\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}
}

func TestJSXBlockWithExpressionAttribute(t *testing.T) {
	src := "<Component value={1 + 2} />\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}
}

func TestJSXBlockWithSpreadAttribute(t *testing.T) {
	src := "<Component {...props} />\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}
}

func TestJSXBlockMemberExpression(t *testing.T) {
	src := "<Foo.Bar />\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}

	// Should contain dot-separated identifiers.
	jsx := green.Children[0].Node
	hasDot := false
	for _, child := range jsx.Children {
		if child.Node != nil {
			for _, gc := range child.Node.Children {
				if gc.Token != nil && gc.Token.Kind == syntax.JSXDot {
					hasDot = true
				}
			}
		}
	}
	if !hasDot {
		t.Error("member expression should contain JSXDot")
	}
}

func TestJSXBlockNamespace(t *testing.T) {
	src := "<xml:tag />\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}
}

func TestJSXBlockMultipleAttributes(t *testing.T) {
	src := `<Component a="1" b={2} c />` + "\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}
}

func TestJSXBlockInterruptsParagraph(t *testing.T) {
	src := "paragraph\n<Component />\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}
	assertChildKinds(t, green, syntax.Paragraph, syntax.JSXBlock)
}

func TestJSXBlockNotInCommonMark(t *testing.T) {
	src := "<Component />\n"
	green := Parse([]byte(src), Options{MDX: false})
	// In CommonMark mode, this could be an HTML block or paragraph.
	for _, child := range green.Children {
		if child.Node != nil && child.Node.Kind == syntax.JSXBlock {
			t.Error("JSX blocks should not be parsed in CommonMark mode")
		}
	}
}

func TestJSXBlockFullDocument(t *testing.T) {
	src := "<Wrapper>\n\nHello world.\n\n</Wrapper>\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}
	// Opening tag, blank, paragraph, blank, closing tag.
	assertChildKinds(t, green,
		syntax.JSXBlock,
		syntax.BlankLineNode,
		syntax.Paragraph,
		syntax.BlankLineNode,
		syntax.JSXBlock,
	)
}

func TestJSXBlockLowercaseHTML(t *testing.T) {
	src := "<div>\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}
	assertChildKinds(t, green, syntax.JSXBlock)
}

func TestJSXBlockWithSingleQuoteAttr(t *testing.T) {
	src := "<Component name='value' />\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}
}

// --- Inline JSX tests ---

func TestJSXInlineInParagraph(t *testing.T) {
	src := "Hello <Component /> world\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}

	// Should have a paragraph with JSXInline.
	para := green.Children[0].Node
	hasJSX := false
	for _, child := range para.Children {
		if child.Node != nil && child.Node.Kind == syntax.JSXInline {
			hasJSX = true
		}
	}
	if !hasJSX {
		t.Error("paragraph should contain JSXInline")
	}
}

func TestJSXInlineFragment(t *testing.T) {
	src := "Hello <> world\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}
}

func TestJSXInlineInHeading(t *testing.T) {
	src := "# Title <Badge />\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}
}

func TestJSXInlineClosingTag(t *testing.T) {
	src := "Hello </Component> world\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}
}

func TestJSXInlineWithAttributes(t *testing.T) {
	src := `Hello <Component name="test" /> world` + "\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}
}

func TestJSXInlineNotInCommonMark(t *testing.T) {
	// In CommonMark mode, < triggers autolink/raw HTML, not JSX.
	src := "Hello <em>world</em>\n"
	green := Parse([]byte(src), Options{MDX: false})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}
	// Should NOT contain JSXInline nodes.
	para := green.Children[0].Node
	for _, child := range para.Children {
		if child.Node != nil && child.Node.Kind == syntax.JSXInline {
			t.Error("JSX inline should not be parsed in CommonMark mode")
		}
	}
}

func TestJSXInlineInvalidTreatedAsText(t *testing.T) {
	// '<' followed by non-identifier should be literal text.
	src := "a < b\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}
}

func TestJSXInlineURLNotJSX(t *testing.T) {
	// URL autolinks should NOT be parsed as JSX.
	src := "<https://example.com>\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}
}

// --- Round-trip tests ---

func TestJSXRoundTrip(t *testing.T) {
	cases := []string{
		"<Component />\n",
		"<Component>\n",
		"</Component>\n",
		"<>\n",
		"</>\n",
		`<C name="val" />` + "\n",
		"<C disabled />\n",
		"<C value={expr} />\n",
		"<C {...props} />\n",
		"<Foo.Bar />\n",
		"<xml:tag />\n",
		`<C a="1" b={2} c />` + "\n",
		"<Wrapper>\nContent\n</Wrapper>\n",
		"Hello <C /> world\n",
		"# Heading <Badge />\n",
		"<div className=\"test\">\n",
		"</div>\n",
		"<Component\n  prop=\"value\"\n/>\n",
		"<C name='single' />\n",
	}

	for _, src := range cases {
		green := Parse([]byte(src), Options{MDX: true})
		got := syntax.FullText(green)
		if got != src {
			t.Errorf("round-trip failed for %q: got %q", src, got)
		}
	}
}

// --- Interaction tests (JSX + expressions) ---

func TestJSXWithExpressionContent(t *testing.T) {
	src := "<Component value={1 + 2}>\n\n{content}\n\n</Component>\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}
}

func TestMixedMDXDocument(t *testing.T) {
	src := `---
title: Test
---

import { Component } from 'lib'

# Hello {name}

<Component value={42}>

Paragraph content.

</Component>

{footer}
`
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed for mixed MDX document")
	}
}
