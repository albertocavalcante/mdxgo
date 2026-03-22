// Package syntax defines the core types for a lossless, trivia-aware
// Concrete Syntax Tree (CST) using the red-green tree pattern.
package syntax

// SyntaxKind identifies every possible node, token, and trivia in the tree.
// Trivia occupies 0–99, tokens 100–399, and composite nodes 400+.
type SyntaxKind uint16

// Trivia kinds (0–99).
const (
	WhitespaceTrivia  SyntaxKind = iota // spaces / tabs within a line
	EndOfLineTrivia                     // \n, \r\n, \r — blank lines between blocks
	LineCommentTrivia                   // HTML/JSX comments used as trivia
)

// Token kinds (100–399).
const (
	// --- generic ---
	TextToken      SyntaxKind = 100 // run of inline text (catch-all)
	EndOfFileToken SyntaxKind = 101
	NewLineToken   SyntaxKind = 102 // structural newline (line ending within a block)
	IndentToken    SyntaxKind = 103 // leading whitespace that determines block nesting
	BlankLineToken SyntaxKind = 104 // entirely blank line
	ErrorToken     SyntaxKind = 105 // unparseable content

	// --- ATX heading ---
	HashToken        SyntaxKind = 110 // the `#`… prefix
	HeadingTextToken SyntaxKind = 111 // heading content (inline text)

	// --- thematic break ---
	ThematicBreakToken SyntaxKind = 120 // `---`, `***`, `___`

	// --- fenced code ---
	FenceOpenToken  SyntaxKind = 130 // opening ``` or ~~~
	FenceCloseToken SyntaxKind = 131 // closing ``` or ~~~
	InfoStringToken SyntaxKind = 132 // language tag after opening fence
	CodeLineToken   SyntaxKind = 133 // line inside fenced code block

	// --- indented code ---
	IndentedCodeToken SyntaxKind = 140 // 4-space-indented code line

	// --- blockquote ---
	BlockQuoteMarker SyntaxKind = 150 // `>` character
	LazyLineToken    SyntaxKind = 151 // lazy continuation line

	// --- list ---
	BulletMarker  SyntaxKind = 160 // `-`, `+`, `*`
	OrderedMarker SyntaxKind = 161 // `1.`, `2)`, etc.

	// --- setext heading ---
	SetextUnderline SyntaxKind = 170 // `===` or `---` under a paragraph

	// --- link reference definition ---
	LinkLabelToken SyntaxKind = 180 // [label]
	LinkDestToken  SyntaxKind = 181 // URL destination
	LinkTitleToken SyntaxKind = 182 // "title"

	// --- HTML ---
	HTMLLineToken SyntaxKind = 190 // line of HTML block content

	// --- MDX frontmatter ---
	FrontmatterFence SyntaxKind = 200 // `---` or `+++` frontmatter delimiter
	FrontmatterLine  SyntaxKind = 201 // content line within frontmatter

	// --- MDX JSX ---
	JSXTextToken     SyntaxKind = 210 // raw JSX text
	JSXOpenAngle     SyntaxKind = 211 // <
	JSXCloseAngle    SyntaxKind = 212 // >
	JSXSlash         SyntaxKind = 213 // /
	JSXEquals        SyntaxKind = 214 // =
	JSXIdentifier    SyntaxKind = 215 // tag or attribute name
	JSXStringLiteral SyntaxKind = 216 // "value" or 'value'
	JSXDot           SyntaxKind = 217 // . in Member.Expression

	// --- MDX expression ---
	ExprOpenBrace    SyntaxKind = 220 // {
	ExprCloseBrace   SyntaxKind = 221 // }
	ExprContentToken SyntaxKind = 222 // JS expression content

	// --- MDX ESM ---
	ESMLineToken SyntaxKind = 230 // import/export line content

	// --- inline tokens (phase 3, reserved) ---
	BacktickToken     SyntaxKind = 300 // ` for code spans
	BackslashToken    SyntaxKind = 301 // \ for escapes
	AmpersandToken    SyntaxKind = 302 // & for entities
	StarToken         SyntaxKind = 303 // *
	UnderscoreToken   SyntaxKind = 304 // _
	OpenBracketToken  SyntaxKind = 305 // [
	CloseBracketToken SyntaxKind = 306 // ]
	OpenParenToken    SyntaxKind = 307 // (
	CloseParenToken   SyntaxKind = 308 // )
	ExclMarkToken     SyntaxKind = 309 // !
	HardBreakToken    SyntaxKind = 310 // trailing 2+ spaces or backslash before newline
	SoftBreakToken    SyntaxKind = 311 // newline in inline context
	AutolinkToken     SyntaxKind = 312 // <uri>
	RawHTMLToken      SyntaxKind = 313 // inline raw HTML
	EntityToken       SyntaxKind = 314 // &amp; etc
)

