package zenui_test

import (
	"fmt"

	"github.com/ha1tch/zenui"
)

// countingRenderer is a Renderer that draws nothing and counts the texts
// it is asked for: what a headless test of a widget needs.
type countingRenderer struct{ texts int }

func (*countingRenderer) FillRect(zenui.Rect, zenui.Colour)                                {}
func (*countingRenderer) StrokeRect(zenui.Rect, zenui.Colour, int)                         {}
func (*countingRenderer) FillRectGradientV(zenui.Rect, zenui.Colour, zenui.Colour)         {}
func (*countingRenderer) FillRectGradientHMultiply(zenui.Rect, zenui.Colour, zenui.Colour) {}
func (*countingRenderer) DrawLine(int, int, int, int, int, zenui.Colour)                   {}
func (r *countingRenderer) DrawText(string, int, int, int, zenui.Colour)                   { r.texts++ }
func (*countingRenderer) TextWidth(s string, scale int) int                                { return len(s) * 8 * scale }
func (*countingRenderer) LineHeight(scale int) int                                         { return 8 * scale }
func (*countingRenderer) Clip(zenui.Rect)                                                  {}
func (*countingRenderer) ClipEnd()                                                         {}

// A widget is given a Renderer to draw with and an Input to react to;
// here a TextModal is drawn once and then closed with Escape.
func ExampleTextModal() {
	r := &countingRenderer{}
	m := zenui.NewTextModal("ABOUT", "# zenui\n\nA widget toolkit.", true)
	m.Draw(r, 800, 600, zenui.DarkTheme())
	fmt.Println("texts drawn:", r.texts)
	fmt.Println("open after a frame with no input:", m.Update(zenui.Input{}))
	fmt.Println("open after Escape:", m.Update(zenui.Input{Keys: []zenui.Key{zenui.KeyEscape}}))
	// Output:
	// texts drawn: 6
	// open after a frame with no input: true
	// open after Escape: false
}
