package zenui

import (
	"fmt"
	"math/rand"
	"testing"
)

// ctlRenderer records what the controls draw. Text is 8 pixels a character at scale 1,
// and a line is 16.
type ctlRenderer struct {
	seq            int
	fills, strokes []ctlRect
	lines          []ctlLine
	texts          []ctlText
	clips          []Rect
}
type ctlRect struct {
	R     Rect
	C     Colour
	Thick int
	Seq   int // the order of drawing, across fills and strokes
}
type ctlLine struct {
	X1, Y1, X2, Y2, Thick int
	C                     Colour
}
type ctlText struct {
	S    string
	X, Y int
	C    Colour
}

func (r *ctlRenderer) FillRect(rc Rect, c Colour) {
	r.seq++
	r.fills = append(r.fills, ctlRect{rc, c, 0, r.seq})
}
func (r *ctlRenderer) StrokeRect(rc Rect, c Colour, t int) {
	r.seq++
	r.strokes = append(r.strokes, ctlRect{rc, c, t, r.seq})
}
func (r *ctlRenderer) FillRectGradientV(Rect, Colour, Colour)         {}
func (r *ctlRenderer) FillRectGradientHMultiply(Rect, Colour, Colour) {}
func (r *ctlRenderer) DrawLine(x1, y1, x2, y2, t int, c Colour) {
	r.lines = append(r.lines, ctlLine{x1, y1, x2, y2, t, c})
}
func (r *ctlRenderer) DrawText(s string, x, y, scale int, c Colour) {
	r.texts = append(r.texts, ctlText{s, x, y, c})
}
func (r *ctlRenderer) TextWidth(s string, scale int) int { return 8 * len(s) * scale }
func (r *ctlRenderer) LineHeight(scale int) int          { return 16 * scale }
func (r *ctlRenderer) Clip(rc Rect)                      { r.clips = append(r.clips, rc) }
func (r *ctlRenderer) ClipEnd()                          {}

const ctlLH = 16

func ctlTextW(s string) int { return 8 * len(s) }

func ctlPress(x, y int) Input {
	return Input{MouseX: x, MouseY: y, MousePressed: true, MouseDown: true}
}

func inside(r, outer Rect) bool {
	return r.X >= outer.X && r.Y >= outer.Y && r.X+r.W <= outer.X+outer.W && r.Y+r.H <= outer.Y+outer.H
}

func overlap(a, b Rect) bool {
	return a.X < b.X+b.W && b.X < a.X+a.W && a.Y < b.Y+b.H && b.Y < a.Y+a.H
}

func TestCheckboxTogglesOnAClickOnTheBoxOrTheLabelAndOnlyThere(t *testing.T) {
	c := &Checkbox{Label: "Hold to fire"}
	area := Rect{X: 20, Y: 30, W: 300, H: 24}
	b := c.Bounds(area, ctlLH, ctlTextW)
	if want := controlBox(ctlLH) + controlGap(ctlLH) + ctlTextW(c.Label); b.W != want || b.H != c.Height(ctlLH) || b.X != area.X || b.Y != area.Y {
		t.Fatalf("bounds %+v, want width %d and height %d at the area's corner", b, want, c.Height(ctlLH))
	}
	if v, ch := c.Update(Input{MouseX: b.X + 2, MouseY: b.Y + 2}, area, ctlLH, ctlTextW); v || ch {
		t.Error("a frame with no press changed it")
	}
	if v, ch := c.Update(ctlPress(b.X+2, b.Y+b.H/2), area, ctlLH, ctlTextW); !v || !ch || !c.Checked {
		t.Errorf("a click on the box gave %v %v", v, ch)
	}
	if v, ch := c.Update(ctlPress(b.X+b.W-2, b.Y+b.H/2), area, ctlLH, ctlTextW); v || !ch || c.Checked {
		t.Errorf("a click on the label's end gave %v %v: want it unchecked again", v, ch)
	}
	for _, p := range [][2]int{{b.X + b.W + 1, b.Y + 2}, {b.X - 1, b.Y + 2}, {b.X + 2, b.Y + b.H}, {b.X + 2, b.Y - 1}} {
		if _, ch := c.Update(ctlPress(p[0], p[1]), area, ctlLH, ctlTextW); ch {
			t.Errorf("a click at %v, outside, changed it", p)
		}
	}
	c.Disabled = true
	if _, ch := c.Update(ctlPress(b.X+2, b.Y+2), area, ctlLH, ctlTextW); ch || c.Checked {
		t.Error("a disabled checkbox changed")
	}
	// a label too long for the area: the bounds stop at the area's edge
	long := &Checkbox{Label: "a label much longer than the room it is given"}
	if lb := long.Bounds(Rect{X: 0, Y: 0, W: 100, H: 24}, ctlLH, ctlTextW); lb.W != 100 {
		t.Errorf("bounds of a long label are %d wide in an area 100 wide", lb.W)
	}
}

