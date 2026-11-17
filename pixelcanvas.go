package zenui

// CanvasEventKind says what the pointer did on a PixelCanvas.
type CanvasEventKind int

// The events a PixelCanvas reports.
const (
	CanvasDown CanvasEventKind = iota + 1 // a button went down on a cell
	CanvasMove                            // the pointer entered another cell with the button held
	CanvasUp                              // the button was let go
)

// CanvasEvent is one thing the pointer did. X and Y are the cell; Right says
// which button.
type CanvasEvent struct {
	Kind  CanvasEventKind
	X, Y  int
	Right bool
}

// PixelCanvas shows a picture of W by H pixels, as large a whole number of
// pixels per cell as fits, and turns the pointer into events. It edits nothing:
// the host reads the picture through Pixel and changes it in answer to events.
type PixelCanvas struct {
	W, H       int
	Pixel      func(x, y int) bool
	Ink, Paper Colour
	Grid       bool // lines between the cells, drawn when the cells are large enough

	held, right  bool
	lastX, lastY int
}

// NewPixelCanvas returns a canvas for a picture of w by h pixels, with the grid
// showing.
func NewPixelCanvas(w, h int) *PixelCanvas { return &PixelCanvas{W: w, H: h, Grid: true} }

// Geometry returns the rectangle the picture occupies inside area, centred, and
// the size of one cell: the largest whole number of pixels that fits, at least 1.
func (c *PixelCanvas) Geometry(area Rect) (pic Rect, cell int) {
	cell = max(1, min(area.W/max(1, c.W), area.H/max(1, c.H)))
	pic = Rect{W: cell * c.W, H: cell * c.H}
	pic.X = area.X + (area.W-pic.W)/2
	pic.Y = area.Y + (area.H-pic.H)/2
	return pic, cell
}

func (c *PixelCanvas) cellAt(pic Rect, cell, px, py int) (x, y int, ok bool) {
	if !pic.Contains(px, py) {
		return 0, 0, false
	}
	return (px - pic.X) / cell, (py - pic.Y) / cell, true
}

// Update turns one frame of input into events. A press on a cell is a Down; the
// pointer entering another cell while the button is held is a Move, once per
// cell; letting go is an Up at the last cell the pointer was on.
func (c *PixelCanvas) Update(in Input, area Rect) []CanvasEvent {
	pic, cell := c.Geometry(area)
	var out []CanvasEvent
	x, y, over := c.cellAt(pic, cell, in.MouseX, in.MouseY)
	if !c.held {
		if over && (in.MousePressed || in.MouseRightPressed) {
			c.held, c.right = true, !in.MousePressed
			c.lastX, c.lastY = x, y
			out = append(out, CanvasEvent{Kind: CanvasDown, X: x, Y: y, Right: c.right})
		}
		return out
	}
	down := in.MouseDown
	if c.right {
		down = in.MouseRightDown
	}
	if !down {
		c.held = false
		return append(out, CanvasEvent{Kind: CanvasUp, X: c.lastX, Y: c.lastY, Right: c.right})
	}
	if over && (x != c.lastX || y != c.lastY) {
		c.lastX, c.lastY = x, y
		out = append(out, CanvasEvent{Kind: CanvasMove, X: x, Y: y, Right: c.right})
	}
	return out
}

// Draw paints the picture inside area: the paper, the runs of ink, the grid, and
// an outline.
func (c *PixelCanvas) Draw(r Renderer, th Theme, area Rect) {
	pic, cell := c.Geometry(area)
	r.FillRect(pic, c.Paper)
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W; {
			if c.Pixel == nil || !c.Pixel(x, y) {
				x++
				continue
			}
			start := x
			for x < c.W && c.Pixel(x, y) {
				x++
			}
			r.FillRect(Rect{X: pic.X + start*cell, Y: pic.Y + y*cell, W: (x - start) * cell, H: cell}, c.Ink)
		}
	}
	if c.Grid && cell >= 4 {
		for x := 1; x < c.W; x++ {
			r.FillRect(Rect{X: pic.X + x*cell, Y: pic.Y, W: 1, H: pic.H}, th.Border)
		}
		for y := 1; y < c.H; y++ {
			r.FillRect(Rect{X: pic.X, Y: pic.Y + y*cell, W: pic.W, H: 1}, th.Border)
		}
	}
	r.StrokeRect(pic, th.Border, 1)
}
