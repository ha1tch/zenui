package zenui

// TextEvent says what a TextField's Update did.
type TextEvent int

// The events of a TextField.
const (
	TextNone      TextEvent = iota
	TextChanged             // the text changed: a rune typed, or Backspace
	TextCommitted           // Enter, or a click away, ended the edit with the text as it is
	TextCancelled           // Escape ended the edit and put back the text from before it began
)

// TextField is a single line of editable text. It ignores the keyboard until it
// is clicked, and stops when Enter, Escape or a click elsewhere ends the edit.
// Its geometry is a pure function of the area and the text metrics, like List's.
type TextField struct {
	Text string
	// MaxLen, if positive, is the most runes the text may hold.
	MaxLen int
	// Accept, if set, says which typed runes are let in.
	Accept   func(rune) bool
	Disabled bool

	active bool
	caret  int // in runes, from 0 to the length of the text
	before string
}

// NewTextField returns an empty field.
func NewTextField() *TextField { return &TextField{} }

// Active reports whether the field is being edited, and so has the keyboard.
func (f *TextField) Active() bool { return f.active }

// Bounds is the field's rectangle inside area: as wide as area, a line and a half
// high.
func (f *TextField) Bounds(area Rect, lineH int) Rect {
	return Rect{X: area.X, Y: area.Y, W: area.W, H: lineH + lineH/2}
}

// Blur ends an edit that is going on, as a commit, for a host whose focus has
// moved elsewhere.
func (f *TextField) Blur() TextEvent {
	if !f.active {
		return TextNone
	}
	f.active = false
	return TextCommitted
}

func (f *TextField) inner(b Rect, lineH int) Rect {
	pad := max(2, lineH/4)
	return Rect{X: b.X + pad, Y: b.Y, W: max(0, b.W-2*pad), H: b.H}
}

// startFor is the index of the first rune shown: when the whole text does not
// fit, the field shows the part that ends at the caret.
func startFor(runes []rune, caret, innerW int, textW func(string) int) int {
	s := 0
	for s < caret && textW(string(runes[s:caret])) > innerW {
		s++
	}
	return s
}

// caretAt is the caret position nearest to a click at x.
func (f *TextField) caretAt(runes []rune, x int, in Rect, textW func(string) int) int {
	s := startFor(runes, f.caret, in.W, textW)
	best, bestD := s, x-in.X
	if bestD < 0 {
		bestD = -bestD
	}
	for i := s + 1; i <= len(runes); i++ {
		d := x - (in.X + textW(string(runes[s:i])))
		if d < 0 {
			d = -d
		}
		if d < bestD {
			best, bestD = i, d
		}
	}
	return best
}

// Update gives the field a frame's input. A press inside starts an edit and
// places the caret; while editing, typed runes go in at the caret, Backspace
// deletes before it, Left and Right move it, Enter commits and Escape restores
// the text the edit began with. A press outside ends the edit as a commit.
func (f *TextField) Update(in Input, area Rect, lineH int, textW func(string) int) TextEvent {
	runes := []rune(f.Text)
	f.caret = clampInt(f.caret, 0, len(runes))
	if f.Disabled {
		f.active = false
		return TextNone
	}
	b := f.Bounds(area, lineH)
	if in.MousePressed {
		if b.Contains(in.MouseX, in.MouseY) {
			if !f.active {
				f.active, f.before = true, f.Text
			}
			f.caret = f.caretAt(runes, in.MouseX, f.inner(b, lineH), textW)
			return TextNone
		}
		if f.active {
			f.active = false
			return TextCommitted
		}
		return TextNone
	}
	if !f.active {
		return TextNone
	}
	ev := TextNone
	for _, k := range in.Keys {
		switch k {
		case KeyBackspace:
			if f.caret > 0 {
				runes = append(runes[:f.caret-1], runes[f.caret:]...)
				f.caret--
				ev = TextChanged
			}
		case KeyLeft:
			f.caret = max(0, f.caret-1)
		case KeyRight:
			f.caret = min(len(runes), f.caret+1)
		case KeyEnter:
			f.Text = string(runes)
			f.active = false
			return TextCommitted
		case KeyEscape:
			f.Text = f.before
			f.active = false
			return TextCancelled
		}
	}
	for _, c := range in.Chars {
		if f.Accept != nil && !f.Accept(c) {
			continue
		}
		if f.MaxLen > 0 && len(runes) >= f.MaxLen {
			continue
		}
		runes = append(runes[:f.caret], append([]rune{c}, runes[f.caret:]...)...)
		f.caret++
		ev = TextChanged
	}
	f.Text = string(runes)
	return ev
}

// Draw paints the field inside area at the given text scale. An active field has
// a heavier border in the selection colour and a caret.
func (f *TextField) Draw(r Renderer, th Theme, area Rect, scale int) {
	lh := r.LineHeight(scale)
	textW := func(s string) int { return r.TextWidth(s, scale) }
	b := f.Bounds(area, lh)
	r.FillRect(b, th.Panel)
	border, w := th.Border, 1
	if f.active {
		border, w = th.SelFill, 2
	}
	r.StrokeRect(b, border, w)
	in := f.inner(b, lh)
	runes := []rune(f.Text)
	caret := clampInt(f.caret, 0, len(runes))
	s := startFor(runes, caret, in.W, textW)
	col := th.Text
	if f.Disabled {
		col = th.Disabled
	}
	y := b.Y + (b.H-lh)/2
	r.Clip(in)
	r.DrawText(string(runes[s:]), in.X, y, scale, col)
	r.ClipEnd()
	if f.active {
		r.FillRect(Rect{X: in.X + textW(string(runes[s:caret])), Y: y, W: max(1, lh/8), H: lh}, th.Text)
	}
}