func TestCheckboxDrawsABoxAndACrossOnlyWhenChecked(t *testing.T) {
	th := LoadTheme(ThemeSpectrum)
	area := Rect{X: 10, Y: 10, W: 300, H: 24}
	for _, checked := range []bool{false, true} {
		r := &ctlRenderer{}
		c := &Checkbox{Label: "Bright", Checked: checked}
		c.Draw(r, th, area, 1)
		b := c.Bounds(area, ctlLH, ctlTextW)
		if len(r.strokes) != 1 || !inside(r.strokes[0].R, b) {
			t.Fatalf("checked %v: strokes %+v, want one box inside the bounds %+v", checked, r.strokes, b)
		}
		if wantLines := map[bool]int{false: 0, true: 2}[checked]; len(r.lines) != wantLines {
			t.Errorf("checked %v: %d lines, want %d: a cross only when checked", checked, len(r.lines), wantLines)
		}
		wantEdge := th.CheckboxEmptyColour
		if checked && th.UseCheckboxColour {
			wantEdge = th.CheckboxColour
		}
		if checked && !th.UseCheckboxColour {
			wantEdge = th.Text
		}
		if r.strokes[0].C != wantEdge {
			t.Errorf("checked %v: the box is %+v, want %+v", checked, r.strokes[0].C, wantEdge)
		}
		if len(r.texts) != 1 || r.texts[0].S != "Bright" || r.texts[0].X < b.X+controlBox(ctlLH) {
			t.Errorf("checked %v: the label is %+v: want it after the box", checked, r.texts)
		}
		for _, l := range r.lines {
			if l.X1 < b.X || l.X2 > b.X+controlBox(ctlLH) {
				t.Errorf("checked %v: the cross %+v runs out of the box", checked, l)
			}
		}
	}
	r := &ctlRenderer{}
	(&Checkbox{Label: "x", Disabled: true, Checked: true}).Draw(r, th, area, 1)
	if r.strokes[0].C != th.Disabled || r.texts[0].C != th.Disabled {
		t.Error("a disabled checkbox is not drawn in the disabled colour")
	}
}

func TestSwitchKnobIsRoundOnTopOfAThinTrackAndCentredOnItsEnd(t *testing.T) {
	th := LoadTheme(ThemeSpectrum)
	for _, lh := range []int{8, 16, 24, 48} {
		base := switchBase(lh)
		kd := knobSize(lh)
		centre := [2]int{}
		for i, on := range []bool{false, true} {
			rec := &ctlRenderer{}
			s := &Switch{Label: "Pixel lock", On: on, mouseX: -1, mouseY: -1}
			area := Rect{X: 50, Y: 50, W: 500, H: 3 * lh}
			s.Draw(&lhRenderer{ctlRenderer: rec, lh: lh}, th, area, 1)
			track, border := rec.fills[0], rec.strokes[0]
			what := func(format string, a ...any) string {
				return fmt.Sprintf("line %d, on %v: ", lh, on) + fmt.Sprintf(format, a...)
			}
			// a thin track: four fifths of what it was, and as long as it was
			if track.R.H*10 > base*8 || track.R.H*10 < base*7 || track.R.H != switchTrackH(lh) || track.R.W != 2*base {
				t.Errorf("%s", what("the track is %d by %d for a base of %d: want 80%% as tall, and twice the base long", track.R.W, track.R.H, base))
			}
			if kd <= base || kd <= track.R.H {
				t.Errorf("%s", what("the knob %d is not bigger than the base %d and the track %d", kd, base, track.R.H))
			}
			wantKnob := th.CheckboxEmptyColour
			if on {
				wantKnob = th.SelText
			}
			edge := track.R.X
			if on {
				edge = track.R.X + track.R.W
			}
			// everything after the track and its border is the knob, drawn over them
			minX, maxX, top, bottom := 1<<30, -1<<30, 1<<30, -1<<30
			var knobFills []ctlRect
			for _, f := range rec.fills[1:] {
				if f.Seq < track.Seq || f.Seq < border.Seq || f.C != wantKnob {
					t.Errorf("%s", what("a knob piece %+v is drawn before the track or its border, or in the wrong colour", f))
				}
				knobFills = append(knobFills, f)
				minX, maxX = min(minX, f.R.X), max(maxX, f.R.X+f.R.W)
				top, bottom = min(top, f.R.Y), max(bottom, f.R.Y+f.R.H)
			}
			box := Rect{X: minX, Y: top, W: maxX - minX, H: bottom - top}
			g := ink(knobFills, box, nil)
			if len(knobFills) < 3 || !roundDisc(g) || box.W != kd || box.H != kd {
				t.Errorf("%s", what("the knob is not a round disc %d across: %d by %d:\n%s", kd, box.W, box.H, show(g)))
			}
			// centred on the end of the track: half over it, half outside, and standing out above and below
			if d := minX + maxX - 2*edge; d < -1 || d > 1 {
				t.Errorf("%s", what("the knob's centre is %d.%d and the track's end is %d", (minX+maxX)/2, (minX+maxX)%2*5, edge))
			}
			if minX >= track.R.X && !on || maxX <= track.R.X+track.R.W && on {
				t.Errorf("%s", what("the knob does not stand out of the end of the track"))
			}
			if top >= track.R.Y || bottom <= track.R.Y+track.R.H {
				t.Errorf("%s", what("the knob does not stand out above and below the track"))
			}
			// and all of it is in the control's bounds, so that a click on the knob counts
			b := s.Bounds(area, lh, func(s string) int { return len(s) * lh / 2 })
			if !inside(box, b) || !inside(track.R, b) {
				t.Errorf("%s", what("the knob %+v or the track %+v is outside the bounds %+v", box, track.R, b))
			}
			// the label comes after the whole mark, clear of the half knob that is outside the track
			if len(rec.texts) != 1 || rec.texts[0].X != b.X+switchMarkW(lh)+controlGap(lh) || rec.texts[0].X < max(maxX, track.R.X+track.R.W)+controlGap(lh) {
				t.Errorf("%s", what("the label is at %+v: want it after the mark and its gap, clear of the knob", rec.texts))
			}
			centre[i] = (minX + maxX) / 2
		}
		if centre[1] <= centre[0] {
			t.Errorf("line %d: the knob is at %d off and %d on: it should move right", lh, centre[0], centre[1])
		}
	}
}

