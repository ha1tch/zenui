package zenui

import "testing"

// barRecorder is a noopRenderer that remembers where text is drawn and whether a clip is on.
type barRecorder struct {
	noopRenderer
	texts *[]Rect // the top-left of each text drawn
	clips *int    // how many Clip calls there have been
	ends  *int    // how many ClipEnd calls
}

func (r barRecorder) DrawText(s string, x, y, scale int, c Colour) {
	*r.texts = append(*r.texts, Rect{X: x, Y: y})
}
func (r barRecorder) Clip(Rect) { *r.clips++ }
func (r barRecorder) ClipEnd()  { *r.ends++ }

func recordBar(mb *MenuBar, th Theme, screenW, height int) (texts []Rect, clips, ends int) {
	mb.Draw(barRecorder{texts: &texts, clips: &clips, ends: &ends}, screenW, 600, 0, height, th)
	return
}

// The left inset moves every label right by that much and nothing else.
func TestMenuBarLeftInsetMovesTheLabels(t *testing.T) {
	th := DarkTheme()
	plain := NewMenuBar(MenuBarConfig{Items: testBarItems()})
	base, _, _ := recordBar(plain, th, 800, 24)
	baseEnd := plain.LabelsEndX()

	moved := NewMenuBar(MenuBarConfig{Items: testBarItems()})
	moved.SetInsets(30, 0)
	got, _, _ := recordBar(moved, th, 800, 24)
	if len(got) != len(base) || len(got) == 0 {
		t.Fatalf("drew %d texts with the inset and %d without", len(got), len(base))
	}
	for i := range got {
		if got[i].X != base[i].X+30 || got[i].Y != base[i].Y {
			t.Errorf("label %d at %+v, want %+v moved 30 right", i, got[i], base[i])
		}
	}
	if moved.LabelsEndX() != baseEnd+30 {
		t.Errorf("LabelsEndX = %d, want %d", moved.LabelsEndX(), baseEnd+30)
	}
}

// The right inset moves the branding left by that much, whichever it is, and leaves the labels.
func TestMenuBarRightInsetMovesTheBranding(t *testing.T) {
	for name, th := range map[string]Theme{"logo": DarkTheme(), "rainbow": SpectrumTheme()} {
		if !th.ShowZSPLogo && !th.ShowBarRainbow {
			t.Skipf("%s: the preset shows no branding", name)
		}
		plain := decorationBar(th, 800, 24)
		mb := NewMenuBar(MenuBarConfig{Items: testBarItems()})
		mb.SetInsets(0, 40)
		mb.Draw(noopRenderer{}, 800, 600, 0, 24, th)
		if got, want := mb.DecorationX(), plain.DecorationX()-40; got != want {
			t.Errorf("%s: DecorationX = %d, want %d (40 left of %d)", name, got, want, plain.DecorationX())
		}
		if mb.LabelsEndX() != plain.LabelsEndX() {
			t.Errorf("%s: the right inset moved the labels: %d, was %d", name, mb.LabelsEndX(), plain.LabelsEndX())
		}
	}
}

// The branding that no longer fits between the labels and the inset is left out, as when the
// window is narrow.
func TestMenuBarInsetsCanSqueezeTheBrandingOut(t *testing.T) {
	th := DarkTheme()
	if !th.ShowZSPLogo {
		t.Skip("the dark preset no longer shows the logo")
	}
	mb := NewMenuBar(MenuBarConfig{Items: testBarItems()})
	mb.Draw(noopRenderer{}, 800, 600, 0, 24, th)
	mb.SetInsets(0, 800)
	mb.Draw(noopRenderer{}, 800, 600, 0, 24, th)
	if got := mb.DecorationX(); got != 800 {
		t.Errorf("DecorationX = %d with no room, want the full width 800", got)
	}
}

// Negative insets are zero, and the defaults change nothing.
func TestMenuBarInsetsDefaultAndClamp(t *testing.T) {
	th := DarkTheme()
	a := NewMenuBar(MenuBarConfig{Items: testBarItems()})
	b := NewMenuBar(MenuBarConfig{Items: testBarItems()})
	b.SetInsets(-5, -5)
	a.Draw(noopRenderer{}, 800, 600, 0, 24, th)
	b.Draw(noopRenderer{}, 800, 600, 0, 24, th)
	if a.LabelsEndX() != b.LabelsEndX() || a.DecorationX() != b.DecorationX() {
		t.Errorf("negative insets changed the layout: %d,%d against %d,%d",
			b.LabelsEndX(), b.DecorationX(), a.LabelsEndX(), a.DecorationX())
	}
}

// The labels offset moves the text vertically, inside a clip, and not the layout.
func TestMenuBarLabelsOffset(t *testing.T) {
	th := DarkTheme()
	plain := NewMenuBar(MenuBarConfig{Items: testBarItems()})
	base, clips, ends := recordBar(plain, th, 800, 24)
	if clips != 0 || ends != 0 {
		t.Fatalf("without an offset the labels are clipped (%d, %d)", clips, ends)
	}
	mb := NewMenuBar(MenuBarConfig{Items: testBarItems()})
	mb.SetLabelsOffset(-24)
	got, clips, ends := recordBar(mb, th, 800, 24)
	if len(got) != len(base) {
		t.Fatalf("drew %d texts, want %d", len(got), len(base))
	}
	for i := range got {
		if got[i].X != base[i].X || got[i].Y != base[i].Y-24 {
			t.Errorf("label %d at %+v, want %+v moved 24 up", i, got[i], base[i])
		}
	}
	if clips != 1 || ends != 1 {
		t.Errorf("clip calls %d and ends %d, want one of each", clips, ends)
	}
	if mb.LabelsEndX() != plain.LabelsEndX() {
		t.Errorf("the offset moved the layout: %d, was %d", mb.LabelsEndX(), plain.LabelsEndX())
	}
}
