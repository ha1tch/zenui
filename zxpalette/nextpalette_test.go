package zxpalette

import "testing"

// TestRGB9ColourLevels: three bits become eight by repeating them, the
// levels CSpect shows.
func TestRGB9ColourLevels(t *testing.T) {
	want := [8]uint8{0, 36, 73, 109, 146, 182, 219, 255}
	for v := uint8(0); v < 8; v++ {
		c := RGB9{R: v, G: 7 - v, B: v}.Colour()
		if c.R != want[v] || c.G != want[7-v] || c.B != want[v] || c.A != 0xFF {
			t.Errorf("level %d: %+v", v, c)
		}
	}
}
