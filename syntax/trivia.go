package syntax

// Trivia represents a single piece of syntactically insignificant text
// attached to a token (leading or trailing whitespace, blank lines, comments).
type Trivia struct {
	Kind SyntaxKind
	Text string
}

// Width returns the byte length of this trivia piece.
func (t Trivia) Width() int { return len(t.Text) }

// TriviaList is an immutable sequence of trivia pieces.
type TriviaList struct {
	pieces []Trivia
}

// NewTriviaList creates a TriviaList from the given pieces.
func NewTriviaList(pieces ...Trivia) TriviaList {
	if len(pieces) == 0 {
		return TriviaList{}
	}
	cp := make([]Trivia, len(pieces))
	copy(cp, pieces)
	return TriviaList{pieces: cp}
}

// Len returns the number of trivia pieces.
func (tl TriviaList) Len() int { return len(tl.pieces) }

// At returns the trivia piece at index i.
func (tl TriviaList) At(i int) Trivia { return tl.pieces[i] }

// Width returns the total byte width of all trivia in the list.
func (tl TriviaList) Width() int {
	w := 0
	for _, t := range tl.pieces {
		w += t.Width()
	}
	return w
}

// IsEmpty reports whether the list contains no trivia.
func (tl TriviaList) IsEmpty() bool { return len(tl.pieces) == 0 }

// Text returns the concatenated text of all trivia pieces.
func (tl TriviaList) Text() string {
	if len(tl.pieces) == 0 {
		return ""
	}
	if len(tl.pieces) == 1 {
		return tl.pieces[0].Text
	}
	n := 0
	for _, t := range tl.pieces {
		n += len(t.Text)
	}
	buf := make([]byte, 0, n)
	for _, t := range tl.pieces {
		buf = append(buf, t.Text...)
	}
	return string(buf)
}

// Append returns a new TriviaList with the given pieces appended.
func (tl TriviaList) Append(pieces ...Trivia) TriviaList {
	newPieces := make([]Trivia, len(tl.pieces)+len(pieces))
	copy(newPieces, tl.pieces)
	copy(newPieces[len(tl.pieces):], pieces)
	return TriviaList{pieces: newPieces}
}
