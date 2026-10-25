# Collection sources

`sources.json` is the input to `openiac-conformance collect`. It is committed;
edit it to change what enters the corpus, then regenerate.

## Shape

| Key | Meaning |
|:--|:--|
| `registry.provider` | Terraform Registry provider to crawl (`aws` for now) |
| `registry.maxPages` | API pages of 100 modules to read; the run stops at the cap and logs it |
| `registry.minDownloads` | Download floor for modules outside the named namespaces |
| `registry.maxModules` | Upper bound after ranking by downloads |
| `registry.namespaces` | Namespaces whose modules bypass the download floor |
| `registry.modules` | Modules always considered, even if the crawl missed them |
| `registry.maxFixturesPerRepo` | Module roots kept per repository |
| `github[]` | Extra repositories: `repo`, `ref`, optional `include` path prefixes and `maxFixtures` |
| `limits` | Archive, per-file, per-fixture and total-tree byte caps |
| `licenses` | SPDX identifiers allowed for a vendored fixture |

Every registry module is resolved through
`/v1/modules/{ns}/{name}/{provider}/{version}/download`, which returns an
`X-Terraform-Get` source. That ref is resolved to an immutable GitHub commit
before the tarball is downloaded, so a fixture always pins the exact revision
that was reviewed.

`include` entries are path prefixes, not globs: `modules/**` and `modules` both
mean "module roots under `modules/`". Local modules a selected root references
are still copied in, even when they sit outside the prefix. When a registry
module's download source names a subdirectory, only that directory becomes a
candidate root, so the published module is what the fixture represents.

## Selection

Within a repository the collector considers every directory that holds
Terraform files, ranked: repository root, then `modules/`, then `examples/`,
then everything else. Directories named `test`, `tests`, `spec`, `specs`, `docs`
or `internal` are excluded.

A candidate is kept only when:

- it parses as Terraform (HCL or JSON);
- it has an AWS provider or is a module wrapper with no direct provider, so at
  least one of the CDK and Pulumi targets applies;
- its own `LICENSE`, or the nearest ancestor license, is recognized and
  allowlisted;
- it contains no provisioners, `data "external"` blocks or credential literals;
- local module references stay inside the candidate root;
- it is under the per-fixture file and byte caps.

Everything else is recorded in the run report with a reason.

## Changing the corpus

1. Edit `sources.json`.
2. Run `make conformance-inputs`.
3. Read `.cache/conformance/collect-report.json` and check the skip reasons.
4. Review `git status`; the collector only replaces each fixture's `input/`,
   `LICENSE`, `NOTICE` and `fixture.json`, never a curated `expected/`.
5. Run `make conformance-verify`.
6. Commit the regenerated fixtures together with the config change.

Removing a seed does not delete existing fixtures. Prune them explicitly if the
corpus should shrink, then run `make conformance-lock` to regenerate
`conformance/corpus.lock.json` and `THIRD_PARTY.md` without re-downloading.

## Limits

`limits.maxFixtureBytes` and `limits.maxFixtureFiles` apply per module root,
after local modules have been copied in. `limits.maxTotalBytes` is enforced
while selecting candidates and checked again against the written tree. A
candidate that exceeds a cap is skipped and recorded in the report rather than
truncated; an individual file larger than `maxFileBytes` rejects its archive.
