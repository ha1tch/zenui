package zenui

import "strings"

const (
	textModalBodyScale = 2
	textModalCols      = 70 // preferred panel width, in characters
)

// TextPanelWidth returns the panel width for a 70-column body at
// the given text scale: the measured text width plus horizontal
// padding and the scrollbar gutter.
func TextPanelWidth(r Renderer, scale int) int {
	pad := r.LineHeight(1)
	cols := strings.Repeat("M", textModalCols)
	return r.TextWidth(cols, scale) + 2*pad + 12
}

// TextBodyScale picks the body text scale: the base scale if a
// 70-column panel fits within the screen width at that scale,
// otherwise 1. Bitmap fonts only look crisp at whole-number scales.
func TextBodyScale(r Renderer, screenW int) int {
	pad := r.LineHeight(1)
	if TextPanelWidth(r, textModalBodyScale) <= screenW-2*pad {
		return textModalBodyScale
	}
	return 1
}

// TextModal is a scrollable text reader for a small, lightweight
// markdown subset -- "# H1"/"## H2" headings, indented lines (two
// leading spaces, dimmer text, for shortcut tables), everything else
// plain body text. Shown over the host's picture with a dim backdrop,
// dismissed via Escape, its own close box, or a click outside the
// panel. It scrolls with the arrows, Page Up and Down, Home and End,
// the wheel, and a scrollbar that drags. With WithLinks, a "[text](address)"
// and a bare http:// or https:// address in the text are links that a click opens.
//
// It began as zenimate's help window and was generalised in zenzx to
// take its title and content and the active Theme, so that Help and
// About share one implementation; it is here so that every host does.
type TextModal struct {
	title  string
	lines  []string
	scroll int // first visible line index

	// autoHeight, if true, sizes the panel to fit every line without
	// scrolling (About: short, fixed-length content, where a
	// full-height panel would leave most of it empty) rather than
	// always filling most of the screen (Help: long enough that it
	// always needs scrolling regardless, and screen-filling gives the
	// most room for that).
	autoHeight bool

	panel    Rect
	body     Rect // text area (inside the panel, below the title)
	closeBx  Rect
	visible  int // lines that fit in the body (set during layout)
	total    int
	bodyLH   int // body line height at the effective scale (set during layout)
	effScale int // effective body text scale for the current screen width

	// scrollbar drag state
	track      Rect
	thumb      Rect
	dragging   bool
	dragOffset int // pointer offset within the thumb at grab time

	mouseX, mouseY int // the pointer at the last Update, for the close box's hover

	// links: openLink, when the host asked for links (WithLinks), is called with the
	// address of a link that is clicked; links is where the links were drawn.
	openLink func(url string)
	links    []linkRect
}

// NewTextModal makes a reader for content under title. With autoHeight the
// panel is as tall as the content (short, fixed text: About); without, it
// fills most of the screen (long text that scrolls anyway: Help).
func NewTextModal(title, content string, autoHeight bool) *TextModal {
	return &TextModal{
		title:      title,
		lines:      strings.Split(strings.TrimRight(content, "\n"), "\n"),
		autoHeight: autoHeight,
	}
}

// Title returns the modal's title.
func (h *TextModal) Title() string { return h.title }

// Lines returns the text, a line an element. The slice is the modal's own.
func (h *TextModal) Lines() []string { return h.lines }

// AutoHeight reports whether the panel is sized to its text.
func (h *TextModal) AutoHeight() bool { return h.autoHeight }

// markdown line classification for the lightweight markdown this reader
// understands.
type textLineKind int

const (
	textLineBody     textLineKind = iota // ordinary prose
	textLineIndented                     // indented (shortcut tables etc.)
	textLineH1                           // "# " heading
	textLineH2                           // "## " heading
)

// classifyTextLine classifies a line and returns the text to draw
// (with any markdown heading marker stripped).
func classifyTextLine(line string) (textLineKind, string) {
	switch {
	case strings.HasPrefix(line, "## "):
		return textLineH2, line[3:]
	case strings.HasPrefix(line, "# "):
		return textLineH1, line[2:]
	case strings.HasPrefix(line, "  "):
		return textLineIndented, line
	default:
		return textLineBody, line
	}
}

