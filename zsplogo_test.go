package zenui

import (
	"testing"

	"github.com/ha1tch/zenui/zxpalette"
)

// The palette is the Spectrum's eight colours, lighter and a little desaturated;
// black is 70% white. The values are pinned, so that a change to them is a
// decision and not a side effect of the arithmetic.
func TestZSPLogoPaletteIsThePastelSpectrum(t *testing.T) {
	want := [8]Colour{
		{0xb3, 0xb3, 0xb3, 0xff}, // black: 70% white
		{0x81, 0x81, 0xdb, 0xff}, // blue
		{0xdc, 0x82, 0x82, 0xff}, // red
		{0xde, 0x84, 0xde, 0xff}, // magenta
		{0x85, 0xdf, 0x85, 0xff}, // green
		{0x87, 0xe1, 0xe1, 0xff}, // cyan
		{0xe2, 0xe2, 0x88, 0xff}, // yellow
		{0xe4, 0xe4, 0xe4, 0xff}, // white
	}
	if zspLogoPalette != want {
		t.Errorf("zspLogoPalette = %+v, want %+v", zspLogoPalette, want)
	}
}

// Each colour is lighter than its Spectrum original, and less saturated: the
// spread between its strongest and weakest channel is smaller than lightening
// alone would leave.
func TestZSPLogoPaletteIsLighterAndSofterThanTheSpectrum(t *testing.T) {
	for i := 1; i < 8; i++ { // black is the exception
		o := zxpalette.RGBA[i]
		got := zspLogoPalette[i]
		if int(got.R)+int(got.G)+int(got.B) <= int(o.R)+int(o.G)+int(o.B) {
			t.Errorf("%s is not lighter than its original: %+v against %+v", zxpalette.Names[i], got, o)
		}
		// Lightening alone takes 0xC8 against 0x00 to 100 apart; desaturating
		// takes a tenth of that off.
		spread := float64(max(got.R, got.G, got.B)) - float64(min(got.R, got.G, got.B))
		lightened := 200 * (1 - zspLogoLighten)
		if i != zxpalette.White && (spread < lightened*(1-zspLogoDesaturate)-1 || spread > lightened*(1-zspLogoDesaturate)+1) {
			t.Errorf("%s is %.0f between its channels, want %.0f: a tenth less than lightening alone leaves", zxpalette.Names[i], spread, lightened*(1-zspLogoDesaturate))
		}
	}
	if zspLogoBlack != (Colour{0xb3, 0xb3, 0xb3, 0xff}) {
		t.Errorf("black is %+v, want 70%% white, 0xb3 in each channel", zspLogoBlack)
	}
}

func TestZSPLogoGeometryBlockProportionIsOneToTwo(t *testing.T) {
	_, blockW, blockH, ok := zspLogoGeometry(100, 800, 24)
	if !ok {
		t.Fatal("expected ok=true with plenty of room")
	}
	if blockH != 8 {
		t.Errorf("blockH = %d, want 8 (24/3, three levels filling the bar height)", blockH)
	}
	if blockW != 4 {
		t.Errorf("blockW = %d, want 4 (half of blockH, matching the reference logo's 45x90 block ratio)", blockW)
	}
}

func TestZSPLogoGeometryNotEnoughSpace(t *testing.T) {
	_, _, _, ok := zspLogoGeometry(790, 800, 24)
	if ok {
		t.Error("expected ok=false with almost no room (logo needs 12 block-widths = 48px at the default block size)")
	}
}

func TestZSPLogoGeometryTooShortForBlocks(t *testing.T) {
	_, _, _, ok := zspLogoGeometry(100, 800, 4) // height=4 -> blockH=1, below the minimum of 2
	if ok {
		t.Error("expected ok=false when the bar is too short for blocks to read as anything")
	}
}

