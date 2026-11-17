package zenui

// ListItem is one entry of a List.
type ListItem struct {
	Name string
	// Thumb, if set, draws the item's picture inside area. The list calls it
	// only for items that are on show.
	Thumb func(r Renderer, area Rect)
}

// ListEventKind says what the person did to a List.
type ListEventKind int

// The events a List reports.
const (
	ListSelect    ListEventKind = iota + 1 // an item was chosen; Index is its position
	ListAdd                                // the add button was pressed
	ListDuplicate                          // the duplicate button was pressed; Index is the selected item
	ListDelete                             // the delete button was pressed; Index is the selected item
	ListMode                               // the list was switched between rows and columns
)

// ListEvent is one thing the person did to a List. Index is -1 where no item is
// meant.
type ListEvent struct {
	Kind  ListEventKind
	Index int
}

// listGridColumns is how many items sit across in the columns mode.
const listGridColumns = 3

// List is a scrolling list of named items, with an optional picture for each, and
// a toolbar to add, duplicate and delete items and to show them as rows or in
// columns. It reports what the person does as events and leaves the items to the
// host.
//
// Rows put the picture beside the name; columns put it above, three across. When
// the items are taller than the window the list scrolls with the wheel, and Top
// and Bottom buttons appear at its foot.
//
// The layout is a pure function of the area and the line height, so Update does
// not depend on a previous Draw: the host passes the same area and line height
// to both, every frame.
type List struct {
	Items    []ListItem
	Selected int // the chosen item, or -1 for none
	// Columns is 1 for rows, or listGridColumns for the grid.
	Columns int
	// The host sets these each frame to match the selection and its rules.
	AddDisabled, DuplicateDisabled, DeleteDisabled bool
	// Unfocused draws the selected item dimmed, because the keyboard focus is
	// elsewhere. The host sets it each frame.
	Unfocused bool
	// Icons draws the toolbar's glyphs. It is a renderer over the icon font in
	// IconsBDF, as zenimate's tool palette draws through one. With none, the
	// toolbar falls back to plain characters.
	Icons Renderer
	// NoToolbar leaves the toolbar out, for a list that is only a palette to choose
	// from.
	NoToolbar bool

	scroll, hover  int
	mouseX, mouseY int

	// A keyboard step switches the hover highlight off until the pointer moves
	// or presses, because the pointer is no longer what the person is using.
	hoverOff           bool
	hoverAtX, hoverAtY int
}

// NewList returns an empty list in rows mode with nothing selected.
func NewList() *List { return &List{Selected: -1, Columns: 1, hover: -1} }

type listGeom struct {
	add, dup, del, rows, cols Rect // the toolbar buttons
	toolbar, body             Rect
	top, bottom               Rect // the foot buttons; zero unless scrollable
	scrollable                bool
	content                   int // the height of all the items
	itemW, itemH, thumb, pad  int
}

func (l *List) geom(area Rect, lh int) listGeom {
	pad := max(2, lh/4)
	btn := lh + lh/2
	g := listGeom{pad: pad}
	g.toolbar = Rect{X: area.X, Y: area.Y, W: area.W, H: btn + 2*pad}
	y := area.Y + pad
	x := area.X + pad
	g.add = Rect{X: x, Y: y, W: btn, H: btn}
	g.dup = Rect{X: x + btn + pad, Y: y, W: btn, H: btn}
	g.del = Rect{X: x + 2*(btn+pad), Y: y, W: btn, H: btn}
	rx := area.X + area.W - pad - btn
	g.cols = Rect{X: rx, Y: y, W: btn, H: btn}
	g.rows = Rect{X: rx - pad - btn, Y: y, W: btn, H: btn}

	if l.NoToolbar {
		g.toolbar = Rect{X: area.X, Y: area.Y, W: area.W}
		g.add, g.dup, g.del, g.rows, g.cols = Rect{}, Rect{}, Rect{}, Rect{}, Rect{}
	}
	body := Rect{X: area.X, Y: area.Y + g.toolbar.H, W: area.W, H: max(0, area.H-g.toolbar.H)}
	cols := max(1, l.Columns)
	if cols == 1 {
		g.thumb = lh * 2
		g.itemW = body.W
		g.itemH = g.thumb + 2*pad
	} else {
		g.itemW = body.W / cols
		g.thumb = max(1, g.itemW-2*pad)
		g.itemH = g.thumb + lh + 3*pad
	}
	g.content = ((len(l.Items) + cols - 1) / cols) * g.itemH
	if g.content > body.H {
		g.scrollable = true
		foot := lh + 2*pad
		body.H = max(0, body.H-foot)
		half := body.W / 2
		g.top = Rect{X: body.X, Y: body.Y + body.H, W: half, H: foot}
		g.bottom = Rect{X: body.X + half, Y: body.Y + body.H, W: body.W - half, H: foot}
	}
	g.body = body
	return g
}

