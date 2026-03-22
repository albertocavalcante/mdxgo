package position_test

import (
	"fmt"

	"github.com/albertocavalcante/mdxgo/position"
)

func ExampleNewLineMap() {
	src := []byte("first\nsecond\nthird")
	lm := position.NewLineMap(src)
	fmt.Println(lm.LineCount())
	// Output:
	// 3
}

func ExampleLineMap_Pos() {
	src := []byte("hello\nworld\n")
	lm := position.NewLineMap(src)

	pos := lm.Pos(6) // byte offset 6 → start of "world"
	fmt.Printf("line %d, col %d\n", pos.Line, pos.Col)
	// Output:
	// line 2, col 1
}

func ExampleLineMap_Offset() {
	src := []byte("hello\nworld\n")
	lm := position.NewLineMap(src)

	off := lm.Offset(2, 1) // line 2, col 1 → byte offset
	fmt.Println(off)
	// Output:
	// 6
}

func ExampleLineMap_LineStart() {
	src := []byte("aaa\nbbb\nccc")
	lm := position.NewLineMap(src)

	fmt.Println(lm.LineStart(1)) // start of first line
	fmt.Println(lm.LineStart(2)) // start of second line
	fmt.Println(lm.LineStart(3)) // start of third line
	// Output:
	// 0
	// 4
	// 8
}
