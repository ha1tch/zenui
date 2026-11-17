# zenui

A renderer-agnostic widget toolkit for ZX Spectrum-style tools, written in Go:
menus and a menu bar, file dialogs, message boxes, a text reader, lists, tabs,
text fields, steppers, a pixel canvas and more ([the catalogue](docs/WIDGETS.md)).
It owns the logic of the widgets and draws nothing itself: the host supplies a
small `Renderer` and an `Input` snapshot each frame, so the same widgets run
under raylib, a software framebuffer or a headless test. zenzx, zenimate and
zenplay share it.

## Packages

| Package | Purpose | Dependencies |
|---|---|---|
| `zenui` | The widgets and the `Renderer` and `Input` contracts | standard library, `zenui/zxpalette` |
| `zenui/zxpalette` | ZX Spectrum and Next colour tables | standard library |
| `zenui/zenuiraylib` | A raylib `Renderer` using the bundled Sinclair bitmap font | raylib-go, zentools `bdf` |

## Requirements

- Go 1.25 or later

## Building

```
make            # build, vet and test (cgo-free, -tags purego)
make help       # every target
```

## Documentation

- [docs/WIDGETS.md](docs/WIDGETS.md): the widgets, and how a host uses one
- [docs/EXTRACTION.md](docs/EXTRACTION.md): provenance, compatibility findings and decisions
- [docs/TRACKING.md](docs/TRACKING.md): the live register of open work
- [docs/KNOWN_ISSUES.md](docs/KNOWN_ISSUES.md): intentional limits and dormant guards
- [CHANGELOG.md](CHANGELOG.md)

## Licence

Apache 2.0, see [LICENSE](LICENSE).

Copyright (c) 2026 haitch (h@ual.li, https://oldbytes.space/@haitchfive)
