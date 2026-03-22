package mdx

import (
	"testing"

	"github.com/albertocavalcante/mdxgo/syntax"
)

func TestParseMDX(t *testing.T) {
	src := `---
title: My Page
---

import { Card } from '@components/Card'

# Hello World

Some text here.
`
	green := Parse([]byte(src))
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed:\ninput:  %q\noutput: %q", src, got)
	}
	if green.Kind != syntax.Document {
		t.Errorf("root kind = %v, want Document", green.Kind)
	}
}

func TestParseCommonMark(t *testing.T) {
	src := "# Heading\n\nParagraph\n"
	green := ParseCommonMark([]byte(src))
	got := syntax.FullText(green)
	if got != src {
		t.Errorf("round-trip failed:\ninput:  %q\noutput: %q", src, got)
	}
}