// Node kinds (400+). Each represents a non-terminal in the syntax tree.
const (
	Document          SyntaxKind = 400
	Paragraph         SyntaxKind = 401
	ATXHeading        SyntaxKind = 402
	SetextHeading     SyntaxKind = 403
	ThematicBreak     SyntaxKind = 404
	FencedCodeBlock   SyntaxKind = 405
	IndentedCodeBlock SyntaxKind = 406
	BlockQuote        SyntaxKind = 407
	BulletList        SyntaxKind = 408
	OrderedList       SyntaxKind = 409
	ListItem          SyntaxKind = 410
	HTMLBlock         SyntaxKind = 411
	LinkReferenceDef  SyntaxKind = 412
	BlankLineNode     SyntaxKind = 413
	ErrorNode         SyntaxKind = 414

	// MDX nodes
	Frontmatter      SyntaxKind = 450
	JSXBlock         SyntaxKind = 451
	JSXInline        SyntaxKind = 452
	ExpressionBlock  SyntaxKind = 453
	ExpressionInline SyntaxKind = 454
	ESMDeclaration   SyntaxKind = 455

	// JSX sub-structure
	JSXOpeningTag     SyntaxKind = 460
	JSXClosingTag     SyntaxKind = 461
	JSXSelfClosingTag SyntaxKind = 462
	JSXAttribute      SyntaxKind = 463
	JSXExprAttribute  SyntaxKind = 464
	JSXFragment       SyntaxKind = 465

	// Inline nodes (phase 3, reserved)
	EmphasisSpan    SyntaxKind = 500
	StrongSpan      SyntaxKind = 501
	CodeSpan        SyntaxKind = 502
	Link            SyntaxKind = 503
	Image           SyntaxKind = 504
	AutolinkSpan    SyntaxKind = 505
	RawHTMLSpan     SyntaxKind = 506
	HardLineBreak   SyntaxKind = 507
	SoftLineBreak   SyntaxKind = 508
	InlineText      SyntaxKind = 509
	BackslashEscape SyntaxKind = 510
	EntityRef       SyntaxKind = 511
)

// Kind range boundaries.
const (
	tokenKindStart SyntaxKind = 100
	nodeKindStart  SyntaxKind = 400
)

// IsTrivia reports whether the kind is a trivia kind.
func (k SyntaxKind) IsTrivia() bool { return k < tokenKindStart }

// IsToken reports whether the kind is a token kind.
func (k SyntaxKind) IsToken() bool { return k >= tokenKindStart && k < nodeKindStart }

// IsNode reports whether the kind is a composite node kind.
func (k SyntaxKind) IsNode() bool { return k >= nodeKindStart }

//go:generate stringer -type=SyntaxKind

// String returns a human-readable name for the kind.
func (k SyntaxKind) String() string {
	if s, ok := kindNames[k]; ok {
		return s
	}
	return "Unknown"
}