func TestDrawZSPLogoFillsTwelveBlocks(t *testing.T) {
	r := newDrawRecorder()
	drawZSPLogo(r, 100, 0, 4, 8, [zspLogoTeeth]Colour{})

	fillCount := 0
	for _, c := range *r.calls {
		if c.kind == "fill" {
			fillCount++
		}
	}
	if fillCount != 12 {
		t.Errorf("fill count = %d, want 12 (3 blocks/tooth * 2 teeth/half * 2 halves)", fillCount)
	}
}

// Each tooth is drawn in the colour its own entry names, all three of its blocks
// alike, and the four teeth are independent: two can share a colour and the
// others differ.
func TestDrawZSPLogoColoursEachToothOnItsOwn(t *testing.T) {
	r := newDrawRecorder()
	p := zspLogoPalette
	teeth := [zspLogoTeeth]Colour{p[2], p[5], p[5], p[7]}
	drawZSPLogo(r, 100, 0, 4, 8, teeth)

	var fills []drawCall
	for _, c := range *r.calls {
		if c.kind == "fill" {
			fills = append(fills, c)
		}
	}
	if len(fills) != 12 {
		t.Fatalf("%d fills, want 12", len(fills))
	}
	for i, c := range fills {
		tooth := i / 3
		if want := teeth[tooth]; c.col != want {
			t.Errorf("block %d (tooth %d) is %+v, want %+v", i, tooth, c.col, want)
		}
		if lo, hi := 100+tooth*3*4, 100+(tooth+1)*3*4+2*4; c.rect.X < lo || c.rect.X >= hi {
			t.Errorf("block %d is at x=%d, outside tooth %d's span", i, c.rect.X, tooth)
		}
	}
}

// A quartet is picked with one random colour for each tooth, never the colour
// the tooth already has, and a first quartet may be any.
func TestNextZSPLogoTeethChangesEveryTooth(t *testing.T) {
	n := len(zspLogoPalette)
	for prevColour := 0; prevColour < n; prevColour++ {
		for k := 0; k < n-1; k++ { // every answer the source can give
			next := nextZSPLogoTeeth([zspLogoTeeth]int{prevColour, prevColour, prevColour, prevColour},
				func(m int) int {
					if m != n-1 {
						t.Fatalf("asked for %d choices, want %d: the colour it has is not one", m, n-1)
					}
					return k
				})
			for i, c := range next {
				if c == prevColour || c < 0 || c >= n {
					t.Errorf("tooth %d: %d after %d with the source giving %d", i, c, prevColour, k)
				}
			}
		}
	}
	// Over all the answers, each of the other colours can come.
	seen := map[int]bool{}
	for k := 0; k < n-1; k++ {
		seen[nextZSPLogoTeeth([zspLogoTeeth]int{3, 3, 3, 3}, func(int) int { return k })[0]] = true
	}
	if len(seen) != n-1 || seen[3] {
		t.Errorf("from colour 3 the next could be %v, want each of the other %d", seen, n-1)
	}
	first := nextZSPLogoTeeth([zspLogoTeeth]int{-1, -1, -1, -1}, func(m int) int {
		if m != n {
			t.Fatalf("a first pick asked for %d choices, want all %d", m, n)
		}
		return m - 1
	})
	if first != [zspLogoTeeth]int{n - 1, n - 1, n - 1, n - 1} {
		t.Errorf("first quartet %v, want the source's answer for every tooth", first)
	}
}

// The four teeth draw their own chance: a source that answers differently each
// time gives four different colours.
func TestNextZSPLogoTeethAreIndependent(t *testing.T) {
	calls := 0
	next := nextZSPLogoTeeth([zspLogoTeeth]int{-1, -1, -1, -1}, func(int) int { calls++; return calls - 1 })
	if calls != zspLogoTeeth || next != [zspLogoTeeth]int{0, 1, 2, 3} {
		t.Errorf("%d picks gave %v, want one pick a tooth: 0 1 2 3", calls, next)
	}
}

