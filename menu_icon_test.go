package zenui

import "testing"

// An item's Icon is drawn in a gutter of its own between the checkmark's and the label.

func TestIconGetsASquareCellInsideItsRowAndTheMenuGrowsByIt(t *testing.T) {
	plain := NewMenu(MenuConfig{Items: []Item{{Label: "Pause"}, {Label: "Reset"}}})
	plain.Draw(newDrawRecorder(), 800, 600, DefaultTheme())

	var got []Rect
	m := NewMenu(MenuConfig{Items: []Item{
		{Label: "Pause", Icon: func(r Renderer, cell Rect, col Colour) { got = append(got, cell) }},
		{Label: "Reset"},
	}})
	m.Draw(newDrawRecorder(), 800, 600, DefaultTheme())

	lh := newDrawRecorder().LineHeight(m.scale)
	if len(got) != 1 {
		t.Fatalf("the icon was drawn %d times, want once (only the item that has one)", len(got))
	}
	row := m.ItemRect(0)
	if got[0].W != lh || got[0].H != lh {
		t.Errorf("cell is %dx%d, want a square of one line (%d)", got[0].W, got[0].H, lh)
	}
	if got[0].Y < row.Y || got[0].Y+got[0].H > row.Y+row.H {
		t.Errorf("cell %v is not inside its row %v", got[0], row)
	}
	if grew := m.bounds.W - plain.bounds.W; grew != lh {
		t.Errorf("the menu is %d wider with an icon, want %d (one gutter)", grew, lh)
	}
}

func TestIconFollowsTheRowsColourAndSkipsSeparatorsAndTitles(t *testing.T) {
	theme := DefaultTheme()
	var cols []Colour
	icon := func(r Renderer, cell Rect, col Colour) { cols = append(cols, col) }
	m := NewMenu(MenuConfig{Items: []Item{
		{Label: "On", Icon: icon},
		{Label: "Off", Icon: icon, Disabled: true},
		{Separator: true, Icon: icon},
		{Title: true, Label: "Group", Icon: icon},
	}})
	m.Draw(newDrawRecorder(), 800, 600, theme)
	if len(cols) != 2 || cols[0] != theme.Text || cols[1] != theme.Disabled {
		t.Errorf("icon colours %v, want the text colour then the disabled one, and none for a separator or a title", cols)
	}
}
