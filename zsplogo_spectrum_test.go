package zenui

import (
	"testing"

	"github.com/ha1tch/zenui/zxpalette"
)

// Every zspLogoRun quartets of random colours, the logo shows the Spectrum's own
// quartet for zspLogoFixedHold seconds, and the random ones begin again.

func TestZSPLogoSpectrumQuartetIsTheRainbowInItsOrder(t *testing.T) {
	want := [zspLogoTeeth]int{zxpalette.Red, zxpalette.Yellow, zxpalette.Green, zxpalette.Cyan}
	if zspLogoSpectrum != want {
		t.Errorf("zspLogoSpectrum = %v, want red, yellow, green, cyan %v", zspLogoSpectrum, want)
	}
	if zspLogoRun != 10 || zspLogoFixedHold != 6.0 {
		t.Errorf("the run is %d quartets and the Spectrum's holds %v s, want 10 and 6", zspLogoRun, zspLogoFixedHold)
	}
}

func TestZSPLogoFixedQuartetComesAfterEachRunOfTen(t *testing.T) {
	var fixed []int
	for seq := 0; seq < 40; seq++ {
		if zspLogoFixedAt(seq) {
			fixed = append(fixed, seq)
		}
		want := zspLogoHold
		if zspLogoFixedAt(seq) {
			want = zspLogoFixedHold
		}
		if got := zspLogoHoldOf(seq); got != want {
			t.Errorf("quartet %d holds %v, want %v", seq, got, want)
		}
	}
	if len(fixed) != 3 || fixed[0] != 10 || fixed[1] != 21 || fixed[2] != 32 {
		t.Errorf("the Spectrum's quartet is number %v of the first 40, want 10, 21 and 32: ten random ones before each", fixed)
	}
}

func TestNextZSPLogoTeethBeforeKeepsOffBothColours(t *testing.T) {
	n := len(zspLogoPalette)
	prev := [zspLogoTeeth]int{0, 1, 2, 3}
	after := [zspLogoTeeth]int{1, 1, 4, -1}
	for k := 0; k < n; k++ {
		calls := 0
		got := nextZSPLogoTeethBefore(prev, after, func(m int) int {
			calls++
			if m < 1 || m > n {
				t.Errorf("asked for a pick of %d", m)
			}
			return k % m
		})
		if calls != zspLogoTeeth {
			t.Errorf("%d picks, want one for each of the %d teeth", calls, zspLogoTeeth)
		}
		for i, c := range got {
			if c == prev[i] || c == after[i] {
				t.Errorf("pick %d, tooth %d: colour %d is one of the two to keep off (%d, %d)", k, i, c, prev[i], after[i])
			}
		}
	}
	// Every other colour can be reached.
	seen := map[int]bool{}
	for k := 0; k < n; k++ {
		seen[nextZSPLogoTeethBefore(prev, after, func(m int) int { return k % m })[0]] = true
	}
	if len(seen) != n-2 {
		t.Errorf("tooth 0 can be %d colours, want the %d that are neither %d nor %d", len(seen), n-2, prev[0], after[0])
	}
}

func TestZSPLogoQuartetBeforeTheSpectrumsSharesNoColourWithIt(t *testing.T) {
	for k := 0; k < len(zspLogoPalette); k++ {
		pick := func(m int) int { return k % m }
		cur := nextZSPLogoTeeth([zspLogoTeeth]int{-1, -1, -1, -1}, pick)
		before := zspLogoQuartetAfter(zspLogoRun-2, cur, pick) // number zspLogoRun-1, the last random one
		for i := range before {
			if before[i] == zspLogoSpectrum[i] || before[i] == cur[i] {
				t.Errorf("pick %d, tooth %d: the last random quartet has colour %d, after %d, before the Spectrum's %d", k, i, before[i], cur[i], zspLogoSpectrum[i])
			}
		}
		if got := zspLogoQuartetAfter(zspLogoRun-1, before, pick); got != zspLogoSpectrum {
			t.Errorf("the quartet after the tenth random one is %v, want the Spectrum's %v", got, zspLogoSpectrum)
		}
		after := zspLogoQuartetAfter(zspLogoRun, zspLogoSpectrum, pick)
		for i := range after {
			if after[i] == zspLogoSpectrum[i] {
				t.Errorf("pick %d, tooth %d: the quartet after the Spectrum's keeps colour %d", k, i, after[i])
			}
		}
	}
}

