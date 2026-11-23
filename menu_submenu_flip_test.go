package zenui

import "testing"

// openedSubmenu draws a one-item menu anchored at anchorX on a screen screenW wide, hovers
// the item to open its submenu, draws again to lay the submenu out, and returns the parent
// menu.
func openedSubmenu(t *testing.T, anchorX, screenW int) *Menu {
	t.Helper()
	m := NewMenu(MenuConfig{
		Anchor: Rect{X: anchorX, Y: 0, W: 0, H: 20},
		Items: []Item{{Label: "Developer", SubItems: []Item{
			{Label: "Debugger"}, {Label: "A rather longer second row"},
		}}},
	})
	m.Draw(newDrawRecorder(), screenW, 600, DefaultTheme())
	rec := m.ItemRect(0)
	m.Update(Input{MouseX: rec.X + 1, MouseY: rec.Y + 1})
	m.Draw(newDrawRecorder(), screenW, 600, DefaultTheme())
	if m.subMenu == nil {
		t.Fatal("submenu did not open")
	}
	return m
}

func TestSubmenuWithRoomOpensToTheRight(t *testing.T) {
	m := openedSubmenu(t, 100, 1000)
	if got, want := m.subMenu.bounds.X, m.bounds.X+m.bounds.W-4; got != want {
		t.Errorf("submenu X = %d, want %d (the parent's right edge less the overlap)", got, want)
	}
}

func TestSubmenuWithNoRoomOnTheRightGrowsLeftWithoutCoveringItsParent(t *testing.T) {
	// A parent whose right edge is 10px short of the screen's: no submenu fits beside it
	// on that side.
	probe := openedSubmenu(t, 600, 1000)
	screenW := probe.bounds.X + probe.bounds.W + 10

	m := openedSubmenu(t, 600, screenW)
	sub, parent := m.subMenu.bounds, m.bounds
	if sub.X+sub.W != parent.X {
		t.Errorf("submenu spans %d..%d, want its right edge on the parent's left edge %d",
			sub.X, sub.X+sub.W, parent.X)
	}
	if sub.X+sub.W > parent.X {
		t.Errorf("submenu covers the parent: %+v over %+v", sub, parent)
	}
	if sub.X < 0 {
		t.Errorf("submenu starts off screen at %d", sub.X)
	}
	if sub.Y != m.ItemRect(0).Y {
		t.Errorf("submenu top = %d, want the parent row's top %d", sub.Y, m.ItemRect(0).Y)
	}
}

func TestSubmenuFlipHoldsAtEveryWidth(t *testing.T) {
	// Whatever the width of the screen, a submenu is on screen; and whenever there is
	// room for it on either side of its parent it covers none of the parent.
	flipped := 0
	for screenW := 460; screenW <= 900; screenW += 7 {
		m := openedSubmenu(t, screenW-150, screenW)
		sub, parent := m.subMenu.bounds, m.bounds
		if sub.X < 0 || sub.X+sub.W > screenW {
			t.Errorf("screen %d: submenu %+v is not on screen", screenW, sub)
		}
		roomLeft := parent.X >= sub.W
		roomRight := parent.X+parent.W-4+sub.W <= screenW
		if (roomLeft || roomRight) && sub.X < parent.X+parent.W-4 && sub.X+sub.W > parent.X {
			t.Errorf("screen %d: submenu %+v covers its parent %+v", screenW, sub, parent)
		}
		if !roomRight && roomLeft {
			flipped++
			if sub.X+sub.W != parent.X {
				t.Errorf("screen %d: flipped submenu %+v is not flush with its parent %+v", screenW, sub, parent)
			}
		}
	}
	if flipped == 0 {
		t.Error("the sweep never flipped a submenu; its geometry no longer tests the flip")
	}
}

func TestSubmenuWithRoomNowhereStaysOnScreen(t *testing.T) {
	// A screen with room for the submenu on neither side of its parent: the submenu is
	// clamped inside the screen, as before, rather than placed off it.
	m := openedSubmenu(t, 0, 500)
	if sub := m.subMenu.bounds; sub.X < 0 || sub.X+sub.W > 500 {
		t.Errorf("submenu %+v is not on a 500px screen", sub)
	}
}