func TestDrawZSPLogoAscendingStaircaseDirection(t *testing.T) {
	// Matches the measured reference exactly: as block level rises
	// (bottom -> middle -> top), x increases and y decreases together.
	r := newDrawRecorder()
	drawZSPLogo(r, 100, 0, 4, 8, [zspLogoTeeth]Colour{})

	var rects []Rect
	for _, c := range *r.calls {
		if c.kind == "fill" {
			rects = append(rects, c.rect)
		}
	}
	if len(rects) < 3 {
		t.Fatal("expected at least 3 fills for the first tooth")
	}
	// First three fills are half=0, tooth=0, level=0,1,2 in that order.
	bottom, middle, top := rects[0], rects[1], rects[2]
	if !(bottom.X < middle.X && middle.X < top.X) {
		t.Errorf("x should increase with level: bottom.X=%d middle.X=%d top.X=%d", bottom.X, middle.X, top.X)
	}
	if !(bottom.Y > middle.Y && middle.Y > top.Y) {
		t.Errorf("y should decrease with level: bottom.Y=%d middle.Y=%d top.Y=%d", bottom.Y, middle.Y, top.Y)
	}
}

// The phase of a cycle gives the colours: the quartet until the hold is over,
// then a straight blend towards the next, which is reached at the end.
func TestZSPLogoShownHoldsThenFades(t *testing.T) {
	cur, next := [zspLogoTeeth]int{1, 2, 3, 4}, [zspLogoTeeth]int{5, 6, 7, 0}
	at := func(phase float64) [zspLogoTeeth]Colour { return zspLogoShown(cur, next, phase) }
	pal := func(q [zspLogoTeeth]int) (c [zspLogoTeeth]Colour) {
		for i, k := range q {
			c[i] = zspLogoPalette[k]
		}
		return
	}
	for _, phase := range []float64{0, 0.5, zspLogoHold} {
		if at(phase) != pal(cur) {
			t.Errorf("phase %v is not the held quartet", phase)
		}
	}
	for _, phase := range []float64{zspLogoCycle, zspLogoCycle + 1, 99} {
		if at(phase) != pal(next) {
			t.Errorf("phase %v is not the next quartet", phase)
		}
	}
	mid := at(zspLogoHold + zspLogoFade/2)
	for i := range mid {
		if want := blendColour(zspLogoPalette[cur[i]], zspLogoPalette[next[i]], 0.5); mid[i] != want {
			t.Errorf("tooth %d half-way is %+v, want %+v", i, mid[i], want)
		}
		if mid[i] == zspLogoPalette[cur[i]] || mid[i] == zspLogoPalette[next[i]] {
			t.Errorf("tooth %d half-way is one of its ends: %+v", i, mid[i])
		}
	}
	if zspLogoHold != 1.0 || zspLogoFade != 0.5 {
		t.Errorf("hold %v and fade %v, want a second and half a second", zspLogoHold, zspLogoFade)
	}
}

func TestBlendColourEnds(t *testing.T) {
	a, b := Colour{10, 200, 30, 0xff}, Colour{250, 0, 31, 0xff}
	if blendColour(a, b, 0) != a || blendColour(a, b, 1) != b {
		t.Errorf("ends: %+v %+v", blendColour(a, b, 0), blendColour(a, b, 1))
	}
	if got := blendColour(a, b, 0.5); got != (Colour{130, 100, 31, 0xff}) {
		t.Errorf("middle %+v", got)
	}
}

