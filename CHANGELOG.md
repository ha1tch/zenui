# Changelog

All notable changes to zenui are recorded here, newest first.

## [0.3.4] - 2026-11-24

A menu row can have a picture.

### Added

- `Item.Icon`: a menu row can draw a small picture before its label. It is a
  function that is given the square cell (one line of text) to fill and the
  colour the label is drawn in, which follows the row being selected or
  disabled. The cell is after the checkmark's gutter and before the label; a
  menu reserves the room only if one of its rows has an icon, as it does for
  `Checked`, so a menu with none draws as before. Not drawn on a separator or
  a title.

## [0.3.3] - 2026-11-23

The ZSP logo shows the Spectrum's own colours now and then.

### Changed

- After every ten quartets of random colours the logo shows a fixed one
  instead: the four bands of the Spectrum's rainbow in their order, red,
  yellow, green and cyan, one for each tooth from the left, in the palette's
  own pastel colours (unchanged). It is faded to as any other quartet is, holds
  for three seconds instead of one, fades to the next random quartet, and ten
  more random ones follow. A run is ten cycles of a second and a half, then the
  Spectrum's quartet's three seconds and half-second fade: eighteen and a
  half seconds in all.
- No tooth keeps its colour into the Spectrum's quartet or out of it: the last
  random quartet before it is picked to keep off its colours as well as its own
  predecessor's.
- `MenuBar` tracks the seconds into the current cycle itself, so that a cycle
  can be as long as its quartet needs; a frame longer than a cycle is still one
  quartet on, not several.

## [0.3.2] - 2026-11-23

The ZSP logo fades from one quartet of colours to the next.

### Changed

- The logo's colours no longer switch all at once every second. A quartet
  holds still for a second, then the four teeth fade together, over half a
  second, to the next quartet (a straight blend, tooth by tooth), which holds
  in its turn; so a cycle is a second and a half. The next quartet is picked
  as the one before it is shown, and as before no tooth keeps its colour from
  one quartet to the next.
- A frame longer than a cycle is one step on, not several.

## [0.3.1] - 2026-11-23

The ZSP logo takes a random colour for each of its four teeth.

### Changed

- The logo at the right of the bar (the Dark and Light themes) no longer
  turns through three fixed pairs of colours. Each of its four teeth is now one
  colour of the palette, picked at random, and once every second all four are
  picked again; a tooth never keeps the colour it had, and two teeth can have
  the same one.
- The palette is the Spectrum's eight colours at the normal level (`0xC8`),
  taken half of the way to white and then a tenth of the way to their own
  grey: a pastel red, green, blue, and so on. Black, which lightening leaves
  dark, is 70% white (`0xB3`). White is a very light grey, and shows little
  on the Light theme's bar.
- `MenuBar` keeps the quartet it is showing, so the colours hold from frame to
  frame and change only when a rotation period has passed.

## [0.3.0] - 2026-11-23

Links in the text of a `TextModal` can be clicked.

### Added

- `TextModal.WithLinks(open)`: a `[text](address)` and a bare `http://` or
  `https://` address in the modal's text become links, drawn in the theme's
  accent colour and underlined (bold under the pointer), and a click on one
  calls `open` with its address and leaves the modal as it is. The host opens
  the address, so that this package opens nothing. It is opt-in, for text the
  host wrote itself: a modal made with `NewTextModal` alone has no links, so
  that a file name or a tape's block name that looks like an address is never
  one. Only the web schemes are links, in either form. `TextModal.OverLink`
  says whether the pointer is over one, for a host's pointing-hand cursor.
- `WrapLinkedText`: `WrapText` for text with `[text](address)` links. It
  counts only what is seen, breaks only at spaces, and carries a link over a
  break by closing it on the one line and opening it again, to the same
  address, on the next.

## [0.2.1] - 2026-11-21

A submenu with no room on the right of its parent now opens to the left of it.

### Fixed

