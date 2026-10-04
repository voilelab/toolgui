package tgutil

import "strings"

// InsertHead put snippet at the start of the html head, after the charset
// meta if it leads. Early, so a CSP meta covers the links after it. ok is
// false when html has no head.
func InsertHead(html, snippet string) (_ string, ok bool) {
	lower := strings.ToLower(html)

	pos := -1
	for i := 0; i < len(lower); {
		j := strings.Index(lower[i:], "<head")
		if j < 0 {
			break
		}
		i += j + len("<head")
		// Not <header>.
		if i < len(lower) && strings.ContainsRune(">/ \t\r\n", rune(lower[i])) {
			end := strings.IndexByte(lower[i:], '>')
			if end >= 0 {
				pos = i + end + 1
			}
			break
		}
	}
	if pos < 0 {
		return html, false
	}

	rest := strings.TrimLeft(lower[pos:], " \t\r\n")
	if strings.HasPrefix(rest, "<meta charset") {
		start := len(lower) - len(rest)
		if end := strings.IndexByte(rest, '>'); end >= 0 {
			pos = start + end + 1
		}
	}

	return html[:pos] + "\n" + snippet + html[pos:], true
}
