package zenui

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"
)

type listFill struct {
	R Rect
	C Colour
}

type listText struct {
	S     string
	X, Y  int
	C     Colour
	Scale int
}

// listCanvas records what a List draws, with glyphs 8 pixels wide and tall at
// scale 1.
type listCanvas struct {
	fills   []listFill
	texts   []listText
	strokes []listStroke
}

// listStroke is an outline drawn: its rectangle, colour and thickness.
type listStroke struct {
	R Rect
	C Colour
	T int
}

func (c *listCanvas) FillRect(r Rect, col Colour) { c.fills = append(c.fills, listFill{r, col}) }
func (c *listCanvas) StrokeRect(r Rect, col Colour, t int) {
	c.strokes = append(c.strokes, listStroke{r, col, t})
}
func (c *listCanvas) FillRectGradientV(Rect, Colour, Colour)         {}
func (c *listCanvas) FillRectGradientHMultiply(Rect, Colour, Colour) {}
func (c *listCanvas) DrawLine(int, int, int, int, int, Colour)       {}
func (c *listCanvas) TextWidth(s string, scale int) int              { return len(s) * 8 * scale }
func (c *listCanvas) LineHeight(scale int) int                       { return 8 * scale }
func (c *listCanvas) Clip(Rect)                                      {}
func (c *listCanvas) ClipEnd()                                       {}
func (c *listCanvas) DrawText(s string, x, y, scale int, col Colour) {
	c.texts = append(c.texts, listText{s, x, y, col, scale})
}

const listLH = 8

var listArea = Rect{X: 0, Y: 0, W: 200, H: 300}

func listOf(n int) *List {
	l := NewList()
	for i := 0; i < n; i++ {
		l.Items = append(l.Items, ListItem{Name: fmt.Sprintf("item %d", i)})
	}
	return l
}

func mid(r Rect) (int, int) { return r.X + r.W/2, r.Y + r.H/2 }

func press(l *List, x, y int) (ListEvent, bool) {
	return l.Update(Input{MouseX: x, MouseY: y, MousePressed: true}, listArea, listLH)
}

func TestListLaysItemsOutInRowsAndInColumns(t *testing.T) {
	l := listOf(5)
	g := l.geom(listArea, listLH)
	if g.itemH != 20 || g.itemW != 200 || g.thumb != 16 {
		t.Fatalf("rows: item %dx%d thumb %d, want 200x20 and 16", g.itemW, g.itemH, g.thumb)
	}
	if got, want := l.cell(g, 3), (Rect{X: 0, Y: 16 + 3*20, W: 200, H: 20}); got != want {
		t.Errorf("row cell 3 = %+v, want %+v", got, want)
	}
	l.Columns = 3
	g = l.geom(listArea, listLH)
	if got, want := l.cell(g, 4), (Rect{X: 66, Y: 16 + 76, W: 66, H: 76}); got != want {
		t.Errorf("grid cell 4 = %+v, want %+v", got, want)
	}
	if got, want := l.cell(g, 2), (Rect{X: 132, Y: 16, W: 66, H: 76}); got != want {
		t.Errorf("grid cell 2 = %+v, want %+v", got, want)
	}
}

func TestClickingAnItemSelectsIt(t *testing.T) {
	l := listOf(5)
	if ev, ok := press(l, 50, 16+2*20+5); !ok || ev != (ListEvent{Kind: ListSelect, Index: 2}) || l.Selected != 2 {
		t.Errorf("click on item 2: %+v %v, selected %d", ev, ok, l.Selected)
	}
	if _, ok := press(l, 50, 16+5*20+5); ok {
		t.Error("a click below the last item selected something")
	}
	if _, ok := press(l, 100, 5); ok {
		t.Error("a click in the toolbar's gap selected something")
	}
	l.Columns = 3
	// Five items in three columns: row 1 holds items 3 and 4.
	if ev, ok := press(l, 100, 16+76+5); !ok || ev.Index != 4 {
		t.Errorf("click on the grid cell in row 1, second column: %+v %v, want item 4", ev, ok)
	}
	if _, ok := press(l, 170, 16+76+5); ok {
		t.Error("a click on the empty third cell of row 1 selected something")
	}
}

