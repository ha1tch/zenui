package zenui

import "strconv"

// Stepper is a number with buttons to step it down and up: "<" and ">" by one
// and, when Big is set, "<<" and ">>" by Big. The value stops at Min and Max.
// Its geometry is a pure function of the area and the text metrics.
type Stepper struct {
	Value, Min, Max int
	// Big, if positive, adds buttons that step by this much.
	Big int
	// Format turns a value into the text shown; nil shows the number.
	Format   func(int) string
	Disabled bool

	mouseX, mouseY int
}

// NewStepper returns a stepper over min to max, at min.
func NewStepper(min, max int) *Stepper { return &Stepper{Value: min, Min: min, Max: max} }

func (s *Stepper) format(v int) string {
	if s.Format != nil {
		return s.Format(v)
	}
	return strconv.Itoa(v)
}

type stepButton struct {
	label string
	delta int
	rect  Rect
}

// layout places the buttons and the box that shows the value, left to right. The
// box is as wide as the widest value, so nothing moves as the value changes.
func (s *Stepper) layout(area Rect, lineH int, textW func(string) int) ([]stepButton, Rect) {
	h := lineH + lineH/2
	pad := max(2, lineH/4)
	vw := 0
	for v := s.Min; v <= s.Max; v++ {
		vw = max(vw, textW(s.format(v)))
	}
	vw += lineH
	down := []stepButton{{label: "<", delta: -1}}
	up := []stepButton{{label: ">", delta: 1}}
	if s.Big > 0 {
		down = append([]stepButton{{label: "<<", delta: -s.Big}}, down...)
		up = append(up, stepButton{label: ">>", delta: s.Big})
	}
	x := area.X
	place := func(b *stepButton) {
		w := max(h, textW(b.label)+2*pad)
		b.rect = Rect{X: x, Y: area.Y, W: w, H: h}
		x += w + pad
	}
	for i := range down {
		place(&down[i])
	}
	value := Rect{X: x, Y: area.Y, W: vw, H: h}
	x += vw + pad
	for i := range up {
		place(&up[i])
	}
	return append(down, up...), value
}

func (s *Stepper) enabled(b stepButton) bool {
	if s.Disabled {
		return false
	}
	if b.delta < 0 {
		return s.Value > s.Min
	}
	return s.Value < s.Max
}

// ButtonRect is the rectangle of the button with a label ("<", ">", "<<" or ">>")
// inside area, for a host that needs to point at it.
func (s *Stepper) ButtonRect(label string, area Rect, lineH int, textW func(string) int) (Rect, bool) {
	btns, _ := s.layout(area, lineH, textW)
	for _, b := range btns {
		if b.label == label {
			return b.rect, true
		}
	}
	return Rect{}, false
}

// Update returns the new value when a button that can act was pressed this frame.
func (s *Stepper) Update(in Input, area Rect, lineH int, textW func(string) int) (int, bool) {
	s.mouseX, s.mouseY = in.MouseX, in.MouseY
	if !in.MousePressed {
		return s.Value, false
	}
	btns, _ := s.layout(area, lineH, textW)
	for _, b := range btns {
		if b.rect.Contains(in.MouseX, in.MouseY) && s.enabled(b) {
			s.Value = clampInt(s.Value+b.delta, s.Min, s.Max)
			return s.Value, true
		}
	}
	return s.Value, false
}

// Draw paints the stepper inside area at the given text scale.
func (s *Stepper) Draw(r Renderer, th Theme, area Rect, scale int) {
	lh := r.LineHeight(scale)
	btns, value := s.layout(area, lh, func(t string) int { return r.TextWidth(t, scale) })
	for _, b := range btns {
		fill, text := th.Button, th.ButtonText
		switch {
		case !s.enabled(b):
			text = th.Disabled
		case b.rect.Contains(s.mouseX, s.mouseY):
			fill = th.ButtonHot
		}
		r.FillRect(b.rect, fill)
		r.StrokeRect(b.rect, th.Border, 1)
		r.DrawText(b.label, b.rect.X+(b.rect.W-r.TextWidth(b.label, scale))/2, b.rect.Y+(b.rect.H-lh)/2, scale, text)
	}
	r.FillRect(value, th.Panel)
	r.StrokeRect(value, th.Border, 1)
	text := th.Text
	if s.Disabled {
		text = th.Disabled
	}
	label := s.format(s.Value)
	r.DrawText(label, value.X+(value.W-r.TextWidth(label, scale))/2, value.Y+(value.H-lh)/2, scale, text)
}
