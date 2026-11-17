package zenui

import "math"

// Tabs chooses which of several panels is showing. The tabs share the area's width
// between them, so they always fit one row and can never wrap onto what is below; a label
// too long for its tab is clipped. The tab chosen is filled, and a rule runs under the
// strip.
type Tabs struct {
	Labels   []string
	Selected int

	mouseX, mouseY int
}

// Height is the height of the strip.
func (t *Tabs) Height(lineH int) int { return controlHeight(lineH) }

// Layout returns each tab's rectangle: equal widths that add up to the area's, with the
// remainder given to the first tabs.
func (t *Tabs) Layout(area Rect, lineH int) []Rect {
	n := len(t.Labels)
	rects := make([]Rect, n)
	if n == 0 {
		return rects
	}
	base, extra := area.W/n, area.W%n
	x := area.X
	for i := range rects {
		w := base
		if i < extra {
			w++
		}
		rects[i] = Rect{X: x, Y: area.Y, W: w, H: t.Height(lineH)}
		x += w
	}
	return rects
}

// Update handles a click on a tab other than the chosen one. It returns the tab chosen
// and whether the click changed it.
func (t *Tabs) Update(in Input, area Rect, lineH int) (int, bool) {
	t.mouseX, t.mouseY = in.MouseX, in.MouseY
	if !in.MousePressed {
		return t.Selected, false
	}
	for i, r := range t.Layout(area, lineH) {
		if r.Contains(in.MouseX, in.MouseY) && i != t.Selected {
			t.Selected = i
			return i, true
		}
	}
	return t.Selected, false
}

// tabCorner is the radius of the rounded ends of the strip: slight, a sixth of a line.
func tabCorner(lineH, h int) int { return min(max(2, min(lineH/6, 6)), h/2) }

// roundInsets says, for each of the h rows of a rectangle with rounded corners of the
// radius, how many pixels the corners take off each end of the row: the top rows and the
// bottom ones, mirrored, and nothing from the rows between.
func roundInsets(h, radius int) []int {
	out := make([]int, h)
	for j := 0; j < radius && j < h; j++ {
		dy := float64(radius) - float64(j) - 0.5
		x := float64(radius) - math.Sqrt(float64(radius*radius)-dy*dy)
		in := int(math.Floor(x + 0.5))
		out[j], out[h-1-j] = in, in
	}
	return out
}

func (t *Tabs) Draw(r Renderer, th Theme, area Rect, scale int) {
	lh := r.LineHeight(scale)
	rects := t.Layout(area, lh)
	h := t.Height(lh)
	ins := roundInsets(h, tabCorner(lh, h))
	last := len(rects) - 1
	// edge is how wide the border is on a row at an end of the strip, for a row that is not
	// the first: a pixel, and more where the corner slopes steeply, so that the rounded
	// border has no gaps. Below the middle the insets only grow, so it is a pixel there.
	edge := func(j int) int { return max(1, ins[j-1]-ins[j]) }
	for i, b := range rects {
		fill, text := th.Panel, th.DimText
		switch {
		case i == t.Selected:
			fill, text = th.SelFill, th.SelText
		case b.Contains(t.mouseX, t.mouseY):
			// a tab the pointer is over must not look chosen, and in some themes the hot colour
			// is the selection colour: then it keeps its fill and only its label darkens
			text = th.ButtonText
			if th.ButtonHot != th.SelFill {
				fill = th.ButtonHot
			}
		}
		for j := 0; j < h; j++ { // a row at a time: the backend has no arc, only rectangles
			l, rr := 0, 0
			if i == 0 {
				l = ins[j]
			}
			if i == last {
				rr = ins[j]
			}
			span := Rect{X: b.X + l, Y: b.Y + j, W: b.W - l - rr, H: 1}
			if span.W <= 0 {
				continue
			}
			r.FillRect(span, fill)
			if j == 0 || j == h-1 {
				r.FillRect(span, th.Border)
				continue
			}
			lw, rw := 1, 1
			if i == 0 {
				lw = edge(j)
			}
			if i == last {
				rw = edge(j)
			}
			r.FillRect(Rect{X: span.X, Y: span.Y, W: min(lw, span.W), H: 1}, th.Border)
			r.FillRect(Rect{X: span.X + span.W - min(rw, span.W), Y: span.Y, W: min(rw, span.W), H: 1}, th.Border)
		}
		if i != t.Selected {
			// a thin line sets a tab that is not chosen apart from the rule under the strip;
			// the chosen tab has none, and runs into the rule
			l, rr := 0, 0
			if i == 0 {
				l = ins[h-3]
			}
			if i == last {
				rr = ins[h-3]
			}
			r.FillRect(Rect{X: b.X + l, Y: b.Y + h - 3, W: b.W - l - rr, H: 1}, th.Border)
		}
		r.Clip(b)
		w := r.TextWidth(t.Labels[i], scale)
		r.DrawText(t.Labels[i], b.X+max(2, (b.W-w)/2), b.Y+(b.H-lh)/2, scale, text)
		r.ClipEnd()
	}
	// the rule under the strip, over the foot of the tabs, with the strip's rounded ends
	for _, j := range []int{h - 2, h - 1} {
		r.FillRect(Rect{X: area.X + ins[j], Y: area.Y + j, W: area.W - 2*ins[j], H: 1}, th.SelFill)
	}
}
