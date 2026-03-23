package parser

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/albertocavalcante/mdxgo/syntax"
)

// ─── Fixture helpers ────────────────────────────────────────────────────────

func loadFixture(b *testing.B, path string) []byte {
	b.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		b.Fatalf("load fixture %s: %v", path, err)
	}
	return data
}

func loadSpecMarkdown(b *testing.B) []string {
	b.Helper()
	data, err := os.ReadFile("../testdata/spec.json")
	if err != nil {
		b.Skipf("spec.json not found: %v", err)
	}
	var examples []specExample
	if err := json.Unmarshal(data, &examples); err != nil {
		b.Fatalf("parse spec.json: %v", err)
	}
	out := make([]string, len(examples))
	for i, ex := range examples {
		out[i] = ex.Markdown
	}
	return out
}

// ─── Synthetic fixtures ─────────────────────────────────────────────────────

var (
	smallInput = []byte("Hello **world**, this is a `test` paragraph.\n")

	mediumInput = []byte(func() string {
		var sb strings.Builder
		sb.WriteString("# Getting Started\n\n")
		sb.WriteString("This is an **introduction** to the API. You can use `code spans` and [links](https://example.com).\n\n")
		sb.WriteString("## Installation\n\n")
		sb.WriteString("```bash\nnpm install @acme/sdk\n```\n\n")
		sb.WriteString("> **Note:** Make sure you have Node.js installed.\n\n")
		sb.WriteString("## Quick Start\n\n")
		for i := range 10 {
			sb.WriteString("- Step ")
			sb.WriteByte(byte('1' + i))
			sb.WriteString(": configure the *settings* for your **project**\n")
		}
		sb.WriteString("\n## Examples\n\n")
		for range 5 {
			sb.WriteString("Here is a paragraph with **bold**, *italic*, `code`, and a [link](https://example.com). ")
			sb.WriteString("It also contains some \\*escaped\\* characters and &amp; entities.\n\n")
		}
		return sb.String()
	}())

	largeInput = []byte(func() string {
		var sb strings.Builder
		sb.WriteString("# Large Document\n\n")
		for i := range 100 {
			sb.WriteString("## Section ")
			sb.WriteString(strings.Repeat("I", i+1)[:min(i+1, 5)])
			sb.WriteString("\n\n")
			sb.WriteString("Paragraph with **bold** and *italic* text and `code spans`.\n\n")
			sb.WriteString("- Item one with [link](https://example.com)\n")
			sb.WriteString("- Item two with `code`\n")
			sb.WriteString("- Item three with **bold**\n\n")
			sb.WriteString("> A blockquote with *emphasis*.\n\n")
			if i%5 == 0 {
				sb.WriteString("```go\nfunc example() {\n\treturn nil\n}\n```\n\n")
			}
		}
		return sb.String()
	}())

	inlineHeavyInput = []byte(func() string {
		var sb strings.Builder
		for range 50 {
			sb.WriteString("This **bold *and italic*** text has `code`, \\*escaped\\*, ")
			sb.WriteString("[links](url \"title\"), ![images](img.png), <https://auto.link>, ")
			sb.WriteString("&amp; entities, and hard  \nline breaks.\n\n")
		}
		return sb.String()
	}())

	jsxHeavyInput = []byte(func() string {
		var sb strings.Builder
		sb.WriteString("---\ntitle: JSX Bench\n---\n\n")
		sb.WriteString("import { Card, Note } from '@components'\n\n")
		for range 30 {
			sb.WriteString("<Card title=\"Test\" icon=\"code\" href=\"/api\">\n")
			sb.WriteString("  Paragraph with **bold** and {expression}.\n")
			sb.WriteString("</Card>\n\n")
			sb.WriteString("<Note>\n  Important note with `code`.\n</Note>\n\n")
			sb.WriteString("{items.map(x => `${x.name}`)}\n\n")
		}
		return sb.String()
	}())
)

// ─── Parse benchmarks ───────────────────────────────────────────────────────

func BenchmarkParseSmall(b *testing.B) {
	for b.Loop() {
		Parse(smallInput, Options{})
	}
	b.SetBytes(int64(len(smallInput)))
}

func BenchmarkParseMedium(b *testing.B) {
	for b.Loop() {
		Parse(mediumInput, Options{})
	}
	b.SetBytes(int64(len(mediumInput)))
}

func BenchmarkParseLarge(b *testing.B) {
	for b.Loop() {
		Parse(largeInput, Options{})
	}
	b.SetBytes(int64(len(largeInput)))
}

func BenchmarkParseCommonMarkSpec(b *testing.B) {
	examples := loadSpecMarkdown(b)
	// Concatenate all examples into one large input for throughput measurement.
	var total int
	inputs := make([][]byte, len(examples))
	for i, ex := range examples {
		inputs[i] = []byte(ex)
		total += len(ex)
	}
	b.SetBytes(int64(total))
	b.ResetTimer()
	for b.Loop() {
		for _, input := range inputs {
			Parse(input, Options{})
		}
	}
}

