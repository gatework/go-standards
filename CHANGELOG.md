# Changelog

## 1.0.1 - 2026-11-11

- `make fmt` judges only the diff on stdout, so `go run` download messages on
  stderr no longer fail the check in CI.

## 1.0.0 - 2026-11-10

- First release of the gatework Go engineering standard (STANDARD.md).
- Shared golangci-lint v2.15.0 configuration and make targets.
- `stdsync` writes and verifies the distributed files; `archcheck` (moved from
  go-netops) enforces declared package import boundaries.
