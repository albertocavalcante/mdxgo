// Package parser implements an MDX v3 parser producing a lossless CST.
package parser

// line represents a single source line with its byte offset information.
type line struct {
	// Start is the byte offset of the first character of this line in the source.
	Start int
	// Content is the raw text of this line including the line ending.
	Content string
}

// scanner splits source text into lines, preserving all bytes exactly.
// Line endings (\n, \r\n, \r) are included in each line's Content.
type scanner struct {
	src   []byte
	lines []line
}

// newScanner creates a scanner from source bytes.
func newScanner(src []byte) *scanner {
	s := &scanner{src: src}
	s.scan()
	return s
}

func (s *scanner) scan() {
	start := 0
	for i := 0; i < len(s.src); i++ {
		switch s.src[i] {
		case '\n':
			s.lines = append(s.lines, line{
				Start:   start,
				Content: string(s.src[start : i+1]),
			})
			start = i + 1
		case '\r':
			end := i + 1
			if end < len(s.src) && s.src[end] == '\n' {
				end++
			}
			s.lines = append(s.lines, line{
				Start:   start,
				Content: string(s.src[start:end]),
			})
			start = end
			i = end - 1 // loop will i++
		}
	}
	// Last line (may not have a trailing newline).
	if start <= len(s.src) {
		remaining := string(s.src[start:])
		if remaining != "" {
			s.lines = append(s.lines, line{
				Start:   start,
				Content: remaining,
			})
		}
	}
}
