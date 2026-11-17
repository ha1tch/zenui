package zenui

// CellGrid is a grid of cells that the host paints, one by one, and that reports
// the pointer's presses and drags on cells the way PixelCanvas does, with the same
// events. A room of tiles is one, and so is the map of rooms. Its geometry is a
// pure function of the area.
type CellGrid struct {
	Cols, Rows int
	// AspectW and AspectH are the shape of a cell, in whole numbers: zero for
	// both is a square. A cell is as many pixels as the largest multiple of that
	// shape that fits.
	AspectW, AspectH int
	// Paint draws the contents of the cell at (x, y) inside cell. It should stay
	// inside it. With none, a cell is plain.
	Paint func(r Renderer, th Theme, cell Rect, x, y int)
	// Lines draws lines between the cells.
	Lines bool
	// SelX and SelY are the cell that is outlined as selected; negative for none.
	SelX, SelY int

	held, right    bool
	lastX, lastY   int
	hoverX, hoverY int
}

// NewCellGrid returns a grid of cols by rows square cells with the lines
// showing and nothing selected.
func NewCellGrid(cols, rows int) *CellGrid {
	return &CellGrid{Cols: cols, Rows: rows, Lines: true, SelX: -1, SelY: -1, hoverX: -1, hoverY: -1}
}

func (g *CellGrid) aspect() (int, int) { return max(1, g.AspectW), max(1, g.AspectH) }

// Geometry returns the rectangle the whole grid occupies inside area, centred,
// and the width and height of one cell.
func (g *CellGrid) Geometry(area Rect) (grid Rect, cellW, cellH int) {
	aw, ah := g.aspect()
	k := max(1, min(area.W/(max(1, g.Cols)*aw), area.H/(max(1, g.Rows)*ah)))
	cellW, cellH = k*aw, k*ah
	grid = Rect{W: cellW * g.Cols, H: cellH * g.Rows}
	grid.X = area.X + (area.W-grid.W)/2
	grid.Y = area.Y + (area.H-grid.H)/2
	return grid, cellW, cellH
}

// CellRect is the rectangle of the cell at (x, y) inside area.
func (g *CellGrid) CellRect(area Rect, x, y int) Rect {
	grid, cw, ch := g.Geometry(area)
	return Rect{X: grid.X + x*cw, Y: grid.Y + y*ch, W: cw, H: ch}
}

// CellAt is the cell under a point inside area, and whether there is one.
func (g *CellGrid) CellAt(area Rect, px, py int) (int, int, bool) {
	grid, cw, ch := g.Geometry(area)
	if !grid.Contains(px, py) {
		return 0, 0, false
	}
	return (px - grid.X) / cw, (py - grid.Y) / ch, true
}

// Hover is the cell the pointer was over at the last Update.
func (g *CellGrid) Hover() (int, int, bool) { return g.hoverX, g.hoverY, g.hoverX >= 0 }

// Update turns a frame's input into events: a press on a cell, the pointer
// entering another cell while the button is held, and the button being let go,
// reported on the last cell the pointer was over. Either button can drag, and
// Right says which.
func (g *CellGrid) Update(in Input, area Rect) []CanvasEvent {
	var out []CanvasEvent
	x, y, over := g.CellAt(area, in.MouseX, in.MouseY)
	if over {
		g.hoverX, g.hoverY = x, y
	} else {
		g.hoverX, g.hoverY = -1, -1
	}
	if !g.held {
		if over && (in.MousePressed || in.MouseRightPressed) {
			g.held, g.right = true, !in.MousePressed
			g.lastX, g.lastY = x, y
			out = append(out, CanvasEvent{Kind: CanvasDown, X: x, Y: y, Right: g.right})
		}
		return out
	}
	down := in.MouseDown
	if g.right {
		down = in.MouseRightDown
	}
	if !down {
		g.held = false
		return append(out, CanvasEvent{Kind: CanvasUp, X: g.lastX, Y: g.lastY, Right: g.right})
	}
	if over && (x != g.lastX || y != g.lastY) {
		g.lastX, g.lastY = x, y
		out = append(out, CanvasEvent{Kind: CanvasMove, X: x, Y: y, Right: g.right})
	}
	return out
}

// Draw paints the grid inside area: each cell, the lines between them, the
// selected cell's outline and an outline round the cell under the pointer.
func (g *CellGrid) Draw(r Renderer, th Theme, area Rect) {
	grid, cw, ch := g.Geometry(area)
	for y := 0; y < g.Rows; y++ {
		for x := 0; x < g.Cols; x++ {
			cell := Rect{X: grid.X + x*cw, Y: grid.Y + y*ch, W: cw, H: ch}
			if g.Paint != nil {
				g.Paint(r, th, cell, x, y)
			} else {
				r.FillRect(cell, th.Panel)
			}
		}
	}
	if g.Lines && cw >= 4 && ch >= 4 {
		for x := 0; x <= g.Cols; x++ {
			r.FillRect(Rect{X: grid.X + x*cw, Y: grid.Y, W: 1, H: grid.H}, th.Border)
		}
		for y := 0; y <= g.Rows; y++ {
			r.FillRect(Rect{X: grid.X, Y: grid.Y + y*ch, W: grid.W, H: 1}, th.Border)
		}
	}
	if g.hoverX >= 0 && (g.hoverX != g.SelX || g.hoverY != g.SelY) {
		r.StrokeRect(g.CellRect(area, g.hoverX, g.hoverY), th.Text, 1)
	}
	if g.SelX >= 0 && g.SelY >= 0 && g.SelX < g.Cols && g.SelY < g.Rows {
		r.StrokeRect(g.CellRect(area, g.SelX, g.SelY), th.SelFill, 2)
	}
}