func TestToolbarEventsAndDisabledButtons(t *testing.T) {
	l := listOf(3)
	g := l.geom(listArea, listLH)
	ax, ay := mid(g.add)
	dx, dy := mid(g.dup)
	bx, by := mid(g.del)

	if ev, ok := press(l, ax, ay); !ok || ev != (ListEvent{Kind: ListAdd, Index: -1}) {
		t.Errorf("add: %+v %v", ev, ok)
	}
	if _, ok := press(l, dx, dy); ok {
		t.Error("duplicate answered with nothing selected")
	}
	if _, ok := press(l, bx, by); ok {
		t.Error("delete answered with nothing selected")
	}

	l.Selected = 1
	if ev, ok := press(l, dx, dy); !ok || ev != (ListEvent{Kind: ListDuplicate, Index: 1}) {
		t.Errorf("duplicate: %+v %v", ev, ok)
	}
	if ev, ok := press(l, bx, by); !ok || ev != (ListEvent{Kind: ListDelete, Index: 1}) {
		t.Errorf("delete: %+v %v", ev, ok)
	}
	l.AddDisabled, l.DuplicateDisabled, l.DeleteDisabled = true, true, true
	for name, p := range map[string][2]int{"add": {ax, ay}, "duplicate": {dx, dy}, "delete": {bx, by}} {
		if _, ok := press(l, p[0], p[1]); ok {
			t.Errorf("a disabled %s button answered", name)
		}
	}

	cx, cy := mid(g.cols)
	rx, ry := mid(g.rows)
	if ev, ok := press(l, cx, cy); !ok || ev.Kind != ListMode || l.Columns != listGridColumns {
		t.Errorf("columns: %+v %v, Columns %d", ev, ok, l.Columns)
	}
	if _, ok := press(l, cx, cy); ok {
		t.Error("switching to the mode already in force reported an event")
	}
	if ev, ok := press(l, rx, ry); !ok || ev.Kind != ListMode || l.Columns != 1 {
		t.Errorf("rows: %+v %v, Columns %d", ev, ok, l.Columns)
	}
}

func TestWheelAndTheFootButtonsScrollALongList(t *testing.T) {
	l := listOf(30)
	g := l.geom(listArea, listLH)
	if !g.scrollable || g.top.W == 0 || g.bottom.W == 0 {
		t.Fatalf("a list of 30 rows should scroll and show Top and Bottom: %+v", g)
	}
	if want := 30*20 - g.body.H; l.maxScroll(g) != want {
		t.Fatalf("max scroll = %d, want %d", l.maxScroll(g), want)
	}
	l.Update(Input{MouseX: 50, MouseY: 100, WheelY: -1}, listArea, listLH)
	if l.scroll != 20 {
		t.Errorf("one wheel step scrolled %d, want one row (20)", l.scroll)
	}
	l.Update(Input{MouseX: 250, MouseY: 100, WheelY: -3}, listArea, listLH)
	if l.scroll != 20 {
		t.Errorf("the wheel outside the list scrolled it to %d", l.scroll)
	}
	l.Update(Input{MouseX: 50, MouseY: 100, WheelY: 50}, listArea, listLH)
	if l.scroll != 0 {
		t.Errorf("scrolling up past the top left %d", l.scroll)
	}
	x, y := mid(g.bottom)
	press(l, x, y)
	if l.scroll != l.maxScroll(g) {
		t.Errorf("Bottom left the scroll at %d, want %d", l.scroll, l.maxScroll(g))
	}
	x, y = mid(g.top)
	press(l, x, y)
	if l.scroll != 0 {
		t.Errorf("Top left the scroll at %d", l.scroll)
	}

	short := listOf(5)
	gs := short.geom(listArea, listLH)
	short.Update(Input{MouseX: 50, MouseY: 100, WheelY: -5}, listArea, listLH)
	if gs.scrollable || gs.top.W != 0 || short.scroll != 0 {
		t.Errorf("a short list scrolled or showed foot buttons: %+v scroll %d", gs, short.scroll)
	}
}

