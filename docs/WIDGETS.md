# zenui — the widgets

Updated: 2026-11-16

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

## Windows and dialogs

| Widget | What it is |
|---|---|
| `Dialog` | File Open/Save: folders, filtering, selection, a file name field, a preview pane |
| `Modal`, `MessageBox` | A framed window with a form and a row of buttons; the usual OK, Yes/No, OK/Cancel boxes |
| `TextModal` | A scrolling reader for a title and a text in a light markdown subset: Help, About |
| `Form` | Labelled fields in a frame, the body of a `Modal` |
| `OptionPanel` | A panel of choices |

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
| `TextField` | One line of editable text |
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
