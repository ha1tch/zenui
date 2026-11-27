package zenui

import "strings"

// Update advances the dialog by one frame of input. It must be called after at
// least one Draw so the layout rects exist; New performs no layout, so on the
// very first frame call Draw first (the typical loop draws every frame anyway).
// Returns the resulting Status.
//
// The name field is always being edited: typed text, paste, the caret keys, and
// Ctrl+A, Ctrl+C and Ctrl+X all work in it, as in a TextField. Up, Down, PageUp and
// PageDown move through the list; Enter opens or accepts; Tab completes the name;
// Ctrl+Tab changes the file type; Ctrl+H shows or hides the hidden files; Ctrl+D
// switches the details; Ctrl+L selects the whole field; Alt+Up goes to the parent
// folder; Escape drops a question being asked, and then cancels.
func (d *Dialog) Update(in Input) Status {
	if d.status != Active {
		return d.status
	}
	d.clock += float64(in.DeltaTime)

	if in.pressed(KeyEscape) {
		switch {
		case d.popup:
			d.popup = false
		case d.confirm != "":
			d.confirm = ""
		default:
			d.status = Cancelled
		}
		return d.status
	}
	if d.popup {
		d.updatePopup(in)
		return d.status
	}

	d.shortcuts(in)
	d.editName(in)
	if in.pressed(KeyTab) {
		if in.Ctrl {
			if in.Shift {
				d.cycleFilter(-1)
			} else {
				d.cycleFilter(+1)
			}
		} else {
			d.complete()
		}
	}
	if in.pressed(KeyEnter) {
		d.enter()
		return d.status
	}
	if in.Alt {
		if in.pressed(KeyUp) {
			d.goUp()
		}
	} else {
		if in.pressed(KeyUp) {
			d.moveSel(-1)
		}
		if in.pressed(KeyDown) {
			d.moveSel(+1)
		}
		if in.pressed(KeyPageUp) {
			d.moveSel(-d.visibleRows())
		}
		if in.pressed(KeyPageDown) {
			d.moveSel(+d.visibleRows())
		}
	}

	// Wheel scrolling over the list.
	if in.WheelY != 0 && d.listRect.Contains(in.MouseX, in.MouseY) {
		d.scroll -= int(in.WheelY)
		d.clampScroll()
	}

	d.mouse(in)
	return d.status
}

// shortcuts are the keys that are not editing the name.
func (d *Dialog) shortcuts(in Input) {
	if !in.Ctrl {
		return
	}
	if in.ctrlChar('h') {
		d.toggleHidden()
	}
	if in.ctrlChar('d') {
		d.toggleDetails()
	}
	if in.ctrlChar('l') {
		d.ed.selectAll(len([]rune(d.name)))
	}
}

// editName applies the keys, typed runes and paste of the frame to the field.
func (d *Dialog) editName(in Input) {
	text, changed := d.ed.edit(d.name, in, editLimits{})
	if !changed {
		return
	}
	d.name = text
	d.typed = true
	d.err, d.hint, d.confirm = "", "", ""
	d.typeAhead()
}

// enter does what Enter does: goes into a selected folder, opens or accepts what
// the field names, or what is selected when the field only echoes the selection.
func (d *Dialog) enter() {
	text := strings.TrimSpace(d.name)
	if d.cfg.Mode == ModeOpen && !(d.typed && text != "") {
		if d.sel >= 0 && d.sel < len(d.entries) {
			if e := d.entries[d.sel]; e.IsDir || e.IsContainer {
				d.openSelection()
				return
			}
			d.commit()
		}
		return
	}
	if text == "" {
		if d.sel >= 0 && d.sel < len(d.entries) {
			if e := d.entries[d.sel]; e.IsDir || e.IsContainer {
				d.openSelection()
			}
		}
		return
	}
	d.commit()
}

