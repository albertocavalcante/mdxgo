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

**v0.3 — Full MDX parsing**

The current release covers block-level and inline CST parsing for CommonMark and MDX, a full tree modification API, and complete MDX extension support. All CommonMark block and inline constructs are supported with 100% spec compliance (652/652 examples). MDX extensions include frontmatter, ESM declarations, JSX blocks/inline, and expression blocks/inline.

Completed phases:
- **Phase 1** — Block-level parsing (all CommonMark blocks, MDX frontmatter/ESM)
- **Phase 2** — Inline parsing (emphasis, code spans, links, images, autolinks, raw HTML, escapes, entities, line breaks)
- **Phase 3** — MDX JSX (block and inline JSX tags with full attribute support: string, boolean, expression, spread)
- **Phase 4** — MDX expressions (block and inline brace-balanced expressions)
- **Phase 5** — Tree modification API (Cursor, Visitor, Walk, Query, Replace, Transform)

MDX JSX tags (block and inline) and MDX expressions (block and inline) are now parsed with full attribute support. See [Roadmap](#roadmap) for remaining work.

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
| `syntax` | Tree modification: `Cursor`, `Walk`, `Visitor`, `FindAll`, `FindFirst`, `ReplaceChild`, `InsertBefore`, `RemoveChild` |
| `parser` | `Parse(src, opts)` — two-pass parser: block-level then inline decomposition, with CommonMark and MDX mode |
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

### Cursor navigation

```go
green := mdx.Parse(src)
cursor := syntax.NewCursor(green)

// Move through the tree
cursor.FirstChild()  // descend to first child
cursor.NextSibling() // move to next sibling
cursor.Parent()      // move back up
```

### Walk and query

```go
green := mdx.Parse(src)

// Walk all nodes depth-first
syntax.Walk(green, func(n *syntax.GreenNode, depth int) syntax.WalkAction {
    fmt.Printf("%s at depth %d\n", n.Kind, depth)
    return syntax.Continue
})

// Find all headings
headings := syntax.FindAll(green, func(n *syntax.GreenNode) bool {
    return n.Kind == syntax.ATXHeading
})
```

### Tree modification

```go
green := mdx.Parse(src)

// Replace a child node (structural sharing — only the spine is copied)
newChild := syntax.TokenElement(syntax.NewGreenToken(syntax.TextToken, "new text\n"))
modified := syntax.ReplaceChild(green, []int{0}, 0, newChild)

// Insert / remove children
modified = syntax.InsertBefore(green, 0, newChild)
modified = syntax.RemoveChild(green, 1)
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
    InlineText [6]
      TextToken "World."
    SoftLineBreak [1]
      SoftBreakToken "\n"
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

| JSX block | — | ✓ |
| JSX inline | — | ✓ |
| Expression block | — | ✓ |
| Expression inline | — | ✓ |

Indented code blocks and HTML blocks are disabled in MDX mode per the [MDX spec](https://mdxjs.com/docs/what-is-mdx/#markdown).

## Inline Constructs

Inline content within paragraphs and headings is decomposed into structured nodes:

| Construct | Node Kind |
|-----------|-----------|
| `*emphasis*` / `_emphasis_` | EmphasisSpan |
| `**strong**` / `__strong__` | StrongSpan |
| `` `code` `` | CodeSpan |
| `[text](url)` / `[text][ref]` | Link |
| `![alt](url)` | Image |
| `<https://url>` | AutolinkSpan |
| `<tag>` (CommonMark only) | RawHTMLSpan |
| `\*` escaped | BackslashEscape |
| `&amp;` entity | EntityRef |
| `{expression}` (MDX only) | ExpressionInline |
| `<Tag />` (MDX only) | JSXInline |
| Hard line break | HardLineBreak |
| Soft line break | SoftLineBreak |

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

### Benchmarks

Run the full benchmark suite:

```bash
just bench                    # runs -count=3, writes bench.txt
go test -bench=. -benchmem ./parser  # single run
```

Compare against a previous baseline:

```bash
cp bench.txt bench-old.txt    # save current baseline
# ... make changes ...
just bench                    # new results in bench.txt
just bench-compare            # benchstat diff
```

Benchmarks can also be run via GitHub Actions (`Bench` workflow, manual dispatch).

#### Baseline (Apple M4, Go 1.26)

| Benchmark | Input | Throughput | Allocs/op |
|-----------|-------|-----------|-----------|
| ParseSmall | 45 B paragraph | 39 MB/s | 62 |
| ParseMedium | ~1.6 KB mixed | 70 MB/s | 999 |
| ParseLarge | ~22 KB doc | 59 MB/s | 16,140 |
| ParseCommonMarkSpec | 652 examples (~15 KB) | 20 MB/s | 31,271 |
| ParseJSXHeavy | ~5.3 KB synthetic | 40 MB/s | 5,239 |
| ParseInlineHeavy | ~8 KB emphasis/links | 40 MB/s | 8,672 |
| ParseJSXHeavyFixture | 1.5 KB testdata | 56 MB/s | 958 |
| ParseLargeMDXFixture | ~14 KB testdata | 40 MB/s | 11,820 |

| Benchmark | Metric | |
|-----------|--------|--|
| FullText (small) | 69 ns/op | 1 alloc |
| FullText (large) | 24 μs/op | 1 alloc |
| Walk (large tree) | 6 μs/op | 0 allocs |
| CountNodes | 7 μs/op | 0 allocs |
| FindAll | 8 μs/op | 8 allocs |
| CursorTraversal | 15 μs/op | 3 allocs |

**Analysis.** Parsing throughput ranges from 20–70 MB/s depending on construct density. The CommonMark spec benchmark is lower throughput because spec examples are tiny fragments with high per-parse overhead. Real documents (~1–20 KB) parse at 40–70 MB/s. FullText reconstruction is single-allocation regardless of tree size. Tree traversal via Walk is zero-alloc; Cursor has a fixed 3-allocation overhead for the stack. The inline-heavy benchmark shows inline parsing roughly doubles cost vs block-only content (55 μs → 177 μs for equivalent size), confirming that inline parsing is the primary optimization target for future work.

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

- [x] Block-level parsing — all CommonMark block constructs, 652/652 spec examples
- [x] Inline parsing — emphasis, code spans, links, images, delimiter stack
- [x] Tree modification API — `Cursor`, `Walk`, `Visitor`, `FindAll`, `ReplaceChild`, `InsertBefore`, `RemoveChild`
- [x] MDX expressions — brace-balanced expression blocks and inline expressions
- [x] MDX JSX blocks/inline — full JSX tag parsing with attribute support
- [ ] Performance — node interning, arena allocation, incremental reparsing
- [ ] LSP support — incremental edits, diagnostics integration

## License

[MIT](LICENSE)
