# Conformance corpus

This directory holds the Terraform inputs used by the conformance harness. The
current phase is **inputs only**: fixtures contain real Terraform modules, the
exact upstream revision and the upstream license. Goldens and the L0–L4 oracle
layers are not written yet.

The design and rollout plan are in [docs/CONFORMANCE.md](../docs/CONFORMANCE.md).

## Snapshot

Collected 2026-10-26 from the Terraform Registry and the configured GitHub
seeds:

| Metric | Value |
|:--|--:|
| Fixtures | 853 |
| Files | 6,471 |
| Bytes | 24,314,566 |
| `smoke` / `core` / `advanced` / `modules` | 20 / 0 / 487 / 346 |
| Registry / GitHub | 811 / 42 |
| Apache-2.0 / MIT / BSD-3-Clause | 633 / 211 / 9 |

Only eleven fixtures use basic features only and all eleven were promoted into
`smoke`, so no fixture currently carries the `core` tier. The remaining nine
`smoke` fixtures come from the smallest-fixture fallback described below.

`conformance/corpus.lock.json` is the authoritative aggregate; the counts here
are descriptive.

## Layout

```text
conformance/
├── fixtures/
│   ├── registry/<namespace>/<name>/<version>[/<module-dir>]/
│   └── github/<owner>/<repo>/<short-sha>[/<module-dir>]/
├── collection/sources.json    # seeds, caps and license allowlist
├── corpus.lock.json           # generated aggregate, verified by CI
└── README.md
```

Each fixture contains:

| Path | Purpose |
|:--|:--|
| `input/` | The Terraform module root, copied from upstream |
| `fixture.json` | ID, tier, provenance, hashes, license, features |
| `LICENSE` | Upstream license text |
| `NOTICE` | Upstream notice, when present |
| `expected/` | Reserved for curated goldens; not present in this phase |

Local module sources (`source = "./modules/x"`) are copied into the same
`input/` tree so the fixture is self-contained. A candidate that references a
path outside its own root is rejected rather than committed broken.

## Collecting

```bash
export GITHUB_TOKEN=$(gh auth token)   # or provide GITHUB_TOKEN yourself
go run ./cmd/openiac-conformance collect -v
```

The command reads [`collection/sources.json`](collection/sources.json), crawls
the Terraform Registry for AWS modules, resolves each module to an immutable
GitHub commit, downloads the archive and writes the accepted fixtures. It also
resolves the configured GitHub seeds, which is how monorepos with many local
modules are captured.

Validate the committed tree without network access:

```bash
make conformance-verify
# or
go run ./cmd/openiac-conformance verify
```

CI runs the same check. It recomputes the lock and attribution, parses every
fixture, and fails on a stale lock, a tampered input, a stray file or an
inconsistent `fixture.json`. It also checks provenance: the fixture ID must
agree with the recorded origin, with the registry coordinates (namespace, name,
version) or the GitHub repository and short commit, and with the source path. It
enforces the license allowlist from `-config` (default
[`collection/sources.json`](collection/sources.json)) and the derived `smoke`
set. Because verification is offline, the recorded commit and archive digest are
checked for consistency, not re-downloaded; a registry ID carries the version,
so its commit is recorded at ingest and not bound by the ID.

Regenerate the lock and attribution from the fixture tree alone, for example
after curating a fixture by hand:

```bash
go run ./cmd/openiac-conformance lock
```

Useful overrides:

```bash
go run ./cmd/openiac-conformance collect -limit 50 -pages 10 -v
```

The collector is deliberately conservative:

- Only `.tf` and `.tf.json` files are kept. State, plans, lockfiles, binaries
  and `.tfvars` are never written.
- Every candidate must parse as Terraform. HCL is parsed, not pattern-matched,
  so a commented-out `source = "../outside"` cannot reject a valid module and a
  comment cannot invent a local module reference.
- Each candidate is governed by its own `LICENSE` or the nearest ancestor
  license. Unknown or copyleft licenses are skipped.
