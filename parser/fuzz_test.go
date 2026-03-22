package parser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/albertocavalcante/mdxgo/syntax"
)

// fuzzSeedsCommonMark returns a comprehensive seed corpus for CommonMark fuzzing.
func fuzzSeedsCommonMark() []string {
	return []string{
		// Empty / minimal
		"",
		"\n",
		"\r\n",
		"\r",
		" ",
		"\t",
		"  \n",
		"\n\n\n",

		// Paragraphs
		"hello\n",
		"hello world\n",
		"line1\nline2\n",
		"paragraph one\n\nparagraph two\n",
		"a\n\nb\n\nc\n",
		"short",                // no trailing newline
		"no newline at end",    // no trailing newline
		"  leading spaces\n",   // leading spaces in paragraph
		"trailing spaces   \n", // trailing spaces
		"word\nword\nword\n",   // multi-line paragraph

		// ATX headings — all levels, edge cases
		"# H1\n",
		"## H2\n",
		"### H3\n",
		"#### H4\n",
		"##### H5\n",
		"###### H6\n",
		"####### Not a heading\n", // 7 hashes
		"# \n",                    // empty heading
		"#\n",                     // hash without space (not heading)
		"# Heading ##\n",          // closing hashes
		"# Heading #####\n",       // many closing hashes
		"# Heading # \n",          // closing hash with trailing space
		" # Heading\n",            // 1-space indent
		"  ## Heading\n",          // 2-space indent
		"   ### Heading\n",        // 3-space indent
		"    # Not heading\n",     // 4-space indent = code
		"#  Two spaces\n",         // double space after hash
		"#\tTab after hash\n",     // tab after hash

		// Setext headings
		"Heading\n=======\n",
		"Heading\n-------\n",
		"Heading\n=\n",
		"Heading\n-\n",
		"Multi\nline\n===\n",
		"  Heading\n  =======\n",
		"   Heading\n   -------\n",

		// Thematic breaks — all variations
		"---\n",
		"***\n",
		"___\n",
		"- - -\n",
		"* * *\n",
		"_ _ _\n",
		"  ---\n",
		"   ---\n",
		"----------\n",
		"***********\n",
		"- -  - -\n",
		"  ***  \n", // trailing spaces

		// Fenced code blocks
		"```\ncode\n```\n",
		"```go\nfunc main() {}\n```\n",
		"~~~\ncode\n~~~\n",
		"````\ncode with ``` inside\n````\n",
		"```\n\n```\n", // empty code block
		"```\nline1\nline2\nline3\n```\n",
		"```\nunclosed fence\n", // unclosed
		"```info string here\ncode\n```\n",
		" ```\n code\n ```\n", // indented fence
		"  ```\n  code\n  ```\n",
		"   ```\n   code\n   ```\n",
		"```\n```\n",    // empty content
		"```\n \n```\n", // whitespace-only content
		"~~~python\ndef f():\n    pass\n~~~\n",
		"```\ttab in fence\ncode\n```\n",

		// Indented code blocks
		"    code line\n",
		"    line 1\n    line 2\n",
		"    line 1\n\n    line 2\n",
		"    a\n    b\n    c\n",
		"\tcode with tab\n",

		// Block quotes
		"> quote\n",
		"> line 1\n> line 2\n",
		"> # Heading in quote\n",
		">  two spaces\n",
		">no space\n",
		">\n", // empty quote line
		"> > nested\n",
		"> > > triple nested\n",
		">   indented content\n",
		">\ttab after marker\n",

		// Lists — bullet
		"- item\n",
		"- item 1\n- item 2\n",
		"- item 1\n- item 2\n- item 3\n",
		"* item\n",
		"+ item\n",
		"- \n", // empty item
		"-  two spaces\n",
		"-\ttab after bullet\n",
		" - indented marker\n",
		"  - two-space indent\n",
		"   - three-space indent\n",

		// Lists — ordered
		"1. item\n",
		"1. first\n2. second\n",
		"1. first\n2. second\n3. third\n",
		"1) paren style\n",
		"0. zero start\n",
		"10. large number\n",
		"999999999. max digits\n",

		// HTML blocks
		"<div>\nfoo\n</div>\n",
		"<!-- comment -->\n",
		"<?php echo 'hi'; ?>\n",
		"<!DOCTYPE html>\n",
		"<![CDATA[\ndata\n]]>\n",
		"<pre>\ncode\n</pre>\n",
		"<script>\njs();\n</script>\n",
		"<style>\n.x{}\n</style>\n",
		"<table>\n<tr><td>cell</td></tr>\n</table>\n",
		"<div>\n\n</div>\n",

		// Mixed constructs
		"# Title\n\nParagraph\n\n---\n",
		"# H1\n\n## H2\n\n### H3\n",
		"- a\n\n- b\n", // loose list
		"> quote\n\nparagraph\n",
		"```\ncode\n```\n\nparagraph\n",
		"para\n\n---\n\npara\n",
		"# Title\n\n- a\n- b\n\n> quote\n\n---\n\n```\ncode\n```\n",

		// Line ending variations
		"hello\r\nworld\r\n",
		"hello\rworld\r",
		"a\nb\r\nc\rd\n",
		"\r\n",
		"\r",

		// Whitespace edge cases
		"\t\n",
		"  \t  \n",
		"\t\t\t\n",
		" \t \t \n",

		// Unicode
		"# 日本語\n",
		"こんにちは\n",
		"🎉\n",
		"é à ü ö\n",
		"# Heading with emoji 🚀\n",

		// Adversarial / pathological
		strings.Repeat("#", 100) + "\n",
		strings.Repeat("-", 100) + "\n",
		strings.Repeat(">", 50) + " deep\n",
		strings.Repeat("- ", 50) + "item\n",
		strings.Repeat("```\n", 10),
		strings.Repeat("  ", 100) + "deep indent\n",
		strings.Repeat("\n", 100),
		strings.Repeat("a", 10000) + "\n",
		"# " + strings.Repeat("x", 5000) + "\n",
	}
}

