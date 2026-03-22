package mdx_test

import (
	"fmt"

	"github.com/albertocavalcante/mdxgo/mdx"
	"github.com/albertocavalcante/mdxgo/syntax"
)

func ExampleParse() {
	src := []byte("# Hello\n\nWorld.\n")
	green := mdx.Parse(src)
	fmt.Println(syntax.FullText(green) == string(src))
	// Output:
	// true
}

func ExampleParse_frontmatter() {
	src := []byte("---\ntitle: Hello\n---\n\n# Hello\n")
	green := mdx.Parse(src)

	root := syntax.NewSyntaxRoot(green)
	for _, child := range root.Children() {
		fmt.Println(child.Kind())
	}
	// Output:
	// Frontmatter
	// BlankLineNode
	// ATXHeading
}

func ExampleParseCommonMark() {
	src := []byte("# Heading\n\nParagraph text.\n")
	green := mdx.ParseCommonMark(src)
	fmt.Println(syntax.FullText(green) == string(src))
	// Output:
	// true
}