func (l *List) maxScroll(g listGeom) int { return max(0, g.content-g.body.H) }

func (l *List) clamp(g listGeom) { l.scroll = clampInt(l.scroll, 0, l.maxScroll(g)) }

func clampInt(v, lo, hi int) int {
	if hi < lo {
		hi = lo
	}
	return max(lo, min(v, hi))
}

// cell is the rectangle of item i, as drawn: shifted up by the scroll.
func (l *List) cell(g listGeom, i int) Rect {
	cols := max(1, l.Columns)
	return Rect{X: g.body.X + (i%cols)*g.itemW, Y: g.body.Y + (i/cols)*g.itemH - l.scroll, W: g.itemW, H: g.itemH}
}

// hit returns the item under a point, or -1.
func (l *List) hit(g listGeom, px, py int) int {
	if !g.body.Contains(px, py) || g.itemW <= 0 || g.itemH <= 0 {
		return -1
	}
	cols := max(1, l.Columns)
	col := (px - g.body.X) / g.itemW
	row := (py - g.body.Y + l.scroll) / g.itemH
	if col >= cols {
		return -1
	}
	if i := row*cols + col; i >= 0 && i < len(l.Items) {
		return i
	}
	return -1
}

// ListButtons are the rectangles of the toolbar buttons.
type ListButtons struct{ Add, Duplicate, Delete, Rows, Columns Rect }

// Buttons returns the toolbar buttons' rectangles inside area, for a host that
// places something against them.
func (l *List) Buttons(area Rect, lineH int) ListButtons {
	g := l.geom(area, lineH)
	return ListButtons{Add: g.add, Duplicate: g.dup, Delete: g.del, Rows: g.rows, Columns: g.cols}
}

// ItemRect returns the rectangle item i occupies as drawn, shifted by the scroll.
func (l *List) ItemRect(i int, area Rect, lineH int) Rect {
	return l.cell(l.geom(area, lineH), i)
}

// Reveal scrolls the list just far enough to show the selected item. The host
// calls it after changing the selection itself, for instance after an add.
func (l *List) Reveal(area Rect, lineH int) {
	g := l.geom(area, lineH)
	if l.Selected >= 0 && l.Selected < len(l.Items) {
		c := l.cell(g, l.Selected)
		if c.Y < g.body.Y {
			l.scroll -= g.body.Y - c.Y
		} else if c.Y+c.H > g.body.Y+g.body.H {
			l.scroll += c.Y + c.H - (g.body.Y + g.body.H)
		}
	}
	l.clamp(g)
}