func TestRevealKeepsTheSelectionInView(t *testing.T) {
	for _, cols := range []int{1, listGridColumns} {
		l := listOf(30)
		l.Columns = cols
		l.Selected = 29
		l.Reveal(listArea, listLH)
		g := l.geom(listArea, listLH)
		c := l.cell(g, 29)
		if c.Y < g.body.Y || c.Y+c.H > g.body.Y+g.body.H {
			t.Errorf("columns=%d: item 29 at %+v is outside the body %+v after Reveal", cols, c, g.body)
		}
		l.Selected = 0
		l.Reveal(listArea, listLH)
		if l.scroll != 0 {
			t.Errorf("columns=%d: revealing the first item left scroll %d", cols, l.scroll)
		}
	}
	// Switching mode keeps the selected item on show.
	l := listOf(30)
	l.Selected = 29
	l.Reveal(listArea, listLH)
	g := l.geom(listArea, listLH)
	x, y := mid(g.cols)
	press(l, x, y)
	g = l.geom(listArea, listLH)
	if c := l.cell(g, 29); c.Y < g.body.Y || c.Y+c.H > g.body.Y+g.body.H {
		t.Errorf("after switching to columns item 29 is at %+v, outside %+v", c, g.body)
	}
}

func TestDrawShowsSelectionHoverAndOnlyTheVisibleThumbnails(t *testing.T) {
	th := DarkTheme()
	l := listOf(30)
	called := map[int]Rect{}
	for i := range l.Items {
		i := i
		l.Items[i].Thumb = func(_ Renderer, a Rect) { called[i] = a }
	}
	l.Selected = 1
	l.Update(Input{MouseX: 50, MouseY: 16 + 3*20 + 5}, listArea, listLH) // the pointer rests on item 3
	cv := &listCanvas{}
	l.Draw(cv, th, listArea, 1)
	g := l.geom(listArea, listLH)

	hasFill := func(r Rect, c Colour) bool {
		for _, f := range cv.fills {
			if f.R == r && f.C == c {
				return true
			}
		}
		return false
	}
	if !hasFill(l.cell(g, 1), th.SelFill) {
		t.Error("the selected item has no selection fill")
	}
	faded := th.SelFill
	faded.A = 96
	if !hasFill(l.cell(g, 3), faded) || hasFill(l.cell(g, 3), th.SelFill) {
		t.Error("the hovered item should have the faded fill only")
	}
	colourOf := func(s string) (Colour, bool) {
		for _, d := range cv.texts {
			if d.S == s {
				return d.C, true
			}
		}
		return Colour{}, false
	}
	if c, ok := colourOf("item 1"); !ok || c != th.SelText {
		t.Error("the selected name is not drawn in the selection text colour")
	}
	if c, ok := colourOf("item 2"); !ok || c != th.Text {
		t.Error("an ordinary name is not drawn in the text colour")
	}
	if _, ok := colourOf("item 29"); ok {
		t.Error("an item far below the window was drawn")
	}
	if len(called) == 0 || len(called) >= 30 {
		t.Errorf("%d thumbnails were painted, want only those on show", len(called))
	}
	for i, a := range called {
		c := l.cell(g, i)
		if a.X < c.X || a.Y < c.Y || a.X+a.W > c.X+c.W || a.Y+a.H > c.Y+c.H {
			t.Errorf("the thumbnail of item %d, %+v, is outside its cell %+v", i, a, c)
		}
	}
	for _, label := range []string{"Top", "Bottom"} {
		if _, ok := colourOf(label); !ok {
			t.Errorf("%s was not drawn on a list that scrolls", label)
		}
	}
}

func TestLongNamesAreShortenedToFit(t *testing.T) {
	th := DarkTheme()
	for _, cols := range []int{1, listGridColumns} {
		l := NewList()
		l.Columns = cols
		long := "a very long tile name indeed, longer than any cell"
		l.Items = []ListItem{{Name: long}}
		cv := &listCanvas{}
		l.Draw(cv, th, listArea, 1)
		g := l.geom(listArea, listLH)
		found := false
		for _, d := range cv.texts {
			if strings.HasPrefix(d.S, "a ") {
				found = true
				if !strings.HasSuffix(d.S, "..") || len(d.S) >= len(long) {
					t.Errorf("columns=%d: name drawn as %q, want it shortened with two dots", cols, d.S)
				}
				if w := len(d.S) * 8; w > g.itemW {
					t.Errorf("columns=%d: shortened name is %d wide, the cell is %d", cols, w, g.itemW)
				}
			}
		}
		if !found {
			t.Errorf("columns=%d: the name was not drawn", cols)
		}
	}
}

