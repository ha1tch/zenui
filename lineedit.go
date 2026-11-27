package zenui

import (
	"strings"
	"unicode"
)

// lineEdit is the editing state that every single-line edit box shares: a caret
// and the other end of a selection (both in runes), the first rune shown, and
// whether a mouse drag that began in the box is going on. TextField, the text
// boxes of a Form and the name field of a Dialog all edit through it, so that
// they all move, select, copy, cut and paste the same way.
//
// The text itself stays with the widget: edit takes it and gives it back.
type lineEdit struct {
	caret  int  // where typing goes, in runes
	anchor int  // the other end of the selection; equal to caret when nothing is selected
	scroll int  // the first rune shown
	set    bool // caret and anchor have been given a place (until then they are at the end)
	last   int  // the length of the text when the caret was last placed or the text last looked at
	drag   bool // a drag that began inside the box is going on
}

// editLimits are the limits a box puts on what goes into it.
type editLimits struct {
	MaxLen int             // if positive, the most runes the text may hold
	Accept func(rune) bool // if set, says which typed or pasted runes are let in
}

// clamp keeps the caret and the anchor inside a text of n runes. A box that has
// not been given a place yet has its caret at the end of its text, where the
// next rune goes.
func (e *lineEdit) clamp(n int) {
	if !e.set {
		e.caret, e.anchor, e.set, e.last = n, n, true, n
		return
	}
	// A caret that was at the end of the text follows the end when the host replaces the text.
	if e.caret == e.last && e.anchor == e.last && n != e.last {
		e.caret, e.anchor = n, n
	}
	e.last = n
	e.caret = clampInt(e.caret, 0, n)
	e.anchor = clampInt(e.anchor, 0, n)
}

// hasSel reports whether some of the text is selected.
func (e *lineEdit) hasSel() bool { return e.caret != e.anchor }

// sel is the selected span, from lo up to hi.
func (e *lineEdit) sel() (lo, hi int) {
	if e.anchor < e.caret {
		return e.anchor, e.caret
	}
	return e.caret, e.anchor
}

// toEnd puts the caret at the end of a text of n runes, with nothing selected.
func (e *lineEdit) toEnd(n int) { e.caret, e.anchor, e.set, e.last = n, n, true, n }

// selectAll selects a text of n runes, with the caret at its end.
func (e *lineEdit) selectAll(n int) { e.anchor, e.caret, e.set, e.last = 0, n, true, n }