// DrawTextLine draws one line of the markdown subset TextModal reads, at
// x, y: a "# " or "## " heading without its marker, in the theme's accent
// colour and bold (a one-pixel horizontal double-strike); a line that
// begins with two spaces dimmer; any other in the text colour. A host
// that lays a window of its own out in lines draws them with this, so
// that its text looks like a TextModal's.
func DrawTextLine(r Renderer, line string, x, y, scale int, theme Theme) {
	switch kind, text := classifyTextLine(line); kind {
	case textLineH1, textLineH2:
		r.DrawText(text, x, y, scale, theme.DirText)
		r.DrawText(text, x+1, y, scale, theme.DirText)
	case textLineIndented:
		r.DrawText(text, x, y, scale, theme.DimText)
	default:
		r.DrawText(text, x, y, scale, theme.Text)
	}
}

func (h *TextModal) layout(r Renderer, screenW, screenH int) {
	lh := r.LineHeight(1)
	pad := lh
	h.effScale = TextBodyScale(r, screenW)

	pw := TextPanelWidth(r, h.effScale)
	if minW := 320; pw < minW {
		pw = minW
	}
	if pw > screenW-2*pad {
		pw = screenW - 2*pad
	}

	titleH := r.LineHeight(2) + 8
	h.bodyLH = r.LineHeight(h.effScale)
	if h.bodyLH < 1 {
		h.bodyLH = 1
	}
	h.total = len(h.lines)

	var ph int
	if h.autoHeight {
		// Estimate the content's own height (every line at the body's
		// line height) and size the panel to fit it exactly, rather
		// than always filling most of the screen -- correct as an
		// estimate rather than a guaranteed exact fit, since actual
		// text rendering could in principle wrap or round slightly
		// differently, but line-count * line-height is what every
		// other size in this layout is already derived from too.
		contentH := h.total * h.bodyLH
		ph = contentH + 2*pad + titleH
		if ph > screenH-2*pad {
			ph = screenH - 2*pad // still cap to the screen, in case content somehow doesn't fit
		}
	} else {
		ph = screenH - 4*pad
	}
	if ph < lh*10 {
		ph = lh * 10
	}
	px := (screenW - pw) / 2
	py := (screenH - ph) / 2
	h.panel = Rect{X: px, Y: py, W: pw, H: ph}

	cb := r.LineHeight(2)
	h.closeBx = Rect{X: px + pw - pad - cb, Y: py + pad, W: cb, H: cb}

	h.body = Rect{
		X: px + pad,
		Y: py + pad + titleH,
		W: pw - 2*pad - 12, // leave room for the scrollbar on the right
		H: ph - 2*pad - titleH,
	}
	h.visible = h.body.H / h.bodyLH
	if h.visible < 1 {
		h.visible = 1
	}
	h.clampScroll()
}

func (h *TextModal) clampScroll() {
	maxScroll := h.total - h.visible
	if maxScroll < 0 {
		maxScroll = 0
	}
	if h.scroll > maxScroll {
		h.scroll = maxScroll
	}
	if h.scroll < 0 {
		h.scroll = 0
	}
}

// Update handles scrolling and dismissal. It returns false when the modal
// should close.
func (h *TextModal) Update(in Input) bool {
	// A click on a link goes to the host and leaves the modal as it is. It is tried first, against
	// where the last Draw put the links, before this frame's scrolling moves them.
	if in.MousePressed {
		if u := h.linkAt(in.MouseX, in.MouseY); u != "" {
			h.mouseX, h.mouseY = in.MouseX, in.MouseY
			h.openLink(u)
			return true
		}
	}
	for _, k := range in.Keys {
		switch k {
		case KeyEscape:
			return false
		case KeyUp:
			h.scroll--
		case KeyDown:
			h.scroll++
		case KeyPageUp:
			h.scroll -= h.visible
		case KeyPageDown:
			h.scroll += h.visible
		case KeyHome:
			h.scroll = 0
		case KeyEnd:
			h.scroll = h.total
		}
	}
	h.mouseX, h.mouseY = in.MouseX, in.MouseY
	if in.WheelY != 0 {
		h.scroll -= int(in.WheelY) * 3
	}

	if in.MousePressed && h.thumb.W > 0 {
		if h.thumb.Contains(in.MouseX, in.MouseY) {
			h.dragging = true
			h.dragOffset = in.MouseY - h.thumb.Y
		} else if h.track.Contains(in.MouseX, in.MouseY) {
			if in.MouseY < h.thumb.Y {
				h.scroll -= h.visible
			} else {
				h.scroll += h.visible
			}
		}
	}
	if h.dragging {
		if !in.MouseDown {
			h.dragging = false
		} else if h.track.H > h.thumb.H {
			rel := float32(in.MouseY-h.dragOffset-h.track.Y) / float32(h.track.H-h.thumb.H)
			if rel < 0 {
				rel = 0
			}
			if rel > 1 {
				rel = 1
			}
			h.scroll = int(rel * float32(h.total-h.visible))
		}
	}
	h.clampScroll()

	if in.MousePressed {
		if h.closeBx.Contains(in.MouseX, in.MouseY) {
			return false
		}
		if !h.panel.Contains(in.MouseX, in.MouseY) {
			return false
		}
	}
	return true
}

