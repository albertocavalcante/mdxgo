package parser

// CommonMark spec-defined constants.
const (
	// maxIndent is the maximum indentation (in spaces) before a construct
	// is treated as an indented code block instead of the intended construct.
	maxIndent = 3

	// indentedCodeIndent is the minimum number of leading spaces for an
	// indented code block.
	indentedCodeIndent = 4

	// maxHeadingLevel is the maximum ATX heading level (######).
	maxHeadingLevel = 6

	// minFenceLength is the minimum number of backticks or tildes for a
	// fenced code block opener/closer.
	minFenceLength = 3

	// minThematicBreakChars is the minimum number of -, *, or _ characters
	// for a thematic break.
	minThematicBreakChars = 3

	// bulletMarkerWidth is the width of a bullet marker plus its trailing space.
	bulletMarkerWidth = 2
)
