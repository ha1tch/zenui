package zenui

import "strings"

// scale is the integer text scale the dialog renders at.
const dlgScale = 2

// recentPlace is the Path of the sidebar's entry for the recent files, which is
// not a folder.
const recentPlace = "\x00recent"

func (d *Dialog) scale() int {
	if d.cfg.Scale > 0 {
		return d.cfg.Scale
	}
	return dlgScale
}

type colKind int

const (
	colName colKind = iota
	colInfo
	colSize
	colDate
	colFolder
)

// dlgCol is a column of the details list, as laid out.
type dlgCol struct {
	kind  colKind
	key   SortKey // what a click on its head sorts by, for the name, size and date columns
	x, w  int
	rect  Rect // the head cell
	label string
}

func (c dlgCol) sortable() bool {
	return c.kind == colName || c.kind == colSize || c.kind == colDate
}

// nameInner is the area of the name field the text is drawn and edited in.
func (d *Dialog) nameInner() Rect {
	n := d.nameRect
	w := n.W - 12
	if w < 0 {
		w = 0
	}
	return Rect{X: n.X + 6, Y: n.Y + 1, W: w, H: n.H - 2}
}

// layout computes the dialog's sub-rectangles for a given outer screen size,
// centring a fixed-ish panel. A favourites sidebar runs down the left; the path
// bar, file list, filename field, file-type selector and buttons occupy the area
// to its right.
func (d *Dialog) layout(r Renderer, screenW, screenH int) {
	sc := d.scale()
	d.textW = func(s string) int { return r.TextWidth(s, sc) }
	lh := r.LineHeight(sc)
	lh1 := r.LineHeight(1)
	pad := lh

	maxW, maxH := 760, 560
	if d.cfg.Width > 0 {
		maxW = d.cfg.Width
	}
	if d.cfg.Height > 0 {
		maxH = d.cfg.Height
	}
	minW, minH := 420, 300
	if maxW < minW {
		minW = maxW
	}
	if maxH < minH {
		minH = maxH
	}
	pw := screenW * 3 / 4
	if pw > maxW {
		pw = maxW
	}
	if pw < minW {
		pw = minW
	}
	ph := screenH * 3 / 4
	if ph > maxH {
		ph = maxH
	}
	if ph < minH {
		ph = minH
	}
	px := (screenW - pw) / 2
	py := (screenH - ph) / 2
	d.bounds = Rect{px, py, pw, ph}

	titleY := py + pad
	btnH := lh + 10
	rowY := py + ph - pad - btnH // button row top
	fieldH := lh + 8
	fieldY := rowY - 8 - fieldH // filename field

	midTop := titleY + lh + 8
	midBottom := fieldY - 8
	d.rowH = lh + 4

	// Favourites sidebar on the left, with the switches at its foot.
	sideW := pw / 4
	if sideW < 110 {
		sideW = 110
	}
	if sideW > 200 {
		sideW = 200
	}
	d.sideRect = Rect{px + pad, midTop, sideW, midBottom - midTop}
	togH := lh1 + 8
	bottom := d.sideRect.Y + d.sideRect.H
	d.detailsRect = Rect{d.sideRect.X + 2, bottom - togH - 2, d.sideRect.W - 4, togH}
	placesBottom := d.detailsRect.Y
	d.hiddenRect = Rect{}
	if d.hiddenSupported() {
		d.hiddenRect = Rect{d.sideRect.X + 2, d.detailsRect.Y - togH, d.sideRect.W - 4, togH}
		placesBottom = d.hiddenRect.Y
	}
	d.placeRects = d.placeRects[:0]
	prowH := lh + 6
	for i := range d.places {
		py2 := d.sideRect.Y + 4 + i*prowH
		if py2+prowH > placesBottom {
			break
		}
		d.placeRects = append(d.placeRects, Rect{d.sideRect.X + 2, py2, d.sideRect.W - 4, prowH})
	}

	// Content area to the right of the sidebar.
	contentX := d.sideRect.X + sideW + pad
	contentW := px + pw - pad - contentX

	pathY := midTop
	d.upRect = Rect{contentX, pathY - 2, lh + 6, lh + 4}

	listY := pathY + lh + 8
	showCols := d.view.Details && !d.inContainer()
	d.headRect = Rect{}
	if showCols {
		d.headRect = Rect{contentX, listY, contentW, d.rowH}
		listY += d.rowH
	}
	listH := midBottom - listY
	// While browsing inside a container, reserve a preview pane on the right of
	// the content area for the selected entry's thumbnail and metadata.
	if d.inContainer() {
		previewW := contentW * 2 / 5
		if previewW < 120 {
			previewW = 120
		}
		listW := contentW - previewW - pad
		d.listRect = Rect{contentX, listY, listW, listH}
		d.previewRect = Rect{contentX + listW + pad, listY, previewW, listH}
	} else {
		d.listRect = Rect{contentX, listY, contentW, listH}
		d.previewRect = Rect{}
	}
	d.layoutCols(r)

	// The name field, with the file-type selector at its right.
	d.nameRect = Rect{contentX, fieldY, contentW, fieldH}
	d.filterRect = Rect{}
	if d.hasSelector() {
		longest := 0
		for _, s := range d.sets {
			if w := r.TextWidth(s.Name, sc); w > longest {
				longest = w
			}
		}
		fw := longest + r.TextWidth(" v", sc) + 16
		if fw > contentW/2 {
			fw = contentW / 2
		}
		d.filterRect = Rect{contentX + contentW - fw, fieldY, fw, fieldH}
		d.nameRect.W = contentW - fw - 8
	}

	bw := 0
	for _, s := range []string{
		pick(d.cfg.Labels.Cancel, "CANCEL"),
		pick(d.cfg.Labels.Open, "OPEN"),
		pick(d.cfg.Labels.Save, "SAVE"),
		pick(d.cfg.Labels.Replace, "REPLACE"),
	} {
		if w := r.TextWidth(strings.ToUpper(s), sc); w > bw {
			bw = w
		}
	}
	bw += 2 * pad
	d.okRect = Rect{px + pw - pad - bw, rowY, bw, btnH}
	d.cancRect = Rect{d.okRect.X - 12 - bw, rowY, bw, btnH}
	d.statusRect = Rect{px + pad, rowY, d.cancRect.X - 12 - (px + pad), btnH}

	// The file-type list, when open: above its button, its rows in order.
	d.popupRows = d.popupRows[:0]
	d.popupRect = Rect{}
	if d.popup && d.hasSelector() {
		w := d.filterRect.W
		n := len(d.sets)
		if maxRows := (d.filterRect.Y - 2 - (py + 4)) / d.rowH; n > maxRows {
			n = maxRows
		}
		h := n * d.rowH
		d.popupRect = Rect{d.filterRect.X + d.filterRect.W - w, d.filterRect.Y - 2 - h, w, h}
		for i := 0; i < n; i++ {
			d.popupRows = append(d.popupRows, Rect{d.popupRect.X, d.popupRect.Y + i*d.rowH, w, d.rowH})
		}
	}
}

