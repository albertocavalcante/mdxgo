package parser

import (
	"testing"

	"github.com/albertocavalcante/mdxgo/syntax"
)

// --- Block expression tests ---

func TestExpressionBlockSimple(t *testing.T) {
	green := Parse([]byte("{expression}\n"), Options{MDX: true})
	assertChildKinds(t, green, syntax.ExpressionBlock)

	expr := green.Children[0].Node
	hasOpen := false
	hasClose := false
	hasContent := false
	for _, child := range expr.Children {
		if child.Token != nil {
			switch child.Token.Kind {
			case syntax.ExprOpenBrace:
				hasOpen = true
			case syntax.ExprCloseBrace:
				hasClose = true
			case syntax.ExprContentToken:
				hasContent = true
				if child.Token.Text != "expression" {
					t.Errorf("content = %q, want %q", child.Token.Text, "expression")
				}
			}
		}
	}
	if !hasOpen {
		t.Error("ExpressionBlock missing ExprOpenBrace")
	}
	if !hasClose {
		t.Error("ExpressionBlock missing ExprCloseBrace")
	}
	if !hasContent {
		t.Error("ExpressionBlock missing ExprContentToken")
	}
}

func TestExpressionBlockMultiline(t *testing.T) {
	src := "{\n  1 + 2\n}\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}
	assertChildKinds(t, green, syntax.ExpressionBlock)
}

func TestExpressionBlockNested(t *testing.T) {
	src := "{obj.map(x => { return x + 1 })}\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}
	assertChildKinds(t, green, syntax.ExpressionBlock)
}

func TestExpressionBlockEmpty(t *testing.T) {
	src := "{}\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}
	assertChildKinds(t, green, syntax.ExpressionBlock)
}

func TestExpressionBlockWithStrings(t *testing.T) {
	src := `{foo("hello { world }")}` + "\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}
	assertChildKinds(t, green, syntax.ExpressionBlock)
}

func TestExpressionBlockUnclosed(t *testing.T) {
	src := "{unclosed\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}
	// Should be ErrorNode for unclosed expression.
	assertChildKinds(t, green, syntax.ErrorNode)
}

func TestExpressionBlockInterruptsParagraph(t *testing.T) {
	src := "paragraph\n{expression}\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}
	assertChildKinds(t, green, syntax.Paragraph, syntax.ExpressionBlock)
}

func TestExpressionBlockNotInCommonMark(t *testing.T) {
	src := "{expression}\n"
	green := Parse([]byte(src), Options{MDX: false})
	// In CommonMark mode, { is just text in a paragraph.
	assertChildKinds(t, green, syntax.Paragraph)
}

func TestExpressionBlockIndented(t *testing.T) {
	src := "  {indented}\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}
	assertChildKinds(t, green, syntax.ExpressionBlock)
}

func TestExpressionBlockWithTemplateLiteral(t *testing.T) {
	src := "{`template ${val} literal`}\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}
	assertChildKinds(t, green, syntax.ExpressionBlock)
}

// --- Inline expression tests ---

func TestExpressionInlineSimple(t *testing.T) {
	src := "Hello {name}!\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}

	// Paragraph should contain InlineText, ExpressionInline, InlineText
	para := green.Children[0].Node
	hasExpr := false
	for _, child := range para.Children {
		if child.Node != nil && child.Node.Kind == syntax.ExpressionInline {
			hasExpr = true
		}
	}
	if !hasExpr {
		t.Error("paragraph should contain ExpressionInline")
	}
}

func TestExpressionInlineInHeading(t *testing.T) {
	src := "# Title {count}\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}
}

func TestExpressionInlineNested(t *testing.T) {
	src := "Value: {a + {b: 1}.b}\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}
}

func TestExpressionInlineUnclosed(t *testing.T) {
	src := "Hello {unclosed\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}
	// Unclosed { should be treated as literal text.
	assertChildKinds(t, green, syntax.Paragraph)
}

func TestExpressionInlineNotInCommonMark(t *testing.T) {
	src := "Hello {name}!\n"
	green := Parse([]byte(src), Options{MDX: false})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}
	// In CommonMark, { is literal text.
	assertChildKinds(t, green, syntax.Paragraph)
}

func TestExpressionInlineEmpty(t *testing.T) {
	src := "Hello {}!\n"
	green := Parse([]byte(src), Options{MDX: true})
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed: %q != %q", got, src)
	}
}

// --- Round-trip tests ---

func TestExpressionRoundTrip(t *testing.T) {
	cases := []string{
		"{expression}\n",
		"{\n  multiline\n}\n",
		"{nested({a: 1, b: 2})}\n",
		"{}\n",
		"{/* comment */}\n",
		"para\n{expr}\n",
		"Hello {name}!\n",
		"{a} and {b}\n",
		"# {title}\n",
		"{unclosed expression\n",
		"  {indented}\n",
		`{foo("string with { inside")}` + "\n",
		"{`template`}\n",
		"  {\n    complex()\n  }\n",
	}

	for _, src := range cases {
		green := Parse([]byte(src), Options{MDX: true})
		got := syntax.FullText(green)
		if got != src {
			t.Errorf("round-trip failed for %q: got %q", src, got)
		}
	}
}