// fuzzSeedsMDX returns MDX-specific seed corpus.
func fuzzSeedsMDX() []string {
	return append(fuzzSeedsCommonMark(), []string{
		// Frontmatter
		"---\ntitle: test\n---\n",
		"---\n---\n",
		"---\na: 1\nb: 2\nc: 3\n---\n",
		"---\ntitle: \"quoted\"\n---\n\n# Heading\n",
		"---\n" + strings.Repeat("key: value\n", 50) + "---\n",
		"+++\ntitle = \"toml\"\n+++\n",

		// ESM
		"import Foo from 'foo'\n",
		"import { Bar } from 'bar'\n",
		"import * as Baz from 'baz'\n",
		"export const x = 1\n",
		"export default Layout\n",
		"import {\n  A,\n  B,\n  C,\n} from 'mod'\n",
		"export const config = {\n  key: 'value',\n}\n",

		// Frontmatter + ESM + content
		"---\ntitle: Page\n---\n\nimport X from 'x'\n\n# Hello\n\nContent.\n",

		// JSX-like content (treated as paragraphs in block-only mode)
		"<Component />\n",
		"<Component prop=\"value\" />\n",
		"<Wrapper>\n  content\n</Wrapper>\n",
		"<A>\n<B>\n<C />\n</B>\n</A>\n",
		"<Note>\n  **Bold** in JSX\n</Note>\n",
		"<Card title=\"Test\" icon={icon}>\n  Body\n</Card>\n",

		// Expressions (treated as paragraphs in block-only mode)
		"{expression}\n",
		"{1 + 2}\n",
		"{variable}\n",
		"{`template ${string}`}\n",

		// Mixed MDX
		"---\ntitle: Full\n---\n\nimport { A } from 'a'\n\nexport const x = 1\n\n# Title\n\nPara.\n\n- List\n\n> Quote\n\n```js\ncode()\n```\n",
	}...)
}

