package zenui

import (
	"reflect"
	"testing"
)

func TestACellGridIsCentredAndAsLargeAsTheAreaAllows(t *testing.T) {
	g := NewCellGrid(4, 3)
	g.AspectW, g.AspectH = 8, 5 // a map square, the shape of a room
	grid, cw, ch := g.Geometry(Rect{X: 10, Y: 20, W: 400, H: 300})
	// 400/(4*8) = 12 and 300/(3*5) = 20, so 12 is the multiple that fits
	if cw != 96 || ch != 60 || grid != (Rect{X: 18, Y: 80, W: 384, H: 180}) {
		t.Errorf("cells %dx%d in %+v, want 96x60 in (18,80) 384x180", cw, ch, grid)
	}
	sq := NewCellGrid(16, 10)
	grid, cw, ch = sq.Geometry(Rect{X: 0, Y: 0, W: 380, H: 500})
	if cw != 23 || ch != 23 || grid != (Rect{X: 6, Y: 135, W: 368, H: 230}) {
		t.Errorf("square cells %dx%d in %+v, want 23x23 in (6,135) 368x230", cw, ch, grid)
	}
	tiny := NewCellGrid(100, 100)
	if _, cw, ch := tiny.Geometry(Rect{W: 40, H: 40}); cw != 1 || ch != 1 {
		t.Errorf("a grid too big for its area has cells %dx%d, want 1x1 at least", cw, ch)
	}
}

func TestCellRectAndCellAtAreInverse(t *testing.T) {
	g := NewCellGrid(5, 4)
	g.AspectW, g.AspectH = 3, 2
	area := Rect{X: 7, Y: 9, W: 200, H: 150}
	for y := 0; y < 4; y++ {
		for x := 0; x < 5; x++ {
			cx, cy := mid(g.CellRect(area, x, y))
			if gx, gy, ok := g.CellAt(area, cx, cy); !ok || gx != x || gy != y {
				t.Errorf("the centre of cell (%d,%d) is over (%d,%d) %v", x, y, gx, gy, ok)
			}
			r := g.CellRect(area, x, y)
			if gx, gy, ok := g.CellAt(area, r.X, r.Y); !ok || gx != x || gy != y {
				t.Errorf("the top left of cell (%d,%d) is over (%d,%d)", x, y, gx, gy)
			}
			if gx, gy, ok := g.CellAt(area, r.X+r.W-1, r.Y+r.H-1); !ok || gx != x || gy != y {
				t.Errorf("the bottom right of cell (%d,%d) is over (%d,%d)", x, y, gx, gy)
			}
		}
	}
	grid, _, _ := g.Geometry(area)
	for _, p := range [][2]int{{grid.X - 1, grid.Y}, {grid.X, grid.Y - 1}, {grid.X + grid.W, grid.Y}, {grid.X, grid.Y + grid.H}} {
		if _, _, ok := g.CellAt(area, p[0], p[1]); ok {
			t.Errorf("the point %v is outside the grid but found a cell", p)
		}
	}
}

// the grid is 40 by 30 at the origin: cells of 10 by 10.
var cgArea = Rect{X: 0, Y: 0, W: 40, H: 30}