// isWordRune is what a word is made of: for jumping by word, a path's separators
// and a file name's dots end a word as spaces do.
func isWordRune(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' }

// wordLeft is where the caret goes for Ctrl+Left from i: to the start of the
// word it is in or just after.
func wordLeft(runes []rune, i int) int {
	for i > 0 && !isWordRune(runes[i-1]) {
		i--
	}
	for i > 0 && isWordRune(runes[i-1]) {
		i--
	}
	return i
}

// wordRight is where the caret goes for Ctrl+Right from i: to the end of the
// word it is in or just before.
func wordRight(runes []rune, i int) int {
	for i < len(runes) && !isWordRune(runes[i]) {
		i++
	}
	for i < len(runes) && isWordRune(runes[i]) {
		i++
	}
	return i
}

// cleanPaste makes what was pasted fit a single line: line ends go, a tab is a
// space, and other control characters are dropped.
func cleanPaste(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == '\n' || r == '\r':
		case r == '\t':
			b.WriteByte(' ')
		case r < 0x20 || r == 0x7f:
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// edit applies a frame's keys, typed runes and paste to text, and returns the new
// text and whether it changed. It handles the keys of an edit box (Left, Right,
// Home, End, Backspace, Delete, with Shift to select and Ctrl to go by word; Ctrl+A
// to select all, Ctrl+C to copy, Ctrl+X to cut, and the paste) and leaves the
// rest (Enter, Escape, Tab, Up, Down, the mouse) to the widget.
func (e *lineEdit) edit(text string, in Input, lim editLimits) (string, bool) {
	runes := []rune(text)
	e.clamp(len(runes))
	changed := false
	word := in.Ctrl || in.Alt // Ctrl on a PC, Option on a Mac: the arrows and Backspace go by word

	deleteSel := func() bool {
		if !e.hasSel() {
			return false
		}
		lo, hi := e.sel()
		runes = append(runes[:lo], runes[hi:]...)
		e.caret, e.anchor = lo, lo
		changed = true
		return true
	}
	deleteRange := func(lo, hi int) {
		if lo >= hi {
			return
		}
		runes = append(runes[:lo], runes[hi:]...)
		e.caret, e.anchor = lo, lo
		changed = true
	}
	move := func(to int) {
		e.caret = clampInt(to, 0, len(runes))
		if !in.Shift {
			e.anchor = e.caret
		}
	}
	insert := func(s []rune) {
		deleteSel()
		if lim.MaxLen > 0 {
			room := lim.MaxLen - len(runes)
			if room <= 0 {
				return
			}
			if len(s) > room {
				s = s[:room]
			}
		}
		if len(s) == 0 {
			return
		}
		out := make([]rune, 0, len(runes)+len(s))
		out = append(out, runes[:e.caret]...)
		out = append(out, s...)
		out = append(out, runes[e.caret:]...)
		runes = out
		e.caret += len(s)
		e.anchor = e.caret
		changed = true
	}
	accepted := func(rs []rune) []rune {
		out := rs[:0:0]
		for _, r := range rs {
			if r < 0x20 || r == 0x7f || (lim.Accept != nil && !lim.Accept(r)) {
				continue
			}
			out = append(out, r)
		}
		return out
	}

	for _, k := range in.Keys {
		switch k {
		case KeyLeft:
			switch {
			case e.hasSel() && !in.Shift:
				lo, _ := e.sel()
				e.caret, e.anchor = lo, lo
			case word:
				move(wordLeft(runes, e.caret))
			default:
				move(e.caret - 1)
			}
		case KeyRight:
			switch {
			case e.hasSel() && !in.Shift:
				_, hi := e.sel()
				e.caret, e.anchor = hi, hi
			case word:
				move(wordRight(runes, e.caret))
			default:
				move(e.caret + 1)
			}
		case KeyHome:
			move(0)
		case KeyEnd:
			move(len(runes))
		case KeyBackspace:
			if !deleteSel() {
				if word {
					deleteRange(wordLeft(runes, e.caret), e.caret)
				} else if e.caret > 0 {
					deleteRange(e.caret-1, e.caret)
				}
			}
		case KeyDelete:
			if !deleteSel() {
				if word {
					deleteRange(e.caret, wordRight(runes, e.caret))
				} else if e.caret < len(runes) {
					deleteRange(e.caret, e.caret+1)
				}
			}
		}
	}

	for _, c := range in.CtrlChars {
		switch c {
		case 'a':
			e.selectAll(len(runes))
		case 'c', 'x':
			if e.hasSel() && in.SetClipboard != nil {
				lo, hi := e.sel()
				in.SetClipboard(string(runes[lo:hi]))
			}
			if c == 'x' {
				deleteSel()
			}
		}
	}

	if !in.Ctrl && len(in.Chars) > 0 {
		insert(accepted(in.Chars))
	}
	if in.Paste != "" {
		insert(accepted([]rune(cleanPaste(in.Paste))))
	}
	e.last = len(runes)
	return string(runes), changed
}

// view is the first rune to show so that the caret is on show, in a box whose
// text area is innerW wide: the box scrolls no further than it must, and shows
// as much of the text as fits when the text is shorter than the box again.
func (e *lineEdit) view(runes []rune, innerW int, textW func(string) int) int {
	n := len(runes)
	e.clamp(n)
	s := clampInt(e.scroll, 0, n)
	if s > e.caret {
		s = e.caret
	}
	for s < e.caret && textW(string(runes[s:e.caret])) > innerW {
		s++
	}
	for s > 0 && textW(string(runes[s-1:])) <= innerW {
		s--
	}
	e.scroll = s
	return s
}

// caretAtX is the caret position nearest to x, in a box whose text area starts
// at innerX and shows the text from rune start.
func caretAtX(runes []rune, start, x, innerX int, textW func(string) int) int {
	best, bestD := start, absInt(x-innerX)
	for i := start + 1; i <= len(runes); i++ {
		d := absInt(x - (innerX + textW(string(runes[start:i]))))
		if d < bestD {
			best, bestD = i, d
		}
	}
	return best
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// pointer applies the mouse to the box: a press inside puts the caret there (and
// starts a selection, unless Shift extends the one there is), a drag that began
// inside moves the caret and keeps the anchor, and the release ends the drag. It
// reports whether the press was inside the box. inner is the text area and box
// the whole box.
func (e *lineEdit) pointer(runes []rune, in Input, box, inner Rect, textW func(string) int) (pressedInside bool) {
	e.clamp(len(runes))
	start := e.view(runes, inner.W, textW)
	switch {
	case in.MousePressed && box.Contains(in.MouseX, in.MouseY):
		e.caret = caretAtX(runes, start, in.MouseX, inner.X, textW)
		if !in.Shift {
			e.anchor = e.caret
		}
		e.drag = true
		return true
	case in.MousePressed:
		e.drag = false
	case e.drag && in.MouseDown:
		e.caret = caretAtX(runes, start, in.MouseX, inner.X, textW)
	default:
		e.drag = false
	}
	return false
}

// drawText draws a box's text from rune start inside inner, with the selected
// part in the selection colours and a caret when the box has the focus.
func (e *lineEdit) drawText(r Renderer, th Theme, inner Rect, runes []rune, start, scale int, col Colour, focused bool) {
	lh := r.LineHeight(scale)
	textW := func(s string) int { return r.TextWidth(s, scale) }
	y := inner.Y + (inner.H-lh)/2
	lo, hi := e.sel()
	lo, hi = clampInt(lo, start, len(runes)), clampInt(hi, start, len(runes))
	r.Clip(inner)
	if focused && lo < hi {
		x0 := inner.X + textW(string(runes[start:lo]))
		r.FillRect(Rect{X: x0, Y: y, W: textW(string(runes[lo:hi])), H: lh}, th.SelFill)
	}
	if focused && lo < hi {
		r.DrawText(string(runes[start:lo]), inner.X, y, scale, col)
		r.DrawText(string(runes[lo:hi]), inner.X+textW(string(runes[start:lo])), y, scale, th.SelText)
		r.DrawText(string(runes[hi:]), inner.X+textW(string(runes[start:hi])), y, scale, col)
	} else {
		r.DrawText(string(runes[start:]), inner.X, y, scale, col)
	}
	r.ClipEnd()
	if focused {
		caret := clampInt(e.caret, start, len(runes))
		r.FillRect(Rect{X: inner.X + textW(string(runes[start:caret])), Y: y, W: max(1, lh/8), H: lh}, col)
	}
}
