package zenui

import "testing"

func TestScaleAlphaBasicMath(t *testing.T) {
	c := Colour{R: 0x11, G: 0x22, B: 0x33, A: 0xff}
	got := ScaleAlpha(c, 0.85)
	if got.R != c.R || got.G != c.G || got.B != c.B {
		t.Errorf("ScaleAlpha should leave R/G/B untouched: got %+v, from %+v", got, c)
	}
	var full float64 = 0xff
	want := uint8(full * 0.85) // truncating conversion, same as the real call site
	if got.A != want {
		t.Errorf("A = %d, want %d (0xff * 0.85)", got.A, want)
	}
}

func TestScaleAlphaZeroFactor(t *testing.T) {
	c := Colour{R: 1, G: 2, B: 3, A: 0xa0}
	got := ScaleAlpha(c, 0)
	if got.A != 0 {
		t.Errorf("A = %d, want 0", got.A)
	}
}

func TestTextModalPanelOpacityIs85Percent(t *testing.T) {
	// Confirms the actual client-area opacity independent of whatever
	// theme.Panel's own alpha is (currently always 0xff across every
	// theme, but this shouldn't assume that stays true).
	for name, theme := range map[string]Theme{
		"Default":  DefaultTheme(),
		"Dark":     DarkTheme(),
		"Light":    LightTheme(),
		"Spectrum": SpectrumTheme(),
	} {
		want := uint8(float64(theme.Panel.A) * ModalPanelOpacity)
		got := ScaleAlpha(theme.Panel, ModalPanelOpacity)
		if got.A != want {
			t.Errorf("%s: panel alpha = %d, want %d (85%% of theme.Panel.A=%d)", name, got.A, want, theme.Panel.A)
		}
	}
}

func TestTextModalBackdropIs20PercentLessDark(t *testing.T) {
	for name, theme := range map[string]Theme{
		"Default":  DefaultTheme(),
		"Dark":     DarkTheme(),
		"Light":    LightTheme(),
		"Spectrum": SpectrumTheme(),
	} {
		want := uint8(float64(theme.Backdrop.A) * ModalBackdropDarkening)
		got := ScaleAlpha(theme.Backdrop, ModalBackdropDarkening)
		if got.A != want {
			t.Errorf("%s: backdrop alpha = %d, want %d (80%% of theme.Backdrop.A=%d, i.e. 20%% less dark)", name, got.A, want, theme.Backdrop.A)
		}
		if got.A >= theme.Backdrop.A {
			t.Errorf("%s: adjusted backdrop alpha (%d) should be strictly less than the theme's own (%d)", name, got.A, theme.Backdrop.A)
		}
	}
}

func TestTextModalDrawUsesAdjustedOpacitiesNotThemeRaw(t *testing.T) {
	theme := DarkTheme()
	m := NewTextModal("TEST", "# Heading\n\nbody text", false)
	r := newDrawRecorder()
	m.Draw(r, 800, 600, theme)

	foundPanel, foundBackdrop := false, false
	for _, c := range *r.calls {
		if c.kind == "fill" && c.rect == m.panel {
			foundPanel = true
			if c.col.A != ScaleAlpha(theme.Panel, ModalPanelOpacity).A {
				t.Errorf("panel fill alpha = %d, want %d", c.col.A, ScaleAlpha(theme.Panel, ModalPanelOpacity).A)
			}
			if c.col.A == theme.Panel.A {
				t.Error("panel fill still uses theme.Panel's own raw alpha, unadjusted")
			}
		}
		if c.kind == "fill" && c.rect.W == 800 && c.rect.H == 600 {
			foundBackdrop = true
			if c.col.A != ScaleAlpha(theme.Backdrop, ModalBackdropDarkening).A {
				t.Errorf("backdrop fill alpha = %d, want %d", c.col.A, ScaleAlpha(theme.Backdrop, ModalBackdropDarkening).A)
			}
		}
	}
	if !foundPanel {
		t.Error("no panel fill call found")
	}
	if !foundBackdrop {
		t.Error("no backdrop fill call found")
	}
}

func TestTextModalTextAndWidgetsStayFullyOpaque(t *testing.T) {
	theme := DarkTheme()
	m := NewTextModal("TEST", "# Heading\n\nbody text", false)
	r := newDrawRecorder()
	m.Draw(r, 800, 600, theme)

	for _, c := range *r.calls {
		if c.kind == "text" {
			if c.col.A != 0xff {
				t.Errorf("text draw call has alpha %d, want 0xff (widgets/text must stay 100%% opaque)", c.col.A)
			}
		}
	}
}

func TestTextModalAutoHeightFitsContentNotFullScreen(t *testing.T) {
	r := noopRenderer{}
	screenW, screenH := 1200, 900

	short := NewTextModal("ABOUT", "# ZenZX\n\nline two\nline three", true)
	short.layout(r, screenW, screenH)

	tall := NewTextModal("HELP", "line", false)
	tall.layout(r, screenW, screenH)

	if short.panel.H >= tall.panel.H {
		t.Errorf("auto-height panel.H = %d, want less than the full-height panel's %d", short.panel.H, tall.panel.H)
	}
	if short.panel.H >= screenH {
		t.Errorf("auto-height panel.H = %d, should not fill the whole screen (screenH=%d)", short.panel.H, screenH)
	}
}

func TestTextModalAutoHeightHeightMatchesContentEstimate(t *testing.T) {
	r := noopRenderer{}
	content := "line one\nline two\nline three\nline four\nline five"
	m := NewTextModal("ABOUT", content, true)
	m.layout(r, 1200, 900)

	lh := r.LineHeight(1)
	pad := lh
	titleH := r.LineHeight(2) + 8
	wantContentH := m.total * m.bodyLH
	wantPanelH := wantContentH + 2*pad + titleH

	if m.panel.H != wantPanelH {
		t.Errorf("panel.H = %d, want %d (content estimate: %d lines * %d line height + padding + title)", m.panel.H, wantPanelH, m.total, m.bodyLH)
	}
}

