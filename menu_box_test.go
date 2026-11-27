package zenui

import "testing"

// Item.Box draws the checkbox glyph but, unlike Item.Toggle, choosing the item closes the menu.
func TestBoxItemDrawsACheckboxAndStillClosesTheMenu(t *testing.T) {
	on, off := true, false
	r := newDrawRecorder()
	m := NewMenu(MenuConfig{Items: []Item{
		{Label: "On", Checked: &on, Box: true},
		{Label: "Off", Checked: &off, Box: true},
	}})
	m.Draw(r, 800, 600, DefaultTheme())
	lh := newDrawRecorder().LineHeight(m.scale)
	boxes := 0
	for _, c := range *r.calls {
		if c.kind == "stroke" && c.rect.W < lh {
			boxes++
		}
	}
	if boxes != 2 {
		t.Errorf("%d boxes drawn for a checked and an unchecked Box item, want 2 (the box is always there)", boxes)
	}
	if got := m.accept(0); got != Accepted {
		t.Errorf("choosing a Box item gave %v, want Accepted: the menu closes", got)
	}
	if !on {
		t.Error("choosing a Box item flipped the bool: the host does that")
	}
}

// The cross of a checked box is the mirror image of itself, left to right and top to bottom, and
// its two diagonals run from corner to corner of the box: it is centred in it, for an even box
// and an odd one.
func TestCheckboxCrossIsCentredInItsBox(t *testing.T) {
	for _, scale := range []int{1, 2, 3} {
		on := true
		r := newDrawRecorder()
		m := NewMenu(MenuConfig{Items: []Item{{Label: "X", Checked: &on, Toggle: true}}, Scale: scale})
		m.Draw(r, 800, 600, DefaultTheme())
		lh := newDrawRecorder().LineHeight(m.scale)
		var box *Rect
		for _, c := range *r.calls {
			if c.kind == "stroke" && c.rect.W < lh {
				rc := c.rect
				box = &rc
				break
			}
		}
		if box == nil {
			t.Fatalf("scale %d: no box", scale)
		}
		have := map[[2]int]bool{}
		for _, c := range *r.calls {
			if c.kind == "fill" && c.rect.W == 1 && c.rect.H == 1 &&
				c.rect.X >= box.X && c.rect.X < box.X+box.W && c.rect.Y >= box.Y && c.rect.Y < box.Y+box.H {
				have[[2]int{c.rect.X - box.X, c.rect.Y - box.Y}] = true
			}
		}
		n := box.W
		for i := 0; i < n; i++ {
			for _, p := range [][2]int{{i, i}, {n - 1 - i, i}} {
				if !have[p] {
					t.Errorf("scale %d, box %d: the diagonal is missing the pixel %v", scale, n, p)
				}
			}
		}
		for p := range have {
			if !have[[2]int{n - 1 - p[0], p[1]}] || !have[[2]int{p[0], n - 1 - p[1]}] {
				t.Errorf("scale %d, box %d: the pixel %v has no mirror image: the cross is off centre", scale, n, p)
			}
		}
		if want := 2*n - n%2; len(have) != want {
			t.Errorf("scale %d, box %d: %d cross pixels, want %d", scale, n, len(have), want)
		}
	}
}
