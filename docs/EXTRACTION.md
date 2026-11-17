# zenui — extraction record

Updated: 2026-10-18

zenui was extracted from two diverged copies of the same widget package: the
one in zenzx (`pkg/zenui`, 22 files) and the one in zenimate (`pkg/zenui`,
13 files).

## Findings

- zenzx's copy is a strict, additive superset of zenimate's. Six types differ
  (`Input`, `Item`, `Menu`, `MenuConfig`, `Renderer`, `Theme`) and every
  difference is an added field or method, checked mechanically as a token
  subsequence. No function signature present in zenimate's copy is missing or
  changed in zenzx's.
- All 49 zenui symbols that zenimate's code uses exist unchanged in zenzx's
  copy.
- zenimate's 60 test functions pass against this module, in isolation. Their
  test doubles needed the three `Renderer` methods added to compile.
- `bdf` is byte-identical in zenimate, zenzx and zentools. `zxpalette.go` is
  identical in zenimate and zenzx; zenzx's copy adds `nextpalette.go`.

## Decisions

| Decision | Choice |
|---|---|
| Base | zenzx's copy, because it is the superset |
| Layout | Core at the module root, `zxpalette` and `zenuiraylib` as subpackages |
| `zxpalette` | Lives in this module, so the core stays free of any other zen* dependency |
| `bdf` | Not copied: zentools' `pkg/bdf` is canonical and the adapter imports it |
| Adapter package name | `zenuiraylib`, unchanged, so existing callers keep their identifiers |

## What consumers must change

Anything that implements `Renderer` must add `FillRectGradientV`,
`FillRectGradientHMultiply` and `DrawLine`, or use `zenuiraylib.Renderer`.
`Input` gains `DeltaTime`, which is optional (zero is valid). Import paths
change to `github.com/ha1tch/zenui` and its subpackages.
