package zenui

import "testing"

func barOf(labels ...string) *ButtonBar {
	b := NewButtonBar()
	for _, l := range labels {
		b.Buttons = append(b.Buttons, BarButton{ID: l, Label: l})
	}
	return b
}

func textW8(s string) int { return len(s) * 8 }

func TestButtonBarWrapsToFitItsArea(t *testing.T) {
	b := barOf("Undo", "Redo", "Pen", "Flip H")
	rects, used := b.Layout(Rect{X: 0, Y: 0, W: 100, H: 200}, 8, textW8)
	want := []Rect{
		{X: 0, Y: 0, W: 40, H: 12},
		{X: 42, Y: 0, W: 40, H: 12},
		{X: 0, Y: 14, W: 32, H: 12}, // 84+32 would pass the edge, so it wraps
		{X: 34, Y: 14, W: 56, H: 12},
	}
	for i := range want {
		if rects[i] != want[i] {
			t.Errorf("button %d = %+v, want %+v", i, rects[i], want[i])
		}
	}
	if used != 26 {
		t.Errorf("height used = %d, want 26", used)
	}
	if _, used := NewButtonBar().Layout(Rect{W: 100, H: 100}, 8, textW8); used != 0 {
		t.Errorf("an empty bar used %d", used)
	}
	// A button wider than the area still gets a row of its own.
	wide := barOf("A very long label indeed")
	if r, _ := wide.Layout(Rect{W: 50, H: 100}, 8, textW8); r[0].X != 0 || r[0].Y != 0 {
		t.Errorf("a too-wide first button is at %+v", r[0])
	}
}

func TestButtonBarCanCentreItsRows(t *testing.T) {
	b := barOf("Undo", "Redo", "Pen", "Flip H")
	b.Centred = true
	area := Rect{X: 10, Y: 0, W: 100, H: 200}
	rects, used := b.Layout(area, 8, textW8)
	// The first row, Undo and Redo, is 82 wide in 100, so it moves 9 right; the
	// second, Pen and Flip H, is 90 wide, so it moves 5.
	want := []Rect{
		{X: 19, Y: 0, W: 40, H: 12},
		{X: 61, Y: 0, W: 40, H: 12},
		{X: 15, Y: 14, W: 32, H: 12},
		{X: 49, Y: 14, W: 56, H: 12},
	}
	for i := range want {
		if rects[i] != want[i] {
			t.Errorf("button %d = %+v, want %+v", i, rects[i], want[i])
		}
	}
	if used != 26 {
		t.Errorf("height used = %d, want it unchanged at 26", used)
	}
	// A press lands on the button where it is drawn.
	x, y := mid(rects[0])
	if id, ok := b.Update(Input{MouseX: x, MouseY: y, MousePressed: true}, area, 8, textW8); !ok || id != "Undo" {
		t.Errorf("a press on the centred Undo gave %q %v", id, ok)
	}
	// A row wider than the area stays at the left edge instead of moving out.
	wide := barOf("A very long label indeed")
	wide.Centred = true
	if r, _ := wide.Layout(Rect{X: 5, W: 50, H: 100}, 8, textW8); r[0].X != 5 {
		t.Errorf("a too-wide button moved to %d, want it to stay at 5", r[0].X)
	}
}

func TestButtonBarShowsIconsAsSquareButtonsWhenAsked(t *testing.T) {
	b := barOf("Undo", "Redo", "Note")
	b.Buttons[0].Icon, b.Buttons[1].Icon = IconUndo, IconRedo // "Note" has none
	area := Rect{X: 0, Y: 0, W: 300, H: 100}
	words, _ := b.Layout(area, 8, textW8)
	if words[0].W != 40 {
		t.Fatalf("with words an Undo button is %d wide, want 40", words[0].W)
	}
	b.UseIcons = true
	if rects, _ := b.Layout(area, 8, textW8); rects[0].W != 40 {
		t.Error("without an icon renderer the buttons must keep their words")
	}
	b.Icons = &iconCanvas{}
	rects, _ := b.Layout(area, 8, textW8)
	if rects[0].W != 12 || rects[1].W != 12 || rects[0].H != 12 {
		t.Errorf("icon buttons are %dx%d and %d wide, want 12 square", rects[0].W, rects[0].H, rects[1].W)
	}
	if rects[2].W != 8*4+8 {
		t.Errorf("a button with no icon is %d wide, want its word's width, 40", rects[2].W)
	}
}

