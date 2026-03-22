package parser

import (
	"strings"

	"github.com/albertocavalcante/mdxgo/syntax"
)

// HTML block start conditions (CommonMark spec §4.6).
// We check a simplified version covering the main cases.

var htmlBlockType1Start = []string{"<pre", "<script", "<style", "<textarea"}
var htmlBlockType1End = []string{"</pre>", "</script>", "</style>", "</textarea>"}

var htmlBlockType6Tags = []string{
	"address", "article", "aside", "base", "basefont", "blockquote", "body",
	"caption", "center", "col", "colgroup", "dd", "details", "dialog", "dir",
	"div", "dl", "dt", "fieldset", "figcaption", "figure", "footer", "form",
	"frame", "frameset", "h1", "h2", "h3", "h4", "h5", "h6", "head", "header",
	"hr", "html", "iframe", "legend", "li", "link", "main", "menu", "menuitem",
	"nav", "noframes", "ol", "optgroup", "option", "p", "param", "search",
	"section", "summary", "table", "tbody", "td", "tfoot", "th", "thead",
	"title", "tr", "track", "ul",
}

// isHTMLBlockStart checks if a line starts an HTML block.
func isHTMLBlockStart(li lineInfo) bool {
	if li.indent > maxIndent || li.blank {
		return false
	}
	s := li.trimmed
	if s == "" || s[0] != '<' {
		return false
	}
	lower := strings.ToLower(s)

	// Type 1: <pre, <script, <style, <textarea
	for _, tag := range htmlBlockType1Start {
		if strings.HasPrefix(lower, tag) {
			rest := lower[len(tag):]
			if rest == "" || rest[0] == ' ' || rest[0] == '>' || rest[0] == '\t' {
				return true
			}
		}
	}

	// Type 2: <!--
	if strings.HasPrefix(s, "<!--") {
		return true
	}

	// Type 3: <?
	if strings.HasPrefix(s, "<?") {
		return true
	}

	// Type 4: <!LETTER
	if len(s) > 2 && s[0] == '<' && s[1] == '!' && isASCIILetter(s[2]) {
		return true
	}

	// Type 5: <![CDATA[
	if strings.HasPrefix(s, "<![CDATA[") {
		return true
	}

	// Type 6: block-level HTML tag
	tagName := extractHTMLTagName(lower)
	if tagName != "" {
		for _, t := range htmlBlockType6Tags {
			if tagName == t {
				return true
			}
		}
	}

	return false
}

func isASCIILetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func extractHTMLTagName(lower string) string {
	if len(lower) < 2 || lower[0] != '<' {
		return ""
	}
	i := 1
	if i < len(lower) && lower[i] == '/' {
		i++
	}
	start := i
	for i < len(lower) && isASCIILetter(lower[i]) {
		i++
	}
	if i == start {
		return ""
	}
	if i < len(lower) && lower[i] != ' ' && lower[i] != '>' && lower[i] != '/' && lower[i] != '\t' && lower[i] != '\n' {
		return ""
	}
	return lower[start:i]
}

// parseHTMLBlock parses an HTML block.
func (p *blockParser) parseHTMLBlock(firstLI lineInfo) {
	p.builder.startNode()

	// Determine which type for end condition.
	lower := strings.ToLower(firstLI.trimmed)
	endCheck := htmlBlockEndDefault

	for i, tag := range htmlBlockType1Start {
		if strings.HasPrefix(lower, tag) {
			endTag := htmlBlockType1End[i]
			endCheck = func(s string) bool {
				return strings.Contains(strings.ToLower(s), endTag)
			}
			break
		}
	}
	switch {
	case strings.HasPrefix(firstLI.trimmed, "<!--"):
		endCheck = func(s string) bool { return strings.Contains(s, "-->") }
	case strings.HasPrefix(firstLI.trimmed, "<?"):
		endCheck = func(s string) bool { return strings.Contains(s, "?>") }
	case len(firstLI.trimmed) > 2 && firstLI.trimmed[0] == '<' && firstLI.trimmed[1] == '!' && isASCIILetter(firstLI.trimmed[2]):
		endCheck = func(s string) bool { return strings.Contains(s, ">") }
	case strings.HasPrefix(firstLI.trimmed, "<![CDATA["):
		endCheck = func(s string) bool { return strings.Contains(s, "]]>") }
	}

	// Emit lines until end condition or blank line (for type 6/7).
	for !p.eof() {
		cl := p.currentLine()
		p.builder.token(syntax.HTMLLineToken, cl.Content)
		p.advance()

		if endCheck(cl.Content) {
			break
		}
	}

	p.builder.finishNode(syntax.HTMLBlock)
}

// htmlBlockEndDefault ends at a blank line (type 6/7 behavior).
func htmlBlockEndDefault(s string) bool {
	trimmed := strings.TrimRight(s, "\r\n")
	return strings.TrimSpace(trimmed) == ""
}
