# go-standards

The gatework Go engineering standard and the tooling that enforces it.

| File | Purpose |
| --- | --- |
| [STANDARD.md](STANDARD.md) | The normative rules, each with a stable ID (for example `NAM-2`). |
| [golangci.yml](golangci.yml) | Shared golangci-lint v2 configuration; each repository carries it as `.golangci.yml`. |
| [standard.mk](standard.mk) | Shared `make check` targets; each repository's Makefile includes it. |
| `cmd/stdsync` | Writes (`stdsync write`) and verifies (`stdsync check`) the distributed copies. |
| `cmd/archcheck` | Enforces the package import boundaries declared in `architecture.yaml`. |

## Adopting the standard

```sh
go -C ../go-standards run ./cmd/stdsync write "$PWD"   # writes .golangci.yml and standard.mk
printf 'include standard.mk\n' >> Makefile               # once
make check STANDARDS_DIR=../go-standards                 # until go-standards is published
```

Then make the first line of `AGENTS.md` link `go-standards/STANDARD.md`, and add an `architecture.yaml`.
