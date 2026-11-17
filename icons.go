package zenui

import _ "embed"

// IconsBDF is the bitmap font of the toolbar icons: sixteen 8 by 8 pictograms
// in the private use area, drawn for this toolkit. A host builds a renderer over
// it, as zenimate does for its tool palette, and gives it to a List or a
// ButtonBar as Icons.
//
//go:embed icons.bdf
var IconsBDF []byte

// The glyphs of IconsBDF.
const (
	IconAdd       rune = 0xE200
	IconDuplicate rune = 0xE201
	IconDelete    rune = 0xE202
	IconRows      rune = 0xE203
	IconColumns   rune = 0xE204
	IconUndo      rune = 0xE205
	IconRedo      rune = 0xE206
	IconPen       rune = 0xE207
	IconMove      rune = 0xE208
	IconFlipH     rune = 0xE209
	IconFlipV     rune = 0xE20A
	IconRotate    rune = 0xE20B
	IconGrid      rune = 0xE20C
	IconFill      rune = 0xE20E
	IconPick      rune = 0xE20F
	IconClear     rune = 0xE20D
)