func TestTextModalAutoHeightNeverNeedsScrolling(t *testing.T) {
	m := NewTextModal("ABOUT", "one\ntwo\nthree", true)
	m.layout(noopRenderer{}, 1200, 900)

	if m.total > m.visible {
		t.Errorf("total=%d, visible=%d -- an auto-height modal should always show every line without scrolling", m.total, m.visible)
	}
}

func TestTextModalAutoHeightRespectsMinimumFloor(t *testing.T) {
	// A single short line shouldn't produce an absurdly short panel --
	// the same lh*10 floor the full-height modal already has.
	m := NewTextModal("ABOUT", "x", true)
	r := noopRenderer{}
	m.layout(r, 1200, 900)

	lh := r.LineHeight(1)
	if m.panel.H < lh*10 {
		t.Errorf("panel.H = %d, want at least %d (the minimum floor)", m.panel.H, lh*10)
	}
}

func TestTextModalAutoHeightCapsAtScreenHeight(t *testing.T) {
	// Pathological case: content taller than the screen itself should
	// still be capped, not overflow.
	var longContent string
	for i := 0; i < 200; i++ {
		longContent += "line\n"
	}
	m := NewTextModal("ABOUT", longContent, true)
	r := noopRenderer{}
	screenH := 400
	m.layout(r, 1200, screenH)

	pad := r.LineHeight(1)
	if m.panel.H > screenH-2*pad {
		t.Errorf("panel.H = %d, should be capped at screenH-2*pad = %d", m.panel.H, screenH-2*pad)
	}
}

func TestTextModalFullHeight(t *testing.T) {
	// Help (autoHeight=false) should be completely unchanged from
	// before: always screenH-4*pad, regardless of content length.
	r := noopRenderer{}
	screenW, screenH := 1200, 900
	m := NewTextModal("HELP", "one line only", false)
	m.layout(r, screenW, screenH)

	pad := r.LineHeight(1)
	want := screenH - 4*pad
	if m.panel.H != want {
		t.Errorf("panel.H = %d, want %d (full-height behaviour unchanged)", m.panel.H, want)
	}
}

// TestTextModalScrollsAndCloses: the keys scroll, Home and End go to the
// ends, and Escape, the close box and a click outside the panel close it.
func TestTextModalScrollsAndCloses(t *testing.T) {
	var content string
	for i := 0; i < 300; i++ {
		content += "line\n"
	}
	m := NewTextModal("HELP", content, false)
	r := noopRenderer{}
	m.Draw(r, 1200, 900, DarkTheme())
	if m.total <= m.visible {
		t.Fatalf("300 lines fit in the window: total %d, visible %d", m.total, m.visible)
	}
	inside := Input{MouseX: m.panel.X + 5, MouseY: m.panel.Y + m.panel.H - 5}
	step := func(keys ...Key) {
		in := inside
		in.Keys = keys
		if !m.Update(in) {
			t.Fatalf("the modal closed on %v", keys)
		}
	}
	step(KeyDown, KeyDown)
	if m.scroll != 2 {
		t.Errorf("after two Downs scroll is %d", m.scroll)
	}
	step(KeyPageDown)
	if m.scroll != 2+m.visible {
		t.Errorf("after Page Down scroll is %d, want %d", m.scroll, 2+m.visible)
	}
	step(KeyEnd)
	if m.scroll != m.total-m.visible {
		t.Errorf("End: scroll %d, want %d", m.scroll, m.total-m.visible)
	}
	step(KeyHome)
	if m.scroll != 0 {
		t.Errorf("Home: scroll %d", m.scroll)
	}
	wheel := inside
	wheel.WheelY = -1
	m.Update(wheel)
	if m.scroll != 3 {
		t.Errorf("one notch of the wheel: scroll %d, want 3", m.scroll)
	}
	if m.Update(Input{Keys: []Key{KeyEscape}}) {
		t.Error("Escape did not close the modal")
	}
	if m.Update(Input{MouseX: m.closeBx.X + 1, MouseY: m.closeBx.Y + 1, MousePressed: true}) {
		t.Error("a click on the close box did not close the modal")
	}
	if m.Update(Input{MouseX: 1, MouseY: 1, MousePressed: true}) {
		t.Error("a click outside the panel did not close the modal")
	}
	if !m.Update(Input{MouseX: m.body.X + 1, MouseY: m.body.Y + 1, MousePressed: true}) {
		t.Error("a click on the text closed the modal")
	}
}

// TestTextModalCloseBoxHover: the close box is lit while the pointer the
// last Update saw is on it.
func TestTextModalCloseBoxHover(t *testing.T) {
	theme := DarkTheme()
	m := NewTextModal("ABOUT", "x", true)
	fillOfCloseBox := func() Colour {
		r := newDrawRecorder()
		m.Draw(r, 800, 600, theme)
		for _, c := range *r.calls {
			if c.kind == "fill" && c.rect == m.closeBx {
				return c.col
			}
		}
		t.Fatal("the close box was not drawn")
		return Colour{}
	}
	if fillOfCloseBox() != theme.Button {
		t.Error("the close box is lit with the pointer elsewhere")
	}
	m.Update(Input{MouseX: m.closeBx.X + 1, MouseY: m.closeBx.Y + 1})
	if fillOfCloseBox() != theme.ButtonHot {
		t.Error("the close box is not lit with the pointer on it")
	}
}
