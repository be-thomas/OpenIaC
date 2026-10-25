# Collection sources

`sources.json` is the input to `openiac-conformance collect`. It is committed;
edit it to change what enters the corpus, then regenerate.

## Shape

| Key | Meaning |
|:--|:--|
| `registry.provider` | Terraform Registry provider to crawl (`aws` for now) |
| `registry.maxPages` | API pages of 100 modules to read |
| `registry.minDownloads` | Download floor for modules outside the named namespaces |
| `registry.maxModules` | Upper bound after ranking by downloads |
| `registry.namespaces` | Namespaces whose modules bypass the download floor |
| `registry.modules` | Modules always considered, even if the crawl missed them |
| `registry.maxFixturesPerRepo` | Module roots kept per repository |
| `github[]` | Extra repositories: `repo`, `ref`, optional `include` globs and `maxFixtures` |
| `limits` | Archive, per-file, per-fixture and total-tree byte caps |
| `licenses` | SPDX identifiers allowed for a vendored fixture |

Every registry module is resolved through
`/v1/modules/{ns}/{name}/{provider}/{version}/download`, which returns an
`X-Terraform-Get` source. That ref is resolved to an immutable GitHub commit
before the tarball is downloaded, so a fixture always pins the exact revision
that was reviewed.

## Changing the corpus

1. Edit `sources.json`.
2. Run `make conformance-inputs`.
3. Read `.cache/conformance/collect-report.json` and check the skip reasons.
4. Review `git status`; the collector only replaces each fixture's `input/`,
   `LICENSE`, `NOTICE` and `fixture.json`, never a curated `expected/`.
5. Commit the regenerated fixtures together with the config change.

Removing a seed does not delete existing fixtures. Prune them explicitly if the
corpus should shrink, then regenerate `conformance/corpus.lock.json`.

## Limits

`limits.maxFixtureBytes` and `limits.maxFixtureFiles` apply per module root,
after local modules have been copied in. `limits.maxTotalBytes` is enforced
across the whole tree. A candidate that exceeds a cap is skipped and recorded in
the report rather than truncated.
