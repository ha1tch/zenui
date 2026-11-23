package zenui

import (
	"math"
	"math/rand/v2"

	"github.com/ha1tch/zenui/zxpalette"
)

// The logo's colours. Each of its four teeth (see drawZSPLogo) is one colour of
// zspLogoPalette, picked at random. A quartet holds still for zspLogoHold, and
// the four teeth then fade together, over zspLogoFade, to the next quartet, in
// which each has another colour than it had; that holds in its turn. After
// every zspLogoRun quartets picked at random the next is not picked: it is the
// Spectrum's own (zspLogoSpectrum), which holds for the longer zspLogoFixedHold
// before the random ones begin again.
//
// The palette is the Spectrum's eight colours at the normal level (the 0xC8 of
// zxpalette.RGBA), lightened almost to a pastel and a little desaturated. Black,
// which lightening leaves a dark colour, is 70% white instead.
const (
	zspLogoLighten    = 0.5 // how far each colour is taken towards white
	zspLogoDesaturate = 0.1 // how far it is then taken towards its own grey
)

// zspLogoBlack is what black is in the palette: 70% white, a light grey.
var zspLogoBlack = Colour{R: 0xb3, G: 0xb3, B: 0xb3, A: 0xff}

// zspLogoPalette is the colours a tooth can be, in the Spectrum's order (black,
// blue, red, magenta, green, cyan, yellow, white).
var zspLogoPalette = func() [8]Colour {
	var p [8]Colour
	for i := range p {
		c := zxpalette.RGBA[i]
		p[i] = softened(c.R, c.G, c.B)
	}
	p[zxpalette.Black] = zspLogoBlack
	return p
}()

// softened is the colour taken zspLogoLighten of the way to white, and then
// zspLogoDesaturate of the way to its own luminance, the grey of the same
// brightness (Rec. 601 weights).
func softened(r, g, b uint8) Colour {
	towards := func(from, to, t float64) float64 { return from + (to-from)*t }
	fr, fg, fb := float64(r), float64(g), float64(b)
	fr, fg, fb = towards(fr, 255, zspLogoLighten), towards(fg, 255, zspLogoLighten), towards(fb, 255, zspLogoLighten)
	y := 0.299*fr + 0.587*fg + 0.114*fb
	fr, fg, fb = towards(fr, y, zspLogoDesaturate), towards(fg, y, zspLogoDesaturate), towards(fb, y, zspLogoDesaturate)
	return Colour{R: uint8(math.Round(fr)), G: uint8(math.Round(fg)), B: uint8(math.Round(fb)), A: 0xff}
}

// How the logo's colours move, in seconds: a quartet holds still for
// zspLogoHold, then fades to the next over zspLogoFade, and the two together are
// one cycle. An earlier rotation of a third of a second, with no fade, read as
// too fast once actually seen running.
const (
	zspLogoHold  = 1.0
	zspLogoFade  = 0.5
	zspLogoCycle = zspLogoHold + zspLogoFade
)

// The Spectrum's quartet. Every zspLogoRun quartets of random colours the logo
// shows the four bands of the Spectrum's own rainbow instead, in their order
// (red, yellow, green, cyan), one for each tooth from the left and each in the
// palette's pastel colour as it is. It holds for zspLogoFixedHold seconds, then
// fades to the next random quartet as any other does, and another run begins.
const (
	zspLogoRun       = 10
	zspLogoFixedHold = 3.0
)

var zspLogoSpectrum = [zspLogoTeeth]int{zxpalette.Red, zxpalette.Yellow, zxpalette.Green, zxpalette.Cyan}

// zspLogoFixedAt reports whether the quartet that is number seq of the logo's
// run of them (the first is 0) is the Spectrum's.
func zspLogoFixedAt(seq int) bool { return seq%(zspLogoRun+1) == zspLogoRun }

// zspLogoHoldOf is how long the quartet number seq holds still.
func zspLogoHoldOf(seq int) float64 {
	if zspLogoFixedAt(seq) {
		return zspLogoFixedHold
	}
	return zspLogoHold
}

// zspLogoGeometry computes where the logo should draw and its block
// size, given the X position just past the bar's last label and the
// screen width. Pure function, no drawing -- mirrors rainbowGeometry's
// own shape. Unlike the rainbow's stripes, the logo's block size isn't
// flexible (it has to stay a clean 1:2 width:height ratio to read as the
// reference shape, not stretch to fill space) -- it's derived once from
// the bar's own height and only checked for whether it fits, not
// resized to available width.
func zspLogoGeometry(afterLabelsX, screenW, height int) (x0, blockW, blockH int, ok bool) {
	const rightMargin = 8

	// Three stacked levels fill the bar's height exactly; block width
	// is half block height, matching the reference logo's own 45x90
	// block proportion.
	blockH = height / 3
	if blockH < 2 {
		return 0, 0, 0, false // too short for the blocks to read as anything
	}
	blockW = blockH / 2
	if blockW < 1 {
		blockW = 1
	}
	// 3 blocks/tooth * 2 teeth/half * 2 halves = 12 block-widths total.
	totalW := blockW * 12

	available := screenW - rightMargin - afterLabelsX
	if totalW > available {
		return 0, 0, 0, false
	}
	x0 = screenW - rightMargin - totalW
	return x0, blockW, blockH, true
}

