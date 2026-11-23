# zenui — live register

Version: 0.3.3
Last reviewed: 2026-11-23

Open, actionable items only. Closed items move verbatim to
`RESOLVED.md` (see `repoman register close`); a closed item still
listed here is itself a defect. Intentional limits and the dormant-guard
table live in `KNOWN_ISSUES.md`.

Status legend: ✓ done · ◐ partial · ☐ not started · ✗ dropped.

## Status table

| ID | Summary | Theme | Priority | Status | Blocks |
|---|---|---|---|---|---|

| T-01 | Migrate zenimate onto the shared zenui module | migration | P1 | ☐ | after the first zenui tag is pushed |
| T-03 | Add tests for the raylib adapter | quality | P2 | ☐ | — |
| T-04 | Decide where the family branding (logo, rainbow, theme presets) belongs | scope | P3 | ☐ | — |
| T-05 | Document the widget set and the Renderer and Input contract | docs | P2 | ◐ | — |
| T-06 | CI: cgo-free tests, vet, gofmt, repoman gate and tagged releases | ci | P2 | ☐ | — |
| T-07 | Decide the long-term home of `zxpalette` | scope | P2 | ☐ | before the first minor release |
| T-08 | zenimate and zenplay carry their own copies of the text reader | migration | P2 | ☐ | — |

## migration

### T-01. Migrate zenimate onto the shared zenui module

Theme: migration · Priority: P1 · Status: ☐ · Blocks/after: after the first zenui tag is pushed

Trigger: zenimate carries its own older copy of `pkg/zenui`, plus copies of `pkg/zxpalette` and `pkg/bdf`.

Scope: replace the three packages with imports of `github.com/ha1tch/zenui`, `github.com/ha1tch/zenui/zxpalette` and zentools' `pkg/bdf`; give `fpRenderer` (and the three test doubles) the methods `FillRectGradientV`, `FillRectGradientHMultiply` and `DrawLine`, or replace `fpRenderer` with `zenuiraylib.Renderer`; delete the local copies.

Done when: zenimate builds with `CGO_ENABLED=0 -tags purego` and its tests pass against the module.

### T-08. zenimate and zenplay carry their own copies of the text reader

Theme: migration · Priority: P2 · Status: ☐

zenimate's help window (cmd/zenimate-gui/helpmodal.go) and zenplay's About window (internal/shell/aboutmodal.go) are the reader that is TextModal here since 0.2.0, each in its own copy. To do: replace both with TextModal, as zenzx did for its Help and About, and compare the windows before and after.

Trigger: TextModal was lifted from zenzx for 0.2.0; the other two hosts were not touched.

## quality

### T-03. Add tests for the raylib adapter

Theme: quality · Priority: P2 · Status: ☐ · Blocks/after: —

Trigger: `zenuiraylib` has no test files; its behaviour is only exercised through zenzx's GUI.

Scope: cover colour and rectangle conversion, font metrics through the bundled bitmap font, and clipping bookkeeping, in tests that run without a display.

Done when: `go test ./zenuiraylib` runs real assertions headless under `-tags purego`.

## scope

### T-04. Decide where the family branding (logo, rainbow, theme presets) belongs

Theme: scope · Priority: P3 · Status: ☐ · Blocks/after: —

Trigger: `zsplogo.go`, `rainbow.go` and `theme_presets.go` carry branding from the zen* family's own tools, inside what is meant to be a general toolkit.

Scope: either keep them as an optional subpackage, or move them to a family-branding package that depends on zenui.

Done when: the core package contains no family branding, or the decision to keep it is recorded in `KNOWN_ISSUES.md`.

### T-07. Decide the long-term home of `zxpalette`

Theme: scope · Priority: P2 · Status: ☐ · Blocks/after: before the first minor release

Trigger: zenzx's emulator core uses only the Next RGB9 functions from `zxpalette` (`nextpalette.go`), while the classic palette is what the widgets need. With `zxpalette` inside this module, the machine depends on the UI module for colour maths.

Scope: decide the long-term home before the first minor release. Candidates: keep it here; move the whole package to zentools (a data-format toolkit both zenzx and zenui already depend on); or split it, classic palette here and the Next RGB9 helpers with the machine.

Done when: the decision is recorded in `docs/EXTRACTION.md` and, if the package moves, every consumer imports it from the new home.

## docs

### T-05. Document the widget set and the Renderer and Input contract

Theme: docs · Priority: P2 · Status: ◐ · Blocks/after: —

Done so far (0.2.0): `docs/WIDGETS.md` is a catalogue of the widgets with how a host uses one, and `example_test.go` drives a `TextModal` with a renderer that records, compiled and run as a test. Not there: the reference itself, every widget's fields and events.

Scope: a reference for every widget, the `Renderer` and `Input` contracts, and one headless example that drives a widget with a recording renderer, so a new host can be written from the document alone.

Done when: `docs/WIDGETS.md` exists and its example compiles as a test.

## ci

### T-06. CI: cgo-free tests, vet, gofmt, repoman gate and tagged releases

Theme: ci · Priority: P2 · Status: ☐ · Blocks/after: —

Scope: GitHub Actions running tests with `CGO_ENABLED=0 -tags purego`, `go vet`, a gofmt check and the repoman gate (register, badcode, gomod), plus a tag-triggered release.

Done when: a pull request runs all of them and a tag publishes a release.