func TestSwitchTogglesOnAClick(t *testing.T) {
	area := Rect{X: 50, Y: 50, W: 300, H: 24}
	s := &Switch{Label: "Pixel lock"}
	b := s.Bounds(area, ctlLH, ctlTextW)
	if v, ch := s.Update(ctlPress(b.X+3, b.Y+3), area, ctlLH, ctlTextW); !v || !ch {
		t.Errorf("a click gave %v %v", v, ch)
	}
	if _, ch := s.Update(ctlPress(b.X+b.W+4, b.Y+3), area, ctlLH, ctlTextW); ch {
		t.Error("a click outside changed it")
	}
	s.Disabled = true
	if _, ch := s.Update(ctlPress(b.X+3, b.Y+3), area, ctlLH, ctlTextW); ch {
		t.Error("a disabled switch changed")
	}
}

func TestRadioGroupStacksOrWrapsAndReportsItsHeight(t *testing.T) {
	g := &RadioGroup{Options: []string{"Bounce", "Stop", "Vanish"}}
	area := Rect{X: 10, Y: 10, W: 200, H: 400}
	rects, used := g.Layout(area, ctlLH, ctlTextW)
	h := controlHeight(ctlLH)
	if len(rects) != 3 || used != 3*h+2*2 {
		t.Fatalf("stacked: %d rects, height %d, want 3 and %d", len(rects), used, 3*h+4)
	}
	for i := 1; i < 3; i++ {
		if rects[i].Y < rects[i-1].Y+rects[i-1].H {
			t.Errorf("stacked options %d and %d overlap", i-1, i)
		}
	}
	// along the row: three that fit, then ones that do not, which wrap and say so
	g.Horizontal = true
	wide := Rect{X: 0, Y: 0, W: 400, H: 100}
	if rects, used = g.Layout(wide, ctlLH, ctlTextW); used != h || rects[0].Y != rects[2].Y {
		t.Errorf("three options in 400 pixels: height %d, rows at %d and %d: want one row of %d", used, rects[0].Y, rects[2].Y, h)
	}
	narrow := Rect{X: 0, Y: 0, W: 150, H: 100}
	rects, used = g.Layout(narrow, ctlLH, ctlTextW)
	if used <= h {
		t.Errorf("three options in 150 pixels took %d: they cannot fit one row of %d", used, h)
	}
	for i, r := range rects {
		if !inside(r, Rect{X: 0, Y: 0, W: 150, H: used}) {
			t.Errorf("option %d at %+v is outside the area and the height reported", i, r)
		}
		for j := 0; j < i; j++ {
			if overlap(r, rects[j]) {
				t.Errorf("options %d and %d overlap", j, i)
			}
		}
	}
	// choosing
	if v, ch := g.Update(ctlPress(rects[1].X+2, rects[1].Y+2), narrow, ctlLH, ctlTextW); v != 1 || !ch || g.Selected != 1 {
		t.Errorf("a click on the second option gave %d %v", v, ch)
	}
	if v, ch := g.Update(ctlPress(rects[1].X+2, rects[1].Y+2), narrow, ctlLH, ctlTextW); v != 1 || ch {
		t.Errorf("a click on the one chosen gave %d %v: no change", v, ch)
	}
	g.Disabled = true
	if _, ch := g.Update(ctlPress(rects[2].X+2, rects[2].Y+2), narrow, ctlLH, ctlTextW); ch || g.Selected != 1 {
		t.Error("a disabled group changed")
	}
}

