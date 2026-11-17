package zenui

// Switch is an on/off choice that takes effect as it is flipped, and reads as a mode
// more than a box to tick: a thin track with a round knob on it, centred on the track's
// left end when off and on its right end when on, so that half the knob is over the track
// and half is outside it. It is used the way Checkbox is.
type Switch struct {
	Label    string
	On       bool
	Disabled bool

	mouseX, mouseY int
}

// switchBase is two thirds of a line: the measure the switch is built from.
func switchBase(lineH int) int { return max(8, lineH*2/3) }

// switchTrackH is the track's height: four fifths of the base, so that the track is thin
// and the knob stands well out of it.
func switchTrackH(lineH int) int { return max(6, switchBase(lineH)*8/10) }

// trackW is the track's length: twice the base.
func trackW(lineH int) int { return 2 * switchBase(lineH) }

// knobOverhang is how far the knob would stand out of a track the height of the base.
func knobOverhang(base int) int { return max(1, base/12) }

// knobSize is the knob's diameter, a little more than the base, and well over the track's
// height.
func knobSize(lineH int) int {
	b := switchBase(lineH)
	return b + 2*knobOverhang(b)
}

// switchMarkW is the width of the mark: the track, and half a knob out at each end of it.
func switchMarkW(lineH int) int { return trackW(lineH) + knobSize(lineH) }

// Height is the height of the control's row.
func (s *Switch) Height(lineH int) int { return controlHeight(lineH) }

// Bounds is the rectangle that answers a click.
func (s *Switch) Bounds(area Rect, lineH int, textW func(string) int) Rect {
	return markedBounds(area, lineH, switchMarkW(lineH), s.Label, textW)
}

// Update handles a click. It returns the value and whether the click changed it.
func (s *Switch) Update(in Input, area Rect, lineH int, textW func(string) int) (bool, bool) {
	s.mouseX, s.mouseY = in.MouseX, in.MouseY
	if s.Disabled || !in.MousePressed {
		return s.On, false
	}
	if s.Bounds(area, lineH, textW).Contains(in.MouseX, in.MouseY) {
		s.On = !s.On
		return s.On, true
	}
	return s.On, false
}

func (s *Switch) Draw(r Renderer, th Theme, area Rect, scale int) {
	lh := r.LineHeight(scale)
	b := s.Bounds(area, lh, func(t string) int { return r.TextWidth(t, scale) })
	kd, trackH := knobSize(lh), switchTrackH(lh)
	// the track starts half a knob in from the bounds, so that the knob at its left end is in them
	track := Rect{X: b.X + kd/2, Y: b.Y + (b.H-trackH)/2, W: trackW(lh), H: trackH}
	fill, knob, text := th.Button, th.CheckboxEmptyColour, th.Text
	edge := track.X // the knob's centre is on the end of the track it is at
	if s.On {
		fill, knob = th.SelFill, th.SelText
		edge = track.X + track.W
	}
	if s.Disabled {
		fill, knob, text = th.Button, th.Disabled, th.Disabled
	}
	if !s.Disabled && b.Contains(s.mouseX, s.mouseY) {
		r.StrokeRect(b, th.Border, 1)
	}
	r.FillRect(track, fill)
	r.StrokeRect(track, th.Border, 1)
	// the knob goes on last, over the track and over its border
	fillDisc(r, Rect{X: edge - kd/2, Y: track.Y + (trackH-kd)/2, W: kd, H: kd}, float64(kd)/2, knob)
	drawControlLabel(r, b, switchMarkW(lh), s.Label, lh, scale, text)
}
