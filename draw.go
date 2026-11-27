package zenui

// Theme controls the dialog's colours. The zero Theme is usable but drab; hosts
// usually set their own. DefaultTheme returns a reasonable dark scheme.
type Theme struct {
	Backdrop Colour // dim layer over the rest of the screen
	Panel    Colour // dialog background
	Sidebar  Colour // favourites sidebar background
	SideText Colour // favourites label
	Border   Colour
	Text     Colour
	DimText  Colour
	DirText  Colour // directory entries
	SelFill  Colour // selected row background
	SelText  Colour // text colour drawn over SelFill -- kept separate from Text since a theme's normal text colour doesn't always contrast against its own selection fill (e.g. Light theme's dark text over its medium-blue selection)
	// BarSelFill/BarSelText are the open top-bar label's background/text
	// -- separate from SelFill/SelText since a theme's dropdown-item
	// selection colour doesn't always suit the bar's own "which menu is
	// open" indicator too (SpectrumTheme's real hardware never uses
	// its cyan highlight on the title strip itself, only inside the
	// menu box). Defaults to matching SelFill/SelText for a theme with
	// no reason to distinguish the two.
	BarSelFill Colour
	BarSelText Colour
	// BorderThickness is the dropdown/dialog border's stroke width, in
	// device pixels -- a flat value, not scaled by a caller's zoom
	// level, so a theme that only looks right thick (or thin) stays
	// that way regardless of the Font menu's zoom setting.
	BorderThickness int
	// DropdownBorderSkipTop omits the top edge of a Menu's border when
	// true -- for a theme whose dropdowns always open directly below
	// the bar strip and shouldn't read as visually separated from it
	// (SpectrumTheme; the real menu box's top edge isn't a border
	// line, it's where the black title strip already ends).
	DropdownBorderSkipTop bool
	// ItemPadYPercent/ItemPadXPercent are a dropdown item row's
	// vertical/horizontal padding, as a percentage of the active line
	// height -- e.g. 25 means padY = lineHeight/4. Every preset sets
	// these explicitly (there's no zero-means-default fallback here,
	// unlike Scale/BorderThickness, since zero padding is itself a
	// meaningful value a theme could legitimately want).
	ItemPadYPercent int
	ItemPadXPercent int
	// ShowBarRainbow signals that the host should draw the classic
	// Sinclair diagonal-stripe rainbow on the right side of the bar
	// strip, space permitting -- a raylib-specific decoration (see
	// menubar_gui.go's drawBarRainbow), so this package only carries
	// the on/off signal, never draws it itself.
	ShowBarRainbow bool
	// DropdownBottomPadding is extra space, in device pixels, added
	// below the last item row and before the panel's own bottom border
	// -- distinct from ItemPadYPercent, which governs the padding
	// between/around every row uniformly. Zero (every preset except
	// SpectrumTheme) means the panel's bottom edge sits exactly at
	// the last row's own bottom edge, the same as before this field
	// existed.
	DropdownBottomPadding int
	// UseBarGradient signals that the bar's background is a vertical
	// gradient (BarGradientTop at the top edge fading to
	// BarGradientBottom at the bottom) rather than a flat Sidebar
	// fill, drawn by MenuBar.Draw itself via FillRectGradientV.
	// Sidebar is unused for a theme with this set.
	UseBarGradient    bool
	BarGradientTop    Colour
	BarGradientBottom Colour
	// UseGradientOverlay, if true, layers a second, horizontal
	// gradient on top of the vertical one (left, right: only
	// meaningful alongside UseBarGradient), composited with multiply
	// blend mode -- darkening the base gradient toward whichever end
	// is closer to black, leaving it unchanged wherever this overlay
	// is white. Covers the region from the bar's last label to the
	// screen's own right edge (the same region ShowBarRainbow/
	// ShowZSPLogo draw into), drawn immediately after the base
	// gradient fill and before labels/rainbow/logo, so those stay at
	// full brightness rather than also being darkened.
	UseGradientOverlay   bool
	GradientOverlayLeft  Colour
	GradientOverlayRight Colour
	// ShowZSPLogo signals the host's other project logo (Zen Spectrum
	// Project) should be drawn on the right side of the bar, in the
	// same slot ShowBarRainbow uses -- the two are mutually exclusive
	// in practice (SpectrumTheme uses the rainbow, Dark/Light use
	// this), never checked together.
	ShowZSPLogo bool
	// CheckboxEmptyColour is an unchecked checkbox's own outline
	// colour -- always distinct from the item's own computed text/
	// selection colour, since an empty checkbox represents "not yet
	// set" rather than active content, and reads better in a
	// consistent neutral tone regardless of theme. Applies uniformly
	// across every theme, not gated by a bool -- there's no case
	// where "use whatever the item's text colour happens to be"
	// reads better for an empty box.
	CheckboxEmptyColour Colour
	// UseCheckmarkColour, if true, means CheckmarkColour overrides a
	// checkmark item's (Toggle=false, checked) tick mark colour,
	// instead of the item's own computed text/selection colour -- a
	// checkmark indicating "the current setting" reads better as a
	// stable colour that doesn't shift with hover. false (every theme
	// except Spectrum) means "use the item's own computed colour",
	// the behaviour every theme had before this field existed.
	UseCheckmarkColour bool
	CheckmarkColour    Colour
	// SeparatorColour draws a Menu separator's own divider line --
	// applies uniformly across every theme, same as CheckboxEmptyColour,
	// no bool gate needed since every theme sets this to a sensible
	// value (usually the same as Border, but not required to be).
	SeparatorColour Colour
	// UseCheckboxColour, if true, means CheckboxColour overrides a
	// checkbox item's (Toggle=true, checked) box outline and cross
	// colour, instead of the item's own computed text/selection
	// colour -- the same "fixed colour, doesn't shift with hover"
	// reasoning UseCheckmarkColour already has, for the checked state
	// specifically (CheckboxEmptyColour already covers the unchecked
	// state unconditionally). false (every theme except Spectrum)
	// means "use the item's own computed colour", the behaviour every
	// theme had before this field existed.
	UseCheckboxColour bool
	CheckboxColour    Colour
	Field             Colour // filename field background
	Button            Colour
	ButtonHot         Colour
	ButtonText        Colour
	Disabled          Colour
}

