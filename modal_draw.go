package zenui

// layout places the window: one frame, centred on screen, holding the form's
// fields and, beneath them, the row of buttons. The form is told to leave room
// for the row inside its own frame and to be wide enough for it, so that the
// buttons are part of the window and not boxes floating below it.
func (m *Modal) layout(r Renderer, screenW, screenH int, theme Theme) {
	lh := r.LineHeight(m.scale)
	gridSize := m.form.gridSize
	pad := gridSize / 2
	buttonH := lh + 2*pad
	gap := gridSize

	widths := make([]int, len(m.cfg.Buttons))
	rowW := 0
	for i, label := range m.cfg.Buttons {
		widths[i] = r.TextWidth(label, m.scale) + 2*pad
		rowW += widths[i]
	}
	if n := len(widths); n > 1 {
		rowW += (n - 1) * (gap / 2)
	}

	m.form.cfg.Anchor = Rect{}
	m.form.cfg.ExtraBottom = gap + buttonH
	m.form.cfg.MinWidth = rowW + 2*pad
	m.form.layout(r, screenW, screenH)
	frame := m.form.bounds

	buttonRowY := frame.Y + frame.H - pad - buttonH
	m.buttonRects = m.buttonRects[:0]
	x := frame.X + frame.W - pad
	for i := len(widths) - 1; i >= 0; i-- {
		x -= widths[i]
		m.buttonRects = append(m.buttonRects, Rect{X: x, Y: buttonRowY, W: widths[i], H: buttonH})
		x -= gap / 2
	}
	// The buttons were placed right to left; put them in the order of
	// cfg.Buttons.
	for i, j := 0, len(m.buttonRects)-1; i < j; i, j = i+1, j-1 {
		m.buttonRects[i], m.buttonRects[j] = m.buttonRects[j], m.buttonRects[i]
	}
	m.bounds = frame
}

// Draw lays out and renders the modal: the backdrop, then one framed window
// holding the form and the button row. Call
// it before Update each frame -- the same convention every zenui widget
// uses.
func (m *Modal) Draw(r Renderer, screenW, screenH int, theme Theme) {
	if m.status != Active {
		return
	}
	m.layout(r, screenW, screenH, theme)

	r.FillRect(Rect{X: 0, Y: 0, W: screenW, H: screenH}, theme.Backdrop)

	m.form.Draw(r, screenW, screenH, theme)

	for i, label := range m.cfg.Buttons {
		rec := m.buttonRects[i]
		bg := theme.Button
		r.FillRect(rec, bg)
		drawBorder(r, rec, theme.Border, 1, false)
		lh := r.LineHeight(m.scale)
		tw := r.TextWidth(label, m.scale)
		r.DrawText(label, rec.X+(rec.W-tw)/2, rec.Y+(rec.H-lh)/2, m.scale, theme.ButtonText)
	}
}

// Update advances the modal's state: delegates to the wrapped Form first
// (text typing, checkbox toggling, tab), then handles button clicks and
// Escape (if CancelButtonIndex is set).
func (m *Modal) Update(in Input) Status {
	if m.status != Active {
		return m.status
	}

	m.form.Update(in)

	if in.pressed(KeyEscape) && m.cfg.CancelButtonIndex >= 0 {
		m.result = m.cfg.CancelButtonIndex
		m.status = Cancelled
		return m.status
	}

	if in.MousePressed {
		for i, rec := range m.buttonRects {
			if rec.Contains(in.MouseX, in.MouseY) {
				m.result = i
				if m.cfg.CancelButtonIndex >= 0 && i == m.cfg.CancelButtonIndex {
					m.status = Cancelled
				} else {
					m.status = Accepted
				}
				return m.status
			}
		}
	}

	return m.status
}
