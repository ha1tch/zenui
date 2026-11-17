package zenui

import "fmt"

// DrawProgressBar draws a horizontal bar in rect: the theme's field
// colour for the whole, its selection colour for done out of total, and
// a border round it. A total of zero or less is a bar with nothing
// filled in, for work whose size is not known yet; done past total fills
// the bar and no more.
func DrawProgressBar(r Renderer, rect Rect, done, total int64, theme Theme) {
	r.FillRect(rect, theme.Field)
	if total > 0 && done > 0 {
		if done > total {
			done = total
		}
		fill := rect
		fill.W = int(int64(rect.W) * done / total)
		r.FillRect(fill, theme.SelFill)
	}
	r.StrokeRect(rect, theme.Border, 1)
}

// ProgressText is the figures to put by a progress bar that counts
// bytes: " 25%   1.0 MB of 4.0 MB" when the total is known, the amount
// done when only that is, and nothing before anything is.
func ProgressText(done, total int64) string {
	mb := func(n int64) string { return fmt.Sprintf("%.1f MB", float64(n)/(1<<20)) }
	switch {
	case total > 0:
		return fmt.Sprintf("%3d%%   %s of %s", done*100/total, mb(done), mb(total))
	case done > 0:
		return mb(done)
	}
	return ""
}
