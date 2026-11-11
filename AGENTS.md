# go-standards

This repository is the source of [the gatework Go engineering standard](STANDARD.md) (go-standards/STANDARD.md). Every rule there applies here too.

- Change rules only by editing STANDARD.md. Bump its version, the `standard-version` stamps in golangci.yml and standard.mk, and CHANGELOG.md in the same commit. Never reuse a rule ID.
- golangci.yml and standard.mk are the originals. `.golangci.yml` here is a distributed copy; refresh it with `make standard-self`.
- After a change, run `make check`, then `stdsync write` on each adopting repository and fix what it reports there.
