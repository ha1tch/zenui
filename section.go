package zenui

// SectionHeight is the height of a section header.
func SectionHeight(lineH int) int { return lineH + lineH/3 }

// DrawSection draws a section's title in the heading colour with a rule after it that
// runs to the right edge of the area, to set a group of controls apart from the next. A
// title too wide for the area is clipped and has no rule.
func DrawSection(r Renderer, th Theme, area Rect, scale int, title string) {
	lh := r.LineHeight(scale)
	h := SectionHeight(lh)
	y := area.Y + (h-lh)/2
	r.Clip(Rect{X: area.X, Y: area.Y, W: area.W, H: h})
	r.DrawText(title, area.X, y, scale, th.DirText)
	r.ClipEnd()
	end := area.X + r.TextWidth(title, scale) + controlGap(lh)
	if end < area.X+area.W {
		mid := area.Y + h/2
		r.DrawLine(end, mid, area.X+area.W, mid, 1, th.SeparatorColour)
	}
}
