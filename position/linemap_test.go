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