func TestScrolledOutItemsAreNeitherDrawnNorPainted(t *testing.T) {
	th := DarkTheme()
	l := listOf(30)
	painted := map[int]bool{}
	for i := range l.Items {
		i := i
		l.Items[i].Thumb = func(Renderer, Rect) { painted[i] = true }
	}
	l.Update(Input{MouseX: 50, MouseY: 100, WheelY: -5}, listArea, listLH) // five rows down: items 0 to 4 are out of sight
	if l.scroll != 100 {
		t.Fatalf("scroll = %d, want 100", l.scroll)
	}
	cv := &listCanvas{}
	l.Draw(cv, th, listArea, 1)
	drawn := map[string]bool{}
	for _, d := range cv.texts {
		drawn[d.S] = true
	}
	for i := 0; i < 5; i++ {
		if drawn[fmt.Sprintf("item %d", i)] || painted[i] {
			t.Errorf("item %d is scrolled out of sight but was drawn or painted", i)
		}
	}
	if !drawn["item 5"] || !painted[5] {
		t.Error("item 5, the first in sight, was not drawn")
	}
}

func TestTheRectanglesAreThoseTheListUses(t *testing.T) {
	l := listOf(5)
	g := l.geom(listArea, listLH)
	b := l.Buttons(listArea, listLH)
	if b.Add != g.add || b.Duplicate != g.dup || b.Delete != g.del || b.Rows != g.rows || b.Columns != g.cols {
		t.Errorf("Buttons = %+v, want the geometry's rectangles", b)
	}
	if got, want := l.ItemRect(3, listArea, listLH), l.cell(g, 3); got != want {
		t.Errorf("ItemRect(3) = %+v, want %+v", got, want)
	}
	x, y := mid(l.ItemRect(2, listArea, listLH))
	if ev, ok := press(l, x, y); !ok || ev.Index != 2 {
		t.Errorf("a click in the middle of ItemRect(2) selected %+v", ev)
	}
	x, y = mid(b.Add)
	if ev, ok := press(l, x, y); !ok || ev.Kind != ListAdd {
		t.Errorf("a click in the middle of the Add rectangle gave %+v", ev)
	}
}

func TestStepMovesTheSelectionWithTheArrowKeys(t *testing.T) {
	l := listOf(7)
	if ev, ok := l.Step(KeyDown); !ok || ev != (ListEvent{Kind: ListSelect, Index: 0}) {
		t.Errorf("with nothing selected Down gave %+v %v, want the first item", ev, ok)
	}
	l.Step(KeyDown)
	l.Step(KeyDown)
	if l.Selected != 2 {
		t.Errorf("after three Downs the selection is %d, want 2", l.Selected)
	}
	l.Step(KeyUp)
	if l.Selected != 1 {
		t.Errorf("Up left the selection at %d, want 1", l.Selected)
	}
	if _, ok := l.Step(KeyLeft); ok {
		t.Error("Left moved a list of rows")
	}
	if _, ok := l.Step(KeyEnter); ok {
		t.Error("an unrelated key moved the selection")
	}
	l.Selected = 0
	if _, ok := l.Step(KeyUp); ok || l.Selected != 0 {
		t.Errorf("Up at the first item reported a change, selection %d", l.Selected)
	}
	l.Selected = 6
	if _, ok := l.Step(KeyDown); ok || l.Selected != 6 {
		t.Errorf("Down at the last item reported a change, selection %d", l.Selected)
	}

	l.Columns = listGridColumns
	l.Selected = 1
	l.Step(KeyDown)
	if l.Selected != 4 {
		t.Errorf("in columns Down from 1 gave %d, want 4: a whole row", l.Selected)
	}
	l.Step(KeyRight)
	l.Step(KeyRight)
	if l.Selected != 6 {
		t.Errorf("Right twice from 4 gave %d, want 6, the last item", l.Selected)
	}
	l.Step(KeyLeft)
	if l.Selected != 5 {
		t.Errorf("Left from 6 gave %d, want 5", l.Selected)
	}
	if _, ok := NewList().Step(KeyDown); ok {
		t.Error("Step answered on an empty list")
	}
}

