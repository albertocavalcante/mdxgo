package parser

import (
	"strings"
	"testing"

	"github.com/albertocavalcante/mdxgo/syntax"
)

// TestAdversarialInputs tests pathological and adversarial inputs that
// might break the parser through edge cases, deep nesting, or large sizes.
func TestAdversarialInputs(t *testing.T) {
	cases := []struct {
		name string
		src  string
		mdx  bool
	}{
		// Deeply repeated markers
		{"100_hashes", strings.Repeat("#", 100) + "\n", false},
		{"100_dashes", strings.Repeat("-", 100) + "\n", false},
		{"100_stars", strings.Repeat("*", 100) + "\n", false},
		{"100_underscores", strings.Repeat("_", 100) + "\n", false},
		{"100_backticks", strings.Repeat("`", 100) + "\n", false},
		{"100_tildes", strings.Repeat("~", 100) + "\n", false},
		{"100_angles", strings.Repeat(">", 100) + "\n", false},
		{"100_equals", strings.Repeat("=", 100) + "\n", false},

		// Deep nesting
		{"50_nested_quotes", strings.Repeat("> ", 50) + "text\n", false},
		{"20_nested_quotes_lines", genNestedQuotes(20), false},

		// Large content
		{"10k_char_line", strings.Repeat("a", 10000) + "\n", false},
		{"10k_char_heading", "# " + strings.Repeat("x", 10000) + "\n", false},
		{"1000_blank_lines", strings.Repeat("\n", 1000), false},
		{"1000_short_paragraphs", genManyParagraphs(1000), false},
		{"500_headings", genManyHeadings(500), false},
		{"200_code_blocks", genManyCodeBlocks(200), false},
		{"200_list_items", genManyListItems(200), false},
		{"100_thematic_breaks", genManyThematicBreaks(100), false},
		{"100_blockquotes", genManyBlockquotes(100), false},

		// Alternating constructs (rapid context switching)
		{"rapid_alternation", genRapidAlternation(50), false},

		// Whitespace torture
		{"only_spaces", strings.Repeat(" ", 1000) + "\n", false},
		{"only_tabs", strings.Repeat("\t", 1000) + "\n", false},
		{"spaces_and_tabs", strings.Repeat(" \t", 500) + "\n", false},
		{"4_space_indent_chain", gen4SpaceChain(100), false},

		// Line ending edge cases
		{"only_crlf", strings.Repeat("\r\n", 500), false},
		{"only_cr", strings.Repeat("\r", 500), false},
		{"only_lf", strings.Repeat("\n", 500), false},
		{"alternating_endings", genAlternatingLineEndings(100), false},
		{"crlf_in_code_block", "```\r\ncode line\r\n```\r\n", false},

		// Fence nesting attacks
		{"fence_in_fence", "````\n```\ninner\n```\n````\n", false},
		{"many_unclosed_fences", strings.Repeat("```\n", 50), false},
		{"fence_chars_in_info", "```go func() { } ```\ncode\n```\n", false},

		// Almost-but-not-quite constructs
		{"almost_heading_7hash", "####### not a heading\n", false},
		{"almost_heading_4indent", "    # not a heading\n", false},
		{"almost_thematic_2dash", "--\n", false},
		{"almost_thematic_mixed", "-*-\n", false},
		{"almost_list_no_space", "-no space\n", false},

		// Setext ambiguity
		{"setext_vs_thematic", "foo\n---\n", false},
		{"setext_vs_thematic_star", "foo\n***\n", false},
		{"setext_indented", "   foo\n   ---\n", false},

		// Empty constructs
		{"empty_string", "", false},
		{"single_newline", "\n", false},
		{"single_space", " ", false},
		{"single_tab", "\t", false},
		{"single_cr", "\r", false},
		{"single_crlf", "\r\n", false},
		{"null_byte", "\x00", false},
		{"control_chars", "\x01\x02\x03\n", false},
		{"bom", "\xef\xbb\xbf# Heading\n", false}, // UTF-8 BOM

		// Unicode stress
		{"cjk_heading", "# 日本語の見出し\n", false},
		{"emoji_list", "- 🎉\n- 🚀\n- ✨\n", false},
		{"rtl_text", "# مرحبا\n\nنص عربي\n", false},
		{"zalgo", "# H̸̡̪̯ẻ̶̬̤a̷̢̝d̴̨̛i̵̬̐n̷̨̛g̷̣̈\n", false},
		{"zero_width", "# test\u200B\u200C\u200D\n", false},
		{"combining_chars", "# e\u0301 a\u0300\n", false},

		// MDX-specific adversarial
		{"unclosed_frontmatter", "---\ntitle: test\n", true},
		{"frontmatter_in_middle", "\n---\ntitle: test\n---\n", true},
		{"triple_frontmatter", "---\na: 1\n---\n\n---\nb: 2\n---\n", true},
		{"esm_like_paragraph", "import but not really\n", true},
		{"export_like_paragraph", "exporting is fun\n", true},
		{"import_no_from", "import\n", true},
		{"export_no_value", "export\n", true},
		{"many_imports", genManyImports(100), true},
		{"mdx_all_constructs", genFullMDXDoc(), true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Must not panic.
			var green *syntax.GreenNode
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("PANIC: %v", r)
					}
				}()
				green = Parse([]byte(tc.src), Options{MDX: tc.mdx})
			}()

			// Must round-trip.
			got := syntax.FullText(green)
			if got != tc.src {
				diverge := firstDivergence(tc.src, got)
				t.Errorf("round-trip failed: input=%d output=%d diverge=%d\n"+
					"  input  context: %q\n"+
					"  output context: %q",
					len(tc.src), len(got), diverge,
					safeSlice(tc.src, max(0, diverge-30), min(len(tc.src), diverge+30)),
					safeSlice(got, max(0, diverge-30), min(len(got), diverge+30)),
				)
			}

			// Root must be Document.
			if green.Kind != syntax.Document {
				t.Errorf("root kind = %v, want Document", green.Kind)
			}

			// Width must match input.
			if green.Width != len(tc.src) {
				t.Errorf("width %d != input length %d", green.Width, len(tc.src))
			}

			// Width consistency.
			if err := checkWidthConsistency(green); err != nil {
				t.Errorf("width inconsistency: %v", err)
			}
		})
	}
}