func TestRadioButtonsAreRoundRingsWithADotInTheOneChosen(t *testing.T) {
	th := LoadTheme(ThemeSpectrum)
	size := controlBox(ctlLH)
	for sel := 0; sel < 3; sel++ {
		g := &RadioGroup{Options: []string{"A", "B", "C"}, Selected: sel, mouseX: -1, mouseY: -1} // the pointer is elsewhere
		area := Rect{X: 50, Y: 50, W: 300, H: 200}
		r := &ctlRenderer{}
		g.Draw(r, th, area, 1)
		rects, _ := g.Layout(area, ctlLH, ctlTextW)
		if len(r.strokes) != 0 || len(r.lines) != 0 || len(r.texts) != 3 {
			t.Fatalf("selected %d: %d strokes, %d lines, %d labels: a round mark is made of rows, and there is a label for each", sel, len(r.strokes), len(r.lines), len(r.texts))
		}
		for i, b := range rects {
			box := Rect{X: b.X, Y: b.Y + (b.H-size)/2, W: size, H: size}
			g := ink(r.fills, box, nil)
			mid := size / 2
			switch {
			case !symmetric(g) || g[0][0] || g[0][size-1] || g[size-1][0] || g[size-1][size-1]:
				t.Errorf("selected %d, option %d: not round (corners filled, or not the same on both sides):\n%s", sel, i, show(g))
			case !g[mid][0] || !g[mid-1][0] || !g[0][mid] || !g[size-1][mid]:
				t.Errorf("selected %d, option %d: the ring does not reach the box's edges:\n%s", sel, i, show(g))
			case g[mid][mid] != (i == sel):
				t.Errorf("selected %d, option %d: the middle is %v, want %v: a dot only in the one chosen:\n%s", sel, i, g[mid][mid], i == sel, show(g))
			}
			if i != sel && g[mid][mid] {
				t.Errorf("selected %d, option %d: an option not chosen has a dot", sel, i)
			}
			// the ring is hollow between its edge and its dot
			if gap := 2 + max(1, size/8); !g[mid][0] || (i != sel && g[mid][gap]) {
				t.Errorf("selected %d, option %d: the ring is not hollow inside:\n%s", sel, i, show(g))
			}
		}
	}
}

func TestMarksAreFourFifthsOfALine(t *testing.T) {
	for _, lh := range []int{10, 12, 16, 20, 24, 32, 48} {
		want := max(8, lh*8/10)
		if got := controlBox(lh); got != want || MarkSize(lh) != want {
			t.Errorf("line %d: the mark is %d, want %d", lh, controlBox(lh), want)
		}
		if lh >= 12 && (controlBox(lh) >= lh || controlBox(lh)*100 < lh*75) {
			t.Errorf("line %d: the mark %d is not about four fifths of it", lh, controlBox(lh))
		}
	}
	th := LoadTheme(ThemeSpectrum)
	r := &ctlRenderer{}
	(&Checkbox{Label: "x"}).Draw(r, th, Rect{X: 5, Y: 5, W: 100, H: 24}, 1)
	if len(r.strokes) != 1 || r.strokes[0].R.W != controlBox(ctlLH) || r.strokes[0].R.H != controlBox(ctlLH) {
		t.Errorf("the checkbox's box is %+v, want a square %d across", r.strokes, controlBox(ctlLH))
	}
}

func TestDiscsAreRoundSymmetricAndInsideTheirBox(t *testing.T) {
	for size := 8; size <= 40; size++ {
		box := Rect{X: 3, Y: 5, W: size, H: size}
		r := &ctlRenderer{}
		fillDisc(r, box, float64(size)/2, Colour{1, 2, 3, 255})
		for _, f := range r.fills {
			if !inside(f.R, box) || f.R.H != 1 {
				t.Fatalf("size %d: a row %+v is outside the box or not one row", size, f.R)
			}
		}
		g := ink(r.fills, box, nil)
		if !roundDisc(g) {
			t.Errorf("size %d is not a round disc:\n%s", size, show(g))
		}
		if size >= 10 {
			want := 3.14159 * float64(size*size) / 4
			if got := float64(count(g)); got < want*0.85 || got > want*1.15 {
				t.Errorf("size %d: %d pixels, want about %.0f", size, count(g), want)
			}
		}
	}
}