func TestAKeyboardStepSwitchesTheHoverOffUntilThePointerMoves(t *testing.T) {
	th := DarkTheme()
	l := listOf(6)
	l.Selected = 1
	g := l.geom(listArea, listLH)
	x, y := mid(l.cell(g, 4))
	faded := th.SelFill
	faded.A = 96
	hovered := func() bool {
		cv := &listCanvas{}
		l.Draw(cv, th, listArea, 1)
		for _, f := range cv.fills {
			if f.R == l.cell(g, 4) && f.C == faded {
				return true
			}
		}
		return false
	}
	l.Update(Input{MouseX: x, MouseY: y}, listArea, listLH) // the pointer rests on item 4
	if !hovered() {
		t.Fatal("the item under the pointer should show the hover highlight")
	}
	l.Step(KeyDown) // the keyboard moves the selection; the pointer does not
	l.Update(Input{MouseX: x, MouseY: y}, listArea, listLH)
	if hovered() {
		t.Error("the hover highlight stayed under a pointer that has not moved, after a keyboard step")
	}
	l.Update(Input{MouseX: x + 1, MouseY: y}, listArea, listLH)
	if !hovered() {
		t.Error("the hover highlight did not come back when the pointer moved")
	}
	l.Step(KeyDown)
	// A left press would select the item and draw it as selected, so the right
	// button is what shows that any press brings the hover back.
	l.Update(Input{MouseX: x + 1, MouseY: y, MouseRightPressed: true}, listArea, listLH)
	if !hovered() {
		t.Error("the hover highlight did not come back when the pointer pressed")
	}
}

func TestAnUnfocusedListDrawsItsSelectionDimmed(t *testing.T) {
	th := DarkTheme()
	l := listOf(5)
	l.Selected = 2
	g := l.geom(listArea, listLH)
	cell := l.cell(g, 2)
	draw := func() *listCanvas {
		cv := &listCanvas{}
		l.Draw(cv, th, listArea, 1)
		return cv
	}
	fillOf := func(cv *listCanvas) (Colour, bool) {
		for _, f := range cv.fills {
			if f.R == cell {
				return f.C, true
			}
		}
		return Colour{}, false
	}
	textOf := func(cv *listCanvas) Colour {
		for _, d := range cv.texts {
			if d.S == "item 2" {
				return d.C
			}
		}
		return Colour{}
	}
	cv := draw()
	if c, ok := fillOf(cv); !ok || c != th.SelFill || textOf(cv) != th.SelText {
		t.Error("a focused list should draw its selection in full")
	}
	l.Unfocused = true
	cv = draw()
	c, ok := fillOf(cv)
	if !ok || c.A != 64 || c.R != th.SelFill.R || c.G != th.SelFill.G || c.B != th.SelFill.B {
		t.Errorf("an unfocused selection fill is %+v, want the selection colour at alpha 64", c)
	}
	if textOf(cv) != th.Text {
		t.Error("an unfocused selection should keep the ordinary text colour")
	}
}

// iconCanvas stands in for a renderer over the icon font, whose glyphs are one
// rune and eight pixels wide.
type iconCanvas struct{ listCanvas }

func (c *iconCanvas) TextWidth(s string, scale int) int { return utf8.RuneCountInString(s) * 8 * scale }

