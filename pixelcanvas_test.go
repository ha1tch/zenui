package zenui

import "testing"

// canvasArea gives a cell size of 10 for an 8 by 8 picture, with the picture at
// the area's corner.
var canvasArea = Rect{X: 0, Y: 0, W: 80, H: 80}

func TestPixelCanvasGeometryIsAWholeNumberOfPixelsAndCentred(t *testing.T) {
	c := NewPixelCanvas(8, 8)
	pic, cell := c.Geometry(Rect{X: 0, Y: 0, W: 100, H: 60})
	if cell != 7 || pic != (Rect{X: 22, Y: 2, W: 56, H: 56}) {
		t.Errorf("cell %d, picture %+v, want 7 and a 56 square at 22,2", cell, pic)
	}
	if _, cell := c.Geometry(Rect{W: 5, H: 5}); cell != 1 {
		t.Errorf("in a tiny area the cell is %d, want 1", cell)
	}
}

func TestPixelCanvasTurnsAPressAndADragIntoEvents(t *testing.T) {
	c := NewPixelCanvas(8, 8)
	var all []CanvasEvent
	step := func(in Input) { all = append(all, c.Update(in, canvasArea)...) }

	step(Input{MouseX: 25, MouseY: 35, MousePressed: true, MouseDown: true})
	step(Input{MouseX: 27, MouseY: 36, MouseDown: true}) // still cell (2,3)
	step(Input{MouseX: 45, MouseY: 35, MouseDown: true})
	step(Input{MouseX: 47, MouseY: 38, MouseDown: true}) // still cell (4,3)
	step(Input{MouseX: 45, MouseY: 35})                  // let go
	step(Input{MouseX: 5, MouseY: 5})                    // idle
	want := []CanvasEvent{
		{Kind: CanvasDown, X: 2, Y: 3},
		{Kind: CanvasMove, X: 4, Y: 3},
		{Kind: CanvasUp, X: 4, Y: 3},
	}
	if len(all) != len(want) {
		t.Fatalf("events = %+v, want %+v", all, want)
	}
	for i := range want {
		if all[i] != want[i] {
			t.Errorf("event %d = %+v, want %+v", i, all[i], want[i])
		}
	}
}

func TestPixelCanvasRightButtonStrokes(t *testing.T) {
	c := NewPixelCanvas(8, 8)
	var all []CanvasEvent
	step := func(in Input) { all = append(all, c.Update(in, canvasArea)...) }
	step(Input{MouseX: 5, MouseY: 5, MouseRightPressed: true, MouseRightDown: true})
	step(Input{MouseX: 15, MouseY: 5, MouseRightDown: true})
	step(Input{MouseX: 25, MouseY: 5, MouseDown: true}) // only the left is down: the right stroke ended
	want := []CanvasEvent{
		{Kind: CanvasDown, X: 0, Y: 0, Right: true},
		{Kind: CanvasMove, X: 1, Y: 0, Right: true},
		{Kind: CanvasUp, X: 1, Y: 0, Right: true},
	}
	if len(all) != len(want) {
		t.Fatalf("events = %+v, want %+v", all, want)
	}
	for i := range want {
		if all[i] != want[i] {
			t.Errorf("event %d = %+v, want %+v", i, all[i], want[i])
		}
	}
}

func TestPixelCanvasIgnoresWhatHappensOutsideThePicture(t *testing.T) {
	c := NewPixelCanvas(8, 8)
	if ev := c.Update(Input{MouseX: 200, MouseY: 20, MousePressed: true, MouseDown: true}, canvasArea); len(ev) != 0 {
		t.Errorf("a press outside the picture gave %+v", ev)
	}
	c.Update(Input{MouseX: 5, MouseY: 5, MousePressed: true, MouseDown: true}, canvasArea)
	if ev := c.Update(Input{MouseX: 200, MouseY: 20, MouseDown: true}, canvasArea); len(ev) != 0 {
		t.Errorf("dragging outside the picture gave %+v, want nothing", ev)
	}
	ev := c.Update(Input{MouseX: 200, MouseY: 20}, canvasArea)
	if len(ev) != 1 || ev[0] != (CanvasEvent{Kind: CanvasUp, X: 0, Y: 0}) {
		t.Errorf("letting go outside gave %+v, want an Up at the last cell", ev)
	}
}

func TestPixelCanvasDrawsRunsOfInkAndTheGrid(t *testing.T) {
	th := DarkTheme()
	c := NewPixelCanvas(8, 8)
	c.Ink, c.Paper = Colour{R: 1, A: 255}, Colour{G: 1, A: 255}
	c.Pixel = func(x, y int) bool { return y == 0 && x >= 2 && x <= 4 || y == 3 && (x == 0 || x == 7) }
	cv := &listCanvas{}
	c.Draw(cv, th, canvasArea)
	var ink, paper, lines []listFill
	for _, f := range cv.fills {
		switch {
		case f.C == c.Ink:
			ink = append(ink, f)
		case f.C == c.Paper:
			paper = append(paper, f)
		case f.C == th.Border:
			lines = append(lines, f)
		}
	}
	if len(paper) != 1 || paper[0].R != canvasArea {
		t.Errorf("paper fills = %+v, want one over the picture", paper)
	}
	if len(ink) != 3 || ink[0].R != (Rect{X: 20, Y: 0, W: 30, H: 10}) {
		t.Errorf("ink fills = %+v, want a run of three cells and two single cells", ink)
	}
	if len(lines) != 7+7 {
		t.Errorf("%d grid lines, want 14 for 8 by 8", len(lines))
	}
	// Cells too small for a grid draw none.
	c.Grid = false
	cv = &listCanvas{}
	c.Draw(cv, th, canvasArea)
	for _, f := range cv.fills {
		if f.C == th.Border {
			t.Error("the grid was drawn although it is switched off")
		}
	}
	small := NewPixelCanvas(8, 8)
	small.Pixel = func(int, int) bool { return false }
	cv = &listCanvas{}
	small.Draw(cv, th, Rect{W: 24, H: 24}) // cells of 3
	for _, f := range cv.fills {
		if f.C == th.Border {
			t.Error("a grid was drawn in cells smaller than four pixels")
		}
	}
}