- A submenu that did not fit to the right of its parent was placed with its
  right edge on the parent's right edge, which put it over the parent's own
  rows. It now opens beside the parent on the left, its right edge on the
  parent's left edge, and covers none of the menu that owns it. A submenu
  that fits on the right is placed as before, and one with room on neither
  side is still clamped inside the screen.

## [0.2.0] - 2026-11-16

What zenzx's own copy of the toolkit gained after the extraction, so that
zenzx can move onto this module (T-02) and lose nothing.

### Added

- `MenuBar.SetLabel` and `MenuBar.LabelFor`: change and read the label of a
  top-level item, for a host whose menu at one place in the bar depends on
  what it is showing (zenzx's Floppy Disk menu becomes SD Card on a Next).
- `TextModal`: a scrolling reader for a title and a text in a light
  markdown subset (`#` and `##` headings, indented lines dimmer), over a
  dimmed backdrop, with a close box and a scrollbar that drags; as tall as
  its text or filling the screen. It was zenzx's Help and About window,
  itself a generalisation of zenimate's help window, and zenplay has a
  third copy; it takes Home and End and lights its close box from `Input`,
  where zenzx's read raylib directly. `Title`, `Lines` and `AutoHeight`
  read it back. `TextPanelWidth`, `TextBodyScale` and `DrawTextLine` are
  its panel width, its text scale and the drawing of one of its lines, for
  a host's own window laid out in lines like it (zenzx's first-start
  dialog for the Next).
- `DrawProgressBar` and `ProgressText`: a bar for done out of total in the
  theme's colours, and the figures to put by one that counts bytes. From
  the dialog that fetches the Next's files in zenzx.
- `WrapText` and `FitTail`: break a text into lines at spaces and path
  separators, and keep the end of a long line in view in a field.
- `ScaleAlpha`, `ModalPanelOpacity` and `ModalBackdropDarkening`: the
  translucent panel and lighter backdrop of zenzx's windows, for any host
  that draws a window of its own in the same style.
- `KeyHome` and `KeyEnd`, which the raylib adapter fills in.
- `Version`: the module's version as a constant, in `version.go`. The
  release's sync step named `version/version.go` as a target; the file was
  not there, so the step could not pass, and it cannot be: Go refuses a
  module that has both `VERSION` and `version/` (names that differ only in
  case), and a Mac's file system cannot hold the two.

- `docs/WIDGETS.md`: a catalogue of the widgets and how a host uses one,
  with `example_test.go`, which is that use compiled and run (T-05, partly).

### Known state

- zenzx 0.27.0 is on this module: its headless suite and its 445
  GUI-tagged tests pass, and its menu, Help, About and first-start
  windows are the same picture, pixel for pixel, as with its own copy
  (run under Xvfb with software OpenGL). T-02 is closed.
- zenimate, zenplay: not moved onto `TextModal` yet (T-08).

### Changed

- `zxpalette.RGB9.Colour` scales three bits to eight by repeating them
  (0, 36, 73, 109, 146, 182, 219, 255), which is what CSpect shows; `v*255/7`
  rounded down gave 72, 145 and 218 for three of the eight. zenzx compared
  the two on the Next's own `show512` demo for its 0.14.1.

## [0.1.0] - 2026-10-18

### Added

- Initial extraction of the shared widget toolkit from zenzx's `pkg/zenui`,
  with its `zxpalette` and raylib adapter (`zenuiraylib`). zenzx's copy is the
  base: it is a strict, additive superset of the copy zenimate carried.
- `List`: a scrolling list of named items with optional pictures, a toolbar to add,
  duplicate and delete and to switch between rows and three columns, a wheel, and
  Top and Bottom buttons when the items outgrow the window. It reports what the
  person does as events; the icons are drawn from rectangles, so no icon font is
  needed. `Buttons` and `ItemRect` give a host the rectangles it uses. The toolbar's icons are five 8 by 8 pictograms in an embedded BDF font, `IconsBDF`, at U+E200 to U+E204, drawn through an optional `Icons` renderer as zenimate draws its tool palette, with plain characters as the fallback. A keyboard step switches the hover highlight off until the pointer moves or presses, so a pointer left resting on an item does not keep it lit. The arrow keys move the selection (`Step`), and a list that does not have the keyboard focus draws its selection dimmed (`Unfocused`).