// ScaleAlpha returns c with its alpha multiplied by factor -- used to
// adjust a window's own panel/backdrop opacity relative to the active
// theme's own values, for that window alone, rather than changing
// theme.Panel/theme.Backdrop themselves, which every other dialog
// (Dialog, Modal, MessageBox) also draws with.
func ScaleAlpha(c Colour, factor float64) Colour {
	c.A = uint8(float64(c.A) * factor)
	return c
}

const (
	// ModalPanelOpacity is the opacity of a translucent window's client
	// area (TextModal's, and a host's own windows in the same style), independent
	// of whatever alpha theme.Panel itself carries (always 0xff/fully
	// opaque across every current theme) -- text and widgets drawn on
	// top afterwards, at their own full opacity, still render fully
	// solid regardless, since each draw call blends independently
	// against whatever's already there.
	ModalPanelOpacity = 0.85
	// ModalBackdropDarkening scales the backdrop's own alpha down by
	// 20%, relative to whatever value the active theme already uses
	// (which varies by theme) -- less alpha means less of the dimming
	// colour shows through, i.e. a less dark overlay.
	ModalBackdropDarkening = 0.8
)

// Draw lays the modal out for the screen and draws it in theme's colours.
func (h *TextModal) Draw(r Renderer, screenW, screenH int, theme Theme) {
	h.layout(r, screenW, screenH)
	lh := r.LineHeight(1)
	pad := lh

	r.FillRect(Rect{X: 0, Y: 0, W: screenW, H: screenH}, ScaleAlpha(theme.Backdrop, ModalBackdropDarkening))
	r.FillRect(h.panel, ScaleAlpha(theme.Panel, ModalPanelOpacity))
	r.StrokeRect(h.panel, theme.Border, 1)

	r.DrawText(h.title, h.panel.X+pad, h.panel.Y+pad, 2, theme.Text)

	mx, my := h.mouseX, h.mouseY // where the last Update saw the pointer
	cbg := theme.Button
	if h.closeBx.Contains(mx, my) {
		cbg = theme.ButtonHot
	}
	r.FillRect(h.closeBx, cbg)
	r.StrokeRect(h.closeBx, theme.Border, 1)
	r.DrawText("x", h.closeBx.X+(h.closeBx.W-r.TextWidth("x", 1))/2,
		h.closeBx.Y+(h.closeBx.H-lh)/2, 1, theme.ButtonText)

	r.Clip(h.body)
	y := h.body.Y
	end := h.scroll + h.visible
	if end > h.total {
		end = h.total
	}
	h.links = h.links[:0]
	for i := h.scroll; i < end; i++ {
		if h.openLink != nil {
			h.links = append(h.links, drawLinkedLine(r, h.lines[i], h.body.X, y, h.effScale, h.bodyLH, theme, mx, my)...)
		} else {
			DrawTextLine(r, h.lines[i], h.body.X, y, h.effScale, theme)
		}
		y += h.bodyLH
	}
	r.ClipEnd()

	if h.total > h.visible {
		h.track = Rect{X: h.body.X + h.body.W + 6, Y: h.body.Y, W: 8, H: h.body.H}
		r.FillRect(h.track, theme.Button)
		frac := float32(h.visible) / float32(h.total)
		thumbH := int(float32(h.track.H) * frac)
		if thumbH < 16 {
			thumbH = 16
		}
		var pos float32
		if h.total-h.visible > 0 {
			pos = float32(h.scroll) / float32(h.total-h.visible)
		}
		thumbY := h.track.Y + int(float32(h.track.H-thumbH)*pos)
		h.thumb = Rect{X: h.track.X, Y: thumbY, W: h.track.W, H: thumbH}
		thumbCol := theme.DirText
		if h.dragging {
			thumbCol = theme.Text
		}
		r.FillRect(h.thumb, thumbCol)
	} else {
		h.track = Rect{}
		h.thumb = Rect{}
	}
}
