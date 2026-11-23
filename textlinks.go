package zenui

import "strings"

// Links in the text of a TextModal: "[text](address)" and a bare address that
// begins http:// or https://. Only those two schemes are links, so that what a
// host puts in a modal can never name a program to run or a file to open.

// textSpan is a run of one line's text: plain when url is empty, a link to
// url otherwise.
type textSpan struct{ text, url string }

// hasWebScheme reports whether s begins with http:// or https:// and has
// something after it.
func hasWebScheme(s string) bool {
	for _, p := range []string{"https://", "http://"} {
		if strings.HasPrefix(s, p) && len(s) > len(p) {
			return true
		}
	}
	return false
}

func isAlnum(b byte) bool {
	return b >= '0' && b <= '9' || b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z'
}

// markdownLink reads "[text](url)" at the start of s, and returns the text, the
// address and the length it took; n is 0 if s does not begin with one. The text
// is not empty and holds no "]", the address is a web address with no space or
// ")" in it.
func markdownLink(s string) (text, url string, n int) {
	if len(s) < 4 || s[0] != '[' {
		return "", "", 0
	}
	j := strings.IndexByte(s, ']')
	if j < 2 || j+1 >= len(s) || s[j+1] != '(' || strings.IndexByte(s[1:j], '[') >= 0 {
		return "", "", 0
	}
	k := strings.IndexByte(s[j+2:], ')')
	if k < 0 {
		return "", "", 0
	}
	url = s[j+2 : j+2+k]
	if !hasWebScheme(url) || strings.ContainsAny(url, " \t") {
		return "", "", 0
	}
	return s[1:j], url, j + 2 + k + 1
}

// bareURLLen is the length of the web address at the start of s, or 0. It runs
// to the next space or angle bracket or quote, less any punctuation that
// closes a sentence (. , ; : ! ?) or a bracket that has no opener inside it.
func bareURLLen(s string) int {
	if !hasWebScheme(s) {
		return 0
	}
	n := strings.IndexAny(s, " \t<>\"")
	if n < 0 {
		n = len(s)
	}
	for n > 0 {
		c := s[n-1]
		if strings.IndexByte(".,;:!?'", c) >= 0 {
			n--
			continue
		}
		if c == ')' && strings.Count(s[:n], "(") < strings.Count(s[:n], ")") {
			n--
			continue
		}
		if c == ']' && strings.Count(s[:n], "[") < strings.Count(s[:n], "]") {
			n--
			continue
		}
		break
	}
	if !hasWebScheme(s[:n]) {
		return 0
	}
	return n
}

// linkSpans splits a line into plain runs and links. With bare, a web address
// standing in the text is a link too; without, only "[text](url)" is.
func linkSpans(s string, bare bool) []textSpan {
	var spans []textSpan
	plain := 0 // where the pending plain run began
	flush := func(end int) {
		if end > plain {
			spans = append(spans, textSpan{text: s[plain:end]})
		}
	}
	for i := 0; i < len(s); {
		if s[i] == '[' {
			if text, url, n := markdownLink(s[i:]); n > 0 {
				flush(i)
				spans = append(spans, textSpan{text: text, url: url})
				i += n
				plain = i
				continue
			}
		}
		if bare && (s[i] == 'h') && (i == 0 || !isAlnum(s[i-1])) {
			if n := bareURLLen(s[i:]); n > 0 {
				flush(i)
				spans = append(spans, textSpan{text: s[i : i+n], url: s[i : i+n]})
				i += n
				plain = i
				continue
			}
		}
		i++
	}
	flush(len(s))
	return spans
}

// linkRect is where a link was drawn, as of the last Draw.
type linkRect struct {
	r   Rect
	url string
}

// WithLinks makes the links in the modal's text work: a "[text](address)" and
// a bare http:// or https:// address are drawn in the theme's accent colour,
// underlined, and a click on one calls open with its address instead of doing
// anything to the modal. A host passes the text of its own, never text it was
// given by a file or a person, since whatever is in it can be clicked. It
// returns the modal, for a chain after NewTextModal.
func (h *TextModal) WithLinks(open func(url string)) *TextModal {
	h.openLink = open
	return h
}