func BenchmarkParseJSXHeavy(b *testing.B) {
	for b.Loop() {
		Parse(jsxHeavyInput, Options{MDX: true})
	}
	b.SetBytes(int64(len(jsxHeavyInput)))
}

func BenchmarkParseInlineHeavy(b *testing.B) {
	for b.Loop() {
		Parse(inlineHeavyInput, Options{})
	}
	b.SetBytes(int64(len(inlineHeavyInput)))
}

func BenchmarkParseJSXHeavyFixture(b *testing.B) {
	data := loadFixture(b, "../testdata/mdx/jsx-heavy.mdx")
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for b.Loop() {
		Parse(data, Options{MDX: true})
	}
}

func BenchmarkParseLargeMDXFixture(b *testing.B) {
	data := loadFixture(b, "../testdata/stress/large-mdx.mdx")
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for b.Loop() {
		Parse(data, Options{MDX: true})
	}
}

// ─── Allocation benchmark ───────────────────────────────────────────────────

func BenchmarkParseAllocation(b *testing.B) {
	b.Run("Small", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			Parse(smallInput, Options{})
		}
	})
	b.Run("Medium", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			Parse(mediumInput, Options{})
		}
	})
	b.Run("Large", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			Parse(largeInput, Options{})
		}
	})
	b.Run("JSXHeavy", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			Parse(jsxHeavyInput, Options{MDX: true})
		}
	})
}

// ─── FullText reconstruction ────────────────────────────────────────────────

func BenchmarkFullText(b *testing.B) {
	b.Run("Small", func(b *testing.B) {
		tree := Parse(smallInput, Options{})
		b.ResetTimer()
		for b.Loop() {
			syntax.FullText(tree)
		}
	})
	b.Run("Medium", func(b *testing.B) {
		tree := Parse(mediumInput, Options{})
		b.ResetTimer()
		for b.Loop() {
			syntax.FullText(tree)
		}
	})
	b.Run("Large", func(b *testing.B) {
		tree := Parse(largeInput, Options{})
		b.ResetTimer()
		for b.Loop() {
			syntax.FullText(tree)
		}
	})
}

// ─── Tree traversal benchmarks ──────────────────────────────────────────────

func BenchmarkCursorTraversal(b *testing.B) {
	tree := Parse(largeInput, Options{})
	b.ResetTimer()
	for b.Loop() {
		c := syntax.NewCursor(tree)
		// Walk entire tree depth-first.
		for {
			if c.FirstChild() {
				continue
			}
			for !c.NextSibling() {
				if !c.Parent() {
					goto done
				}
			}
		}
	done:
	}
}

func BenchmarkWalk(b *testing.B) {
	tree := Parse(largeInput, Options{})
	b.ResetTimer()
	for b.Loop() {
		syntax.Walk(tree, func(_ *syntax.GreenNode, _ int) syntax.WalkAction {
			return syntax.Continue
		})
	}
}

func BenchmarkFindAll(b *testing.B) {
	tree := Parse(largeInput, Options{})
	b.ResetTimer()
	for b.Loop() {
		syntax.FindAll(tree, func(n *syntax.GreenNode) bool {
			return n.Kind == syntax.Paragraph
		})
	}
}

func BenchmarkFindAllHeadings(b *testing.B) {
	tree := Parse(largeInput, Options{})
	b.ResetTimer()
	for b.Loop() {
		syntax.FindAll(tree, func(n *syntax.GreenNode) bool {
			return n.Kind == syntax.ATXHeading
		})
	}
}

func BenchmarkCountNodes(b *testing.B) {
	tree := Parse(largeInput, Options{})
	b.ResetTimer()
	for b.Loop() {
		syntax.CountNodes(tree, func(n *syntax.GreenNode) bool {
			return n.Kind == syntax.Paragraph
		})
	}
}

// ─── Block-only vs full parse ───────────────────────────────────────────────

func BenchmarkBlockOnlyVsFull(b *testing.B) {
	// This benchmark compares parse cost with and without inline parsing
	// by parsing content that has no inline markup characters.
	plainInput := []byte(func() string {
		var sb strings.Builder
		for range 100 {
			sb.WriteString("This is a plain paragraph without any special characters at all\n\n")
		}
		return sb.String()
	}())

	b.Run("Plain", func(b *testing.B) {
		b.SetBytes(int64(len(plainInput)))
		for b.Loop() {
			Parse(plainInput, Options{})
		}
	})
	b.Run("InlineHeavy", func(b *testing.B) {
		b.SetBytes(int64(len(inlineHeavyInput)))
		for b.Loop() {
			Parse(inlineHeavyInput, Options{})
		}
	})
}
