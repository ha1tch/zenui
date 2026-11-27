# zenui — the widgets

Updated: 2026-11-27

A catalogue: what there is and where to read on. Each widget's own doc
comment is its reference. A full reference with every field is register
item T-05.

## How a widget is used

A widget draws nothing and reads no device itself. Each frame the host
gives it an `Input` (the pointer, the buttons, the wheel, the keys and
characters of that frame, the seconds since the last one) and a `Renderer`
(rectangles, two gradients, lines, text and its metrics, clipping), with a
`Theme` for the colours. `zenui/zenuiraylib` is both for raylib:
`zenuiraylib.Input()` and a `Renderer` that draws BDF bitmap fonts.

Most widgets have `Update(in Input)` and `Draw(r Renderer, ...)`; the
simplest are one function that draws and reports what a click did. Layout
is a pure function of the area and the text metrics, so a widget can be
tested with a renderer that only records what it is asked to draw:

```go
m := zenui.NewTextModal("ABOUT", "# zenui\n\nA widget toolkit.", true)
m.Draw(r, 800, 600, zenui.DarkTheme())      // r: any Renderer
open := m.Update(zenuiraylib.Input())       // false: close it
```

`example_test.go` is that, compiled and run as a test.

Keyboard shortcuts reach a widget through `Input`: `Ctrl` is the shortcut
modifier (Control, or Command on a Mac), `CtrlChars` the letters pressed with
it, `Paste` the text pasted this frame, and `SetClipboard` where a copy goes.
The raylib adapter fills them in, so the edit boxes have Ctrl or Cmd with A, C,
X and V without the host doing anything.

## Windows and dialogs

| Widget | What it is |
|---|---|
| `Dialog` | File Open/Save: a name field that takes typed and pasted paths, folders, file types, hidden files, details and order, recent files, a preview pane; see below |
| `Modal`, `MessageBox` | A framed window with a form and a row of buttons; the usual OK, Yes/No, OK/Cancel boxes |
| `TextModal` | A scrolling reader for a title and a text in a light markdown subset: Help, About; `WithLinks` makes its `[text](address)` links and bare web addresses clickable |
| `Form` | Labelled fields in a frame, the body of a `Modal` |
| `OptionPanel` | A panel of choices |

### The file dialog

```go
d := zenui.NewDialog(zenui.DialogConfig{
	Mode:       zenui.ModeOpen,
	StartDir:   lastFolder,
	FilterSets: []zenui.FilterSet{{Name: "Tapes", Exts: []string{"tap", "tzx"}}},
	Recent:     recentPaths, // most recent first
	View:       savedView,   // from an earlier d.View()
})
// each frame:
d.Draw(r, w, h, theme)
switch d.Update(zenuiraylib.Input()) {
case zenui.Accepted:
	open(d.Result()); savedView = d.View(); lastFolder = d.Dir()
case zenui.Cancelled:
	// closed
}
```

The name field is always being edited, in Open as in Save. Type or paste a
path and Enter opens it: absolute, `~`, relative to the folder on show, with
`..`, in quotes, a `file://` address, or in other letter case than the file's
own. A folder is gone into, a bundle is entered, a file is accepted; what
cannot be opened is said in the status line. A name inside a bundle names an
entry. A typed `*.tap` limits the list to it (`*` takes the limit away).

| Key | What it does |
|---|---|
| Tab | completes the name: a single match, what several share, or a hint with how many |
| Up, Down, PageUp, PageDown | move the selection; Enter opens it when the field only echoes it |
| Ctrl+Tab, Ctrl+Shift+Tab | the next and the previous file type |
| Ctrl+H | shows or hides the hidden files (a filesystem that is `HiddenAware`) |
| Ctrl+D | the details view: size, date and `Entry.Detail`; a column head sorts |
| Ctrl+L | selects the whole field |
| Alt+Up | the parent folder |
| Escape | drops a question being asked (a popup, a replace), then cancels |

What a host chooses, in `DialogConfig`:

| Field | Effect |
|---|---|
| `Filters`, or `FilterSets` and `NoAllFiles` | one fixed file type, or a choice of named ones; `FilterSet.DefaultExt` is what a Save adds while that one is chosen |
| `View` | how the person last looked: hidden files, details, sort key and direction, file type. Read back with `Dialog.View` |
| `Recent`, `Places`, `ExtraPlaces` | a Recent place in the sidebar (the host keeps the list); the favourites, replaced or added to |
| `Name`, `StartDir`, `StartContainer` | the starting text, folder, or bundle to be inside |
| `Validate` | refuses a chosen path with a message, in Open and in Save |
| `NoOverwriteConfirm` | a Save replaces without asking (it asks by default) |
| `Labels` | every word the dialog draws |
| `Scale`, `Width`, `Height`, `DoubleClickSeconds` | size and timing |
| `Preview` | a thumbnail for the selected entry |

A filesystem is an `FS`; `OSFS` is the real one. One that also has
`ReadDirAll` (`HiddenAware`) can list hidden entries, and `Entry.Size`,
`ModTime` and `Detail` fill the details view.

## Menus

| Widget | What it is |
|---|---|
| `MenuBar` | A strip of labels that open dropdowns, with the family's logo or the Spectrum rainbow at its right; `SetItems`, `SetLabel`, `DecorationX` for a host that changes it |
| `Menu` | A dropdown: items, separators, checkboxes, submenus, titles |

## Controls

| Widget | What it is |
|---|---|
| `ButtonBar` | A row of text or icon buttons that wraps to fit |
| `Tabs`, `RadioGroup`, `Checkbox`, `Switch` | What a settings page is made of |
| `TextField` | One line of editable text, with a caret, a selection and the clipboard; the form's text box and the dialog's name field share its editor |
| `Stepper` | A number with buttons that step it |
| `Swatches` | A row of colour squares, one selected |
| `List` | A scrolling list of named items with pictures, a toolbar, rows or columns |
| `DrawSection` | A title with a rule after it |
| `DrawProgressBar`, `ProgressText` | A bar for done out of total, and the figures to put by it |

## Editors' parts

| Widget | What it is |
|---|---|
| `PixelCanvas` | A picture at the largest whole zoom that fits, with a grid, and press, move and release events |
| `CellGrid` | A grid of cells the host paints, with selection and the same events |
| `ToolPalette` | A palette of tools |
| `ZXClassicPaletteChooser` | Ink, paper, bright and flash, as the Spectrum has them |
| `PreviewPane` | A picture beside a file list |

## Text and colour helpers

`WrapText`, `FitTail`, `DrawTextLine`, `TextPanelWidth`, `TextBodyScale`;
`ScaleAlpha`, `ModalPanelOpacity`, `ModalBackdropDarkening`; the themes
`DarkTheme`, `LightTheme`, `SpectrumTheme`; the icon font `IconsBDF`.
`zenui/zxpalette` has the Spectrum's and the Next's colours.