// iconArt is what the icon font must hold, one string per row of the glyph.
var iconArt = map[rune][]string{
	IconAdd:       {"...##...", "...##...", "...##...", "########", "########", "...##...", "...##...", "...##..."},
	IconDuplicate: {"..######", "..#....#", "######.#", "#....#.#", "#....#.#", "#....###", "#....#..", "######.."},
	IconDelete:    {"...##...", "########", ".#.##.#.", ".#.##.#.", ".#.##.#.", ".#.##.#.", ".######.", "........"},
	IconRows:      {"########", "########", "........", "########", "########", "........", "########", "########"},
	IconColumns:   {"##.##.##", "##.##.##", "........", "##.##.##", "##.##.##", "........", "##.##.##", "##.##.##"},
	IconUndo:      {"..#.....", ".##.....", "#######.", ".##...#.", "..#...#.", "......#.", "...####.", "........"},
	IconRedo:      {".....#..", ".....##.", ".#######", ".#...##.", ".#...#..", ".#......", ".####...", "........"},
	IconPen:       {"......##", ".....###", "....###.", "...###..", "..###...", ".###....", "##......", "#......."},
	IconMove:      {"...##...", "..####..", ".#.##.#.", "########", "########", ".#.##.#.", "..####..", "...##..."},
	IconFlipH:     {"...##...", "...##...", "#..##..#", "##.##.##", "##.##.##", "#..##..#", "...##...", "...##..."},
	IconFlipV:     {"..####..", "...##...", "........", "########", "########", "........", "...##...", "..####.."},
	IconRotate:    {"..####..", ".#....#.", "#......#", "#....###", "#.....#.", ".#......", "..####..", "........"},
	IconGrid:      {"########", "#..##..#", "#..##..#", "########", "########", "#..##..#", "#..##..#", "########"},
	IconClear:     {"#......#", ".#....#.", "..#..#..", "...##...", "...##...", "..#..#..", ".#....#.", "#......#"},
	IconFill:      {"..####..", ".#....#.", ".#....#.", "########", ".######.", ".######.", "..####..", "........"},
	IconPick:      {"...##...", "...##...", "........", "##....##", "##....##", "........", "...##...", "...##..."},
}

// bdfGlyphs reads the encoding and bitmap rows of every glyph in a BDF file.
func bdfGlyphs(t *testing.T, bdf []byte) map[int][]string {
	t.Helper()
	out := map[int][]string{}
	enc, inBitmap := -1, false
	for _, line := range strings.Split(string(bdf), "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "ENCODING "):
			n, err := strconv.Atoi(strings.TrimPrefix(line, "ENCODING "))
			if err != nil {
				t.Fatal(err)
			}
			enc = n
		case line == "BITMAP":
			inBitmap = true
			out[enc] = nil
		case line == "ENDCHAR":
			inBitmap = false
		case inBitmap:
			out[enc] = append(out[enc], line)
		}
	}
	return out
}

func TestTheIconFontHoldsTheIconsAsDrawn(t *testing.T) {
	glyphs := bdfGlyphs(t, IconsBDF)
	if len(glyphs) != len(iconArt) {
		t.Fatalf("the icon font holds %d glyphs, want %d", len(glyphs), len(iconArt))
	}
	for r, art := range iconArt {
		rows, ok := glyphs[int(r)]
		if !ok || len(rows) != 8 {
			t.Errorf("U+%04X: %d rows (present %v), want 8", r, len(rows), ok)
			continue
		}
		for y, row := range rows {
			v, err := strconv.ParseUint(row, 16, 8)
			if err != nil {
				t.Fatal(err)
			}
			got := ""
			for x := 7; x >= 0; x-- {
				if v&(1<<uint(x)) != 0 {
					got += "#"
				} else {
					got += "."
				}
			}
			if got != art[y] {
				t.Errorf("U+%04X row %d is %s, want %s", r, y, got, art[y])
			}
		}
	}
}

