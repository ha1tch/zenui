# zenui — resolution record

Closed register items, full text as at closure, newest first, each stamped
with the closing version and date. Append-only: entries are written by
`repoman register close` and never edited afterwards.

## [0.4.0] T-11 — Edit boxes: caret, selection, clipboard and the shortcut keys (v0.4.0, 2026-11-27)

Theme: widgets · closed 0.4.0 · 2026-11-27
Trigger: the edit boxes (TextField, the form's text box, and a dialog's name field) edited by bytes, had no caret movement, no selection and no clipboard, and the raylib adapter passed no shortcut keys.

Scope: one shared single-line editor with a caret and selection (keys, mouse and drag), copy, cut and paste, word and line movement, Delete; the Input contract gains `Shift`, `Ctrl`, `Alt`, `CtrlChars`, `Paste`, `SetClipboard`; the raylib adapter reads Ctrl or Cmd with C, X, V and A, Shift+Insert, Ctrl+Insert, Shift+Delete, Cmd+Left and Right as Home and End, Alt or Option for word jumps, key repeat for the editing keys, and keeps AltGr text as text.

Resolved in 0.4.0: `lineedit.go`, `textfield.go`, `form.go`, `zenuiraylib/keys.go`; tests in `lineedit_test.go`, `textfield_edit_test.go`, `form_caret_test.go`, `zenuiraylib/keys_test.go`.

Cross-ref: CHANGELOG 0.4.0.

## [0.4.0] T-10 — Make the file dialog configurable and its view rememberable (v0.4.0, 2026-11-27)

Theme: widgets · closed 0.4.0 · 2026-11-27
Trigger: the dialog had one fixed look: a single extension filter, no hidden-file choice, one order, no recent files, English-only words, one size, and a double click that never timed out.

Scope: make the dialog configurable by the host and let the person's choices be kept: `FilterSets` with a selector and Ctrl+Tab, `View` (hidden files, details, order, file type) given back by `Dialog.View`, `Recent`, `Places`/`ExtraPlaces`, `Name`, `Validate`, `NoOverwriteConfirm`, `Labels`, `Scale`, `Width`/`Height`, `DoubleClickSeconds`, wildcards typed in the field, a details view with size and date columns and sortable heads, natural name order, and `HiddenAware` and the extra `Entry` fields for a filesystem that can say more.

Resolved in 0.4.0: `dialog.go`, `dialog_view.go`, `dialog_draw.go`, `fs.go`. Multi-select is not provided; see KNOWN_ISSUES.

Cross-ref: CHANGELOG 0.4.0.

## [0.4.0] T-09 — The Open and Save dialog takes typed and pasted paths (v0.4.0, 2026-11-27)

Theme: widgets · closed 0.4.0 · 2026-11-27
Trigger: zenzx's Open dialog (zenzx T-104, T-107) could not take a typed or pasted path: its name field was filled only by the selection, edited by bytes with no caret, and Open had no field at all.

Scope: an always-active name/path field in Open and Save, edited like a TextField; a path resolves when it is absolute, `~`, relative, with `..`, quoted, a `file://` address, or differs only in case; a folder goes in, a bundle is entered, a file is accepted; Tab completes (single match, common prefix, a hint with the number of matches); type-ahead selects the matching row; a Save asks before replacing.

Resolved in 0.4.0: `dialog_path.go` and `dialog_input.go`; tests in `dialog_path_test.go`.

Cross-ref: CHANGELOG 0.4.0.

## [0.2.0] T-02 — Migrate zenzx onto the shared zenui module (v0.2.0, 2026-11-16)

Theme: migration · closed 0.2.0 · 2026-11-16
Trigger: zenzx is where the widget code was developed and still holds the only copy of `pkg/zenui`, `pkg/zenuiraylib`, `pkg/zxpalette` and `pkg/bdf`.

Scope: switch the root package's imports to the module (menu bar, logo menu, modals, OSD); use zentools' `pkg/bdf`; delete the four local packages. Coordinate with zenzx's open items (T-33 and T-34 wait on a tagged release), so the migration lands in its own release.

Done when: zenzx's headless suite and its GUI-tagged tests pass on the module, and the four local packages are gone.

Cross-ref: CHANGELOG 0.2.0.

## Record format

Each entry is headed `## [version] T-nn — summary` and carries the item's
detail text verbatim, as it stood in the register at closure.
