package zenui

import "math"

// The renderer has rectangles and lines but no circle, so circles are drawn a row at a
// time: each row of a disc is one rectangle, its ends worked out from the circle. The
// ends are mirrored, so a disc is the same left and right, and it is the same top and
// bottom because the rows are measured from the middle.

// discSpans gives, for each row of a size by size box, how far in from the left edge the
// disc of the radius begins on that row; it ends as far in from the right edge. A row the
// disc does not reach is -1. The disc is centred in the box.
func discSpans(size int, radius float64) []int {
	c := float64(size) / 2
	out := make([]int, size)
	for j := range out {
		dy := float64(j) + 0.5 - c
		if radius*radius <= dy*dy {
			out[j] = -1
			continue
		}
		in := int(math.Floor(c - math.Sqrt(radius*radius-dy*dy) + 0.5))
		in = max(in, 0)
		if 2*in >= size {
			in = -1
		}
		out[j] = in
	}
	return out
}

// fillDisc fills the disc of the radius, centred in the box, which is square.
func fillDisc(r Renderer, box Rect, radius float64, c Colour) {
	for j, in := range discSpans(box.W, radius) {
		if in >= 0 {
			r.FillRect(Rect{X: box.X + in, Y: box.Y + j, W: box.W - 2*in, H: 1}, c)
		}
	}
}

// strokeCircle draws a ring that fills the box and is thick pixels thick. It is drawn as
// what is left of the outer disc outside the inner, so that nothing has to be painted
// over the middle and whatever is behind shows through.
func strokeCircle(r Renderer, box Rect, thick int, c Colour) {
	half := float64(box.W) / 2
	outer := discSpans(box.W, half)
	inner := discSpans(box.W, half-float64(thick))
	for j, o := range outer {
		if o < 0 {
			continue
		}
		i := inner[j]
		if i < 0 || i <= o {
			r.FillRect(Rect{X: box.X + o, Y: box.Y + j, W: box.W - 2*o, H: 1}, c)
			continue
		}
		r.FillRect(Rect{X: box.X + o, Y: box.Y + j, W: i - o, H: 1}, c)
		r.FillRect(Rect{X: box.X + box.W - i, Y: box.Y + j, W: i - o, H: 1}, c)
	}
}
