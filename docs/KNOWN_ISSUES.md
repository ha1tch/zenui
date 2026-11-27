# zenui — known issues and limits

Version: 0.3.5
Last reviewed: 2026-11-25

Intentional limits, invariant boundaries and recorded decisions. Open,
actionable work lives in `TRACKING.md`.

## Intentional limits

- `zsplogo.go`, `rainbow.go` and `theme_presets.go` carry branding from the
  zen* family's own tools; whether they belong in a general toolkit is open
  (T-04).
- The file dialog picks one file. There is no multi-select: Open returns one
  path, and a host that wants several opens the dialog again.
- The dialog's hidden files are those whose names begin with a dot, on every
  system. `OSFS` does not read the Windows hidden attribute; a host whose
  filesystem knows better implements `HiddenAware` and sets `Entry.Hidden` itself.
- The dialog does not keep the recent files or the person's `View`. The host
  passes `DialogConfig.Recent` and `DialogConfig.View` in, and reads the view
  back with `Dialog.View`, and stores both where it likes.
- The edit boxes are one line, and take the runes and the paste the host gives
  them. There is no input-method composition and no undo.
- A double click is timed from `Input.DeltaTime`; a host that leaves it zero has
  no time limit on the second click.

## Dormant guards

Verification that does not run in the default `go test ./...` invocation.
A guard's existence is not evidence; only its execution record is.
`repoman guards stale` lists guards not exercised since the previous release.
Each guard is listed here in the same change that writes it.

None yet.