func TestButtonBarDrawsIconsThroughTheIconRenderer(t *testing.T) {
	th := DarkTheme()
	b := barOf("Undo", "Redo", "Pen", "Note")
	for i, g := range []rune{IconUndo, IconRedo, IconPen} {
		b.Buttons[i].Icon = g
	}
	b.Buttons[0].Disabled = true
	b.Buttons[2].On = true
	b.UseIcons = true
	ic := &iconCanvas{}
	b.Icons = ic
	area := Rect{X: 0, Y: 0, W: 300, H: 100}
	rects, _ := b.Layout(area, 8, textW8)
	b.Update(Input{MouseX: -50, MouseY: -50}, area, 8, textW8) // no pointer on any button
	cv := &listCanvas{}
	b.Draw(cv, th, area, 1)
	want := map[string]struct {
		rect Rect
		col  Colour
	}{
		string(IconUndo): {rects[0], th.Disabled},
		string(IconRedo): {rects[1], th.ButtonText},
		string(IconPen):  {rects[2], th.SelText},
	}
	seen := 0
	for _, d := range ic.texts {
		w, ok := want[d.S]
		if !ok {
			t.Errorf("the icon renderer drew %q", d.S)
			continue
		}
		seen++
		if d.X != w.rect.X+(w.rect.W-8)/2 || d.Y != w.rect.Y+(w.rect.H-8)/2 || d.C != w.col {
			t.Errorf("icon U+%04X at %d,%d in %+v, want it centred in %+v in %+v", []rune(d.S)[0], d.X, d.Y, d.C, w.rect, w.col)
		}
	}
	if seen != 3 {
		t.Errorf("%d of the three icons were drawn", seen)
	}
	words := map[string]bool{}
	for _, d := range cv.texts {
		words[d.S] = true
	}
	if words["Undo"] || words["Redo"] || words["Pen"] || !words["Note"] {
		t.Errorf("words drawn in icon mode: %v, want only the button that has no icon", words)
	}
}

func TestAnIconButtonNamesItselfWhileThePointerRestsOnIt(t *testing.T) {
	th := DarkTheme()
	b := barOf("Undo", "Redo")
	b.Buttons[0].Icon, b.Buttons[1].Icon = IconUndo, IconRedo
	b.UseIcons, b.Icons = true, &iconCanvas{}
	area := Rect{X: 0, Y: 0, W: 300, H: 100}
	rects, _ := b.Layout(area, 8, textW8)
	tipFor := func(label string) *listText {
		cv := &listCanvas{}
		b.Draw(cv, th, area, 1)
		for i := range cv.texts {
			if cv.texts[i].S == label {
				return &cv.texts[i]
			}
		}
		return nil
	}
	if tipFor("Redo") != nil {
		t.Fatal("a tip was shown with no pointer over any button")
	}
	x, y := mid(rects[1])
	b.Update(Input{MouseX: x, MouseY: y}, area, 8, textW8)
	tip := tipFor("Redo")
	if tip == nil {
		t.Fatal("no tip named the button under the pointer")
	}
	if tip.Y <= rects[1].Y+rects[1].H || tip.X < area.X {
		t.Errorf("the tip is at %d,%d, want it below the button %+v", tip.X, tip.Y, rects[1])
	}
	if tipFor("Undo") != nil {
		t.Error("a tip named a button that has no pointer on it")
	}
	b.UseIcons = false
	if tipFor("Redo") == nil {
		t.Fatal("in word mode the button shows its word")
	}
	cv := &listCanvas{}
	b.Draw(cv, th, area, 1)
	count := 0
	for _, d := range cv.texts {
		if d.S == "Redo" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("in word mode %q was drawn %d times, want once and no tip", "Redo", count)
	}
}