- Provisioners, `data "external"` and credential-like literals reject a
  candidate fixture. Credential-looking *placeholders* are allowed.
- A candidate must use AWS or be a module wrapper, so at least one output target
  applies. Providers with no target are rejected.
- Directories named `test`, `tests`, `spec`, `specs`, `docs` or `internal` are
  not module roots. `examples/` is kept: those are user-facing configurations.
- A fixture is capped at 512 KiB and 200 files; the whole tree is capped at
  100 MiB and checked again after writing.
- Identical input trees are stored once, deduplicated by content hash.
- Local module references must stay inside the fixture root.

Every run writes a detailed audit to `.cache/conformance/collect-report.json`
(gitignored). `corpus.lock.json` and `THIRD_PARTY.md` are generated from the
committed tree.

## Fixture metadata

```json
{
  "schema": 1,
  "id": "registry/terraform-aws-modules/vpc/5.1.0",
  "tier": "modules",
  "origin": "registry",
  "source": {
    "url": "https://registry.terraform.io/modules/terraform-aws-modules/vpc/aws/5.1.0",
    "ref": "5.1.0",
    "commit": "<40-hex git commit>",
    "gitOwner": "terraform-aws-modules",
    "gitRepo": "terraform-aws-modules",
    "path": ".",
    "archiveSha256": "…",
    "inputSha256": "…",
    "licensePath": "LICENSE"
  },
  "license": "Apache-2.0",
  "providers": ["aws"],
  "features": ["modules", "outputs", "variables"],
  "targets": ["cdk", "pulumi"],
  "support": "partial"
}
```

`ref` is the registry version or the requested git ref, and `commit` is the
immutable commit the archive was cut from. `licensePath` records which upstream
file supplied the license. `inputSha256` is the deterministic digest of the
input tree and is the unit of deduplication. `archiveSha256` pins the exact
downloaded tarball. `targets` lists only the backends that can consume the
fixture; today that means `cdk` and `pulumi` for AWS modules.

## Tiers

| Tier | Meaning |
|:--|:--|
| `smoke` | Derived 20-fixture set; selected for every pull request |
| `core` | Resources, variables and outputs only |
| `advanced` | Data sources, locals, expressions, count/for_each |
| `modules` | Module composition |

If fewer than 20 fixtures are basic, the smoke tier is filled with the smallest
remaining fixtures and the run logs how many came from the fallback. Tier is
metadata, not a directory. Re-tiering never moves a fixture.

The smoke set is deterministic: basic modules first, then the smallest
remaining fixtures, ties broken by fixture ID. `verify` recomputes it and
rejects any `fixture.json` whose tier disagrees, so re-tiering requires
re-running the collector or matching the derived rule exactly. Other tiers are
re-derived from the parsed features.

## Adding a source

1. Add a registry namespace/module or a GitHub seed to
   `collection/sources.json`.
2. Run the collector.
3. Review `git status` and the new `fixture.json` files.
4. Run `make conformance-verify` and `go test ./...`, then commit the fixtures
   with the code that consumes them.

Do not edit vendored Terraform files. If an upstream module needs a fix, add a
note and wait for a new upstream revision. Hand-curated fixtures under a future
`authored/` directory are planned; the current verifier accepts only `registry`
and `github` origins.

## Licensing

OpenIaC is GPL-3.0. Vendored fixtures remain under their upstream licenses;
`LICENSE` in each fixture is authoritative. The generated
[`THIRD_PARTY.md`](../THIRD_PARTY.md) lists every fixture and its upstream.

The allowlist in `sources.json` accepts Apache-2.0, MIT, MIT-0, BSD-2-Clause,
BSD-3-Clause, ISC, MPL-2.0, Unlicense and CC0-1.0. License detection is
deliberately strict: text that adds restrictions to a permissive license (the
JSON license, Commons Clause, advertising clauses, SSPL/BUSL) is treated as
unknown and skipped, and a file that quotes both a permissive and a copyleft
license is not vendored.
