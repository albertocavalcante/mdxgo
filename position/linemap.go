// Package position provides offset↔line:col conversion for source text.
package position

// Pos represents a line and column position in source text.
// Both Line and Col are 1-based.
type Pos struct {
	Line int
	Col  int
}

// LineMap supports bidirectional conversion between byte offsets
// and line:column positions. It is built once from source text.
type LineMap struct {
	// lineStarts[i] is the byte offset where line i+1 begins.
	// lineStarts[0] is always 0.
	lineStarts []int
	size       int
}

// NewLineMap builds a LineMap from source text.
func NewLineMap(src []byte) *LineMap {
	starts := []int{0}
	for i := 0; i < len(src); i++ {
		switch src[i] {
		case '\n':
			starts = append(starts, i+1)
		case '\r':
			if i+1 < len(src) && src[i+1] == '\n' {
				starts = append(starts, i+2) //nolint:mnd // \r\n offset
				i++                          // skip the \n
			} else {
				starts = append(starts, i+1)
			}
		}
	}
	return &LineMap{lineStarts: starts, size: len(src)}
}

// LineCount returns the number of lines in the source.
func (lm *LineMap) LineCount() int { return len(lm.lineStarts) }

// Pos converts a byte offset to a line:col position.
// Both line and column are 1-based. Col counts bytes, not runes.
func (lm *LineMap) Pos(offset int) Pos {
	if offset < 0 {
		offset = 0
	}
	if offset > lm.size {
		offset = lm.size
	}
	// Binary search for the line.
	lo, hi := 0, len(lm.lineStarts)-1
	for lo < hi {
		mid := (lo + hi + 1) / 2 //nolint:mnd // binary search midpoint
		if lm.lineStarts[mid] <= offset {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	return Pos{Line: lo + 1, Col: offset - lm.lineStarts[lo] + 1}
}

// Offset converts a 1-based line:col to a byte offset.
// Returns -1 if the position is out of range.
func (lm *LineMap) Offset(line, col int) int {
	if line < 1 || line > len(lm.lineStarts) || col < 1 {
		return -1
	}
	off := lm.lineStarts[line-1] + col - 1
	if off > lm.size {
		return -1
	}
	return off
}

// LineStart returns the byte offset of the start of the given 1-based line.
func (lm *LineMap) LineStart(line int) int {
	if line < 1 || line > len(lm.lineStarts) {
		return -1
	}
	return lm.lineStarts[line-1]
}