// --- Generators for adversarial test data ---

func genNestedQuotes(depth int) string {
	var b strings.Builder
	for i := 1; i <= depth; i++ {
		b.WriteString(strings.Repeat("> ", i))
		b.WriteString("level ")
		b.WriteByte(byte('0' + i%10))
		b.WriteByte('\n')
	}
	return b.String()
}

func genManyParagraphs(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteString("Paragraph ")
		b.WriteString(strings.Repeat("word ", 5))
		b.WriteString("\n\n")
	}
	return b.String()
}

func genManyHeadings(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		level := (i % 6) + 1
		b.WriteString(strings.Repeat("#", level))
		b.WriteString(" Heading\n\n")
	}
	return b.String()
}

func genManyCodeBlocks(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteString("```\ncode line\n```\n\n")
	}
	return b.String()
}

func genManyListItems(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteString("- item\n")
	}
	return b.String()
}

func genManyThematicBreaks(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteString("---\n\n")
	}
	return b.String()
}

func genManyBlockquotes(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteString("> quote\n\n")
	}
	return b.String()
}

func genRapidAlternation(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteString("# Heading\n\n")
		b.WriteString("Paragraph.\n\n")
		b.WriteString("---\n\n")
		b.WriteString("> Quote\n\n")
		b.WriteString("- Item\n\n")
		b.WriteString("```\ncode\n```\n\n")
	}
	return b.String()
}

func gen4SpaceChain(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteString("    code line\n")
	}
	return b.String()
}

func genAlternatingLineEndings(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteString("line")
		switch i % 3 {
		case 0:
			b.WriteByte('\n')
		case 1:
			b.WriteString("\r\n")
		case 2:
			b.WriteByte('\r')
		}
	}
	return b.String()
}

func genManyImports(n int) string {
	var b strings.Builder
	b.WriteString("---\ntitle: test\n---\n\n")
	for i := 0; i < n; i++ {
		b.WriteString("import { Comp")
		b.WriteString(strings.Repeat("x", i%20))
		b.WriteString(" } from 'mod'\n")
	}
	b.WriteString("\n# Content\n")
	return b.String()
}

func genFullMDXDoc() string {
	return `---
title: Full MDX Doc
description: Tests all MDX constructs together
tags:
  - test
  - mdx
---

import { Card } from '@components/Card'
import { Note } from '@components/Note'

export const metadata = { version: "1.0" }

# Main Title

Regular paragraph with text.

## Section One

<Note>
  Important note here.
</Note>

- List item one
- List item two
- List item three

> Blockquote text.

` + "```" + `javascript
function hello() {
  return "world";
}
` + "```" + `

---

## Section Two

1. First
2. Second
3. Third

Another paragraph here.

<Card title="Example" icon="code">
  Card content with **bold** and *italic*.
</Card>

### Subsection

Final paragraph.
`
}