var kindNames = map[SyntaxKind]string{
	WhitespaceTrivia:  "WhitespaceTrivia",
	EndOfLineTrivia:   "EndOfLineTrivia",
	LineCommentTrivia: "LineCommentTrivia",

	TextToken:          "TextToken",
	EndOfFileToken:     "EndOfFileToken",
	NewLineToken:       "NewLineToken",
	IndentToken:        "IndentToken",
	BlankLineToken:     "BlankLineToken",
	ErrorToken:         "ErrorToken",
	HashToken:          "HashToken",
	HeadingTextToken:   "HeadingTextToken",
	ThematicBreakToken: "ThematicBreakToken",
	FenceOpenToken:     "FenceOpenToken",
	FenceCloseToken:    "FenceCloseToken",
	InfoStringToken:    "InfoStringToken",
	CodeLineToken:      "CodeLineToken",
	IndentedCodeToken:  "IndentedCodeToken",
	BlockQuoteMarker:   "BlockQuoteMarker",
	LazyLineToken:      "LazyLineToken",
	BulletMarker:       "BulletMarker",
	OrderedMarker:      "OrderedMarker",
	SetextUnderline:    "SetextUnderline",
	LinkLabelToken:     "LinkLabelToken",
	LinkDestToken:      "LinkDestToken",
	LinkTitleToken:     "LinkTitleToken",
	HTMLLineToken:      "HTMLLineToken",
	FrontmatterFence:   "FrontmatterFence",
	FrontmatterLine:    "FrontmatterLine",
	JSXTextToken:       "JSXTextToken",
	JSXOpenAngle:       "JSXOpenAngle",
	JSXCloseAngle:      "JSXCloseAngle",
	JSXSlash:           "JSXSlash",
	JSXEquals:          "JSXEquals",
	JSXIdentifier:      "JSXIdentifier",
	JSXStringLiteral:   "JSXStringLiteral",
	JSXDot:             "JSXDot",
	ExprOpenBrace:      "ExprOpenBrace",
	ExprCloseBrace:     "ExprCloseBrace",
	ExprContentToken:   "ExprContentToken",
	ESMLineToken:       "ESMLineToken",
	BacktickToken:      "BacktickToken",
	BackslashToken:     "BackslashToken",
	AmpersandToken:     "AmpersandToken",
	StarToken:          "StarToken",
	UnderscoreToken:    "UnderscoreToken",
	OpenBracketToken:   "OpenBracketToken",
	CloseBracketToken:  "CloseBracketToken",
	OpenParenToken:     "OpenParenToken",
	CloseParenToken:    "CloseParenToken",
	ExclMarkToken:      "ExclMarkToken",
	HardBreakToken:     "HardBreakToken",
	SoftBreakToken:     "SoftBreakToken",
	AutolinkToken:      "AutolinkToken",
	RawHTMLToken:       "RawHTMLToken",
	EntityToken:        "EntityToken",

	Document:          "Document",
	Paragraph:         "Paragraph",
	ATXHeading:        "ATXHeading",
	SetextHeading:     "SetextHeading",
	ThematicBreak:     "ThematicBreak",
	FencedCodeBlock:   "FencedCodeBlock",
	IndentedCodeBlock: "IndentedCodeBlock",
	BlockQuote:        "BlockQuote",
	BulletList:        "BulletList",
	OrderedList:       "OrderedList",
	ListItem:          "ListItem",
	HTMLBlock:         "HTMLBlock",
	LinkReferenceDef:  "LinkReferenceDef",
	BlankLineNode:     "BlankLineNode",
	ErrorNode:         "ErrorNode",
	Frontmatter:       "Frontmatter",
	JSXBlock:          "JSXBlock",
	JSXInline:         "JSXInline",
	ExpressionBlock:   "ExpressionBlock",
	ExpressionInline:  "ExpressionInline",
	ESMDeclaration:    "ESMDeclaration",
	JSXOpeningTag:     "JSXOpeningTag",
	JSXClosingTag:     "JSXClosingTag",
	JSXSelfClosingTag: "JSXSelfClosingTag",
	JSXAttribute:      "JSXAttribute",
	JSXExprAttribute:  "JSXExprAttribute",
	JSXFragment:       "JSXFragment",
	EmphasisSpan:      "EmphasisSpan",
	StrongSpan:        "StrongSpan",
	CodeSpan:          "CodeSpan",
	Link:              "Link",
	Image:             "Image",
	AutolinkSpan:      "AutolinkSpan",
	RawHTMLSpan:       "RawHTMLSpan",
	HardLineBreak:     "HardLineBreak",
	SoftLineBreak:     "SoftLineBreak",
	InlineText:        "InlineText",
	BackslashEscape:   "BackslashEscape",
	EntityRef:         "EntityRef",
}