// zspLogoTeeth is the number of teeth, the logo's elements: two in each of its
// two halves.
const zspLogoTeeth = 4

// nextZSPLogoTeeth picks the colours of the next quartet, one for each tooth
// and each a random colour of the palette (two teeth can have the same one),
// other than the one that tooth has in prev, so that every tooth changes at
// each step. A tooth that has none yet, which prev marks with -1, may have any.
// pick(n) returns a random integer from 0 to n-1.
func nextZSPLogoTeeth(prev [zspLogoTeeth]int, pick func(n int) int) [zspLogoTeeth]int {
	return nextZSPLogoTeethBefore(prev, [zspLogoTeeth]int{-1, -1, -1, -1}, pick)
}

// nextZSPLogoTeethBefore is nextZSPLogoTeeth for a quartet that is itself
// followed by after: each tooth's colour is also other than the one it has in
// after (-1 marks none), so that no tooth keeps its colour into that quartet
// either. It calls pick once for each tooth.
func nextZSPLogoTeethBefore(prev, after [zspLogoTeeth]int, pick func(n int) int) [zspLogoTeeth]int {
	var next [zspLogoTeeth]int
	for i := range next {
		skip := func(c int) bool { return c == prev[i] || c == after[i] }
		free := 0
		for c := range zspLogoPalette {
			if !skip(c) {
				free++
			}
		}
		k := pick(free)
		for c := range zspLogoPalette {
			if skip(c) {
				continue
			}
			if k == 0 {
				next[i] = c
				break
			}
			k--
		}
	}
	return next
}

// zspLogoQuartetAfter is the quartet that follows quartet number seq, which is
// cur: the Spectrum's when that is its turn, and otherwise a random one that,
// when the Spectrum's is to come after it, also keeps off the Spectrum's colours.
func zspLogoQuartetAfter(seq int, cur [zspLogoTeeth]int, pick func(n int) int) [zspLogoTeeth]int {
	switch {
	case zspLogoFixedAt(seq + 1):
		return zspLogoSpectrum
	case zspLogoFixedAt(seq + 2):
		return nextZSPLogoTeethBefore(cur, zspLogoSpectrum, pick)
	}
	return nextZSPLogoTeeth(cur, pick)
}

// zspLogoRandom is the bar's source of chance for the logo: a random integer
// from 0 to n-1. A test replaces it on its own bar.
func zspLogoRandom(n int) int { return rand.IntN(n) }

// blendColour is the colour t of the way from a to b, for t from 0 to 1.
func blendColour(a, b Colour, t float64) Colour {
	mix := func(x, y uint8) uint8 { return uint8(math.Round(float64(x) + (float64(y)-float64(x))*t)) }
	return Colour{R: mix(a.R, b.R), G: mix(a.G, b.G), B: mix(a.B, b.B), A: 0xff}
}

// zspLogoShown is the colour of each tooth phase seconds into a cycle that began
// with the quartet cur and ends with next (indices into zspLogoPalette): cur
// until zspLogoHold has passed, then a straight blend towards next, which it
// is at zspLogoCycle.
func zspLogoShown(cur, next [zspLogoTeeth]int, phase float64) [zspLogoTeeth]Colour {
	return zspLogoShownHeld(cur, next, zspLogoHold, phase)
}

// zspLogoShownHeld is zspLogoShown for a quartet that holds for hold seconds
// before it fades.
func zspLogoShownHeld(cur, next [zspLogoTeeth]int, hold, phase float64) [zspLogoTeeth]Colour {
	t := (phase - hold) / zspLogoFade
	t = math.Max(0, math.Min(1, t))
	var out [zspLogoTeeth]Colour
	for i := range out {
		out[i] = blendColour(zspLogoPalette[uint(cur[i])%uint(len(zspLogoPalette))],
			zspLogoPalette[uint(next[i])%uint(len(zspLogoPalette))], t)
	}
	return out
}

// drawZSPLogo draws the logo: two halves, each of two ascending three-block
// "teeth" (bottom-left to top-right, matching the measured reference exactly --
// x increases and y decreases together as the block level rises). teeth is the
// colour of each tooth, from the left. Entirely through Renderer (FillRect), so
// it works identically under any host, not just raylib.
func drawZSPLogo(r Renderer, x0, barY, blockW, blockH int, teeth [zspLogoTeeth]Colour) {
	for tooth := 0; tooth < zspLogoTeeth; tooth++ {
		for level := 0; level < 3; level++ {
			x := x0 + tooth*3*blockW + level*blockW
			y := barY + (2-level)*blockH
			r.FillRect(Rect{X: x, Y: y, W: blockW, H: blockH}, teeth[tooth])
		}
	}
}