func TestRingsAreHollowRoundAndSymmetric(t *testing.T) {
	for size := 10; size <= 40; size++ {
		for _, thick := range []int{1, 2, 3} {
			box := Rect{X: 3, Y: 5, W: size, H: size}
			r := &ctlRenderer{}
			strokeCircle(r, box, thick, Colour{1, 2, 3, 255})
			g := ink(r.fills, box, nil)
			mid := size / 2
			if !symmetric(g) || g[0][0] || g[size-1][size-1] || !g[mid][0] || !g[0][mid] {
				t.Errorf("size %d thick %d: not a round ring touching the box:\n%s", size, thick, show(g))
			}
			if g[mid][mid] || g[mid][thick] && size > 2*thick+2 || g[mid-1][thick] && size > 2*thick+2 {
				t.Errorf("size %d thick %d: not hollow inside:\n%s", size, thick, show(g))
			}
			if !g[mid][thick-1] {
				t.Errorf("size %d thick %d: the ring is thinner than %d at its side:\n%s", size, thick, thick, show(g))
			}
		}
	}
}

// ink rasterises the fills that land in a box, those of a colour the filter accepts if it
// is given, into a grid of what is covered.
func ink(fills []ctlRect, box Rect, only func(Colour) bool) [][]bool {
	g := make([][]bool, box.H)
	for y := range g {
		g[y] = make([]bool, box.W)
	}
	for _, f := range fills {
		if only != nil && !only(f.C) {
			continue
		}
		for y := max(f.R.Y, box.Y); y < min(f.R.Y+f.R.H, box.Y+box.H); y++ {
			for x := max(f.R.X, box.X); x < min(f.R.X+f.R.W, box.X+box.W); x++ {
				g[y-box.Y][x-box.X] = true
			}
		}
	}
	return g
}

func count(g [][]bool) int {
	n := 0
	for _, row := range g {
		for _, c := range row {
			if c {
				n++
			}
		}
	}
	return n
}

func symmetric(g [][]bool) bool {
	h := len(g)
	for y := range g {
		w := len(g[y])
		for x := range g[y] {
			if g[y][x] != g[y][w-1-x] || g[y][x] != g[h-1-y][x] {
				return false
			}
		}
	}
	return true
}

// roundDisc says a grid is a filled disc: the same on both axes, the middle filled and
// the corners not, and reaching the middle of each edge.
func roundDisc(g [][]bool) bool {
	h, w := len(g), len(g[0])
	return h > 0 && symmetric(g) && !g[0][0] && !g[0][w-1] && !g[h-1][0] && !g[h-1][w-1] &&
		g[h/2][w/2] && g[h/2][0] && g[0][w/2] && g[h-1][w/2] && g[h/2][w-1]
}

func show(g [][]bool) string {
	out := ""
	for _, row := range g {
		for _, c := range row {
			if c {
				out += "#"
			} else {
				out += "."
			}
		}
		out += "\n"
	}
	return out
}

func TestTabsShareTheWidthExactlyAndNeverWrap(t *testing.T) {
	for _, n := range []int{1, 2, 3, 5} {
		for _, w := range []int{50, 100, 380, 401} {
			tb := &Tabs{Labels: make([]string, n)}
			area := Rect{X: 7, Y: 9, W: w, H: 100}
			rects := tb.Layout(area, ctlLH)
			sum := 0
			for i, r := range rects {
				sum += r.W
				if r.Y != area.Y || r.H != tb.Height(ctlLH) {
					t.Errorf("n %d w %d: tab %d is %+v: all must be one row", n, w, i, r)
				}
				if i > 0 && r.X != rects[i-1].X+rects[i-1].W {
					t.Errorf("n %d w %d: tab %d does not start where the last ends", n, w, i)
				}
				if i > 0 && rects[i-1].W-r.W > 1 {
					t.Errorf("n %d w %d: widths differ by more than a pixel", n, w)
				}
			}
			if sum != w || rects[0].X != area.X {
				t.Errorf("n %d w %d: the tabs add up to %d starting at %d", n, w, sum, rects[0].X)
			}
		}
	}
	if len((&Tabs{}).Layout(Rect{W: 100}, ctlLH)) != 0 {
		t.Error("no labels gave tabs")
	}
}

func TestTabsChooseOnAClickAndClipALongLabel(t *testing.T) {
	th := LoadTheme(ThemeSpectrum)
	tb := &Tabs{Labels: []string{"Look", "Does", "a label far too long for its tab"}}
	area := Rect{X: 0, Y: 0, W: 300, H: 40}
	rects := tb.Layout(area, ctlLH)
	if i, ch := tb.Update(ctlPress(rects[1].X+3, rects[1].Y+3), area, ctlLH); i != 1 || !ch || tb.Selected != 1 {
		t.Errorf("a click on the second tab gave %d %v", i, ch)
	}
	if i, ch := tb.Update(ctlPress(rects[1].X+3, rects[1].Y+3), area, ctlLH); i != 1 || ch {
		t.Errorf("a click on the chosen tab gave %d %v: no change", i, ch)
	}
	if _, ch := tb.Update(ctlPress(rects[0].X+3, rects[0].Y+tb.Height(ctlLH)+5), area, ctlLH); ch {
		t.Error("a click under the strip chose a tab")
	}
	if _, ch := tb.Update(Input{MouseX: rects[0].X + 3, MouseY: rects[0].Y + 3}, area, ctlLH); ch {
		t.Error("a frame with no press chose a tab")
	}
	tb.mouseX, tb.mouseY = -1, -1 // the pointer is elsewhere
	r := &ctlRenderer{}
	tb.Draw(r, th, area, 1)
	if len(r.clips) != 3 {
		t.Fatalf("%d clips, want one for each label", len(r.clips))
	}
	for i, c := range r.clips {
		if c != rects[i] {
			t.Errorf("label %d is clipped to %+v, want its tab %+v", i, c, rects[i])
		}
	}
	h := tb.Height(ctlLH)
	for i, b := range rects {
		c, ok := colourAt(r.fills, b.X+b.W/2, b.Y+h/2)
		if !ok || (c == th.SelFill) != (i == 1) {
			t.Errorf("tab %d: its middle is %+v: only the chosen tab, 1, is in the selection colour", i, c)
		}
	}
}

