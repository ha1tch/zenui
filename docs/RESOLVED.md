# zenui — resolution record

Closed register items, full text as at closure, newest first, each stamped
with the closing version and date. Append-only: entries are written by
`repoman register close` and never edited afterwards.

## [0.2.0] T-02 — Migrate zenzx onto the shared zenui module (v0.2.0, 2026-11-16)

Theme: migration · closed 0.2.0 · 2026-11-16
Trigger: zenzx is where the widget code was developed and still holds the only copy of `pkg/zenui`, `pkg/zenuiraylib`, `pkg/zxpalette` and `pkg/bdf`.

Scope: switch the root package's imports to the module (menu bar, logo menu, modals, OSD); use zentools' `pkg/bdf`; delete the four local packages. Coordinate with zenzx's open items (T-33 and T-34 wait on a tagged release), so the migration lands in its own release.

Done when: zenzx's headless suite and its GUI-tagged tests pass on the module, and the four local packages are gone.

Cross-ref: CHANGELOG 0.2.0.

## Record format

Each entry is headed `## [version] T-nn — summary` and carries the item's
detail text verbatim, as it stood in the register at closure.
