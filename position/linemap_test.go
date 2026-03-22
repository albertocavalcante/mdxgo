package position

import "testing"

func TestLineMap(t *testing.T) {
	src := "first\nsecond\nthird"
	lm := NewLineMap([]byte(src))

	if lm.LineCount() != 3 {
		t.Fatalf("LineCount() = %d, want 3", lm.LineCount())
	}

	tests := []struct {
		offset int
		line   int
		col    int
	}{
		{0, 1, 1},  // 'f'
		{5, 1, 6},  // '\n'
		{6, 2, 1},  // 's'
		{12, 2, 7}, // '\n'
		{13, 3, 1}, // 't'
		{17, 3, 5}, // 'd'
	}

	for _, tt := range tests {
		pos := lm.Pos(tt.offset)
		if pos.Line != tt.line || pos.Col != tt.col {
			t.Errorf("Pos(%d) = %v, want {%d, %d}", tt.offset, pos, tt.line, tt.col)
		}
	}
}

func TestLineMapOffset(t *testing.T) {
	src := "ab\ncd\n"
	lm := NewLineMap([]byte(src))

	if off := lm.Offset(1, 1); off != 0 {
		t.Errorf("Offset(1,1) = %d, want 0", off)
	}
	if off := lm.Offset(2, 1); off != 3 {
		t.Errorf("Offset(2,1) = %d, want 3", off)
	}
	if off := lm.Offset(2, 2); off != 4 {
		t.Errorf("Offset(2,2) = %d, want 4", off)
	}
}

func TestLineMapCRLF(t *testing.T) {
	src := "line1\r\nline2\r\n"
	lm := NewLineMap([]byte(src))

	if lm.LineCount() != 3 {
		t.Fatalf("LineCount() = %d, want 3", lm.LineCount())
	}

	pos := lm.Pos(7)
	if pos.Line != 2 || pos.Col != 1 {
		t.Errorf("Pos(7) = %v, want {2, 1}", pos)
	}
}

// ---------------------------------------------------------------------------
// Additional coverage
// ---------------------------------------------------------------------------

func TestLineMapEmpty(t *testing.T) {
	lm := NewLineMap([]byte(""))

	if lm.LineCount() != 1 {
		t.Errorf("LineCount() = %d, want 1 (single empty line)", lm.LineCount())
	}
	pos := lm.Pos(0)
	if pos.Line != 1 || pos.Col != 1 {
		t.Errorf("Pos(0) = %v, want {1, 1}", pos)
	}
}

func TestLineMapSingleLine(t *testing.T) {
	lm := NewLineMap([]byte("hello"))

	if lm.LineCount() != 1 {
		t.Errorf("LineCount() = %d, want 1", lm.LineCount())
	}
	pos := lm.Pos(4)
	if pos.Line != 1 || pos.Col != 5 {
		t.Errorf("Pos(4) = %v, want {1, 5}", pos)
	}
}

func TestLineMapSingleNewline(t *testing.T) {
	lm := NewLineMap([]byte("\n"))

	if lm.LineCount() != 2 {
		t.Errorf("LineCount() = %d, want 2", lm.LineCount())
	}
	p0 := lm.Pos(0)
	if p0.Line != 1 || p0.Col != 1 {
		t.Errorf("Pos(0) = %v, want {1, 1}", p0)
	}
	p1 := lm.Pos(1)
	if p1.Line != 2 || p1.Col != 1 {
		t.Errorf("Pos(1) = %v, want {2, 1}", p1)
	}
}

func TestLineMapCROnly(t *testing.T) {
	src := "aaa\rbbb\r"
	lm := NewLineMap([]byte(src))

	if lm.LineCount() != 3 {
		t.Fatalf("LineCount() = %d, want 3", lm.LineCount())
	}
	// After first \r, line 2 starts at offset 4.
	pos := lm.Pos(4)
	if pos.Line != 2 || pos.Col != 1 {
		t.Errorf("Pos(4) = %v, want {2, 1}", pos)
	}
}