func TestSectionHeaderHasATitleAndARuleToTheRightEdge(t *testing.T) {
	th := LoadTheme(ThemeSpectrum)
	area := Rect{X: 10, Y: 20, W: 300, H: 40}
	r := &ctlRenderer{}
	DrawSection(r, th, area, 1, "Ladders")
	if len(r.texts) != 1 || r.texts[0].S != "Ladders" || r.texts[0].C != th.DirText {
		t.Fatalf("texts %+v: want the title in the heading colour", r.texts)
	}
	if len(r.lines) != 1 || r.lines[0].X2 != area.X+area.W || r.lines[0].X1 <= area.X+ctlTextW("Ladders") || r.lines[0].C != th.SeparatorColour {
		t.Errorf("rule %+v: want it from after the title to the right edge in the separator colour", r.lines)
	}
	if h := SectionHeight(ctlLH); r.lines[0].Y1 < area.Y || r.lines[0].Y1 > area.Y+h {
		t.Errorf("the rule is at %d, outside the header's %d high", r.lines[0].Y1, h)
	}
	r = &ctlRenderer{}
	DrawSection(r, th, Rect{X: 0, Y: 0, W: 40, H: 40}, 1, "a title wider than the area")
	if len(r.lines) != 0 {
		t.Error("a title wider than the area has a rule")
	}
}

// Whatever the labels, the sizes and the room, a control's rectangles stay inside the
// area it is given in width, its rows are the height it says, and what it reports is
// what a click does.
func TestControlsStayInsideTheirAreaWhateverTheLabelsAndSizes(t *testing.T) {
	rng := rand.New(rand.NewSource(5))
	word := func() string {
		b := make([]byte, 1+rng.Intn(30))
		for i := range b {
			b[i] = byte('a' + rng.Intn(26))
		}
		return string(b)
	}
	for i := 0; i < 400; i++ {
		lh := 8 + 2*rng.Intn(20)
		area := Rect{X: rng.Intn(50), Y: rng.Intn(50), W: 30 + rng.Intn(500), H: 600}
		tw := func(s string) int { return len(s) * lh / 2 }
		var opts []string
		for n := 1 + rng.Intn(5); n > 0; n-- {
			opts = append(opts, word())
		}
		g := &RadioGroup{Options: opts, Horizontal: rng.Intn(2) == 0}
		rects, used := g.Layout(area, lh, tw)
		for j, r := range rects {
			if r.X < area.X || r.X+r.W > area.X+area.W || r.Y < area.Y || r.Y+r.H > area.Y+used {
				t.Fatalf("case %d: radio option %d at %+v is outside %+v with height %d", i, j, r, area, used)
			}
			for k := 0; k < j; k++ {
				if overlap(r, rects[k]) {
					t.Fatalf("case %d: radio options %d and %d overlap", i, k, j)
				}
			}
		}
		c := &Checkbox{Label: word()}
		if b := c.Bounds(area, lh, tw); b.X != area.X || b.W > area.W || b.H != c.Height(lh) {
			t.Fatalf("case %d: checkbox bounds %+v in %+v", i, b, area)
		}
		tb := &Tabs{Labels: opts}
		sum := 0
		for _, r := range tb.Layout(area, lh) {
			sum += r.W
		}
		if sum != area.W {
			t.Fatalf("case %d: the tabs add up to %d in a width of %d", i, sum, area.W)
		}
	}
}

// lhRenderer draws through a recorder that has a line height of its own, to try a control
// at several sizes of text.
type lhRenderer struct {
	*ctlRenderer
	lh int
}

func (r *lhRenderer) LineHeight(scale int) int          { return r.lh * scale }
func (r *lhRenderer) TextWidth(s string, scale int) int { return len(s) * r.lh / 2 * scale }

// colourAt is the colour of the last fill that covers a pixel, which is what shows.
func colourAt(fills []ctlRect, x, y int) (Colour, bool) {
	for i := len(fills) - 1; i >= 0; i-- {
		if r := fills[i].R; x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H {
			return fills[i].C, true
		}
	}
	return Colour{}, false
}

