package parser

import (
	"unicode"
	"unicode/utf8"

	"github.com/albertocavalcante/mdxgo/syntax"
)

// delimiter represents an opening or closing delimiter run on the delimiter stack.
// This implements the CommonMark delimiter stack algorithm (appendix).
type delimiter struct {
	// char is the delimiter character ('*' or '_').
	char byte
	// count is the number of delimiter characters in the run.
	count int
	// origCount is the original number of delimiters (before processing).
	origCount int
	// pos is the position of the first delimiter character in the source text.
	pos int
	// canOpen and canClose indicate whether this delimiter run can open/close emphasis.
	canOpen  bool
	canClose bool
	// active is set to false when a delimiter is deactivated.
	active bool
	// outputIdx is the index in the output slice where this delimiter's InlineText was placed.
	outputIdx int
}

// tryEmphasis handles * and _ characters that may begin or close emphasis/strong spans.
// Instead of immediately emitting, it pushes a delimiter onto the stack and emits
// the delimiter text as InlineText. The delimiter stack is processed after all
// inline content is scanned.
func (ip *inlineProcessor) tryEmphasis() {
	s := ip.scanner
	ch := s.text[s.pos]

	// Count the run of identical delimiter characters.
	start := s.pos
	count := 0
	for s.pos+count < len(s.text) && s.text[s.pos+count] == ch {
		count++
	}

	delimText := s.text[start : start+count]

	// Determine left-flanking and right-flanking per CommonMark spec §6.2.
	charBefore := charBeforePos(s.text, start)
	charAfter := charAfterPos(s.text, start+count)

	leftFlanking := isLeftFlanking(charBefore, charAfter)
	rightFlanking := isRightFlanking(charBefore, charAfter)

	var canOpen, canClose bool
	if ch == '*' {
		canOpen = leftFlanking
		canClose = rightFlanking
	} else {
		// _ has stricter rules.
		canOpen = leftFlanking && (!rightFlanking || isUnicodePunctuation(charBefore))
		canClose = rightFlanking && (!leftFlanking || isUnicodePunctuation(charAfter))
	}

	// Flush any accumulated plain text.
	ip.flushText()

	// Emit the delimiter text as an InlineText node placeholder.
	inlineText := syntax.NewGreenNode(syntax.InlineText, []syntax.GreenElement{
		syntax.TokenElement(syntax.NewGreenToken(syntax.TextToken, delimText)),
	})
	outputIdx := len(ip.output)
	ip.output = append(ip.output, syntax.NodeElement(inlineText))

	// Push delimiter onto the stack.
	ip.delimiters = append(ip.delimiters, &delimiter{
		char:      ch,
		count:     count,
		origCount: count,
		pos:       start,
		canOpen:   canOpen,
		canClose:  canClose,
		active:    true,
		outputIdx: outputIdx,
	})

	s.advance(count)
}

