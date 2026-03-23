package parser

// inlineMarkupChars is a pre-computed lookup table for characters that
// can start an inline construct, used by containsInlineMarkup.
//
//nolint:gochecknoglobals // lookup table is intentionally global for performance
var inlineMarkupChars = func() [256]bool {
	var t [256]bool
	for _, b := range []byte{'`', '\\', '&', '*', '_', '[', ']', '!', '<', '{', '\n', '\r'} {
		t[b] = true
	}
	return t
}()

// inlineScanner provides character-level scanning over inline content
// that may span multiple tokens (e.g., multi-line paragraphs).
// It operates on a concatenated view of the text but tracks which
// original token each character belongs to for proper round-trip.
type inlineScanner struct {
	text string // concatenated inline content
	pos  int    // current position in text
}

// newInlineScanner creates a scanner over the given inline text.
func newInlineScanner(text string) *inlineScanner {
	return &inlineScanner{text: text}
}

// eof reports whether the scanner has consumed all input.
func (s *inlineScanner) eof() bool {
	return s.pos >= len(s.text)
}

// peek returns the current byte without advancing, or 0 at EOF.
func (s *inlineScanner) peek() byte {
	if s.pos >= len(s.text) {
		return 0
	}
	return s.text[s.pos]
}

// peekAt returns the byte at offset i ahead of current position, or 0 if out of bounds.
func (s *inlineScanner) peekAt(i int) byte {
	p := s.pos + i
	if p < 0 || p >= len(s.text) {
		return 0
	}
	return s.text[p]
}

// advance moves the position forward by n bytes.
func (s *inlineScanner) advance(n int) {
	s.pos += n
	if s.pos > len(s.text) {
		s.pos = len(s.text)
	}
}

// remaining returns the unconsumed portion of the text.
func (s *inlineScanner) remaining() string {
	if s.pos >= len(s.text) {
		return ""
	}
	return s.text[s.pos:]
}

// indexOf returns the index of the first occurrence of b starting from the
// current position, or -1 if not found. The returned index is absolute
// within s.text.
func (s *inlineScanner) indexOf(b byte) int {
	for i := s.pos; i < len(s.text); i++ {
		if s.text[i] == b {
			return i
		}
	}
	return -1
}

// indexOfString returns the index of the first occurrence of sub starting
// from the current position, or -1 if not found.
func (s *inlineScanner) indexOfString(sub string) int {
	if sub == "" {
		return s.pos
	}
	for i := s.pos; i <= len(s.text)-len(sub); i++ {
		if s.text[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// isNewline checks if the byte at the given absolute position is a newline character.
func isNewline(b byte) bool {
	return b == '\n' || b == '\r'
}

// isASCIIPunctuation reports whether b is an ASCII punctuation character
// per the CommonMark spec.
func isASCIIPunctuation(b byte) bool {
	switch b {
	case '!', '"', '#', '$', '%', '&', '\'', '(', ')', '*', '+', ',', '-', '.', '/',
		':', ';', '<', '=', '>', '?', '@', '[', '\\', ']', '^', '_', '`', '{', '|', '}', '~':
		return true
	}
	return false
}

// isWhitespace reports whether b is an ASCII whitespace character.
func isWhitespace(b byte) bool {
	switch b {
	case ' ', '\t', '\n', '\r', '\f':
		return true
	}
	return false
}