// layoutCols works out the columns of the details list: the name takes what the
// others leave, and the others are left out, the date first, when the list is
// too narrow to give the name room.
func (d *Dialog) layoutCols(r Renderer) {
	d.cols = d.cols[:0]
	if d.headRect.W == 0 {
		return
	}
	h := d.headRect
	cell := func(kind colKind, key SortKey, x, w int, label string) {
		d.cols = append(d.cols, dlgCol{kind: kind, key: key, x: x, w: w, rect: Rect{x, h.Y, w, h.H}, label: label})
	}
	l := d.cfg.Labels
	cw := r.TextWidth("0", 1)
	if cw < 1 {
		cw = 1
	}
	const pad = 8
	if d.recent {
		w := h.W * 45 / 100
		cell(colName, SortName, h.X, h.W-w, pick(l.Name, "Name"))
		cell(colFolder, SortName, h.X+h.W-w, w, pick(l.Folder, "Folder"))
		return
	}
	nameMin := 12 * r.TextWidth("0", d.scale())
	sizeW := r.TextWidth("1023.9 KB", 1) + 2*pad
	dateW := r.TextWidth(dateLayout, 1) + 2*pad
	infoW := 0
	if d.infoChars > 0 {
		infoW = d.infoChars*cw + 2*pad
		if infoW > h.W/3 {
			infoW = h.W / 3
		}
	}
	rem := h.W
	useSize, useDate, useInfo := false, false, false
	if rem-sizeW >= nameMin {
		useSize, rem = true, rem-sizeW
	}
	if rem-dateW >= nameMin {
		useDate, rem = true, rem-dateW
	}
	if infoW > 0 && rem-infoW >= nameMin {
		useInfo, rem = true, rem-infoW
	}
	x := h.X
	cell(colName, SortName, x, rem, pick(l.Name, "Name"))
	x += rem
	if useInfo {
		cell(colInfo, SortName, x, infoW, pick(l.Info, "Info"))
		x += infoW
	}
	if useSize {
		cell(colSize, SortSize, x, sizeW, pick(l.Size, "Size"))
		x += sizeW
	}
	if useDate {
		cell(colDate, SortDate, x, dateW, pick(l.Modified, "Modified"))
	}
}