// DefaultTheme is a dark scheme that fits most editors.
func DefaultTheme() Theme {
	return Theme{
		Backdrop:              Colour{0, 0, 0, 0xa0},
		Panel:                 Colour{0x20, 0x20, 0x2a, 0xff},
		Sidebar:               Colour{0x18, 0x18, 0x20, 0xff},
		SideText:              Colour{0xc0, 0xc0, 0xd0, 0xff},
		Border:                Colour{0x6a, 0x6a, 0x78, 0xff},
		Text:                  Colour{0xe0, 0xe0, 0xe8, 0xff},
		DimText:               Colour{0x90, 0x90, 0xa0, 0xff},
		DirText:               Colour{0xff, 0xd0, 0x60, 0xff},
		SelFill:               Colour{0x30, 0x50, 0x80, 0xff},
		SelText:               Colour{0xf8, 0xf8, 0xfc, 0xff},
		BarSelFill:            Colour{0x30, 0x50, 0x80, 0xff},
		BarSelText:            Colour{0xf8, 0xf8, 0xfc, 0xff},
		BorderThickness:       1,
		ItemPadYPercent:       25, // matches the previous fixed lh/4 formula
		ItemPadXPercent:       50, // matches the previous fixed lh/2 formula
		ShowBarRainbow:        false,
		DropdownBottomPadding: 0,
		UseBarGradient:        false,
		UseGradientOverlay:    false,
		ShowZSPLogo:           false,
		CheckboxEmptyColour:   Colour{0x80, 0x80, 0x80, 0xff}, // neutral grey
		SeparatorColour:       Colour{0x6a, 0x6a, 0x78, 0xff}, // same as Border
		UseCheckboxColour:     false,
		Field:                 Colour{0x14, 0x14, 0x1c, 0xff},
		Button:                Colour{0x30, 0x30, 0x40, 0xff},
		ButtonHot:             Colour{0x44, 0x44, 0x58, 0xff},
		ButtonText:            Colour{0xf0, 0xf0, 0xf8, 0xff},
		Disabled:              Colour{0x50, 0x50, 0x5a, 0xff},
	}
}