// processEmphasis implements the CommonMark emphasis processing algorithm.
// It is called after all inline content has been scanned to match opening
// and closing delimiters and wrap the matched content in EmphasisSpan/StrongSpan nodes.
func (ip *inlineProcessor) processEmphasis() {
	if len(ip.delimiters) == 0 {
		return
	}

	// Process from left to right, looking for closers.
	for closerIdx := 0; closerIdx < len(ip.delimiters); closerIdx++ {
		closer := ip.delimiters[closerIdx]
		if !closer.canClose || !closer.active || closer.count == 0 || closer.outputIdx < 0 {
			continue
		}

		// Search backwards for a matching opener.
		openerIdx := closerIdx - 1
		found := false
		for openerIdx >= 0 {
			opener := ip.delimiters[openerIdx]
			if opener.char == closer.char && opener.canOpen && opener.active && opener.count > 0 && opener.outputIdx >= 0 {
				// Rule of three: if the closer can both open and close, or the opener
				// can both open and close, then the sum of lengths of the delimiter
				// runs must not be a multiple of 3 unless both lengths are multiples of 3.
				if (opener.canOpen && opener.canClose) || (closer.canOpen && closer.canClose) {
					if (opener.origCount+closer.origCount)%3 == 0 && opener.origCount%3 != 0 && closer.origCount%3 != 0 {
						openerIdx--
						continue
					}
				}
				found = true
				break
			}
			openerIdx--
		}

		if !found {
			// No matching opener. If closer can't open, deactivate it.
			if !closer.canOpen {
				closer.active = false
			}
			continue
		}

		opener := ip.delimiters[openerIdx]

		// Determine if strong (2+ delimiters) or regular emphasis (1 delimiter).
		useCount := 1
		if opener.count >= 2 && closer.count >= 2 {
			useCount = 2
		}

		emphKind := syntax.EmphasisSpan
		if useCount == 2 {
			emphKind = syntax.StrongSpan
		}

		// Build the emphasis node from the output elements between opener and closer.
		// The opener's InlineText must be trimmed (or removed) by useCount characters.
		// The closer's InlineText must be trimmed (or removed) by useCount characters.
		openOutputIdx := opener.outputIdx
		closeOutputIdx := closer.outputIdx

		// Get delimiter token text.
		openDelimChar := string(opener.char)
		var openDelimText string
		if useCount == 1 {
			openDelimText = openDelimChar
		} else {
			openDelimText = openDelimChar + openDelimChar
		}
		closeDelimText := openDelimText

		// Build children: open delimiter token + content + close delimiter token.
		var emphChildren []syntax.GreenElement
		emphChildren = append(emphChildren, syntax.TokenElement(
			syntax.NewGreenToken(delimTokenKind(opener.char), openDelimText),
		))

		// Content between opener and closer.
		for i := openOutputIdx + 1; i < closeOutputIdx; i++ {
			emphChildren = append(emphChildren, ip.output[i])
		}

		emphChildren = append(emphChildren, syntax.TokenElement(
			syntax.NewGreenToken(delimTokenKind(closer.char), closeDelimText),
		))

		emphNode := syntax.NewGreenNode(emphKind, emphChildren)

		// Update opener: reduce count.
		opener.count -= useCount
		// Update closer: reduce count.
		closer.count -= useCount

		// Rebuild the output slice:
		// 1. Everything before opener's output position.
		// 2. Updated opener InlineText (if count > 0) or nothing.
		// 3. Emphasis node.
		// 4. Updated closer InlineText (if count > 0) or nothing.
		// 5. Everything after closer's output position.
		var newOutput []syntax.GreenElement
		newOutput = append(newOutput, ip.output[:openOutputIdx]...)

		// Opener remainder.
		if opener.count > 0 {
			remainText := makeDelimText(opener.char, opener.count)
			newOutput = append(newOutput, syntax.NodeElement(
				syntax.NewGreenNode(syntax.InlineText, []syntax.GreenElement{
					syntax.TokenElement(syntax.NewGreenToken(syntax.TextToken, remainText)),
				}),
			))
			opener.outputIdx = len(newOutput) - 1
		} else {
			opener.outputIdx = -1
		}

		// Emphasis node.
		newOutput = append(newOutput, syntax.NodeElement(emphNode))

		// Closer remainder.
		if closer.count > 0 {
			remainText := makeDelimText(closer.char, closer.count)
			newOutput = append(newOutput, syntax.NodeElement(
				syntax.NewGreenNode(syntax.InlineText, []syntax.GreenElement{
					syntax.TokenElement(syntax.NewGreenToken(syntax.TextToken, remainText)),
				}),
			))
			closer.outputIdx = len(newOutput) - 1
		} else {
			closer.outputIdx = -1
		}

		// Everything after the closer.
		if closeOutputIdx+1 < len(ip.output) {
			newOutput = append(newOutput, ip.output[closeOutputIdx+1:]...)
		}

		// Deactivate delimiters between opener and closer.
		for i := openerIdx + 1; i < closerIdx; i++ {
			ip.delimiters[i].active = false
		}

		// Calculate the shift in output indices.
		oldLen := closeOutputIdx - openOutputIdx + 1
		newLen := len(newOutput) - len(ip.output) + oldLen

		// Update output indices for delimiters after the closer.
		shift := newLen - oldLen
		for i := closerIdx + 1; i < len(ip.delimiters); i++ {
			if ip.delimiters[i].outputIdx >= 0 {
				ip.delimiters[i].outputIdx += shift
			}
		}

		ip.output = newOutput

		// If closer still has remaining count, reprocess from closer.
		if closer.count > 0 {
			closerIdx-- // will be incremented by loop
		} else {
			closer.active = false
		}
		if opener.count == 0 {
			opener.active = false
		}
	}
}

func delimTokenKind(ch byte) syntax.SyntaxKind {
	if ch == '*' {
		return syntax.StarToken
	}
	return syntax.UnderscoreToken
}

func makeDelimText(ch byte, count int) string {
	b := make([]byte, count)
	for i := range b {
		b[i] = ch
	}
	return string(b)
}

// charBeforePos returns the Unicode rune before position pos in text,
// or a newline character if at the start of the text.
func charBeforePos(text string, pos int) rune {
	if pos <= 0 {
		return '\n' // treat start of text as if preceded by newline
	}
	r, _ := utf8.DecodeLastRuneInString(text[:pos])
	return r
}

// charAfterPos returns the Unicode rune after position pos in text,
// or a newline character if at the end of the text.
func charAfterPos(text string, pos int) rune {
	if pos >= len(text) {
		return '\n' // treat end of text as if followed by newline
	}
	r, _ := utf8.DecodeRuneInString(text[pos:])
	return r
}

// isLeftFlanking checks if a delimiter run is left-flanking per CommonMark.
// A left-flanking delimiter run:
//   - is not followed by Unicode whitespace, AND
//   - is not followed by a Unicode punctuation character, OR is preceded by
//     Unicode whitespace or a Unicode punctuation character.
func isLeftFlanking(before, after rune) bool {
	if unicode.IsSpace(after) {
		return false
	}
	if !isUnicodePunctuation(after) {
		return true
	}
	return unicode.IsSpace(before) || isUnicodePunctuation(before)
}

// isRightFlanking checks if a delimiter run is right-flanking per CommonMark.
func isRightFlanking(before, after rune) bool {
	if unicode.IsSpace(before) {
		return false
	}
	if !isUnicodePunctuation(before) {
		return true
	}
	return unicode.IsSpace(after) || isUnicodePunctuation(after)
}

// isUnicodePunctuation reports whether r is a Unicode punctuation character
// per the CommonMark spec (Pc, Pd, Pe, Pf, Pi, Po, Ps, Sc, Sk, Sm, So).
func isUnicodePunctuation(r rune) bool {
	return unicode.IsPunct(r) || unicode.IsSymbol(r)
}