// visibleRows is how many entries fit in the list area.
func (d *Dialog) visibleRows() int {
	if d.rowH <= 0 {
		return 0
	}
	n := d.listRect.H / d.rowH
	if n < 0 {
		return 0
	}
	return n
}

// fit shortens s with ".." until it is no wider than w.
func fit(r Renderer, s string, w, scale int) string {
	if r.TextWidth(s, scale) <= w {
		return s
	}
	rs := []rune(s)
	for len(rs) > 0 {
		rs = rs[:len(rs)-1]
		if t := string(rs) + ".."; r.TextWidth(t, scale) <= w {
			return t
		}
	}
	return ""
}

// statusLine is what the line beside the buttons says: the question being asked,
// the error, a hint, or how much is listed.
func (d *Dialog) statusLine(th Theme) (string, Colour) {
	switch {
	case d.confirm != "":
		return baseOf(d.confirm) + " exists. Replace it?  Enter replaces, Esc keeps", th.Text
	case d.err != "":
		return "! " + d.err, th.Text
	case d.hint != "":
		return d.hint, th.DimText
	case d.custom != "":
		return "Showing " + d.custom, th.DimText
	}
	switch n := len(d.entries); n {
	case 1:
		return "1 item", th.DimText
	default:
		return itoa(n) + " items", th.DimText
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

// Draw renders the dialog. screenW/screenH are the host surface dimensions so
// the dialog can centre itself and dim the backdrop. theme controls colours.
func (d *Dialog) Draw(r Renderer, screenW, screenH int, theme Theme) {
	d.layout(r, screenW, screenH)
	sc := d.scale()

	// Backdrop over the whole screen.
	r.FillRect(Rect{0, 0, screenW, screenH}, theme.Backdrop)

	b := d.bounds
	r.FillRect(b, theme.Panel)
	r.StrokeRect(b, theme.Border, 1)

	lh := r.LineHeight(sc)
	lh1 := r.LineHeight(1)
	pad := lh
	innerX := b.X + pad

	// Title.
	r.DrawText(strings.ToUpper(d.cfg.Title), innerX, b.Y+pad, sc, theme.Text)

	// Up button and current path.
	r.FillRect(d.upRect, theme.Button)
	r.StrokeRect(d.upRect, theme.Border, 1)
	upGlyph := ".."
	r.DrawText(upGlyph, d.upRect.X+(d.upRect.W-r.TextWidth(upGlyph, sc))/2,
		d.upRect.Y+(d.upRect.H-lh)/2, sc, theme.Text)
	pathX := d.upRect.X + d.upRect.W + 8
	pathLabel := d.dir
	if d.recent {
		pathLabel = pick(d.cfg.Labels.Recent, "Recent")
	}
	d.drawClipped(r, pathLabel, pathX, d.upRect.Y+(d.upRect.H-lh)/2, b.X+b.W-pad-pathX, theme.DimText)

	// Favourites sidebar.
	r.FillRect(d.sideRect, theme.Sidebar)
	r.Clip(d.sideRect)
	for i, pr := range d.placeRects {
		if i >= len(d.places) {
			break
		}
		here := d.places[i].Path == d.dir && !d.recent
		if d.places[i].Path == recentPlace {
			here = d.recent
		}
		if here {
			r.FillRect(pr, theme.SelFill)
		}
		r.DrawText(d.places[i].Label, pr.X+6, pr.Y+(pr.H-lh)/2, sc, theme.SideText)
	}
	if d.hiddenRect.W > 0 {
		d.drawSwitch(r, d.hiddenRect, pick(d.cfg.Labels.Hidden, "Hidden"), d.view.ShowHidden, true, theme)
	}
	d.drawSwitch(r, d.detailsRect, pick(d.cfg.Labels.Details, "Details"), d.view.Details, !d.inContainer(), theme)
	r.ClipEnd()
	r.StrokeRect(d.sideRect, theme.Border, 1)

	// Column heads.
	if d.headRect.W > 0 {
		for _, c := range d.cols {
			r.FillRect(c.rect, theme.Button)
			r.StrokeRect(c.rect, theme.Border, 1)
			label := c.label
			if c.sortable() && !d.recent && c.key == d.view.Sort {
				if d.view.Descending {
					label += " v"
				} else {
					label += " ^"
				}
			}
			r.Clip(c.rect)
			r.DrawText(label, c.rect.X+6, c.rect.Y+(c.rect.H-lh1)/2, 1, theme.ButtonText)
			r.ClipEnd()
		}
	}

	// File list.
	r.FillRect(d.listRect, theme.Field)
	r.Clip(d.listRect)
	vr := d.visibleRows()
	for i := 0; i < vr; i++ {
		idx := d.scroll + i
		if idx >= len(d.entries) {
			break
		}
		e := d.entries[idx]
		rowRect := Rect{d.listRect.X, d.listRect.Y + i*d.rowH, d.listRect.W, d.rowH}
		if idx == d.sel {
			r.FillRect(rowRect, theme.SelFill)
		}
		col := theme.Text
		label := e.Name
		if e.IsDir {
			col = theme.DirText
			label = label + "/"
		} else if e.IsContainer {
			col = theme.DirText
			label = label + "  [bundle]"
		}
		nameW := rowRect.W
		if len(d.cols) > 0 {
			nameW = d.cols[0].w
		}
		r.DrawText(fit(r, label, nameW-10, sc), rowRect.X+6, rowRect.Y+(d.rowH-lh)/2, sc, col)
		for _, c := range d.cols {
			var s string
			switch c.kind {
			case colSize:
				if !e.IsDir && (e.Size > 0 || !e.ModTime.IsZero()) {
					s = humanSize(e.Size)
				}
			case colDate:
				if !e.ModTime.IsZero() {
					s = e.ModTime.Format(dateLayout)
				}
			case colInfo, colFolder:
				s = e.Detail
			}
			if s == "" {
				continue
			}
			s = fit(r, s, c.w-16, 1)
			x := rowRect.X + (c.x - d.listRect.X) + 8
			if c.kind == colSize {
				x = rowRect.X + (c.x - d.listRect.X) + c.w - 8 - r.TextWidth(s, 1)
			}
			r.DrawText(s, x, rowRect.Y+(d.rowH-lh1)/2, 1, theme.DimText)
		}
	}
	r.ClipEnd()
	r.StrokeRect(d.listRect, theme.Border, 1)

	// Preview pane (only while browsing inside a container).
	if d.inContainer() && d.previewRect.W > 0 {
		d.drawPreview(r, theme)
	}

	// The name field: always being edited.
	nr := d.nameRect
	r.FillRect(nr, theme.Field)
	r.StrokeRect(nr, theme.SelFill, 2)
	inner := d.nameInner()
	runes := []rune(d.name)
	start := d.ed.view(runes, inner.W, d.textW)
	d.ed.drawText(r, theme, inner, runes, start, sc, theme.Text, true)
	if len(runes) == 0 {
		hint := pick(d.cfg.Labels.OpenHint, "Type or paste a path")
		if d.cfg.Mode == ModeSave {
			hint = pick(d.cfg.Labels.SaveHint, "File name")
		}
		d.drawClipped(r, hint, inner.X+lh/4+2, inner.Y+(inner.H-lh)/2, inner.W-lh/4-2, theme.Disabled)
	}

	// The file-type selector.
	if d.hasSelector() {
		fr := d.filterRect
		r.FillRect(fr, theme.Button)
		r.StrokeRect(fr, theme.Border, 1)
		d.drawClipped(r, d.filterLabel()+" v", fr.X+8, fr.Y+(fr.H-lh)/2, fr.W-12, theme.ButtonText)
	}

	// Status line and buttons.
	st, stCol := d.statusLine(theme)
	if d.statusRect.W > 0 {
		r.Clip(d.statusRect)
		r.DrawText(fit(r, st, d.statusRect.W-4, 1), d.statusRect.X, d.statusRect.Y+(d.statusRect.H-lh1)/2, 1, stCol)
		r.ClipEnd()
	}
	d.drawButton(r, d.cancRect, pick(d.cfg.Labels.Cancel, "CANCEL"), true, theme)
	d.drawButton(r, d.okRect, d.okLabel(), d.canAccept(), theme)

	// The file-type list, over everything else.
	if len(d.popupRows) > 0 {
		r.FillRect(d.popupRect, theme.Panel)
		for i, pr := range d.popupRows {
			if i == d.popupSel {
				r.FillRect(pr, theme.SelFill)
			}
			col := theme.Text
			if i == d.view.Filter && d.custom == "" {
				col = theme.DirText
			}
			r.DrawText(d.sets[i].Name, pr.X+8, pr.Y+(pr.H-lh)/2, sc, col)
		}
		r.StrokeRect(d.popupRect, theme.Border, 1)
	}
}

// drawSwitch draws a small on/off row of the sidebar.
func (d *Dialog) drawSwitch(r Renderer, rec Rect, label string, on, enabled bool, theme Theme) {
	mark := "[ ] "
	if on {
		mark = "[x] "
	}
	col := theme.SideText
	if !enabled {
		col = theme.Disabled
	}
	r.DrawText(mark+label, rec.X+6, rec.Y+(rec.H-r.LineHeight(1))/2, 1, col)
}

func (d *Dialog) drawButton(r Renderer, rec Rect, label string, enabled bool, theme Theme) {
	r.FillRect(rec, theme.Button)
	r.StrokeRect(rec, theme.Border, 1)
	col := theme.ButtonText
	if !enabled {
		col = theme.Disabled
	}
	label = strings.ToUpper(label)
	sc := d.scale()
	lh := r.LineHeight(sc)
	r.DrawText(label, rec.X+(rec.W-r.TextWidth(label, sc))/2,
		rec.Y+(rec.H-lh)/2, sc, col)
}

// drawClipped draws s at (x,y) clipped to maxW pixels (no ellipsis; the host's
// Clip handles overflow). It clips to a 1-line-high band.
func (d *Dialog) drawClipped(r Renderer, s string, x, y, maxW int, c Colour) {
	if maxW <= 0 {
		return
	}
	sc := d.scale()
	r.Clip(Rect{x, y - 2, maxW, r.LineHeight(sc) + 4})
	r.DrawText(s, x, y, sc, c)
	r.ClipEnd()
}

// drawPreview renders the preview pane for the selected in-container entry: an
// optional thumbnail (from DialogConfig.Preview) above its metadata lines.
func (d *Dialog) drawPreview(r Renderer, theme Theme) {
	pr := d.previewRect
	r.FillRect(pr, theme.Field)
	r.StrokeRect(pr, theme.Border, 1)

	if d.sel < 0 || d.sel >= len(d.entries) {
		r.DrawText("select an entry", pr.X+6, pr.Y+6, 1, theme.DimText)
		return
	}
	e := d.entries[d.sel]
	pad := 8
	x := pr.X + pad
	y := pr.Y + pad

	// Thumbnail, if the host supplies one.
	if d.cfg.Preview != nil {
		if pv := d.cfg.Preview(d.container, e); pv != nil && pv.W > 0 && pv.H > 0 && len(pv.Pixels) >= pv.W*pv.H {
			// Scale to fit the pane width (and a capped height), integer zoom >= 1.
			maxW := pr.W - 2*pad
			maxH := pr.H/2 - pad
			zoom := maxW / pv.W
			if zh := maxH / pv.H; zh < zoom {
				zoom = zh
			}
			if zoom < 1 {
				zoom = 1
			}
			r.Clip(pr)
			for py := 0; py < pv.H; py++ {
				for px := 0; px < pv.W; px++ {
					c := pv.Pixels[py*pv.W+px]
					if c.A == 0 {
						continue
					}
					r.FillRect(Rect{x + px*zoom, y + py*zoom, zoom, zoom}, c)
				}
			}
			r.ClipEnd()
			y += pv.H*zoom + pad
		}
	}

	// Name + metadata lines.
	lh := r.LineHeight(1)
	r.Clip(pr)
	r.DrawText(e.Name, x, y, 1, theme.Text)
	y += lh + 4
	for _, m := range e.Meta {
		r.DrawText(m.Key+": "+m.Value, x, y, 1, theme.DimText)
		y += lh + 2
	}
	r.ClipEnd()
}
