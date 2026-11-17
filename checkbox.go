package zenui

// Checkbox is an independent on/off choice with its own label: a box, with a cross in it
// when checked, drawn as the menus draw theirs. A host sets Checked from what it holds
// each frame, and acts when Update says a click changed it. A click on the label counts.
type Checkbox struct {
	Label    string
	Checked  bool
	Disabled bool

	mouseX, mouseY int
}

// Height is the height of the control's row.
func (c *Checkbox) Height(lineH int) int { return controlHeight(lineH) }

// Bounds is the rectangle that answers a click.
func (c *Checkbox) Bounds(area Rect, lineH int, textW func(string) int) Rect {
	return markedBounds(area, lineH, controlBox(lineH), c.Label, textW)
}

// Update handles a click. It returns the value and whether the click changed it.
func (c *Checkbox) Update(in Input, area Rect, lineH int, textW func(string) int) (bool, bool) {
	c.mouseX, c.mouseY = in.MouseX, in.MouseY
	if c.Disabled || !in.MousePressed {
		return c.Checked, false
	}
	if c.Bounds(area, lineH, textW).Contains(in.MouseX, in.MouseY) {
		c.Checked = !c.Checked
		return c.Checked, true
	}
	return c.Checked, false
}

func (c *Checkbox) Draw(r Renderer, th Theme, area Rect, scale int) {
	lh := r.LineHeight(scale)
	b := c.Bounds(area, lh, func(s string) int { return r.TextWidth(s, scale) })
	size := controlBox(lh)
	box := Rect{X: b.X, Y: b.Y + (b.H-size)/2, W: size, H: size}
	text, edge := th.Text, th.CheckboxEmptyColour
	if c.Checked {
		edge = checkedColour(th, text)
	}
	if c.Disabled {
		text, edge = th.Disabled, th.Disabled
	}
	if !c.Disabled && b.Contains(c.mouseX, c.mouseY) {
		r.StrokeRect(b, th.Border, 1)
	}
	thick := 1
	if size >= 12 {
		thick = 2
	}
	r.StrokeRect(box, edge, thick)
	if c.Checked {
		if size >= 12 {
			// The cross runs from corner to corner of the box. The box covers size pixels,
			// from X to X+size, so a line between its corners is centred in it; clipped to
			// the box, a thick line's square ends cannot poke out of the corners.
			r.Clip(box)
			r.DrawLine(box.X, box.Y, box.X+size, box.Y+size, thick, edge)
			r.DrawLine(box.X+size, box.Y, box.X, box.Y+size, thick, edge)
			r.ClipEnd()
		} else { // the smallest box keeps its cross a pixel in from its edge
			in := thick + 1
			r.DrawLine(box.X+in, box.Y+in, box.X+size-1-in, box.Y+size-1-in, thick, edge)
			r.DrawLine(box.X+size-1-in, box.Y+in, box.X+in, box.Y+size-1-in, thick, edge)
		}
	}
	drawControlLabel(r, b, size, c.Label, lh, scale, text)
}
