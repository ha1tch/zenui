# zenui — known issues and limits

Version: 0.3.5
Last reviewed: 2026-11-25

Intentional limits, invariant boundaries and recorded decisions. Open,
actionable work lives in `TRACKING.md`.

## Intentional limits

- `zsplogo.go`, `rainbow.go` and `theme_presets.go` carry branding from the
  zen* family's own tools; whether they belong in a general toolkit is open
  (T-04).

## Dormant guards

Verification that does not run in the default `go test ./...` invocation.
A guard's existence is not evidence; only its execution record is.
`repoman guards stale` lists guards not exercised since the previous release.
Each guard is listed here in the same change that writes it.

None yet.
