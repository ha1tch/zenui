package zenui

// Swatches is a row of colour squares, one of them selected, that wraps onto
// further rows to fit its area. Its geometry is a pure function of the area and
// the line height.
type Swatches struct {
	Colours  []Colour
	Selected int
	Disabled bool
}

// Layout returns the square of each swatch inside area, flowing left to right
// and wrapping, and the total height used.
func (s *Swatches) Layout(area Rect, lineH int) ([]Rect, int) {
	pad := max(2, lineH/4)
	h := lineH + lineH/2
	rects := make([]Rect, len(s.Colours))
	x, y := area.X, area.Y
	for i := range s.Colours {
		if x > area.X && x+h > area.X+area.W {
			x = area.X
			y += h + pad
		}
		rects[i] = Rect{X: x, Y: y, W: h, H: h}
		x += h + pad
	}
	used := 0
	if len(rects) > 0 {
		used = y + h - area.Y
	}
	return rects, used
}

// Update returns the index of the swatch that was pressed this frame, with true
// when that changed the selection.
func (s *Swatches) Update(in Input, area Rect, lineH int) (int, bool) {
	if !in.MousePressed || s.Disabled {
		return s.Selected, false
	}
	rects, _ := s.Layout(area, lineH)
	for i, r := range rects {
		if r.Contains(in.MouseX, in.MouseY) {
			if i == s.Selected {
				return i, false
			}
			s.Selected = i
			return i, true
		}
	}
	return s.Selected, false
}

// Draw paints the swatches inside area at the given text scale. The selected one
// has a ring round it, in the text colour, which shows against any swatch and any
// background.
func (s *Swatches) Draw(r Renderer, th Theme, area Rect, scale int) {
	rects, _ := s.Layout(area, r.LineHeight(scale))
	for i, rc := range rects {
		r.FillRect(rc, s.Colours[i])
		r.StrokeRect(rc, th.Border, 1)
		if i == s.Selected {
			r.StrokeRect(Rect{X: rc.X - 3, Y: rc.Y - 3, W: rc.W + 6, H: rc.H + 6}, th.Text, 2)
		}
	}
}
