package zenui

import "testing"

// The stepper is at (10,20): with 8-pixel glyphs its buttons are 12 square, with 2
// between them, and "<<" and ">>" are 20 wide.
var stpArea = Rect{X: 10, Y: 20, W: 300, H: 40}

func stpButton(t *testing.T, s *Stepper, label string) Rect {
	t.Helper()
	btns, _ := s.layout(stpArea, 8, tfW)
	for _, b := range btns {
		if b.label == label {
			return b.rect
		}
	}
	t.Fatalf("the stepper has no %q button", label)
	return Rect{}
}

func stpClick(t *testing.T, s *Stepper, label string) (int, bool) {
	t.Helper()
	x, y := mid(stpButton(t, s, label))
	return s.Update(Input{MouseX: x, MouseY: y, MousePressed: true}, stpArea, 8, tfW)
}

func TestAStepperStepsByOneAndStopsAtItsLimits(t *testing.T) {
	s := NewStepper(0, 3)
	if v, ok := stpClick(t, s, "<"); ok || v != 0 {
		t.Errorf("stepping down at the minimum gave %d, %v", v, ok)
	}
	for want := 1; want <= 3; want++ {
		if v, ok := stpClick(t, s, ">"); !ok || v != want {
			t.Errorf("stepping up gave %d, %v, want %d", v, ok, want)
		}
	}
	if v, ok := stpClick(t, s, ">"); ok || v != 3 {
		t.Errorf("stepping up at the maximum gave %d, %v", v, ok)
	}
	if v, ok := stpClick(t, s, "<"); !ok || v != 2 || s.Value != 2 {
		t.Errorf("stepping down gave %d, %v, value %d, want 2", v, ok, s.Value)
	}
}

func TestABigStepperAlsoStepsInTens(t *testing.T) {
	s := NewStepper(0, 99)
	s.Big = 10
	s.Value = 95
	if v, ok := stpClick(t, s, ">>"); !ok || v != 99 {
		t.Errorf(">> from 95 gave %d, %v, want 99, the maximum", v, ok)
	}
	if v, ok := stpClick(t, s, ">>"); ok || v != 99 {
		t.Errorf(">> at the maximum gave %d, %v", v, ok)
	}
	if v, ok := stpClick(t, s, "<<"); !ok || v != 89 {
		t.Errorf("<< from 99 gave %d, %v, want 89", v, ok)
	}
	s.Value = 5
	if v, ok := stpClick(t, s, "<<"); !ok || v != 0 {
		t.Errorf("<< from 5 gave %d, %v, want 0, the minimum", v, ok)
	}
	if r := stpButton(t, s, "<<"); r.W != 20 {
		t.Errorf("the << button is %d wide, want 20", r.W)
	}
	if plain := NewStepper(0, 9); len(func() []stepButton { b, _ := plain.layout(stpArea, 8, tfW); return b }()) != 2 {
		t.Error("a stepper with no Big has more than two buttons")
	}
}

func TestButtonRectFindsEachButtonByItsLabel(t *testing.T) {
	s := NewStepper(0, 99)
	s.Big = 10
	for _, label := range []string{"<<", "<", ">", ">>"} {
		got, ok := s.ButtonRect(label, stpArea, 8, tfW)
		if want := stpButton(t, s, label); !ok || got != want {
			t.Errorf("ButtonRect(%q) = %+v, %v, want %+v", label, got, ok, want)
		}
	}
	if _, ok := s.ButtonRect("?", stpArea, 8, tfW); ok {
		t.Error("ButtonRect found a button that is not there")
	}
	if _, ok := NewStepper(0, 9).ButtonRect("<<", stpArea, 8, tfW); ok {
		t.Error("a stepper with no Big has a << button")
	}
}

func TestADisabledStepperDoesNotMove(t *testing.T) {
	s := NewStepper(0, 9)
	s.Value, s.Disabled = 4, true
	if v, ok := stpClick(t, s, ">"); ok || v != 4 {
		t.Errorf("a disabled stepper gave %d, %v", v, ok)
	}
}

func TestAStepperKeepsItsLayoutWhateverItsValue(t *testing.T) {
	s := NewStepper(0, 99)
	s.Big = 10
	s.Format = func(v int) string {
		if v == 0 {
			return "none"
		}
		return "#" + string(rune('0'+v%10))
	}
	a, av := s.layout(stpArea, 8, tfW)
	s.Value = 77
	b, bv := s.layout(stpArea, 8, tfW)
	if av != bv || len(a) != len(b) {
		t.Fatalf("the value box moved: %+v then %+v", av, bv)
	}
	for i := range a {
		if a[i] != b[i] {
			t.Errorf("button %d moved from %+v to %+v as the value changed", i, a[i], b[i])
		}
	}
	if av.W < 4*8 {
		t.Errorf("the value box is %d wide, want room for the widest text, none (32)", av.W)
	}
	down, up := stpButton(t, s, "<"), stpButton(t, s, ">")
	if !(down.X+down.W <= av.X && av.X+av.W <= up.X) {
		t.Errorf("want < then the value then >, got %+v %+v %+v", down, av, up)
	}
}

func TestAStepperDrawsTheFormattedValueAndGreysWhatCannotAct(t *testing.T) {
	th := DarkTheme()
	s := NewStepper(0, 16)
	s.Format = func(v int) string {
		if v == 0 {
			return "none"
		}
		return "n"
	}
	textColour := func(cv *listCanvas, label string) Colour {
		for _, d := range cv.texts {
			if d.S == label {
				return d.C
			}
		}
		t.Fatalf("%q was not drawn", label)
		return Colour{}
	}
	cv := &listCanvas{}
	s.Draw(cv, th, stpArea, 1)
	if textColour(cv, "none") != th.Text {
		t.Error("the value should be in the text colour")
	}
	if textColour(cv, "<") != th.Disabled || textColour(cv, ">") != th.ButtonText {
		t.Error("at the minimum only the down button should be greyed")
	}
	s.Disabled = true
	cv = &listCanvas{}
	s.Draw(cv, th, stpArea, 1)
	if textColour(cv, ">") != th.Disabled || textColour(cv, "none") != th.Disabled {
		t.Error("a disabled stepper should grey every part of itself")
	}
}