// The bar draws the cycle: a quartet held for a second, a fade of half a second
// to the next, which is then held, and so on; a new quartet is picked once for
// each cycle, and every tooth of it has another colour than it had.
func TestZSPLogoHoldsFadesAndHoldsAgain(t *testing.T) {
	mb := NewMenuBar(MenuBarConfig{Items: testBarItems()})
	theme := DefaultTheme()
	theme.ShowZSPLogo = true
	picks := 0
	mb.logoRandom = func(n int) int { picks++; return picks % n }

	// drawn is what the bar fills each tooth with, the first block of it.
	drawn := func() (q [zspLogoTeeth]Colour) {
		r := newDrawRecorder()
		mb.Draw(r, 800, 600, 0, 24, theme)
		var fills []Colour
		for _, c := range *r.calls {
			if c.kind == "fill" && c.rect.W < 20 && c.rect.H > 4 && c.rect.H < 12 {
				fills = append(fills, c.col)
			}
		}
		if len(fills) < 3*zspLogoTeeth {
			t.Fatalf("%d block fills, want the logo's %d", len(fills), 3*zspLogoTeeth)
		}
		fills = fills[len(fills)-3*zspLogoTeeth:] // the logo is drawn last of them
		for i := range q {
			q[i] = fills[3*i]
		}
		return
	}
	pal := func(q [zspLogoTeeth]int) (c [zspLogoTeeth]Colour) {
		for i, k := range q {
			c[i] = zspLogoPalette[k]
		}
		return
	}

	first := drawn()
	if picks != 2*zspLogoTeeth {
		t.Fatalf("%d picks at the start, want the quartet and the next (%d)", picks, 2*zspLogoTeeth)
	}
	cur, next := mb.logoTeeth, mb.logoNext
	if first != pal(cur) {
		t.Errorf("starts with %+v, want the first quartet %+v", first, pal(cur))
	}
	for i := range cur {
		if cur[i] == next[i] {
			t.Errorf("tooth %d: the next quartet keeps colour %d", i, cur[i])
		}
	}

	// Still in the hold: nothing moves, nothing is picked.
	mb.Update(Input{DeltaTime: float32(zspLogoHold) - 0.01})
	if got := drawn(); got != first || picks != 2*zspLogoTeeth {
		t.Errorf("moved within the hold: %+v after %d picks", got, picks)
	}
	// In the fade: each tooth is between its two colours, by how far in it is.
	mb.Update(Input{DeltaTime: 0.26})
	phase := mb.logoElapsed
	tt := (phase - zspLogoHold) / zspLogoFade
	if tt <= 0.4 || tt >= 0.6 {
		t.Fatalf("the test's own timing is off: %v of the way through the fade", tt)
	}
	got := drawn()
	for i := range got {
		if want := blendColour(zspLogoPalette[cur[i]], zspLogoPalette[next[i]], tt); got[i] != want {
			t.Errorf("tooth %d in the fade is %+v, want %+v", i, got[i], want)
		}
	}
	if picks != 2*zspLogoTeeth {
		t.Errorf("%d picks during the fade, want none new", picks)
	}
	// The cycle is over: the quartet it faded to holds, and one more is picked after it.
	mb.Update(Input{DeltaTime: float32(zspLogoCycle) - float32(phase) + 0.05})
	if got := drawn(); got != pal(next) {
		t.Errorf("after the fade shows %+v, want the next quartet %+v", got, pal(next))
	}
	if picks != 3*zspLogoTeeth || mb.logoTeeth != next {
		t.Errorf("%d picks, current %v: want %d, and the old next as current %v", picks, mb.logoTeeth, 3*zspLogoTeeth, next)
	}
	for i := range next {
		if mb.logoNext[i] == next[i] {
			t.Errorf("tooth %d: the quartet after keeps colour %d", i, next[i])
		}
	}
	// A long frame is one step on, not many.
	mb.Update(Input{DeltaTime: 20})
	drawn()
	if picks != 4*zspLogoTeeth {
		t.Errorf("%d picks after a long frame, want one more quartet (%d)", picks, 4*zspLogoTeeth)
	}
}

func TestZSPLogoDrawnOnlyWhenThemeRequestsIt(t *testing.T) {
	mb := NewMenuBar(MenuBarConfig{Items: testBarItems()})

	themeOff := DefaultTheme()
	themeOff.ShowZSPLogo = false
	r1 := newDrawRecorder()
	mb.Draw(r1, 800, 600, 0, 24, themeOff)
	countOff := len(*r1.calls)

	themeOn := DefaultTheme()
	themeOn.ShowZSPLogo = true
	r2 := newDrawRecorder()
	mb.Draw(r2, 800, 600, 0, 24, themeOn)
	countOn := len(*r2.calls)

	if countOn <= countOff {
		t.Errorf("draw call count with ShowZSPLogo=true (%d) should exceed false (%d)", countOn, countOff)
	}
}
