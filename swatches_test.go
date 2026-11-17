package zenui

import "testing"

func swColours(n int) []Colour {
	out := make([]Colour, n)
	for i := range out {
		out[i] = Colour{R: uint8(i * 30), G: uint8(255 - i*30), B: 7, A: 255}
	}
	return out
}

// Swatches are 12 square with 2 between them, so four fit across 56 pixels.
var swArea = Rect{X: 10, Y: 20, W: 56, H: 100}

func TestSwatchesAreSquareAndWrapToFitTheArea(t *testing.T) {
	s := &Swatches{Colours: swColours(8)}
	rects, used := s.Layout(swArea, 8)
	if len(rects) != 8 {
		t.Fatalf("%d rects, want 8", len(rects))
	}
	if rects[0] != (Rect{X: 10, Y: 20, W: 12, H: 12}) || rects[3] != (Rect{X: 52, Y: 20, W: 12, H: 12}) {
		t.Errorf("the first row is %+v ... %+v", rects[0], rects[3])
	}
	if rects[4] != (Rect{X: 10, Y: 34, W: 12, H: 12}) || used != 26 {
		t.Errorf("the second row starts at %+v, height used %d, want (10,34) and 26", rects[4], used)
	}
}

func TestClickingASwatchSelectsItAndOnlyAChangeCounts(t *testing.T) {
	s := &Swatches{Colours: swColours(8)}
	rects, _ := s.Layout(swArea, 8)
	press := func(x, y int) (int, bool) {
		return s.Update(Input{MouseX: x, MouseY: y, MousePressed: true}, swArea, 8)
	}
	x, y := mid(rects[5])
	if i, ok := press(x, y); !ok || i != 5 || s.Selected != 5 {
		t.Errorf("a click on swatch 5 gave %d, %v, selected %d", i, ok, s.Selected)
	}
	if i, ok := press(x, y); ok || i != 5 {
		t.Errorf("a second click on the same swatch gave %d, %v, want no change", i, ok)
	}
	if _, ok := press(500, 500); ok || s.Selected != 5 {
		t.Error("a click outside changed the selection")
	}
	if _, ok := s.Update(Input{MouseX: x, MouseY: y}, swArea, 8); ok {
		t.Error("the pointer resting on a swatch, with no press, changed something")
	}
	s.Disabled = true
	x, y = mid(rects[2])
	if _, ok := press(x, y); ok || s.Selected != 5 {
		t.Error("a disabled row of swatches changed")
	}
}

func TestTheSelectedSwatchHasARingTheOthersDoNot(t *testing.T) {
	th := DarkTheme()
	s := &Swatches{Colours: swColours(8), Selected: 2}
	cv := &listCanvas{}
	s.Draw(cv, th, swArea, 1)
	rects, _ := s.Layout(swArea, 8)
	for i, rc := range rects {
		found := false
		for _, f := range cv.fills {
			if f.R == rc && f.C == s.Colours[i] {
				found = true
			}
		}
		if !found {
			t.Errorf("swatch %d was not filled with its colour", i)
		}
	}
	rings := 0
	for _, st := range cv.strokes {
		if st.C == th.Text && st.T == 2 {
			rings++
			want := Rect{X: rects[2].X - 3, Y: rects[2].Y - 3, W: 18, H: 18}
			if st.R != want {
				t.Errorf("the ring is %+v, want %+v round swatch 2", st.R, want)
			}
		}
	}
	if rings != 1 {
		t.Errorf("%d rings drawn, want one", rings)
	}
}