// The bar, run in quarter-second frames (exact in binary, so the cycle edges fall on
// frames): ten cycles of a second and a half of random colours, then the Spectrum's
// quartet from 15 s to 18 s, its fade to 18.5 s, and ten more random cycles.
func TestZSPLogoShowsTheSpectrumQuartetForThreeSecondsAfterTenCycles(t *testing.T) {
	mb := NewMenuBar(MenuBarConfig{Items: testBarItems()})
	theme := DefaultTheme()
	theme.ShowZSPLogo = true
	picks := 0
	mb.logoRandom = func(n int) int { picks++; return picks % n }

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
		fills = fills[len(fills)-3*zspLogoTeeth:]
		for i := range q {
			q[i] = fills[3*i]
		}
		return
	}
	var spectrum [zspLogoTeeth]Colour
	for i, k := range zspLogoSpectrum {
		spectrum[i] = zspLogoPalette[k]
	}

	const step = 0.25
	runEnd := float64(zspLogoRun) * zspLogoCycle // 15 s: the Spectrum's quartet takes over
	holdEnd := runEnd + zspLogoFixedHold         // 21 s
	fadeEnd := holdEnd + zspLogoFade             // 21.5 s
	second := fadeEnd + runEnd                   // 36.5 s: the next run's Spectrum quartet
	drawn()
	var beforeFade [zspLogoTeeth]Colour
	for at := step; at <= second+zspLogoFixedHold+step; at += step {
		mb.Update(Input{DeltaTime: step})
		got := drawn()
		isSpectrum := got == spectrum
		switch {
		case at >= runEnd && at <= holdEnd, at >= second && at <= second+zspLogoFixedHold:
			if !isSpectrum {
				t.Errorf("%.2f s: %+v, want the Spectrum's quartet %+v", at, got, spectrum)
			}
		case at > holdEnd && at < fadeEnd:
			if isSpectrum {
				t.Errorf("%.2f s: still the Spectrum's quartet in its fade", at)
			}
		case at > second+zspLogoFixedHold:
			if isSpectrum {
				t.Errorf("%.2f s: the Spectrum's quartet is held for more than %v s", at, zspLogoFixedHold)
			}
		default:
			// A random quartet, or the fade into the Spectrum's: never wholly the Spectrum's
			// before its time, nor after its six seconds.
			if isSpectrum {
				t.Errorf("%.2f s: the Spectrum's quartet is shown outside its six seconds", at)
			}
		}
		if at == runEnd-step {
			beforeFade = got
		}
	}
	if beforeFade == spectrum {
		t.Error("the quartet before the Spectrum's is the Spectrum's")
	}
	// The picks: two at the start, one for each quartet entered after that except the
	// Spectrum's, which is not picked.
	if picks%zspLogoTeeth != 0 || picks == 0 {
		t.Errorf("%d picks is not a whole number of quartets", picks)
	}
}

// The Spectrum's quartet is held for the longer time: after its fade begins, 3 s on,
// and not before, the bar leaves it.
func TestZSPLogoSpectrumFadeBeginsOnlyAfterItsThreeSeconds(t *testing.T) {
	mb := NewMenuBar(MenuBarConfig{Items: testBarItems()})
	theme := DefaultTheme()
	theme.ShowZSPLogo = true
	mb.logoRandom = func(n int) int { return 1 % n }
	r := newDrawRecorder()
	mb.Draw(r, 800, 600, 0, 24, theme)
	// Go to the first frame of the Spectrum's quartet.
	for mb.logoSeq < zspLogoRun {
		mb.Update(Input{DeltaTime: 0.25})
		mb.Draw(r, 800, 600, 0, 24, theme)
	}
	if mb.logoTeeth != zspLogoSpectrum {
		t.Fatalf("quartet %d is %v, want the Spectrum's %v", mb.logoSeq, mb.logoTeeth, zspLogoSpectrum)
	}
	// A quarter second before its hold is over: held. A quarter second after: fading (the quartet is the same until the fade is over).
	mb.Update(Input{DeltaTime: zspLogoFixedHold - 0.25 - float32(mb.logoPhase)})
	mb.Draw(r, 800, 600, 0, 24, theme)
	if got := zspLogoShownHeld(mb.logoTeeth, mb.logoNext, zspLogoHoldOf(mb.logoSeq), mb.logoPhase); got[0] != zspLogoPalette[zspLogoSpectrum[0]] {
		t.Errorf("fading a quarter second before the end of the hold of the Spectrum's quartet")
	}
	mb.Update(Input{DeltaTime: 0.5})
	mb.Draw(r, 800, 600, 0, 24, theme)
	if got := zspLogoShownHeld(mb.logoTeeth, mb.logoNext, zspLogoHoldOf(mb.logoSeq), mb.logoPhase); got[0] == zspLogoPalette[zspLogoSpectrum[0]] && mb.logoSeq == zspLogoRun {
		t.Errorf("not fading a quarter second after the hold of the Spectrum's quartet: %+v", got)
	}
	// And the random ones begin again after it.
	mb.Update(Input{DeltaTime: 0.5})
	mb.Draw(r, 800, 600, 0, 24, theme)
	if zspLogoFixedAt(mb.logoSeq) {
		t.Errorf("still on the Spectrum's quartet, number %d, long after its hold", mb.logoSeq)
	}
}
