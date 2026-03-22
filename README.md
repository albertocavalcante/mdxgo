# mdxgo

A lossless, trivia-aware Concrete Syntax Tree (CST) parser for [MDX v3](https://mdxjs.com/) and [CommonMark](https://commonmark.org/) in Go, following the red-green tree architecture from [Roslyn](https://github.com/dotnet/roslyn) and [SwiftSyntax](https://github.com/swiftlang/swift-syntax).

```go
package main

import (
	"fmt"
	"github.com/albertocavalcante/mdxgo/mdx"
	"github.com/albertocavalcante/mdxgo/syntax"
)

func main() {
	src := []byte("# Hello\n\nWorld.\n")
	green := mdx.Parse(src)
	fmt.Println(syntax.FullText(green) == string(src)) // true — always
}
```

## Status

**v0.1 — Block-level parsing**

The current release covers **block-level** CST parsing for CommonMark and MDX. All CommonMark block constructs are supported with 100% spec compliance (652/652 examples). MDX extensions include frontmatter and ESM declarations.

Inline parsing (emphasis, code spans, links, images), full JSX tag parsing, MDX expressions, and tree modification APIs are planned for future releases — see [Roadmap](#roadmap).

## Why

Existing Go Markdown parsers (goldmark, blackfriday) produce ASTs that discard whitespace, trivia, and exact source positions. This makes them unsuitable for:

- **Source-to-source transforms** that must preserve formatting
- **LSP servers** that need precise byte offsets for diagnostics and completions
- **Structural refactoring** tools that modify specific nodes without touching the rest
- **MDX tooling** that needs to handle JSX, expressions, ESM, and frontmatter

mdxgo solves this by parsing into a lossless CST where `FullText(Parse(src)) == src` for *all* input, including malformed files.

## Architecture

### Red-Green Tree

The tree has two layers:

**Green nodes** are the immutable data tree. They store relative widths (not positions), have no parent pointers, and are constructed bottom-up. Structurally identical subtrees can be shared across edits.

**Red nodes** are ephemeral wrappers manufactured on-the-fly during traversal. They provide parent pointers and absolute byte positions computed from green widths. They cost nothing when you don't traverse.

```
Green (data)              Red (navigation)
─────────────             ────────────────
GreenNode                 SyntaxNode
  kind: Document            green: →GreenNode
  width: 16                 parent: nil
  children: [...]           offset: 0
                            ↓ .Children() manufactures red wrappers on demand
```

### Trivia

Whitespace that doesn't affect structure is attached to tokens as leading or trailing **trivia** — spaces between tokens, blank lines between blocks, etc.

Whitespace that *does* affect structure — indentation that determines block nesting, `>` blockquote prefixes, `##` heading prefixes — is part of the token text.

| Pattern | Classification |
|---------|---------------|
| Indentation determining block structure | Token text |
| `>` blockquote prefix | Token (trailing trivia: space) |
| `##` heading prefix | Token (trailing trivia: space) |
| Blank lines between blocks | Leading trivia on next token |
| Trailing 0–1 spaces on a line | Trailing trivia |
| Tabs | Preserved as-is (never expanded) |

## Packages

| Package | Description |
|---------|-------------|
| `syntax` | Core types: `SyntaxKind`, `GreenNode`, `GreenToken`, `SyntaxNode`, `TriviaList`, `FullText` |
| `parser` | `Parse(src, opts)` — block-level parser with CommonMark and MDX mode |
| `mdx` | Convenience entry points: `mdx.Parse(src)`, `mdx.ParseCommonMark(src)` |
| `position` | `LineMap` for offset ↔ line:col conversion |

## Usage

### Parse MDX

```go
green := mdx.Parse(src)
```

### Parse CommonMark

```go
green := mdx.ParseCommonMark(src)
// or
green := parser.Parse(src, parser.Options{MDX: false})
```

### Round-trip

```go
green := mdx.Parse(src)
output := syntax.FullText(green) // output == string(src), always
```

### Navigate the tree

```go
green := mdx.Parse(src)
root := syntax.NewSyntaxRoot(green)

for i, child := range root.Children() {
    fmt.Printf("child %d: %s at offset %d\n", i, child.Kind(), child.Offset())
}

// Find token at a byte offset
tok := root.TokenAt(42)
fmt.Printf("token: %s %q at %d\n", tok.Kind(), tok.Text(), tok.TextOffset())
```

### Debug dump

```go
fmt.Println(syntax.DebugDump(green))
```

```
Document [45]
  ATXHeading [8]
    HashToken "#" (lead=0, trail=1)
    HeadingTextToken "Hello"
    NewLineToken "\n"
  BlankLineNode [1]
    BlankLineToken "\n"
  Paragraph [7]
    TextToken "World.\n"
```

## Block-Level Constructs

The parser handles all CommonMark block constructs plus MDX extensions:

| Construct | CommonMark | MDX |
|-----------|:---:|:---:|
| ATX heading | ✓ | ✓ |
| Setext heading | ✓ | ✓ |
| Thematic break | ✓ | ✓ |
| Fenced code block | ✓ | ✓ |
| Indented code block | ✓ | — |
| Block quote | ✓ | ✓ |
| Bullet list | ✓ | ✓ |
| Ordered list | ✓ | ✓ |
| HTML block | ✓ | — |
| Paragraph | ✓ | ✓ |
| Frontmatter | — | ✓ |
| ESM (import/export) | — | ✓ |

Indented code blocks and HTML blocks are disabled in MDX mode per the [MDX spec](https://mdxjs.com/docs/what-is-mdx/#markdown).

## Testing

```bash
go test ./...                               # all tests (~2300 cases)
go test -v -run TestSpecBySection ./parser  # CommonMark spec pass rates
go test -v -run TestAdversarial ./parser    # pathological inputs
go test -v -run TestCorpus ./parser         # testdata fixtures
go test -v -run TestLiveDoczel ./parser     # doczel MDX files (if available)
```

### Fuzz testing

```bash
go test -fuzz='^FuzzRoundTrip$' -fuzztime=60s ./parser
go test -fuzz='^FuzzRoundTripMDX$' -fuzztime=60s ./parser
go test -fuzz='^FuzzWidthConsistency$' -fuzztime=60s ./parser
go test -fuzz='^FuzzDoubleParseIdempotent$' -fuzztime=60s ./parser
go test -fuzz='^FuzzBothModes$' -fuzztime=60s ./parser
go test -fuzz='^FuzzFromCorpus$' -fuzztime=60s ./parser   # seeds from spec.json + testdata
```

### Live repo testing

Clone real-world repos and round-trip every MDX/Markdown file:

```bash
MDXGO_LIVE_REPOS=1 go test -v -run TestLiveRepos -timeout 10m ./parser
```

Test against a local directory:

```bash
MDXGO_LOCAL_CORPUS=~/docs go test -v -run TestLiveRepoLocal ./parser
```

Debug a single file with tree dump:

```bash
MDXGO_TEST_FILE=path/to/file.mdx go test -v -run TestLiveSingleFile ./parser
```

## CommonMark Spec Compliance

The parser round-trips all 652 examples from the [CommonMark 0.31.2 spec](https://spec.commonmark.org/0.31.2/):

```
Tabs                            11/ 11 passed
Backslash escapes               13/ 13 passed
Entity and numeric character references  17/ 17 passed
Precedence                       1/  1 passed
Thematic breaks                 19/ 19 passed
ATX headings                    18/ 18 passed
Setext headings                 27/ 27 passed
Indented code blocks            12/ 12 passed
Fenced code blocks              29/ 29 passed
HTML blocks                     44/ 44 passed
Link reference definitions      27/ 27 passed
Paragraphs                       8/  8 passed
Blank lines                      1/  1 passed
Block quotes                    25/ 25 passed
List items                      48/ 48 passed
Lists                           26/ 26 passed
Inlines                          1/  1 passed
Code spans                      22/ 22 passed
Emphasis and strong emphasis   132/132 passed
Links                           90/ 90 passed
Images                          22/ 22 passed
Autolinks                       19/ 19 passed
Raw HTML                        20/ 20 passed
Hard line breaks                15/ 15 passed
Soft line breaks                 2/  2 passed
Textual content                  3/  3 passed
```

## Roadmap

- [ ] Inline parsing — emphasis, code spans, links, images, delimiter stack
- [ ] MDX JSX blocks/inline — full JSX tag parsing with attribute support
- [ ] MDX expressions — brace-balanced JS expression blocks
- [ ] Tree modification API — `ReplaceNode`, `ReplaceDescendant`, `Visitor`, `Cursor`
- [ ] Performance — node interning, arena allocation, incremental reparsing
- [ ] LSP support — incremental edits, diagnostics integration

## License

[MIT](LICENSE)