func TestLineMapMixedLineEndings(t *testing.T) {
	src := "a\nb\r\nc\rd"
	lm := NewLineMap([]byte(src))

	if lm.LineCount() != 4 {
		t.Fatalf("LineCount() = %d, want 4", lm.LineCount())
	}

	tests := []struct {
		offset int
		line   int
		col    int
	}{
		{0, 1, 1}, // 'a'
		{2, 2, 1}, // 'b'
		{5, 3, 1}, // 'c'
		{7, 4, 1}, // 'd'
	}
	for _, tt := range tests {
		pos := lm.Pos(tt.offset)
		if pos.Line != tt.line || pos.Col != tt.col {
			t.Errorf("Pos(%d) = %v, want {%d, %d}", tt.offset, pos, tt.line, tt.col)
		}
	}
}

func TestLineMapPosClampNegative(t *testing.T) {
	lm := NewLineMap([]byte("abc"))

	pos := lm.Pos(-5)
	if pos.Line != 1 || pos.Col != 1 {
		t.Errorf("Pos(-5) = %v, want {1, 1}", pos)
	}
}

func TestLineMapPosClampBeyondEnd(t *testing.T) {
	lm := NewLineMap([]byte("abc"))

	pos := lm.Pos(100)
	// Should clamp to size (3), which is line 1 col 4.
	if pos.Line != 1 || pos.Col != 4 {
		t.Errorf("Pos(100) = %v, want {1, 4}", pos)
	}
}

func TestLineMapOffsetOutOfRange(t *testing.T) {
	lm := NewLineMap([]byte("ab\ncd"))

	tests := []struct {
		name string
		line int
		col  int
		want int
	}{
		{"line 0", 0, 1, -1},
		{"negative line", -1, 1, -1},
		{"col 0", 1, 0, -1},
		{"negative col", 1, -1, -1},
		{"line too high", 99, 1, -1},
		{"col beyond end", 1, 100, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := lm.Offset(tt.line, tt.col); got != tt.want {
				t.Errorf("Offset(%d, %d) = %d, want %d", tt.line, tt.col, got, tt.want)
			}
		})
	}
}

func TestLineMapLineStart(t *testing.T) {
	lm := NewLineMap([]byte("abc\ndef\nghi"))

	tests := []struct {
		line int
		want int
	}{
		{1, 0},
		{2, 4},
		{3, 8},
	}
	for _, tt := range tests {
		if got := lm.LineStart(tt.line); got != tt.want {
			t.Errorf("LineStart(%d) = %d, want %d", tt.line, got, tt.want)
		}
	}
}

func TestLineMapLineStartOutOfRange(t *testing.T) {
	lm := NewLineMap([]byte("abc"))

	if got := lm.LineStart(0); got != -1 {
		t.Errorf("LineStart(0) = %d, want -1", got)
	}
	if got := lm.LineStart(-1); got != -1 {
		t.Errorf("LineStart(-1) = %d, want -1", got)
	}
	if got := lm.LineStart(99); got != -1 {
		t.Errorf("LineStart(99) = %d, want -1", got)
	}
}

func TestLineMapOffsetRoundTrip(t *testing.T) {
	src := "hello\nworld\nfoo bar\n"
	lm := NewLineMap([]byte(src))

	for offset := range len(src) {
		pos := lm.Pos(offset)
		got := lm.Offset(pos.Line, pos.Col)
		if got != offset {
			t.Errorf("roundtrip offset %d → Pos %v → Offset %d", offset, pos, got)
		}
	}
}

func TestLineMapMultipleCRLF(t *testing.T) {
	src := "a\r\nb\r\nc\r\n"
	lm := NewLineMap([]byte(src))

	if lm.LineCount() != 4 {
		t.Fatalf("LineCount() = %d, want 4", lm.LineCount())
	}

	// Line starts: 0, 3, 6, 9
	tests := []struct {
		line  int
		start int
	}{
		{1, 0},
		{2, 3},
		{3, 6},
		{4, 9},
	}
	for _, tt := range tests {
		if got := lm.LineStart(tt.line); got != tt.start {
			t.Errorf("LineStart(%d) = %d, want %d", tt.line, got, tt.start)
		}
	}
}

func TestLineMapConsecutiveNewlines(t *testing.T) {
	src := "a\n\n\nb"
	lm := NewLineMap([]byte(src))

	if lm.LineCount() != 4 {
		t.Fatalf("LineCount() = %d, want 4", lm.LineCount())
	}

	pos := lm.Pos(4) // 'b'
	if pos.Line != 4 || pos.Col != 1 {
		t.Errorf("Pos(4) = %v, want {4, 1}", pos)
	}
}
