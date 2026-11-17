package zenui

// RadioGroup is a choice of exactly one among a few options, all shown: each has a
// box, with a filled square in it for the one chosen, so that it cannot be taken for a
// Checkbox, whose mark is a cross. The options stack, or with Horizontal flow along the
// row and wrap when they do not fit, and Layout says how tall that came to, so that
// nothing placed under the group can be covered by it.
type RadioGroup struct {
	Options    []string
	Selected   int
	Horizontal bool
	Disabled   bool

	mouseX, mouseY int
}

// Layout returns the rectangle that answers a click for each option, and the height
// all of them take together.
func (g *RadioGroup) Layout(area Rect, lineH int, textW func(string) int) ([]Rect, int) {
	h, size := controlHeight(lineH), controlBox(lineH)
	rects := make([]Rect, len(g.Options))
	x, y := area.X, area.Y
	for i, o := range g.Options {
		w := min(size+controlGap(lineH)+textW(o), area.W)
		if g.Horizontal {
			if x > area.X && x+w > area.X+area.W {
				x, y = area.X, y+h+2
			}
			rects[i] = Rect{X: x, Y: y, W: w, H: h}
			x += w + lineH
			continue
		}
		rects[i] = Rect{X: area.X, Y: y, W: w, H: h}
		y += h + 2
	}
	used := 0
	if len(rects) > 0 {
		last := rects[len(rects)-1]
		used = last.Y + last.H - area.Y
	}
	return rects, used
}

// Update handles a click on an option. It returns the option chosen and whether the
// click changed it.
func (g *RadioGroup) Update(in Input, area Rect, lineH int, textW func(string) int) (int, bool) {
	g.mouseX, g.mouseY = in.MouseX, in.MouseY
	if g.Disabled || !in.MousePressed {
		return g.Selected, false
	}
	rects, _ := g.Layout(area, lineH, textW)
	for i, r := range rects {
		if r.Contains(in.MouseX, in.MouseY) && i != g.Selected {
			g.Selected = i
			return i, true
		}
	}
	return g.Selected, false
}

func (g *RadioGroup) Draw(r Renderer, th Theme, area Rect, scale int) {
	lh := r.LineHeight(scale)
	rects, _ := g.Layout(area, lh, func(s string) int { return r.TextWidth(s, scale) })
	size := controlBox(lh)
	for i, b := range rects {
		text, edge := th.Text, th.CheckboxEmptyColour
		if i == g.Selected {
			edge = tickColour(th, text)
		}
		if g.Disabled {
			text, edge = th.Disabled, th.Disabled
		}
		if !g.Disabled && b.Contains(g.mouseX, g.mouseY) {
			r.StrokeRect(b, th.Border, 1)
		}
		box := Rect{X: b.X, Y: b.Y + (b.H-size)/2, W: size, H: size}
		strokeCircle(r, box, max(2, size/8), edge) // round, so that it cannot be taken for a checkbox
		if i == g.Selected {
			fillDisc(r, box, float64(size)/2*0.55, edge)
		}
		drawControlLabel(r, b, size, g.Options[i], lh, scale, text)
	}
}
