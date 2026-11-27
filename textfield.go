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
	ed     lineEdit // the caret, the selection and the scroll
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
	f.active, f.ed.drag = false, false
	return TextCommitted
}

func (f *TextField) inner(b Rect, lineH int) Rect {
	pad := max(2, lineH/4)
	return Rect{X: b.X + pad, Y: b.Y, W: max(0, b.W-2*pad), H: b.H}
}

// limits are what the field lets in.
func (f *TextField) limits() editLimits { return editLimits{MaxLen: f.MaxLen, Accept: f.Accept} }

// Update gives the field a frame's input. A press inside starts an edit and
// places the caret; a drag from there selects. While editing, typed runes go in
// at the caret (over the selection, if there is one); Left, Right, Home and End
// move it, with Shift to select and Ctrl to go by word; Backspace and Delete
// remove a rune, or a word with Ctrl, or the selection; Ctrl+A selects all,
// Ctrl+C copies, Ctrl+X cuts and a paste goes in at the caret. Enter commits and
// Escape restores the text the edit began with. A press outside ends the edit as
// a commit.
func (f *TextField) Update(in Input, area Rect, lineH int, textW func(string) int) TextEvent {
	runes := []rune(f.Text)
	f.ed.clamp(len(runes))
	if f.Disabled {
		f.active, f.ed.drag = false, false
		return TextNone
	}
	b := f.Bounds(area, lineH)
	inner := f.inner(b, lineH)
	if in.MousePressed {
		if b.Contains(in.MouseX, in.MouseY) {
			if !f.active {
				f.active, f.before = true, f.Text
			}
			f.ed.pointer(runes, in, b, inner, textW)
			return TextNone
		}
		if f.active {
			f.active, f.ed.drag = false, false
			return TextCommitted
		}
		return TextNone
	}
	if !f.active {
		return TextNone
	}
	f.ed.pointer(runes, in, b, inner, textW) // a drag, or its release
	ev := TextNone
	if text, changed := f.ed.edit(f.Text, in, f.limits()); changed {
		f.Text, ev = text, TextChanged
	}
	for _, k := range in.Keys {
		switch k {
		case KeyEnter:
			f.active, f.ed.drag = false, false
			return TextCommitted
		case KeyEscape:
			f.Text = f.before
			f.active, f.ed.drag = false, false
			f.ed.toEnd(len([]rune(f.Text)))
			return TextCancelled
		}
	}
	return ev
}

// Draw paints the field inside area at the given text scale. An active field has
// a heavier border in the selection colour, a caret, and its selection drawn in
// the selection colours.
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
	f.ed.clamp(len(runes))
	s := f.ed.view(runes, in.W, textW)
	col := th.Text
	if f.Disabled {
		col = th.Disabled
	}
	f.ed.drawText(r, th, in, runes, s, scale, col, f.active)
}