func TestATipIsOneZoomStepSmallerThanTheBarAndNeverBelowOne(t *testing.T) {
	th := DarkTheme()
	for _, c := range []struct{ scale, want int }{{1, 1}, {2, 1}, {3, 2}} {
		b := barOf("Undo", "Redo")
		b.Buttons[0].Icon, b.Buttons[1].Icon = IconUndo, IconRedo
		b.UseIcons, b.Icons = true, &iconCanvas{}
		area := Rect{X: 0, Y: 0, W: 400, H: 200}
		lh := 8 * c.scale
		rects, _ := b.Layout(area, lh, textW8)
		x, y := mid(rects[1])
		b.Update(Input{MouseX: x, MouseY: y}, area, lh, textW8)
		cv := &listCanvas{}
		b.Draw(cv, th, area, c.scale)
		var tip *listText
		for i := range cv.texts {
			if cv.texts[i].S == "Redo" {
				tip = &cv.texts[i]
			}
		}
		if tip == nil {
			t.Fatalf("zoom %d: no tip was drawn", c.scale)
		}
		if tip.Scale != c.want {
			t.Errorf("zoom %d: the tip is drawn at scale %d, want %d", c.scale, tip.Scale, c.want)
		}
		found := false
		for _, f := range cv.fills {
			if f.C == th.Panel {
				found = true
				if want := 8*c.want + 4*c.want; f.R.H != want { // a line and a half of the tip's own size
					t.Errorf("zoom %d: the tip's box is %d high, want %d", c.scale, f.R.H, want)
				}
			}
		}
		if !found {
			t.Errorf("zoom %d: no tip box was drawn", c.scale)
		}
	}
}

func TestATipStaysInsideTheAreaNearTheRightEdge(t *testing.T) {
	th := DarkTheme()
	b := barOf("Undo", "Redo", "Pen")
	for i, g := range []rune{IconUndo, IconRedo, IconPen} {
		b.Buttons[i].Icon = g
	}
	b.UseIcons, b.Icons = true, &iconCanvas{}
	area := Rect{X: 10, Y: 0, W: 56, H: 100}
	rects, _ := b.Layout(area, 8, textW8) // the third button ends at 50, the area at 66
	x, y := mid(rects[2])
	b.Update(Input{MouseX: x, MouseY: y}, area, 8, textW8)
	cv := &listCanvas{}
	b.Draw(cv, th, area, 1)
	found := false
	for _, f := range cv.fills {
		if f.C == th.Panel {
			found = true
			if f.R.X < area.X || f.R.X+f.R.W > area.X+area.W {
				t.Errorf("the tip %+v runs outside the area %+v", f.R, area)
			}
		}
	}
	if !found {
		t.Error("no tip was drawn")
	}
}

func TestButtonBarAnswersAPressOnAnEnabledButton(t *testing.T) {
	b := barOf("Undo", "Redo")
	b.Buttons[1].Disabled = true
	area := Rect{X: 0, Y: 0, W: 200, H: 100}
	rects, _ := b.Layout(area, 8, textW8)
	x, y := mid(rects[0])
	if id, ok := b.Update(Input{MouseX: x, MouseY: y, MousePressed: true}, area, 8, textW8); !ok || id != "Undo" {
		t.Errorf("a press on Undo gave %q %v", id, ok)
	}
	x, y = mid(rects[1])
	if _, ok := b.Update(Input{MouseX: x, MouseY: y, MousePressed: true}, area, 8, textW8); ok {
		t.Error("a disabled button answered")
	}
	if _, ok := b.Update(Input{MouseX: x, MouseY: y}, area, 8, textW8); ok {
		t.Error("hovering without a press answered")
	}
	if _, ok := b.Update(Input{MouseX: 190, MouseY: 90, MousePressed: true}, area, 8, textW8); ok {
		t.Error("a press on empty space answered")
	}
}

func TestButtonBarDrawsChosenDisabledAndHotButtons(t *testing.T) {
	th := DarkTheme()
	b := barOf("Pen", "Move", "Undo")
	b.Buttons[0].On = true
	b.Buttons[2].Disabled = true
	area := Rect{X: 0, Y: 0, W: 300, H: 100}
	rects, _ := b.Layout(area, 8, textW8)
	x, y := mid(rects[1])
	b.Update(Input{MouseX: x, MouseY: y}, area, 8, textW8) // the pointer rests on Move
	cv := &listCanvas{}
	b.Draw(cv, th, area, 1)
	fill := func(r Rect) Colour {
		for _, f := range cv.fills {
			if f.R == r {
				return f.C
			}
		}
		return Colour{}
	}
	if fill(rects[0]) != th.SelFill {
		t.Error("the chosen button is not drawn with the selection fill")
	}
	if fill(rects[1]) != th.ButtonHot {
		t.Error("the button under the pointer is not drawn hot")
	}
	if fill(rects[2]) != th.Button {
		t.Error("a disabled button should keep the plain fill")
	}
	text := map[string]Colour{}
	for _, d := range cv.texts {
		text[d.S] = d.C
	}
	if text["Pen"] != th.SelText || text["Move"] != th.ButtonText || text["Undo"] != th.Disabled {
		t.Errorf("label colours = %+v", text)
	}
}
