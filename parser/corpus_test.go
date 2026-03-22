package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/albertocavalcante/mdxgo/syntax"
)

// TestCorpusMDX runs round-trip tests against all .mdx files in testdata/mdx/.
func TestCorpusMDX(t *testing.T) {
	runCorpus(t, "../testdata/mdx", "*.mdx", Options{MDX: true})
}

// TestCorpusCommonMark runs round-trip tests against all .md files in testdata/commonmark/.
func TestCorpusCommonMark(t *testing.T) {
	runCorpus(t, "../testdata/commonmark", "*.md", Options{MDX: false})
}

// TestCorpusStressMD runs round-trip tests against .md stress test files.
func TestCorpusStressMD(t *testing.T) {
	runCorpus(t, "../testdata/stress", "*.md", Options{MDX: false})
}

// TestCorpusStressMDX runs round-trip tests against .mdx stress test files.
func TestCorpusStressMDX(t *testing.T) {
	runCorpus(t, "../testdata/stress", "*.mdx", Options{MDX: true})
}

// runCorpus scans a directory for files matching a glob pattern, reads each
// file, and asserts the round-trip invariant.
func runCorpus(t *testing.T, dir, pattern string, opts Options) {
	t.Helper()

	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		t.Fatalf("glob error: %v", err)
	}
	if len(matches) == 0 {
		t.Skipf("no files matching %s in %s", pattern, dir)
	}

	t.Logf("Testing %d files in %s", len(matches), dir)

	for _, path := range matches {
		name := filepath.Base(path)
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read %s: %v", path, err)
			}

			src := string(data)
			green := Parse(data, opts)

			// 1. Round-trip invariant.
			got := syntax.FullText(green)
			if got != src {
				diverge := firstDivergence(src, got)
				t.Errorf("round-trip failed for %s:\n"+
					"  input  length: %d bytes\n"+
					"  output length: %d bytes\n"+
					"  first divergence at byte %d\n"+
					"  context: input[%d:%d]=%q vs output[%d:%d]=%q",
					name, len(src), len(got), diverge,
					max(0, diverge-20), min(len(src), diverge+20),
					safeSlice(src, max(0, diverge-20), min(len(src), diverge+20)),
					max(0, diverge-20), min(len(got), diverge+20),
					safeSlice(got, max(0, diverge-20), min(len(got), diverge+20)),
				)
			}

			// 2. Document root.
			if green.Kind != syntax.Document {
				t.Errorf("root kind = %v, want Document", green.Kind)
			}

			// 3. Width consistency.
			if err := checkWidthConsistency(green); err != nil {
				t.Errorf("width inconsistency: %v", err)
			}

			// 4. Non-empty input produces non-empty tree.
			if len(data) > 0 && len(green.Children) == 0 {
				t.Errorf("non-empty input produced empty document")
			}

			// 5. Total width equals input length.
			if green.Width != len(data) {
				t.Errorf("root width %d != input length %d", green.Width, len(data))
			}

			// 6. No node has zero width (except possibly ErrorNode).
			checkNoZeroWidthNodes(t, green, name)
		})
	}
}

// checkNoZeroWidthNodes recursively checks that no non-error node has zero width.
func checkNoZeroWidthNodes(t *testing.T, n *syntax.GreenNode, file string) {
	t.Helper()
	for _, child := range n.Children {
		if child.Node != nil {
			if child.Node.Width == 0 && child.Node.Kind != syntax.ErrorNode {
				t.Errorf("%s: zero-width node of kind %v", file, child.Node.Kind)
			}
			checkNoZeroWidthNodes(t, child.Node, file)
		}
	}
}

func firstDivergence(a, b string) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return n
}

func safeSlice(s string, lo, hi int) string {
	if lo < 0 {
		lo = 0
	}
	if hi > len(s) {
		hi = len(s)
	}
	if lo >= hi {
		return ""
	}
	return s[lo:hi]
}

// TestCorpusParseDoesNotPanic runs Parse on all corpus files and verifies
// no panics occur, even on potentially adversarial input.
func TestCorpusParseDoesNotPanic(t *testing.T) {
	dirs := []struct {
		path    string
		pattern string
		opts    Options
	}{
		{"../testdata/mdx", "*.mdx", Options{MDX: true}},
		{"../testdata/commonmark", "*.md", Options{MDX: false}},
		{"../testdata/stress", "*.md", Options{MDX: false}},
		{"../testdata/stress", "*.mdx", Options{MDX: true}},
	}

	for _, d := range dirs {
		matches, _ := filepath.Glob(filepath.Join(d.path, d.pattern))
		for _, path := range matches {
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}

			// Run in both modes — neither should panic.
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Parse panicked on %s: %v", path, r)
					}
				}()
				Parse(data, d.opts)
			}()

			// Also try MDX mode on .md files and vice versa.
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Parse panicked on %s (opposite mode): %v", path, r)
					}
				}()
				Parse(data, Options{MDX: !d.opts.MDX})
			}()
		}
	}
}

// TestCorpusDoubleParseIdempotent verifies that parsing the round-tripped
// output produces the exact same tree width (structural idempotency).
func TestCorpusDoubleParseIdempotent(t *testing.T) {
	dirs := []struct {
		path    string
		pattern string
		opts    Options
	}{
		{"../testdata/mdx", "*.mdx", Options{MDX: true}},
		{"../testdata/commonmark", "*.md", Options{MDX: false}},
		{"../testdata/stress", "*.md", Options{MDX: false}},
		{"../testdata/stress", "*.mdx", Options{MDX: true}},
	}

	for _, d := range dirs {
		matches, _ := filepath.Glob(filepath.Join(d.path, d.pattern))
		for _, path := range matches {
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			name := filepath.Base(path)
			t.Run(name, func(t *testing.T) {
				green1 := Parse(data, d.opts)
				output1 := syntax.FullText(green1)

				green2 := Parse([]byte(output1), d.opts)
				output2 := syntax.FullText(green2)

				if output1 != output2 {
					t.Errorf("double parse not idempotent for %s: first=%d bytes, second=%d bytes",
						name, len(output1), len(output2))
				}
			})
		}
	}
}

// TestCorpusTreeDepth sanity-checks that tree depth is reasonable
// (no runaway recursion in parser).
func TestCorpusTreeDepth(t *testing.T) {
	dirs := []struct {
		path    string
		pattern string
		opts    Options
	}{
		{"../testdata/mdx", "*.mdx", Options{MDX: true}},
		{"../testdata/commonmark", "*.md", Options{MDX: false}},
		{"../testdata/stress", "*", Options{MDX: false}},
	}

	for _, d := range dirs {
		matches, _ := filepath.Glob(filepath.Join(d.path, d.pattern))
		for _, path := range matches {
			if strings.HasSuffix(path, ".json") {
				continue
			}
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			name := filepath.Base(path)
			t.Run(name, func(t *testing.T) {
				green := Parse(data, d.opts)
				depth := maxTreeDepth(green, 0)
				// Reasonable max depth — if this fires, there's a recursion bug.
				if depth > 100 {
					t.Errorf("%s: tree depth %d exceeds 100, possible runaway nesting", name, depth)
				}
			})
		}
	}
}

func maxTreeDepth(n *syntax.GreenNode, current int) int {
	maxD := current
	for _, child := range n.Children {
		if child.Node != nil {
			d := maxTreeDepth(child.Node, current+1)
			if d > maxD {
				maxD = d
			}
		}
	}
	return maxD
}
