package zenui

import (
	"strings"
	"unicode/utf8"
)

// FitTail returns the end of text that fits in width, so that what is
// being typed at the end of a long line (an address, a path) stays in view.
func FitTail(r Renderer, text string, scale, width int) string {
	for text != "" && r.TextWidth(text, scale) > width {
		_, n := utf8.DecodeRuneInString(text)
		text = text[n:]
	}
	return text
}

// WrapText breaks s into lines of at most width characters, at spaces
// and at the separators of a path where it can; each line starts with
// indent, which is not counted in width.
func WrapText(s string, width int, indent string) []string {
	var lines []string
	for width > 1 && len(s) > width {
		cut := width
		for i := width; i > width/2; i-- {
			if strings.ContainsRune(" /\\", rune(s[i-1])) {
				cut = i
				break
			}
		}
		lines = append(lines, indent+strings.TrimRight(s[:cut], " "))
		s = strings.TrimLeft(s[cut:], " ")
	}
	return append(lines, indent+s)
}
