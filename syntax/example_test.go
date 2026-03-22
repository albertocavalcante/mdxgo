package syntax_test

import (
	"fmt"

	"github.com/albertocavalcante/mdxgo/syntax"
)

func ExampleFullText() {
	// Build a small green tree manually.
	heading := syntax.NewGreenNode(syntax.ATXHeading, []syntax.GreenElement{
		syntax.TokenElement(syntax.NewGreenTokenTrivia(
			syntax.HashToken,
			syntax.TriviaList{},
			"#",
			syntax.NewTriviaList(syntax.Trivia{Kind: syntax.WhitespaceTrivia, Text: " "}),
		)),
		syntax.TokenElement(syntax.NewGreenToken(syntax.HeadingTextToken, "Hello")),
		syntax.TokenElement(syntax.NewGreenToken(syntax.NewLineToken, "\n")),
	})
	doc := syntax.NewGreenNode(syntax.Document, []syntax.GreenElement{
		syntax.NodeElement(heading),
	})

	fmt.Print(syntax.FullText(doc))
	// Output:
	// # Hello
}

func ExampleDebugDump() {
	tok := syntax.NewGreenToken(syntax.TextToken, "hello")
	para := syntax.NewGreenNode(syntax.Paragraph, []syntax.GreenElement{
		syntax.TokenElement(tok),
	})
	doc := syntax.NewGreenNode(syntax.Document, []syntax.GreenElement{
		syntax.NodeElement(para),
	})

	fmt.Print(syntax.DebugDump(doc))
	// Output:
	// Document [5]
	//   Paragraph [5]
	//     TextToken "hello"
}

func ExampleNewSyntaxRoot() {
	tok := syntax.NewGreenToken(syntax.TextToken, "hi\n")
	doc := syntax.NewGreenNode(syntax.Document, []syntax.GreenElement{
		syntax.TokenElement(tok),
	})

	root := syntax.NewSyntaxRoot(doc)
	fmt.Println(root.Kind())
	fmt.Println(root.Offset())
	fmt.Println(root.Width())
	// Output:
	// Document
	// 0
	// 3
}

func ExampleSyntaxNode_Children() {
	heading := syntax.NewGreenNode(syntax.ATXHeading, []syntax.GreenElement{
		syntax.TokenElement(syntax.NewGreenToken(syntax.HashToken, "#")),
		syntax.TokenElement(syntax.NewGreenToken(syntax.NewLineToken, "\n")),
	})
	doc := syntax.NewGreenNode(syntax.Document, []syntax.GreenElement{
		syntax.NodeElement(heading),
	})

	root := syntax.NewSyntaxRoot(doc)
	for i, child := range root.Children() {
		fmt.Printf("child %d: %s\n", i, child.Kind())
	}
	// Output:
	// child 0: ATXHeading
}

func ExampleSyntaxNode_TokenAt() {
	tok1 := syntax.NewGreenToken(syntax.HashToken, "# ")
	tok2 := syntax.NewGreenToken(syntax.HeadingTextToken, "Hi")
	tok3 := syntax.NewGreenToken(syntax.NewLineToken, "\n")
	heading := syntax.NewGreenNode(syntax.ATXHeading, []syntax.GreenElement{
		syntax.TokenElement(tok1),
		syntax.TokenElement(tok2),
		syntax.TokenElement(tok3),
	})
	doc := syntax.NewGreenNode(syntax.Document, []syntax.GreenElement{
		syntax.NodeElement(heading),
	})

	root := syntax.NewSyntaxRoot(doc)
	found := root.TokenAt(2) // byte offset 2 → inside "Hi"
	fmt.Printf("%s %q\n", found.Kind(), found.Text())
	// Output:
	// HeadingTextToken "Hi"
}

func ExampleGreenNode_ReplaceChild() {
	old := syntax.NewGreenToken(syntax.TextToken, "old")
	doc := syntax.NewGreenNode(syntax.Document, []syntax.GreenElement{
		syntax.TokenElement(old),
	})

	replacement := syntax.NewGreenToken(syntax.TextToken, "new")
	doc2 := doc.ReplaceChild(0, syntax.TokenElement(replacement))

	fmt.Println(syntax.FullText(doc))  // original unchanged
	fmt.Println(syntax.FullText(doc2)) // new version
	// Output:
	// old
	// new
}

func ExampleTriviaList() {
	tl := syntax.NewTriviaList(
		syntax.Trivia{Kind: syntax.WhitespaceTrivia, Text: "  "},
		syntax.Trivia{Kind: syntax.EndOfLineTrivia, Text: "\n"},
	)
	fmt.Println(tl.Len())
	fmt.Println(tl.Width())
	fmt.Printf("%q\n", tl.Text())
	// Output:
	// 2
	// 3
	// "  \n"
}