func TestACellGridReportsPressDragAndRelease(t *testing.T) {
	g := NewCellGrid(4, 3)
	press := func(x, y int, down bool) []CanvasEvent {
		return g.Update(Input{MouseX: x, MouseY: y, MousePressed: down, MouseDown: down}, cgArea)
	}
	if ev := g.Update(Input{MouseX: 15, MouseY: 5}, cgArea); len(ev) != 0 {
		t.Errorf("the pointer resting on a cell gave %v", ev)
	}
	if ev := press(15, 5, true); !reflect.DeepEqual(ev, []CanvasEvent{{Kind: CanvasDown, X: 1, Y: 0}}) {
		t.Errorf("a press gave %v, want a down on (1,0)", ev)
	}
	hold := func(x, y int) []CanvasEvent { return g.Update(Input{MouseX: x, MouseY: y, MouseDown: true}, cgArea) }
	if ev := hold(17, 6); len(ev) != 0 {
		t.Errorf("moving inside the same cell gave %v", ev)
	}
	if ev := hold(25, 5); !reflect.DeepEqual(ev, []CanvasEvent{{Kind: CanvasMove, X: 2, Y: 0}}) {
		t.Errorf("entering another cell gave %v, want a move to (2,0)", ev)
	}
	if ev := hold(500, 500); len(ev) != 0 {
		t.Errorf("dragging outside the grid gave %v", ev)
	}
	if ev := g.Update(Input{MouseX: 500, MouseY: 500}, cgArea); !reflect.DeepEqual(ev, []CanvasEvent{{Kind: CanvasUp, X: 2, Y: 0}}) {
		t.Errorf("letting go outside gave %v, want an up on the last cell, (2,0)", ev)
	}

	right := func(x, y int, pressed, down bool) []CanvasEvent {
		return g.Update(Input{MouseX: x, MouseY: y, MouseRightPressed: pressed, MouseRightDown: down}, cgArea)
	}
	if ev := right(5, 25, true, true); !reflect.DeepEqual(ev, []CanvasEvent{{Kind: CanvasDown, X: 0, Y: 2, Right: true}}) {
		t.Errorf("a right press gave %v", ev)
	}
	if ev := right(15, 25, false, true); !reflect.DeepEqual(ev, []CanvasEvent{{Kind: CanvasMove, X: 1, Y: 2, Right: true}}) {
		t.Errorf("a right drag gave %v", ev)
	}
	if ev := g.Update(Input{MouseX: 15, MouseY: 25, MouseDown: true, MouseRightDown: true}, cgArea); len(ev) != 0 {
		t.Errorf("the left button held during a right drag gave %v", ev)
	}
	if ev := right(15, 25, false, false); !reflect.DeepEqual(ev, []CanvasEvent{{Kind: CanvasUp, X: 1, Y: 2, Right: true}}) {
		t.Errorf("letting go of the right button gave %v", ev)
	}
	if ev := press(500, 5, true); len(ev) != 0 {
		t.Errorf("a press outside the grid gave %v", ev)
	}
}

func TestACellGridKnowsWhichCellThePointerIsOver(t *testing.T) {
	g := NewCellGrid(4, 3)
	if _, _, ok := g.Hover(); ok {
		t.Error("a new grid has a hover")
	}
	g.Update(Input{MouseX: 35, MouseY: 25}, cgArea)
	if x, y, ok := g.Hover(); !ok || x != 3 || y != 2 {
		t.Errorf("hover = %d,%d %v, want 3,2", x, y, ok)
	}
	g.Update(Input{MouseX: 400, MouseY: 25}, cgArea)
	if _, _, ok := g.Hover(); ok {
		t.Error("the hover stayed after the pointer left the grid")
	}
}

func TestACellGridPaintsEveryCellAndMarksTheSelectedAndHoveredOnes(t *testing.T) {
	th := DarkTheme()
	g := NewCellGrid(4, 3)
	painted := map[[2]int]Rect{}
	g.Paint = func(r Renderer, _ Theme, cell Rect, x, y int) { painted[[2]int{x, y}] = cell }
	g.SelX, g.SelY = 1, 1
	g.Update(Input{MouseX: 35, MouseY: 25}, cgArea) // hovering (3,2)
	cv := &listCanvas{}
	g.Draw(cv, th, cgArea)
	if len(painted) != 12 {
		t.Fatalf("%d cells painted, want 12", len(painted))
	}
	for x := 0; x < 4; x++ {
		for y := 0; y < 3; y++ {
			if want := (Rect{X: x * 10, Y: y * 10, W: 10, H: 10}); painted[[2]int{x, y}] != want {
				t.Errorf("cell (%d,%d) was painted in %+v, want %+v", x, y, painted[[2]int{x, y}], want)
			}
		}
	}
	lines := 0
	for _, f := range cv.fills {
		if f.C == th.Border {
			lines++
		}
	}
	if lines != 5+4 {
		t.Errorf("%d lines drawn, want 5 vertical and 4 horizontal", lines)
	}
	var sel, hov bool
	for _, s := range cv.strokes {
		if s.C == th.SelFill && s.T == 2 && s.R == (Rect{X: 10, Y: 10, W: 10, H: 10}) {
			sel = true
		}
		if s.C == th.Text && s.T == 1 && s.R == (Rect{X: 30, Y: 20, W: 10, H: 10}) {
			hov = true
		}
	}
	if !sel || !hov {
		t.Errorf("selected outlined %v, hovered outlined %v: want both", sel, hov)
	}
	g.Lines = false
	cv = &listCanvas{}
	g.Draw(cv, th, cgArea)
	for _, f := range cv.fills {
		if f.C == th.Border {
			t.Error("lines were drawn with Lines off")
		}
	}
	plain := NewCellGrid(2, 2)
	cv = &listCanvas{}
	plain.Draw(cv, th, cgArea) // no Paint: the cells are plain
	filled := 0
	for _, f := range cv.fills {
		if f.C == th.Panel {
			filled++
		}
	}
	if filled != 4 {
		t.Errorf("a grid with no Paint filled %d cells, want 4", filled)
	}
}
