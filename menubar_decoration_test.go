package zenui

import "testing"

func decorationBar(th Theme, screenW, height int) *MenuBar {
	mb := NewMenuBar(MenuBarConfig{Items: testBarItems()})
	mb.Draw(noopRenderer{}, screenW, 600, 0, height, th)
	return mb
}

func TestDecorationXIsZeroBeforeTheFirstDraw(t *testing.T) {
	if got := NewMenuBar(MenuBarConfig{Items: testBarItems()}).DecorationX(); got != 0 {
		t.Errorf("DecorationX before any Draw = %d, want 0", got)
	}
}

// The logo is right-anchored 8 pixels in, twelve block-widths wide, its block
// height a third of the bar's.
func TestDecorationXIsTheLogosLeftEdge(t *testing.T) {
	th := DarkTheme()
	if !th.ShowZSPLogo {
		t.Skip("the dark preset no longer shows the logo")
	}
	const w, h = 800, 24
	blockW := (h / 3) / 2
	want := w - 8 - 12*blockW
	if got := decorationBar(th, w, h).DecorationX(); got != want {
		t.Errorf("DecorationX = %d, want %d (the logo's left edge)", got, want)
	}
}

// The rainbow is right-anchored 8 pixels in; its rows lean one pixel further
// left each, so its leftmost pixel is scanlines-1 left of its first row.
func TestDecorationXIsTheRainbowsLeftmostPixel(t *testing.T) {
	th := SpectrumTheme()
	if !th.ShowBarRainbow {
		t.Skip("the Spectrum preset no longer shows the rainbow")
	}
	const w, h = 800, 24
	scan := RainbowScanlines
	if h < scan {
		scan = h
	}
	want := w - 8 - RainbowBandWidth*len(rainbowColours) - (scan - 1)
	if got := decorationBar(th, w, h).DecorationX(); got != want {
		t.Errorf("DecorationX = %d, want %d (the rainbow's leftmost pixel)", got, want)
	}
}

func TestDecorationXIsTheFullWidthWhenThereIsNoBranding(t *testing.T) {
	th := DarkTheme()
	th.ShowZSPLogo, th.ShowBarRainbow = false, false
	const w, h = 800, 24
	if got := decorationBar(th, w, h).DecorationX(); got != w {
		t.Errorf("DecorationX = %d, want the full width %d", got, w)
	}
}

func TestDecorationXIsTheFullWidthWhenTheBrandingDoesNotFit(t *testing.T) {
	// Narrower than the labels plus the logo: the logo is not drawn at all.
	const w, h = 70, 24
	if got := decorationBar(DarkTheme(), w, h).DecorationX(); got != w {
		t.Errorf("DecorationX = %d, want the full width %d when the logo does not fit", got, w)
	}
}

// A host that packs its items against DecorationX must find the branding
// still at the right of the strip whatever the strip's width.
func TestDecorationXFollowsTheStripWidth(t *testing.T) {
	th := DarkTheme()
	a := decorationBar(th, 800, 24).DecorationX()
	b := decorationBar(th, 1000, 24).DecorationX()
	if b-a != 200 {
		t.Errorf("a 200-pixel wider strip moved the branding by %d, want 200", b-a)
	}
}