func FuzzRoundTrip(f *testing.F) {
	for _, s := range fuzzSeedsCommonMark() {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, src string) {
		green := Parse([]byte(src), Options{})
		got := syntax.FullText(green)
		if got != src {
			t.Errorf("round-trip failed:\ninput  (%d bytes): %q\noutput (%d bytes): %q",
				len(src), truncate(src, 100), len(got), truncate(got, 100))
		}
	})
}

func FuzzRoundTripMDX(f *testing.F) {
	for _, s := range fuzzSeedsMDX() {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, src string) {
		green := Parse([]byte(src), Options{MDX: true})
		got := syntax.FullText(green)
		if got != src {
			t.Errorf("MDX round-trip failed:\ninput  (%d bytes): %q\noutput (%d bytes): %q",
				len(src), truncate(src, 100), len(got), truncate(got, 100))
		}
	})
}

// FuzzWidthConsistency verifies that node widths stay consistent under fuzzing.
func FuzzWidthConsistency(f *testing.F) {
	for _, s := range fuzzSeedsCommonMark() {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, src string) {
		green := Parse([]byte(src), Options{})
		if green.Width != len(src) {
			t.Errorf("root width %d != input length %d", green.Width, len(src))
		}
		if err := checkWidthConsistency(green); err != nil {
			t.Errorf("width inconsistency: %v", err)
		}
	})
}

// FuzzDoubleParseIdempotent verifies parse(fulltext(parse(src))) == parse(src)
// at the output level.
func FuzzDoubleParseIdempotent(f *testing.F) {
	for _, s := range fuzzSeedsCommonMark() {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, src string) {
		green1 := Parse([]byte(src), Options{})
		out1 := syntax.FullText(green1)
		green2 := Parse([]byte(out1), Options{})
		out2 := syntax.FullText(green2)
		if out1 != out2 {
			t.Errorf("double-parse not idempotent:\n  first:  %q\n  second: %q", truncate(out1, 100), truncate(out2, 100))
		}
	})
}

// FuzzBothModes verifies that parsing in both CommonMark and MDX mode
// doesn't panic and always round-trips.
func FuzzBothModes(f *testing.F) {
	for _, s := range fuzzSeedsCommonMark() {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, src string) {
		for _, mdx := range []bool{false, true} {
			green := Parse([]byte(src), Options{MDX: mdx})
			got := syntax.FullText(green)
			if got != src {
				mode := "CommonMark"
				if mdx {
					mode = "MDX"
				}
				t.Errorf("%s round-trip failed:\ninput  (%d bytes): %q\noutput (%d bytes): %q",
					mode, len(src), truncate(src, 100), len(got), truncate(got, 100))
			}
		}
	})
}

// FuzzFromCorpus seeds the fuzzer from corpus files if available.
func FuzzFromCorpus(f *testing.F) {
	// Seed from testdata files.
	dirs := []string{"../testdata/mdx", "../testdata/commonmark", "../testdata/stress"}
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				continue
			}
			f.Add(string(data))
		}
	}

	// Also seed from spec.json if available.
	if data, err := os.ReadFile("../testdata/spec.json"); err == nil {
		type specEx struct {
			Markdown string `json:"markdown"`
		}
		var examples []specEx
		if err := json.Unmarshal(data, &examples); err == nil {
			for _, ex := range examples {
				f.Add(ex.Markdown)
			}
		}
	}

	f.Fuzz(func(t *testing.T, src string) {
		green := Parse([]byte(src), Options{MDX: true})
		got := syntax.FullText(green)
		if got != src {
			t.Errorf("corpus-fuzz round-trip failed:\ninput  (%d bytes): %q\noutput (%d bytes): %q",
				len(src), truncate(src, 100), len(got), truncate(got, 100))
		}
	})
}
