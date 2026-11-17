package zenui

// Shared measures and colours of the labelled controls: Checkbox, Switch, RadioGroup.
// Every size comes from the line height, so that the controls follow the text.

// controlBox is the side of a control's mark: four fifths of a line, as tall as the text
// allows with a fifth to spare, so that the mark is easy to see and to hit.
func controlBox(lineH int) int { return max(8, lineH*8/10) }

// MarkSize is the side of the mark of a Checkbox or a RadioGroup option at this line
// height, for a host that has to line something up with them. A Switch is smaller.
func MarkSize(lineH int) int { return controlBox(lineH) }

// controlGap is the space between a mark and its label.
func controlGap(lineH int) int { return max(4, lineH/3) }

// controlHeight is the height of a control's row: a line and a half, as a TextField's.
func controlHeight(lineH int) int { return lineH + lineH/2 }

// markedBounds is the rectangle that answers a click for a mark with a label beside it:
// the mark, the gap and the label, as wide as that and no wider than the area.
func markedBounds(area Rect, lineH, markW int, label string, textW func(string) int) Rect {
	w := markW + controlGap(lineH) + textW(label)
	return Rect{X: area.X, Y: area.Y, W: min(w, area.W), H: controlHeight(lineH)}
}

// checkedColour is the colour of a checked mark: the theme's own, if it has one.
func checkedColour(th Theme, def Colour) Colour {
	if th.UseCheckboxColour {
		return th.CheckboxColour
	}
	return def
}

// tickColour is the colour of the sign of the current choice among several: the
// theme's own, if it has one, as in a menu.
func tickColour(th Theme, def Colour) Colour {
	if th.UseCheckmarkColour {
		return th.CheckmarkColour
	}
	return def
}

// drawControlLabel draws a label after a mark, a line high and centred in the bounds,
// clipped to them.
func drawControlLabel(r Renderer, b Rect, markW int, label string, lineH, scale int, col Colour) {
	r.Clip(b)
	r.DrawText(label, b.X+markW+controlGap(lineH), b.Y+(b.H-lineH)/2, scale, col)
	r.ClipEnd()
}