func TestTheCrossRunsCornerToCornerInTheMediumAndLargeBoxesAndKeepsItsInsetInTheSmall(t *testing.T) {
	th := LoadTheme(ThemeSpectrum)
	for _, lh := range []int{8, 16, 24} {
		rec := &ctlRenderer{}
		c := &Checkbox{Label: "x", Checked: true, mouseX: -1, mouseY: -1}
		area := Rect{X: 50, Y: 50, W: 300, H: 3 * lh}
		c.Draw(&lhRenderer{ctlRenderer: rec, lh: lh}, th, area, 1)
		b := c.Bounds(area, lh, func(s string) int { return len(s) * lh / 2 })
		size := controlBox(lh)
		box := Rect{X: b.X, Y: b.Y + (b.H-size)/2, W: size, H: size}
		if len(rec.lines) != 2 {
			t.Fatalf("line %d: %d lines in the cross, want 2", lh, len(rec.lines))
		}
		clipped := false
		for _, cl := range rec.clips {
			clipped = clipped || cl == box
		}
		if size >= 12 {
			want := [2][4]int{{box.X, box.Y, box.X + size, box.Y + size}, {box.X + size, box.Y, box.X, box.Y + size}}
			for i, l := range rec.lines {
				if got := [4]int{l.X1, l.Y1, l.X2, l.Y2}; got != want[i] || l.Thick != 2 {
					t.Errorf("line %d: diagonal %d is %v thick %d, want %v thick 2: corner to corner", lh, i, got, l.Thick, want[i])
				}
				if l.X1+l.X2 != 2*box.X+size || l.Y1+l.Y2 != 2*box.Y+size {
					t.Errorf("line %d: diagonal %d is not centred in the box", lh, i)
				}
			}
			if !clipped {
				t.Errorf("line %d: the cross is not clipped to its box, so a thick line's ends could poke out", lh)
			}
		} else {
			in := 2
			want := [2][4]int{{box.X + in, box.Y + in, box.X + size - 1 - in, box.Y + size - 1 - in}, {box.X + size - 1 - in, box.Y + in, box.X + in, box.Y + size - 1 - in}}
			for i, l := range rec.lines {
				if got := [4]int{l.X1, l.Y1, l.X2, l.Y2}; got != want[i] || l.Thick != 1 {
					t.Errorf("line %d: the small box's diagonal %d is %v thick %d, want %v thick 1 as it was", lh, i, got, l.Thick, want[i])
				}
			}
			if clipped {
				t.Errorf("line %d: the small box's cross is clipped, and was not", lh)
			}
		}
	}
}

func TestTheTabStripHasSlightlyRoundedEndsSquareJoinsAndAVisibleRule(t *testing.T) {
	th := LoadTheme(ThemeSpectrum)
	for _, lh := range []int{8, 16, 24, 40, 48} { // the larger ones are where the corner steps more than a pixel a row
		tb := &Tabs{Labels: []string{"Look", "Does", "Fire"}, Selected: 1, mouseX: -1, mouseY: -1}
		area := Rect{X: 40, Y: 30, W: 300, H: 60}
		rec := &ctlRenderer{}
		tb.Draw(&lhRenderer{ctlRenderer: rec, lh: lh}, th, area, 1)
		h, W := tb.Height(lh), area.W
		strip := Rect{X: area.X, Y: area.Y, W: W, H: h}
		g := ink(rec.fills, strip, nil)
		rad := tabCorner(lh, h)
		if rad < 2 || rad > 6 {
			t.Errorf("line %d: the corner radius is %d: slight means 2 to 6 pixels", lh, rad)
		}
		if g[0][0] || g[0][W-1] || g[h-1][0] || g[h-1][W-1] {
			t.Errorf("line %d: a corner of the strip is not rounded off", lh)
		}
		if !g[h/2][0] || !g[h/2][W-1] || !g[0][W/2] || !g[h-1][W/2] || !g[0][rad] || !g[0][W-1-rad] {
			t.Errorf("line %d: the strip does not reach its own edges, or is cut by more than the radius %d", lh, rad)
		}
		for y := range g {
			for x := range g[y] {
				if g[y][x] != g[y][W-1-x] {
					t.Fatalf("line %d: the strip is not the same at both ends, at (%d,%d)", lh, x, y)
				}
			}
		}
		rects := tb.Layout(area, lh)
		for i := 1; i < len(rects); i++ { // where one tab meets the next the corners stay square
			x := rects[i].X - strip.X
			if !g[0][x] || !g[0][x-1] || !g[h-1][x] || !g[h-1][x-1] {
				t.Errorf("line %d: the join between tabs %d and %d is rounded", lh, i-1, i)
			}
		}
		// the border follows the curve with no gap, at both ends
		ins := roundInsets(h, rad)
		for j := 1; j < h-2; j++ { // not the last two rows: the rule is over them
			from, to := ins[j], ins[j-1] // the top half: the border runs from this row's edge out to the last row's
			if j >= h/2 {
				to = ins[j+1] // the bottom half, mirrored
			}
			for x := from; x < max(to, from+1); x++ {
				if c, ok := colourAt(rec.fills, strip.X+x, strip.Y+j); !ok || c != th.Border {
					t.Errorf("line %d: a gap in the border on row %d at %d", lh, j, x)
				}
				if c, ok := colourAt(rec.fills, strip.X+W-1-x, strip.Y+j); !ok || c != th.Border {
					t.Errorf("line %d: a gap in the border on row %d at %d from the right", lh, j, x)
				}
			}
		}
		// the rule under the strip shows, under the tabs that are not chosen too
		for _, tab := range []int{0, 1, 2} {
			for _, j := range []int{h - 2, h - 1} {
				if c, ok := colourAt(rec.fills, rects[tab].X+rects[tab].W/2, strip.Y+j); !ok || c != th.SelFill {
					t.Errorf("line %d: the rule is not showing under tab %d on row %d (%+v)", lh, tab, j, c)
				}
			}
		}
	}
}