// OverLink reports whether the pointer, as of the last Update, is over a link:
// a host can show its pointing-hand cursor then.
func (h *TextModal) OverLink() bool { return h.linkAt(h.mouseX, h.mouseY) != "" }

// linkAt is the address of the link drawn at x, y, or "".
func (h *TextModal) linkAt(x, y int) string {
	if h.openLink == nil || !h.body.Contains(x, y) {
		return ""
	}
	for _, l := range h.links {
		if l.r.Contains(x, y) {
			return l.url
		}
	}
	return ""
}

// drawLinkedLine draws one line like DrawTextLine and, where the line has
// links, draws them in the link style and returns where they are. hover is the
// pointer, to draw the link under it bold.
func drawLinkedLine(r Renderer, line string, x, y, scale, lineH int, theme Theme, mx, my int) []linkRect {
	kind, text := classifyTextLine(line)
	spans := linkSpans(text, true)
	var out []linkRect
	for _, sp := range spans {
		w := r.TextWidth(sp.text, scale)
		if sp.url == "" {
			switch kind {
			case textLineH1, textLineH2:
				r.DrawText(sp.text, x, y, scale, theme.DirText)
				r.DrawText(sp.text, x+1, y, scale, theme.DirText)
			case textLineIndented:
				r.DrawText(sp.text, x, y, scale, theme.DimText)
			default:
				r.DrawText(sp.text, x, y, scale, theme.Text)
			}
			x += w
			continue
		}
		box := Rect{X: x, Y: y, W: w, H: lineH}
		r.DrawText(sp.text, x, y, scale, theme.DirText)
		if box.Contains(mx, my) {
			r.DrawText(sp.text, x+1, y, scale, theme.DirText) // bold, as a heading is
		}
		r.FillRect(Rect{X: x, Y: y + lineH - scale, W: w, H: 1}, theme.DirText)
		out = append(out, linkRect{r: box, url: sp.url})
		x += w
	}
	return out
}

// WrapLinkedText wraps s, which may hold "[text](address)" links, to width
// columns of what is seen: the markup does not count. It breaks only at
// spaces, and a link that falls across a break is closed on the one line and
// opened again, to the same address, on the next, so that each part of it
// works. Each line is led by indent.
func WrapLinkedText(s string, width int, indent string) []string {
	type cell struct {
		r   byte
		url string
	}
	var cells []cell
	for _, sp := range linkSpans(s, false) {
		for i := 0; i < len(sp.text); i++ {
			cells = append(cells, cell{sp.text[i], sp.url})
		}
	}
	// serialise writes a line's cells, a run to the same address as a link.
	serialise := func(line []cell) string {
		var b strings.Builder
		b.WriteString(indent)
		for i := 0; i < len(line); {
			j := i
			for j < len(line) && line[j].url == line[i].url {
				j++
			}
			run := make([]byte, 0, j-i)
			for _, c := range line[i:j] {
				run = append(run, c.r)
			}
			if line[i].url == "" {
				b.Write(run)
			} else {
				b.WriteString("[" + string(run) + "](" + line[i].url + ")")
			}
			i = j
		}
		return b.String()
	}

	var lines []string
	var line []cell
	for i := 0; i < len(cells); {
		for i < len(cells) && cells[i].r == ' ' {
			i++
		}
		j := i
		for j < len(cells) && cells[j].r != ' ' {
			j++
		}
		if i == j {
			break
		}
		word := cells[i:j]
		if len(line) > 0 && len(line)+1+len(word) > width {
			lines = append(lines, serialise(line))
			line = nil
		}
		if len(line) > 0 {
			// The space between two words of one link belongs to it.
			sp := cell{' ', ""}
			if line[len(line)-1].url != "" && line[len(line)-1].url == word[0].url {
				sp.url = word[0].url
			}
			line = append(line, sp)
		}
		line = append(line, word...)
		i = j
	}
	return append(lines, serialise(line))
}