- `ButtonBar`: a row of text buttons that wraps onto further rows to fit its area, with chosen and disabled states, and an option to centre each row.
- `Tabs`, `RadioGroup`, `Checkbox`, `Switch` and `DrawSection`: the controls a settings form is made of, each with the one behaviour its name says, which `ButtonBar` was being made to stand in for. `Tabs` share the width of their area, so they cannot wrap onto what is below; the two ends of the strip are slightly rounded, a row at a time, since the backend has no arc, and the rule under the strip is drawn over the tabs' feet so that it shows, with a thin line of the border colour between it and each tab that is not chosen; the chosen tab has none, and runs into the rule. A tab the pointer is over never looks chosen, even in a theme whose hot and selection colours are the same. `RadioGroup` is a choice of one among options that are all shown, stacked or in a row that wraps, and its `Layout` reports the height it took; each option is a round ring, with a round dot in the one chosen. `Checkbox` and `Switch` are on/off choices drawn as the menus draw theirs, and the label is part of what answers a click. The marks of a checkbox and a radio button are four fifths of a line high (`MarkSize` says how many pixels), and a switch is a thin track, four fifths as tall as two thirds of a line, with a round knob painted over it, centred on the track's left end when off and on its right end when on, so that half of the knob is outside the track and the control's bounds and label make room for it. In the medium and large sizes the cross of a checked box runs from corner to corner, centred and clipped to the box. The renderer has no circle, so rings and discs are drawn a row at a time, mirrored so that they are the same on both sides. Each reports the new value and whether a click changed it, as `Stepper` does. `DrawSection` is a title with a rule after it, to set a group of controls apart.
- `CellGrid`: a grid of cells that the host paints, with a cell shape of its own (a room is square cells, a map of rooms is cells shaped like a room), the lines between the cells, an outline for the selected cell and one for the cell under the pointer, and the same press, drag and release events as `PixelCanvas`, from either button. `List.NoToolbar` leaves a list's toolbar out, for a palette. The icon font gains a bucket (fill) and a crosshair (pick), sixteen glyphs in all.
- `Stepper.ButtonRect` gives a host the rectangle of one of its buttons.
- `TextField`, `Stepper` and `Swatches`: a line of editable text (typing, Backspace, the arrows, Enter to commit, Escape to restore, a click away to commit, an optional length limit and filter), a number with buttons that step it by one and, if asked, by a larger amount, and a row of colour squares with a ring round the selected one. Their geometry is a pure function of the area and the text metrics, like the other widgets'.
- `Modal`, and so `MessageBox`, draw one framed window around the title, the message and the buttons: the button row is inside the frame, which is as wide as the buttons need, and not a separate row of boxes floating beneath it. `FormConfig` gained `MinWidth` and `ExtraBottom` for this. A button bar's tip is drawn one zoom step smaller than its text, never below 1.
- `ButtonBar` can show the glyphs of the icon font instead of words (`Icons`, `UseIcons`, `BarButton.Icon`): icon buttons are square, a button with no icon or a bar with no icon renderer keeps its word, and an icon button names itself in a small tip while the pointer rests on it. The icon font grows to fourteen glyphs: undo, redo, pen, move, flip H, flip V, rotate, grid and clear join the list's five.
- `PixelCanvas`: shows a picture of any size at the largest whole number of pixels per cell that fits, with a grid, and turns the pointer into press, move and release events for either button; the host does the editing.
- `Input.MouseRightDown`: the right button held, which the raylib adapter now fills in.
- `MenuBar.DecorationX`: where the right-hand branding (the ZSP logo, or the
  Spectrum rainbow) starts, so a host can pack its own items against it and
  keep the branding the rightmost element of the strip.
- `docs/EXTRACTION.md`: provenance, the compatibility findings and the
  decisions taken for the extraction.
- Version sync, manifest-driven release with an archive gate, and a live
  register (`docs/TRACKING.md`).