// Update applies one frame of input to the list inside area. It returns an
// event when the person did something the host must act on.
func (l *List) Update(in Input, area Rect, lineH int) (ListEvent, bool) {
	g := l.geom(area, lineH)
	l.clamp(g)
	l.mouseX, l.mouseY = in.MouseX, in.MouseY
	if l.hoverOff && (in.MouseX != l.hoverAtX || in.MouseY != l.hoverAtY || in.MousePressed || in.MouseRightPressed) {
		l.hoverOff = false
	}
	if in.WheelY != 0 && area.Contains(in.MouseX, in.MouseY) && g.scrollable {
		l.scroll -= int(in.WheelY * float32(g.itemH))
		l.clamp(g)
	}
	l.hover = -1
	if !l.hoverOff {
		l.hover = l.hit(g, in.MouseX, in.MouseY)
	}
	if !in.MousePressed {
		return ListEvent{}, false
	}
	x, y := in.MouseX, in.MouseY
	switch {
	case g.add.Contains(x, y):
		if !l.AddDisabled {
			return ListEvent{Kind: ListAdd, Index: -1}, true
		}
	case g.dup.Contains(x, y):
		if !l.DuplicateDisabled && l.Selected >= 0 {
			return ListEvent{Kind: ListDuplicate, Index: l.Selected}, true
		}
	case g.del.Contains(x, y):
		if !l.DeleteDisabled && l.Selected >= 0 {
			return ListEvent{Kind: ListDelete, Index: l.Selected}, true
		}
	case g.rows.Contains(x, y):
		return l.setColumns(1, area, lineH)
	case g.cols.Contains(x, y):
		return l.setColumns(listGridColumns, area, lineH)
	case g.scrollable && g.top.Contains(x, y):
		l.scroll = 0
	case g.scrollable && g.bottom.Contains(x, y):
		l.scroll = l.maxScroll(g)
	default:
		if i := l.hit(g, x, y); i >= 0 {
			l.Selected = i
			return ListEvent{Kind: ListSelect, Index: i}, true
		}
	}
	return ListEvent{}, false
}

func (l *List) setColumns(n int, area Rect, lineH int) (ListEvent, bool) {
	if l.Columns == n {
		return ListEvent{}, false
	}
	l.Columns = n
	l.Reveal(area, lineH)
	return ListEvent{Kind: ListMode, Index: -1}, true
}

// fadedFill is the highlight under the pointer: the selection colour, faded, so
// it never reads as the selection itself.
func fadedFill(th Theme) Colour {
	c := th.SelFill
	c.A = 96
	return c
}

// dimmedFill is the mark of a selection whose list does not have the keyboard
// focus: fainter than the highlight under the pointer.
func dimmedFill(th Theme) Colour {
	c := th.SelFill
	c.A = 64
	return c
}

// Step moves the selection for an arrow key and reports it as a select event.
// Up and Down move by a row, which in columns is a step of three items; Left and
// Right move by one item, in columns only. With nothing selected, an arrow key
// selects the first item. The selection stops at the ends; it reports false when
// nothing changed.
func (l *List) Step(k Key) (ListEvent, bool) {
	n := len(l.Items)
	if n == 0 {
		return ListEvent{}, false
	}
	cols := max(1, l.Columns)
	delta := 0
	switch k {
	case KeyUp:
		delta = -cols
	case KeyDown:
		delta = cols
	case KeyLeft:
		if cols > 1 {
			delta = -1
		}
	case KeyRight:
		if cols > 1 {
			delta = 1
		}
	}
	if delta == 0 {
		return ListEvent{}, false
	}
	i := 0
	if l.Selected >= 0 {
		i = clampInt(l.Selected+delta, 0, n-1)
	}
	if i == l.Selected {
		return ListEvent{}, false
	}
	l.Selected = i
	l.hoverOff, l.hoverAtX, l.hoverAtY = true, l.mouseX, l.mouseY
	l.hover = -1
	return ListEvent{Kind: ListSelect, Index: i}, true
}

