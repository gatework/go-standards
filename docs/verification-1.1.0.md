# Standard 1.1.0 local verification

The standard, lint configuration and make targets carry version 1.1.0.
Before publication, consumers use STANDARDS_DIR to verify the local revision. The pre-existing ARC-3 module-name correction is preserved.

Architecture regressions cover empty configuration, additional YAML documents
and unmatched rules. The pinned golangci-lint 2.15.0 executable verifies the
configuration; enum switches require explicit known cases, and revive uses
package-naming with standard-library name collisions permitted.

## Results

- go-standards: format, tidy, vet, lint, architecture and shuffled race tests
  pass, including govulncheck (no vulnerabilities found). The local scan used
  a loopback bridge that fetched official vuln.go.dev data through TLS-verified
  curl because native Go certificate verification is unavailable in this process.
- set and parser: the same local checks pass up to the vulnerability download.
- All six adopting root modules received exact stdsync copies and pass
  stdsync check: go-netops, firemon, net-connector, set, motor-core, parser.
- go-netops: import formatting corrected; tidy cannot download the pinned
  net-connector v0.6.0 dependency due to DNS failure.
- firemon: vet fails on concurrent riskEvaluation declarations in audit/risk.go
  and audit/risk_expression.go. No business code was changed for this upgrade.
- net-connector: enum cases made explicit; intentional partial selectors carry
  explained exhaustive exceptions. Lint and architecture pass. Targeted race
  tests pass for config, internal/cli, internal/nilcheck and internal/progress;
  connector SSH fixtures fail because the host reports no user for UID 501.
- motor-core: logging enum cases made explicit. Existing sdk/pkg/utils and
  tools/utils violate PKG-1; package migration remains tracked in that module's
  docs/standards.md. No blanket exclusion or public package rename was applied.

Adopting repositories contain concurrent work. These results describe local
verification only, not commits, publication, CI or real-device validation.

## Publication prerequisite

The HTTPS remote can be fetched, but authenticated GitHub requests return 401
and the CLI cannot retrieve its credential from the macOS keychain in this
process. Local user lookup and computer automation are unavailable too.
Push, PR, same-commit CI and Release verification remain pending valid access.