// updatePopup handles a frame while the file-type list is open.
func (d *Dialog) updatePopup(in Input) {
	n := len(d.sets)
	if n == 0 {
		d.popup = false
		return
	}
	if in.pressed(KeyUp) {
		d.popupSel = (d.popupSel - 1 + n) % n
	}
	if in.pressed(KeyDown) {
		d.popupSel = (d.popupSel + 1) % n
	}
	if in.pressed(KeyEnter) {
		d.setFilter(d.popupSel)
		d.popup = false
		return
	}
	if in.MousePressed {
		for i, rr := range d.popupRows {
			if rr.Contains(in.MouseX, in.MouseY) {
				d.setFilter(i)
				break
			}
		}
		d.popup = false
	}
}

// dblLimit is the longest time between two clicks that make a double click.
func (d *Dialog) dblLimit() float64 {
	switch s := d.cfg.DoubleClickSeconds; {
	case s == 0:
		return 0.5
	case s < 0:
		return 1e18
	default:
		return s
	}
}

// mouse handles the pointer: the field (caret, selection, drag), the switches, the
// column heads, the sidebar, the list and the buttons.
func (d *Dialog) mouse(in Input) {
	if d.textW != nil && d.ed.pointer([]rune(d.name), in, d.nameRect, d.nameInner(), d.textW) {
		d.hint = ""
		return
	}
	if !in.MousePressed {
		return
	}
	x, y := in.MouseX, in.MouseY
	switch {
	case d.hasSelector() && d.filterRect.Contains(x, y):
		d.popup = true
		d.popupSel = d.view.Filter
	case d.hiddenRect.Contains(x, y) && d.hiddenSupported():
		d.toggleHidden()
	case d.detailsRect.Contains(x, y):
		d.toggleDetails()
	case d.clickHead(x, y):
		// sorted by the column
	case d.clickPlace(x, y):
		// handled: navigated to a favourite
	case d.upRect.Contains(x, y):
		d.goUp()
	case d.okRect.Contains(x, y):
		if d.canAccept() {
			d.accept()
		}
	case d.cancRect.Contains(x, y):
		d.status = Cancelled
	case d.listRect.Contains(x, y):
		row := (y - d.listRect.Y) / d.rowH
		idx := d.scroll + row
		if idx >= 0 && idx < len(d.entries) {
			double := idx == d.lastIdx && d.clock-d.lastClick <= d.dblLimit()
			d.lastIdx, d.lastClick = idx, d.clock
			if idx == d.sel && double {
				d.lastIdx = -1
				d.openSelection() // second click on the same row = open/enter
			} else {
				d.selectIndex(idx)
			}
		}
	case !d.bounds.Contains(x, y):
		d.status = Cancelled // click outside dismisses
	}
}

// clickHead sorts by the column whose head is at (x,y).
func (d *Dialog) clickHead(x, y int) bool {
	if !d.headRect.Contains(x, y) {
		return false
	}
	for _, c := range d.cols {
		if c.rect.Contains(x, y) && c.sortable() {
			d.sortBy(c.key)
		}
	}
	return true
}

// clickPlace navigates to a favourite if (x,y) hits one of the sidebar rows.
// Returns true if a place was clicked.
func (d *Dialog) clickPlace(x, y int) bool {
	for i, pr := range d.placeRects {
		if pr.Contains(x, y) && i < len(d.places) {
			if d.places[i].Path == recentPlace {
				d.showRecent()
				return true
			}
			d.setDir(d.places[i].Path)
			if d.err == "" && d.cfg.Mode == ModeOpen {
				d.setName("", false)
			}
			return true
		}
	}
	return false
}

func (d *Dialog) moveSel(delta int) {
	if len(d.entries) == 0 {
		return
	}
	i := d.sel
	if i < 0 {
		i = 0
	} else {
		i += delta
	}
	d.selectIndex(clampInt(i, 0, len(d.entries)-1))
}

// ensureVisible scrolls the list so that the selected row shows.
func (d *Dialog) ensureVisible() {
	if d.sel < 0 {
		return
	}
	if d.sel < d.scroll {
		d.scroll = d.sel
	}
	if vr := d.visibleRows(); vr > 0 && d.sel >= d.scroll+vr {
		d.scroll = d.sel - vr + 1
	}
}

func (d *Dialog) clampScroll() {
	max := len(d.entries) - d.visibleRows()
	if max < 0 {
		max = 0
	}
	if d.scroll > max {
		d.scroll = max
	}
	if d.scroll < 0 {
		d.scroll = 0
	}
}
