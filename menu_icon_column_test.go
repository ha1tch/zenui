package zenui

import "testing"

// A menu that has checkboxes and icons in different rows puts both in one column, each centred in
// it, and the labels start after it: the two kinds of indicator are at one indent, not two.
func TestChecksAndIconsShareOneColumnAndOneLabelIndent(t *testing.T) {
	var cells []Rect
	icon := func(r Renderer, cell Rect, col Colour) { cells = append(cells, cell) }
	on := false
	r := newDrawRecorder()
	m := NewMenu(MenuConfig{Items: []Item{
		{Label: "Auto", Checked: &on, Toggle: true},
		{Label: "Play", Icon: icon},
		{Label: "Stop", Icon: icon},
	}})
	m.Draw(r, 800, 600, DefaultTheme())

	lh := newDrawRecorder().LineHeight(m.scale)
	if m.iconGutter != 0 || m.checkGutter != lh+lh/2 {
		t.Fatalf("gutters are check %d, icon %d; want one column of %d and no second one", m.checkGutter, m.iconGutter, lh+lh/2)
	}
	if len(cells) != 2 {
		t.Fatalf("%d icons drawn, want 2", len(cells))
	}
	var box *Rect
	for _, c := range *r.calls {
		if c.kind == "stroke" && c.rect.W < lh {
			rc := c.rect
			box = &rc
			break
		}
	}
	if box == nil {
		t.Fatal("no checkbox was drawn")
	}
	centre := func(rc Rect) int { return rc.X + rc.W/2 }
	for i, cell := range cells {
		if d := centre(cell) - centre(*box); d < -1 || d > 1 {
			t.Errorf("icon %d is centred at x=%d and the checkbox at x=%d: they are at different indents", i, centre(cell), centre(*box))
		}
	}
	one := NewMenu(MenuConfig{Items: []Item{{Label: "Auto", Checked: &on, Toggle: true}, {Label: "Play"}, {Label: "Stop"}}})
	one.Draw(newDrawRecorder(), 800, 600, DefaultTheme())
	if one.checkGutter != lh+lh/2 {
		t.Errorf("a menu with checks only has a column of %d, want a line and a half (%d)", one.checkGutter, lh+lh/2)
	}
}

// An item that has both a check and an icon cannot share: the two stay side by side.
func TestAnItemWithACheckAndAnIconKeepsTwoColumns(t *testing.T) {
	var cell Rect
	on := true
	m := NewMenu(MenuConfig{Items: []Item{{Label: "Both", Checked: &on, Icon: func(r Renderer, c Rect, col Colour) { cell = c }}}})
	m.Draw(newDrawRecorder(), 800, 600, DefaultTheme())
	lh := newDrawRecorder().LineHeight(m.scale)
	if m.checkGutter != lh || m.iconGutter != lh {
		t.Fatalf("gutters are check %d, icon %d, want a line each", m.checkGutter, m.iconGutter)
	}
	if want := m.bounds.X + m.padX + lh; cell.X != want {
		t.Errorf("the icon starts at x=%d, want %d (after the check's column)", cell.X, want)
	}
}

// A menu checkbox is a fifth larger than half a line (0.6 of it), so it reads beside the label's
// glyphs; its tick, for the mutually exclusive rows, has the same size.
func TestMenuCheckboxIsSixTenthsOfALine(t *testing.T) {
	on := true
	r := newDrawRecorder()
	m := NewMenu(MenuConfig{Items: []Item{{Label: "Auto", Checked: &on, Toggle: true}}})
	m.Draw(r, 800, 600, DefaultTheme())
	lh := newDrawRecorder().LineHeight(m.scale)
	want := (lh*3 + 2) / 5
	if want <= lh/2 {
		t.Fatalf("test font is too small to tell: line %d, box %d", lh, want)
	}
	for _, c := range *r.calls {
		if c.kind == "stroke" && c.rect.W < lh {
			if c.rect.W != want || c.rect.H != want {
				t.Errorf("the checkbox is %dx%d, want %dx%d (line height %d)", c.rect.W, c.rect.H, want, want, lh)
			}
			return
		}
	}
	t.Fatal("no checkbox was drawn")
}