func TestAHoveredTabNeverLooksChosen(t *testing.T) {
	for _, th := range []Theme{LoadTheme(ThemeSpectrum), LoadTheme(ThemeDark), LoadTheme(ThemeLight)} {
		tb := &Tabs{Labels: []string{"Look", "Does", "Fire"}, Selected: 1}
		area := Rect{X: 40, Y: 30, W: 300, H: 60}
		rects := tb.Layout(area, ctlLH)
		tb.mouseX, tb.mouseY = rects[0].X+5, rects[0].Y+5 // over tab 0, which is not the chosen one
		r := &ctlRenderer{}
		tb.Draw(r, th, area, 1)
		h := tb.Height(ctlLH)
		hot, _ := colourAt(r.fills, rects[0].X+rects[0].W/2, rects[0].Y+h/2)
		chosen, _ := colourAt(r.fills, rects[1].X+rects[1].W/2, rects[1].Y+h/2)
		if hot == chosen {
			t.Errorf("hot %+v, selection %+v: the hovered tab is drawn as the chosen one", th.ButtonHot, th.SelFill)
		}
		var label Colour
		for _, tx := range r.texts {
			if tx.S == "Look" {
				label = tx.C
			}
		}
		if label != th.ButtonText {
			t.Errorf("the hovered tab's label is %+v, want the button text colour %+v", label, th.ButtonText)
		}
	}
}

func TestTabsThatAreNotChosenAreSetApartFromTheRuleByAThinLineAndTheChosenOneRunsIntoIt(t *testing.T) {
	th := LoadTheme(ThemeSpectrum)
	for _, lh := range []int{8, 16, 24, 48} {
		for sel := 0; sel < 3; sel++ {
			tb := &Tabs{Labels: []string{"Look", "Does", "Fire"}, Selected: sel, mouseX: -1, mouseY: -1}
			area := Rect{X: 40, Y: 30, W: 300, H: 80}
			rec := &ctlRenderer{}
			tb.Draw(&lhRenderer{ctlRenderer: rec, lh: lh}, th, area, 1)
			h := tb.Height(lh)
			rects := tb.Layout(area, lh)
			line, above := area.Y+h-3, area.Y+h-4
			for i, b := range rects {
				mid := b.X + b.W/2
				cLine, _ := colourAt(rec.fills, mid, line)
				cAbove, _ := colourAt(rec.fills, mid, above)
				cRule, _ := colourAt(rec.fills, mid, area.Y+h-2)
				if cRule != th.SelFill {
					t.Errorf("line %d chosen %d tab %d: the rule is %+v under it", lh, sel, i, cRule)
				}
				if i == sel {
					if cLine != th.SelFill || cAbove != th.SelFill {
						t.Errorf("line %d: the chosen tab %d has %+v on the row above the rule: it should run into the rule", lh, i, cLine)
					}
					continue
				}
				if cLine != th.Border {
					t.Errorf("line %d chosen %d: tab %d has %+v on the row above the rule, want the border colour: a thin line", lh, sel, i, cLine)
				}
				if cAbove == th.Border || cAbove == th.SelFill {
					t.Errorf("line %d chosen %d: tab %d has %+v above its line: the line should be one pixel", lh, sel, i, cAbove)
				}
				// all the way along the tab, not only in the middle
				for x := b.X + 3; x < b.X+b.W-3; x++ {
					if c, _ := colourAt(rec.fills, x, line); c != th.Border {
						t.Errorf("line %d chosen %d: tab %d has a gap in its line at %d (%+v)", lh, sel, i, x, c)
						break
					}
				}
			}
		}
	}
}
