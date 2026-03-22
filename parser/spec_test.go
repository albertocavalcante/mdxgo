package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/albertocavalcante/mdxgo/syntax"
)

// specExample represents a single CommonMark spec test example.
type specExample struct {
	Markdown  string `json:"markdown"`
	HTML      string `json:"html"`
	Example   int    `json:"example"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	Section   string `json:"section"`
}

func loadSpecExamples(t *testing.T) []specExample {
	t.Helper()
	data, err := os.ReadFile("../testdata/spec.json")
	if err != nil {
		t.Skipf("spec.json not found: %v (run from repo root or download CommonMark spec)", err)
	}
	var examples []specExample
	if err := json.Unmarshal(data, &examples); err != nil {
		t.Fatalf("failed to parse spec.json: %v", err)
	}
	return examples
}

// TestSpecRoundTrip verifies the round-trip invariant for every CommonMark
// spec example: FullText(Parse(markdown)) == markdown.
//
// This is the single most important test in the suite. If a lossless parser
// cannot round-trip the spec examples, it is broken.
func TestSpecRoundTrip(t *testing.T) {
	examples := loadSpecExamples(t)
	if len(examples) == 0 {
		t.Fatal("spec.json has no examples")
	}
	t.Logf("Testing round-trip for %d CommonMark spec examples", len(examples))

	failed := 0
	for _, ex := range examples {
		name := fmt.Sprintf("example_%d_%s", ex.Example, ex.Section)
		t.Run(name, func(t *testing.T) {
			green := Parse([]byte(ex.Markdown), Options{MDX: false})
			got := syntax.FullText(green)
			if got != ex.Markdown {
				failed++
				// Show first divergence point for debugging.
				diverge := -1
				for i := 0; i < len(got) && i < len(ex.Markdown); i++ {
					if got[i] != ex.Markdown[i] {
						diverge = i
						break
					}
				}
				if diverge == -1 {
					diverge = min(len(got), len(ex.Markdown))
				}

				t.Errorf("round-trip failed for example %d (section: %s)\n"+
					"  input  (%d bytes): %q\n"+
					"  output (%d bytes): %q\n"+
					"  first divergence at byte %d",
					ex.Example, ex.Section,
					len(ex.Markdown), truncate(ex.Markdown, 200),
					len(got), truncate(got, 200),
					diverge,
				)
			}
		})
	}
}

// TestSpecDocumentRoot verifies that every spec example parses into
// a Document-rooted tree.
func TestSpecDocumentRoot(t *testing.T) {
	examples := loadSpecExamples(t)
	for _, ex := range examples {
		green := Parse([]byte(ex.Markdown), Options{MDX: false})
		if green.Kind != syntax.Document {
			t.Errorf("example %d: root kind = %v, want Document", ex.Example, green.Kind)
		}
	}
}

// TestSpecNonEmpty verifies that every non-empty spec example produces
// at least one child node/token.
func TestSpecNonEmpty(t *testing.T) {
	examples := loadSpecExamples(t)
	for _, ex := range examples {
		if ex.Markdown == "" {
			continue
		}
		green := Parse([]byte(ex.Markdown), Options{MDX: false})
		if len(green.Children) == 0 {
			t.Errorf("example %d: parsed to empty document from non-empty input %q",
				ex.Example, truncate(ex.Markdown, 80))
		}
	}
}

// TestSpecWidthConsistency verifies that every node's Width equals the
// sum of its children's widths.
func TestSpecWidthConsistency(t *testing.T) {
	examples := loadSpecExamples(t)
	for _, ex := range examples {
		green := Parse([]byte(ex.Markdown), Options{MDX: false})
		if err := checkWidthConsistency(green); err != nil {
			t.Errorf("example %d: width inconsistency: %v", ex.Example, err)
		}
	}
}

func checkWidthConsistency(n *syntax.GreenNode) error {
	computedWidth := 0
	for _, child := range n.Children {
		computedWidth += child.Width()
		if child.Node != nil {
			if err := checkWidthConsistency(child.Node); err != nil {
				return err
			}
		}
	}
	if computedWidth != n.Width {
		return fmt.Errorf("node %v: cached width %d != computed width %d",
			n.Kind, n.Width, computedWidth)
	}
	return nil
}

// TestSpecBySection groups spec examples by section and reports pass rates.
func TestSpecBySection(t *testing.T) {
	examples := loadSpecExamples(t)
	sections := make(map[string]struct{ total, passed int })

	for _, ex := range examples {
		green := Parse([]byte(ex.Markdown), Options{MDX: false})
		got := syntax.FullText(green)
		stats := sections[ex.Section]
		stats.total++
		if got == ex.Markdown {
			stats.passed++
		}
		sections[ex.Section] = stats
	}

	for section, stats := range sections {
		t.Logf("%-30s %3d/%3d passed", section, stats.passed, stats.total)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