func TestToolbarIconsAreDrawnFromTheIconRenderer(t *testing.T) {
	th := DarkTheme()
	l := listOf(3)
	l.Selected = 1
	l.DuplicateDisabled = true
	ic := &iconCanvas{}
	l.Icons = ic
	cv := &listCanvas{}
	l.Draw(cv, th, listArea, 1)
	b := l.Buttons(listArea, listLH)
	rect := map[string]Rect{string(IconAdd): b.Add, string(IconDuplicate): b.Duplicate, string(IconDelete): b.Delete, string(IconRows): b.Rows, string(IconColumns): b.Columns}
	colour := map[string]Colour{
		string(IconAdd): th.ButtonText, string(IconDuplicate): th.Disabled, string(IconDelete): th.ButtonText,
		string(IconRows): th.SelText, string(IconColumns): th.ButtonText,
	}
	seen := map[string]bool{}
	for _, d := range ic.texts {
		r, ok := rect[d.S]
		if !ok {
			t.Errorf("the icon renderer drew %q, which is not a toolbar icon", d.S)
			continue
		}
		seen[d.S] = true
		if d.X != r.X+(r.W-8)/2 || d.Y != r.Y+(r.H-8)/2 || d.X < r.X || d.X+8 > r.X+r.W {
			t.Errorf("icon U+%04X is at %d,%d, not centred in its button %+v", []rune(d.S)[0], d.X, d.Y, r)
		}
		if d.C != colour[d.S] {
			t.Errorf("icon U+%04X is drawn in %+v, want %+v", []rune(d.S)[0], d.C, colour[d.S])
		}
	}
	if len(seen) != 5 {
		t.Errorf("%d of the five icons were drawn", len(seen))
	}
	for _, d := range cv.texts {
		if r := []rune(d.S); len(r) > 0 && r[0] >= 0xE000 && r[0] <= 0xF8FF {
			t.Errorf("an icon glyph went through the text renderer: %q", d.S)
		}
	}
}

func TestWithoutAnIconRendererTheToolbarFallsBackToPlainCharacters(t *testing.T) {
	l := listOf(3)
	l.Selected = 1
	cv := &listCanvas{}
	l.Draw(cv, DarkTheme(), listArea, 1)
	got := map[string]bool{}
	for _, d := range cv.texts {
		got[d.S] = true
	}
	for _, want := range []string{"+", "c", "x", "=", "#"} {
		if !got[want] {
			t.Errorf("the fallback did not draw %q", want)
		}
	}
}

func TestAListWithNoToolbarHasNoButtonsAndUsesTheWholeArea(t *testing.T) {
	th := DarkTheme()
	with, without := listOf(5), listOf(5)
	without.NoToolbar = true
	gw, gn := with.geom(listArea, listLH), without.geom(listArea, listLH)
	if gn.toolbar.H != 0 || gn.body.Y != listArea.Y || gn.body.H != listArea.H {
		t.Errorf("with no toolbar the body is %+v and the toolbar %+v, want the body to start at the top and fill the area", gn.body, gn.toolbar)
	}
	if gw.body.Y == gn.body.Y || gn.add != (Rect{}) || gn.dup != (Rect{}) || gn.del != (Rect{}) || gn.rows != (Rect{}) || gn.cols != (Rect{}) {
		t.Error("the buttons of a list with no toolbar have rectangles")
	}
	cv := &listCanvas{}
	without.Draw(cv, th, listArea, 1)
	for _, f := range cv.fills {
		if f.C == th.Button || f.C == th.ButtonHot {
			t.Errorf("a button was drawn: %+v", f)
		}
	}
	// the items sit where the toolbar was, and a press there selects, not adds
	r := without.ItemRect(0, listArea, listLH)
	x, y := mid(r)
	ev, ok := without.Update(Input{MouseX: x, MouseY: y, MousePressed: true}, listArea, listLH)
	if !ok || ev.Kind != ListSelect || ev.Index != 0 {
		t.Errorf("a press on the first item gave %+v %v, want it selected", ev, ok)
	}
	if r.Y != listArea.Y {
		t.Errorf("the first item starts at %d, want the top of the area", r.Y)
	}
}

func TestAListThatFitsShowsNoFootButtonsAndNoScrollThumb(t *testing.T) {
	th := DarkTheme()
	l := listOf(5)
	cv := &listCanvas{}
	l.Draw(cv, th, listArea, 1)
	for _, d := range cv.texts {
		if d.S == "Top" || d.S == "Bottom" {
			t.Errorf("%q was drawn on a list that fits", d.S)
		}
	}
	for _, f := range cv.fills {
		if f.C == th.DimText && f.R.W == 2 {
			t.Error("a scroll thumb was drawn on a list that fits")
		}
	}
}