// Draw paints the list inside area at the given text scale. It does not fill the
// area's background; the host does.
func (l *List) Draw(r Renderer, th Theme, area Rect, scale int) {
	lh := r.LineHeight(scale)
	g := l.geom(area, lh)
	l.clamp(g)

	if !l.NoToolbar {
		l.drawToolbar(r, th, g, scale)
	}

	r.Clip(g.body)
	cols := max(1, l.Columns)
	first := max(0, l.scroll/max(1, g.itemH)) * cols
	for i := first; i < len(l.Items); i++ {
		c := l.cell(g, i)
		if c.Y >= g.body.Y+g.body.H {
			break
		}
		item := l.Items[i]
		textCol := th.Text
		switch {
		case i == l.Selected && l.Unfocused:
			r.FillRect(c, dimmedFill(th))
		case i == l.Selected:
			r.FillRect(c, th.SelFill)
			textCol = th.SelText
		case i == l.hover:
			r.FillRect(c, fadedFill(th))
		}
		var pic Rect
		if cols == 1 {
			pic = Rect{X: c.X + g.pad, Y: c.Y + g.pad, W: g.thumb, H: g.thumb}
			name := fitText(r, item.Name, c.W-g.thumb-3*g.pad, scale)
			r.DrawText(name, pic.X+pic.W+g.pad, c.Y+(c.H-lh)/2, scale, textCol)
		} else {
			pic = Rect{X: c.X + g.pad, Y: c.Y + g.pad, W: g.thumb, H: g.thumb}
			name := fitText(r, item.Name, c.W-2*g.pad, scale)
			r.DrawText(name, c.X+(c.W-r.TextWidth(name, scale))/2, pic.Y+pic.H+g.pad, scale, textCol)
		}
		if item.Thumb != nil {
			item.Thumb(r, pic)
		}
		r.StrokeRect(pic, th.Border, 1)
	}
	if g.scrollable {
		w := max(2, lh/6)
		span := l.maxScroll(g)
		thumbH := max(lh, g.body.H*g.body.H/max(1, g.content))
		y := g.body.Y + (g.body.H-thumbH)*l.scroll/max(1, span)
		r.FillRect(Rect{X: g.body.X + g.body.W - w - 1, Y: y, W: w, H: thumbH}, th.DimText)
	}
	r.ClipEnd()

	if g.scrollable {
		for _, b := range []struct {
			rect  Rect
			label string
		}{{g.top, "Top"}, {g.bottom, "Bottom"}} {
			fill := th.Button
			if b.rect.Contains(l.mouseX, l.mouseY) {
				fill = th.ButtonHot
			}
			r.FillRect(b.rect, fill)
			r.StrokeRect(b.rect, th.Border, 1)
			r.DrawText(b.label, b.rect.X+(b.rect.W-r.TextWidth(b.label, scale))/2, b.rect.Y+(b.rect.H-lh)/2, scale, th.ButtonText)
		}
	}
}

// fitText shortens s with two dots so that it fits in width.
func fitText(r Renderer, s string, width, scale int) string {
	if r.TextWidth(s, scale) <= width {
		return s
	}
	rs := []rune(s)
	for len(rs) > 0 {
		rs = rs[:len(rs)-1]
		if t := string(rs) + ".."; r.TextWidth(t, scale) <= width {
			return t
		}
	}
	return ""
}

func (l *List) drawToolbar(r Renderer, th Theme, g listGeom, scale int) {
	button := func(b Rect, on, disabled bool) Colour {
		fill := th.Button
		switch {
		case on:
			fill = th.SelFill
		case !disabled && b.Contains(l.mouseX, l.mouseY):
			fill = th.ButtonHot
		}
		r.FillRect(b, fill)
		r.StrokeRect(b, th.Border, 1)
		if disabled {
			return th.Disabled
		}
		if on {
			return th.SelText
		}
		return th.ButtonText
	}
	noSel := l.Selected < 0
	l.icon(r, scale, g.add, IconAdd, "+", button(g.add, false, l.AddDisabled))
	l.icon(r, scale, g.dup, IconDuplicate, "c", button(g.dup, false, l.DuplicateDisabled || noSel))
	l.icon(r, scale, g.del, IconDelete, "x", button(g.del, false, l.DeleteDisabled || noSel))
	l.icon(r, scale, g.rows, IconRows, "=", button(g.rows, l.Columns == 1, false))
	l.icon(r, scale, g.cols, IconColumns, "#", button(g.cols, l.Columns != 1, false))
}

// icon draws a toolbar glyph centred in its button: from the icon renderer, or
// as a plain character in the text face when the host gave none.
func (l *List) icon(r Renderer, scale int, b Rect, glyph rune, plain string, c Colour) {
	if l.Icons != nil {
		s := string(glyph)
		l.Icons.DrawText(s, b.X+(b.W-l.Icons.TextWidth(s, scale))/2, b.Y+(b.H-l.Icons.LineHeight(scale))/2, scale, c)
		return
	}
	r.DrawText(plain, b.X+(b.W-r.TextWidth(plain, scale))/2, b.Y+(b.H-r.LineHeight(scale))/2, scale, c)
}
