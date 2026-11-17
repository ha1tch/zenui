package zenui

// BarButton is one button of a ButtonBar.
type BarButton struct {
	ID    string
	Label string
	// Icon is the glyph that stands for the button when the bar shows icons.
	// Zero means the button always shows its label.
	Icon     rune
	On       bool // drawn as chosen, for a tool or a switch
	Disabled bool
}

// ButtonBar is a row of buttons that wraps onto further rows to fit its area.
// Its buttons show words, or icons when UseIcons is set. Like List, its layout is
// a pure function of the area and the text metrics, so Update does not depend on
// a previous Draw.
type ButtonBar struct {
	Buttons []BarButton
	// Centred puts each row of buttons in the middle of the area's width instead
	// of at its left edge.
	Centred bool
	// Icons draws the buttons' glyphs, through a renderer over an icon font, as
	// List's toolbar does. UseIcons asks for the glyphs instead of the words. A
	// button with no Icon, or a bar with no Icons renderer, still shows its word,
	// so no button is ever blank. An icon button names itself in a small tip while
	// the pointer rests on it.
	Icons    Renderer
	UseIcons bool

	mouseX, mouseY int
}

// NewButtonBar returns an empty bar.
func NewButtonBar() *ButtonBar { return &ButtonBar{} }

func (b *ButtonBar) iconMode(btn BarButton) bool {
	return b.UseIcons && b.Icons != nil && btn.Icon != 0
}

// Layout returns the rectangle of each button inside area, flowing left to
// right and wrapping, and the total height used. An icon button is square; a
// word button is as wide as its word. With Centred, each row is moved to the
// middle of the area's width; a row wider than the area stays at its left edge.
func (b *ButtonBar) Layout(area Rect, lineH int, textW func(string) int) ([]Rect, int) {
	pad := max(2, lineH/4)
	h := lineH + lineH/2
	rects := make([]Rect, len(b.Buttons))
	x, y := area.X, area.Y
	rowStart := 0
	finish := func(end int) { // the row of buttons rowStart to end-1 is complete
		if !b.Centred || end == rowStart {
			return
		}
		used := rects[end-1].X + rects[end-1].W - area.X
		shift := max(0, (area.W-used)/2)
		for i := rowStart; i < end; i++ {
			rects[i].X += shift
		}
	}
	for i, btn := range b.Buttons {
		w := textW(btn.Label) + lineH
		if b.iconMode(btn) {
			w = h
		}
		if x > area.X && x+w > area.X+area.W {
			finish(i)
			rowStart = i
			x = area.X
			y += h + pad
		}
		rects[i] = Rect{X: x, Y: y, W: w, H: h}
		x += w + pad
	}
	finish(len(b.Buttons))
	used := 0
	if len(rects) > 0 {
		used = y + h - area.Y
	}
	return rects, used
}

// Update returns the ID of the button the person pressed this frame, if any.
// A disabled button does not answer.
func (b *ButtonBar) Update(in Input, area Rect, lineH int, textW func(string) int) (string, bool) {
	b.mouseX, b.mouseY = in.MouseX, in.MouseY
	if !in.MousePressed {
		return "", false
	}
	rects, _ := b.Layout(area, lineH, textW)
	for i, r := range rects {
		if r.Contains(in.MouseX, in.MouseY) && !b.Buttons[i].Disabled {
			return b.Buttons[i].ID, true
		}
	}
	return "", false
}

// Draw paints the bar inside area at the given text scale.
func (b *ButtonBar) Draw(r Renderer, th Theme, area Rect, scale int) {
	lh := r.LineHeight(scale)
	rects, _ := b.Layout(area, lh, func(s string) int { return r.TextWidth(s, scale) })
	tip := -1
	for i, btn := range b.Buttons {
		fill, text := th.Button, th.ButtonText
		switch {
		case btn.Disabled:
			text = th.Disabled
		case btn.On:
			fill, text = th.SelFill, th.SelText
		case rects[i].Contains(b.mouseX, b.mouseY):
			fill = th.ButtonHot
		}
		r.FillRect(rects[i], fill)
		r.StrokeRect(rects[i], th.Border, 1)
		if b.iconMode(btn) {
			g := string(btn.Icon)
			b.Icons.DrawText(g, rects[i].X+(rects[i].W-b.Icons.TextWidth(g, scale))/2, rects[i].Y+(rects[i].H-b.Icons.LineHeight(scale))/2, scale, text)
			if rects[i].Contains(b.mouseX, b.mouseY) {
				tip = i
			}
			continue
		}
		r.DrawText(btn.Label, rects[i].X+(rects[i].W-r.TextWidth(btn.Label, scale))/2, rects[i].Y+(rects[i].H-lh)/2, scale, text)
	}
	if tip >= 0 {
		b.drawTip(r, th, area, rects[tip], b.Buttons[tip].Label, scale)
	}
}

// drawTip names an icon button in a small box just below it, kept inside area.
// The tip's text is one zoom step smaller than the bar's, and never below 1.
func (b *ButtonBar) drawTip(r Renderer, th Theme, area, btn Rect, label string, scale int) {
	scale = max(1, scale-1)
	lh := r.LineHeight(scale)
	w := r.TextWidth(label, scale) + lh
	h := lh + lh/2
	x := clampInt(btn.X, area.X, area.X+max(0, area.W-w))
	box := Rect{X: x, Y: btn.Y + btn.H + 2, W: w, H: h}
	r.FillRect(box, th.Panel)
	r.StrokeRect(box, th.Border, 1)
	r.DrawText(label, box.X+lh/2, box.Y+(h-lh)/2, scale, th.Text)
}
