# Conformance Process

> **Status:** inputs phase. The collector and committed fixture tree exist.
> The L1–L4 harness, goldens and CI gates below are still to be built.

Conformance answers one question: **does OpenIaC preserve the meaning of real
Terraform when it emits AWS CDK or Pulumi?** Coverage alone cannot answer that,
and neither can examples written by the same person who wrote the emitter.

## Principles

| # | Principle | Consequence |
|:--:|:--|:--|
| 1 | **Inputs are collected, not invented** | The corpus comes from public Terraform sources with pinned provenance. |
| 2 | **Expectations come from independent oracles** | Terraform, `tsc`, CDK and Pulumi decide correctness — not OpenIaC, not an LLM. |
| 3 | **Warnings are assertions** | Every unsupported construct must produce a stable warning code. Silent drops fail the fixture. |
| 4 | **Synthetic fixtures never count** | LLM-generated cases may probe gaps but are excluded from headline metrics. |
| 5 | **Everything is pinned and reproducible** | Corpus SHAs, tool versions and schema snapshots are recorded per run. |
| 6 | **One repository** | Compiler, fixtures and expectations share one revision; no submodule or second corpus repo. |

There is no reliable public corpus of Terraform → CDK/Pulumi *pairs*. AWS
publishes the two kinds of examples separately. The plan is therefore: collect
real Terraform inputs, then derive expected behaviour from tools that can
validate or execute them independently.

## Inputs phase

The first implementation step is the corpus itself. `cmd/openiac-conformance`
crawls the Terraform Registry for AWS modules, resolves every module version to
an immutable GitHub commit, downloads the archive and writes the accepted module
roots under `conformance/fixtures/`. Configured GitHub seeds add monorepos that
contain many local modules. The source list, caps and license allowlist live in
[`conformance/collection/sources.json`](../conformance/collection/sources.json).

```bash
make conformance-inputs
# or
go run ./cmd/openiac-conformance collect -v
```

The collector is intentionally conservative. It keeps only `.tf` and
`.tf.json` files, rejects provisioners, external data sources and credential
literals, and requires a recognized permissive license on the candidate or its
nearest ancestor. Module roots are discovered by parsing HCL, not by matching
text, so comments cannot invent or suppress a local reference. A module that
references a local path outside its own fixture root is skipped. Local modules
referenced inside the root are copied in, so the fixture is self-contained.
`verify` re-derives the lock, parses every fixture and checks its metadata
without network access; CI runs it on every pull request. It reads the same
`sources.json` as the collector, so the license allowlist cannot diverge between
the two commands. `lock` rewrites `corpus.lock.json` and `THIRD_PARTY.md` from
the fixture tree alone, for the case where fixtures were curated by hand. The
full rules and the regeneration workflow are in
[`conformance/README.md`](../conformance/README.md).

Because verification is offline, `source.commit` and `source.archiveSha256`
cannot be re-derived. Verify checks their shape and, for GitHub-origin fixtures,
that the short commit in the fixture ID matches `source.commit`; registry IDs
carry the registry version and source path instead, and their commit is recorded
at ingest. Authenticity is established only by the collector at ingest time.

No expected IR, warnings, CDK or Pulumi output is generated in this phase. The
L0–L4 layers below remain the contract for the next phase.

## Conformance layers

Each layer isolates a class of failure.

| Layer | Question | Oracle | Cadence |
|:--|:--|:--|:--|
| **L0 Mapping schema** | Do names in [`aws.json`](../pkg/mappings/aws.json) exist? | `terraform providers schema -json`, CloudFormation resource spec, Pulumi AWS schema | Every PR |
| **L1 Parse** | Does valid HCL produce the expected IR? | `terraform validate`, reviewed golden IR, metamorphic properties | Every PR |
| **L2 Emit** | Does the generated TypeScript compile? | `tsc --noEmit` | Every PR |
| **L3 Synth** | Does the target tool accept the program? | `cdk synth`, `pulumi preview --json` | Nightly |
| **L4 Semantic** | Is it the same infrastructure? | `terraform plan -json` vs normalized target config | Nightly, curated set |

### L0 — Mapping schema

Every entry in the mapping registry is resolved against three pinned schemas:

- Terraform provider schema (attribute names and types);
- CloudFormation resource specification (L1 property names);
- Pulumi AWS provider schema (module, class and property names).

This catches typos, renamed properties and wrong construct names without
building or deploying anything. It is the cheapest layer and the first one to
build.

### L1 — Parse

- `terraform validate` proves the fixture is an accepted Terraform module.
- Golden IR files are generated once and **reviewed by a human**, then frozen.
- Metamorphic checks need no golden: renaming a resource must update every
  reference, and reordering independent blocks must not change the graph.

### L2 — Emit

Generated files are compiled with `tsc --noEmit` against pinned `aws-cdk-lib`
and `@pulumi/aws` versions. A file that emits but does not compile is a
failure, not a warning.

### L3 — Synth

- `cdk synth` must produce a CloudFormation template.
- `pulumi preview --json` must complete against a local file backend with fake
  credentials and `--refresh=false`.
- Fixtures that require live AWS APIs (data sources, lookups) belong in a
  separate tier and use recorded responses or LocalStack.

### L4 — Semantic equivalence

For a small curated set, compare Terraform's plan with the target's
synthesized configuration.

Normalization rules:

| Rule | Reason |
|:--|:--|
| Compare only attributes traceable to HCL | Terraform and target providers apply different defaults |
| Apply the declared property mapping (snake → camel, renames) | The mapping is part of what is under test |
| Report unmapped attributes separately | Distinguishes “not mapped” from “mapped wrong” |
| Disable asset staging, fix stack/project names | Keeps synthesised output stable |
| Diff stable fields only | Avoids autonames, hashes and timestamps |

A semantic mismatch requires a human decision: fix the mapping, add a warning,
or declare the fixture out of scope. It is never auto-accepted.

## Corpus

The corpus is **committed to this repository** under `conformance/fixtures/`.
One clone gives a contributor the compiler, the fixtures and the expectations at
the same revision. There is no second repository, submodule or release artifact
for inputs.

Snapshot at 2026-10-26: **853 fixtures**, 6,471 files, 24,314,566 bytes; 811
from the Terraform Registry and 42 from GitHub seeds. Tiers: 20 `smoke`, 0
`core`, 487 `advanced` and 346 `modules`. Licenses: 633 Apache-2.0, 211 MIT and
9 BSD-3-Clause. Only eleven fixtures use basic features only; all were promoted
into `smoke`, leaving `core` empty. The generated
[`corpus.lock.json`](../conformance/corpus.lock.json) is authoritative.

The smoke set is derived, not curated. The collector prefers feature-basic
modules, then fills the remaining slots with the smallest other fixtures
(falling back when fewer than 20 basic modules exist), breaking ties by fixture
ID. `verify` re-runs that rule and rejects a fixture whose `tier` is `smoke`
without being in the derived set, or outside it while claiming `smoke`. Every
other fixture is classified from its parsed features. A hand-re-tiered corpus
therefore fails verification instead of silently drifting.

Provider, CloudFormation and Pulumi schemas are dependencies rather than source:
they are pinned by lockfile and cached outside the tree.

### Parse summary

`openiac-conformance status` parses every fixture and writes three committed
files: [`status.json`](../conformance/status.json) holds the corpus size, the
parse counts and the minimum-version distribution,
[`badge.json`](../conformance/badge.json) is the shields.io endpoint the README
renders as a badge, and the chart between the `conformance:summary` markers in
the [`README`](../README.md) is regenerated in place. CI runs `status -check`,
so the summary cannot drift from the corpus without failing a pull request.

```bash
make conformance-status        # re-parse the corpus and rewrite all three
make conformance-status-check  # fail when the committed summary is stale
```

The minimum version is the oldest Terraform release a fixture needs, derived
from its `required_version` constraint and the language features it uses (see
[TERRAFORM-VERSIONS.md](TERRAFORM-VERSIONS.md)); it measures parse coverage, not
conversion completeness.

### Sources

| Source | Access | License posture |
|:--|:--|:--|
| Terraform Registry modules | Registry API → versioned archive | Varies; record SPDX per module |
| `terraform-aws-modules/*` | GitHub archive | Mostly Apache-2.0 |
| `cloudposse/*` | GitHub archive | Mostly Apache-2.0 |
| `hashicorp/terraform-provider-aws` test configs | GitHub archive | MPL-2.0 |
| GitHub public code dataset / code search for `*.tf` | BigQuery or API, rate-limited | Per repository; allowlist only |
| AWS samples repositories | GitHub archive | MIT-0 / Apache-2.0 |

Prefer official APIs and archives over scraping. Respect rate limits and terms
of service.

### Selection rules

- AWS provider only for the first corpus. A candidate needs an AWS provider, or
  no direct provider and at least one module call, so at least one output target
  applies.
- No `.tfstate`, no `.tfvars` containing secrets, no credentials.
- No `local-exec`, `remote-exec`, `external` data sources or dynamic providers.
- Parse-clean: every file must parse as HCL or JSON before it is considered.
- Directories named `test`, `tests`, `spec`, `specs`, `docs` or `internal` are
  excluded. Examples are kept because they are real user configurations.
- Deduplicate by the content hash of the input tree. There is no AST
  normalization and no stratified sampling: the corpus is every candidate that
  survives ranking, caps and the total budget.
- Never execute corpus code outside the sandbox described below.

### Tiers

| Tier | Size target | Purpose |
|:--|:--|:--|
| `smoke` | ~20 | Every pull request |
| `core` | ~100 | Literal resources, variables, outputs, references |
| `advanced` | ~80 | Locals, data sources, interpolation, blocks |
| `modules` | ~20 | Module calls, `count`, `for_each` |
| `synthetic` | as needed | Planned: LLM-generated probes; excluded from headline metrics |

Only the first four tiers exist in the corpus today; `synthetic` is reserved by
the design and rejected by the current verifier, which accepts only
`smoke`, `core`, `advanced` and `modules`.

Initial targets: 200 fixtures at v0.1 and ~2,000 by v0.3. These are goals, not
guarantees.

### Fixture layout

Everything lives under `conformance/fixtures/`:

```text
conformance/
├── fixtures/
│   ├── authored/                     # planned: maintained by this project
│   │   ├── smoke/<id>/
│   │   ├── core/<id>/
│   │   ├── advanced/<id>/
│   │   ├── modules/<id>/
│   │   └── synthetic/<id>/           # planned: excluded from headline metrics
│   ├── registry/<org>/<module>/<version>/
│   └── github/<org>/<repo>/<short-sha>[/<subdir>...]/
├── collection/
│   ├── sources.json                  # seeds, path prefixes, caps, license allowlist
│   └── README.md
├── policy/
│   └── feature-support.json          # planned: feature → expected warning codes
├── oracle/
│   ├── normalize/<resource_type>.json  # planned
│   ├── allowlist.json                # planned
│   └── README.md                     # planned
├── container/
│   ├── Dockerfile                    # planned
│   └── entrypoint.sh                 # planned
├── baselines/<date>.json             # planned
├── schemas.lock.json                 # planned
├── toolchain.lock.json               # planned
├── corpus.lock.json                  # generated aggregate; verified by CI
└── reports/                          # gitignored
```

Today the tree contains `fixtures/` (registry and github origins only),
`collection/`, `README.md` and `corpus.lock.json`; every entry marked *planned*
belongs to a later phase.

Every fixture has the same shape:

```text
registry/terraform-aws-modules/vpc/5.1.0/
├── fixture.json                      # metadata and provenance
├── input/                            # Terraform module root, copied verbatim
│   ├── main.tf
│   └── variables.tf
├── expected/                         # present only when curated
│   ├── ir.json
│   ├── warnings.json
│   ├── cdk/index.ts
│   └── pulumi/index.ts
├── LICENSE                           # vendored fixtures only
├── NOTICE                            # upstream notice, when present
└── NOTES.md                          # optional
```

The fixture ID is its path relative to `conformance/fixtures/`, for example
`registry/terraform-aws-modules/vpc/5.1.0`. A new upstream version creates a new
fixture; upstream files are never edited in place.

`fixture.json`:

```json
{
  "schema": 1,
  "id": "registry/terraform-aws-modules/vpc/5.1.0",
  "tier": "core",
  "origin": "registry",
  "source": {
    "url": "https://registry.terraform.io/modules/terraform-aws-modules/vpc/aws/5.1.0",
    "ref": "5.1.0",
    "commit": "<40-hex git commit>",
    "gitOwner": "terraform-aws-modules",
    "gitRepo": "terraform-aws-modules",
    "path": ".",
    "archiveSha256": "<digest>",
    "inputSha256": "<digest>",
    "licensePath": "LICENSE"
  },
  "license": "Apache-2.0",
  "providers": ["aws"],
  "features": ["variables", "outputs"],
  "targets": ["cdk", "pulumi"],
  "support": "full"
}
```

| Field | Meaning |
|:--|:--|
| `id` | Fixture path under `fixtures/` |
| `tier` | `smoke`, `core`, `advanced` or `modules` (`synthetic` is planned); `verify` recomputes it |
| `origin` | `registry` or `github` (`authored` is planned) |
| `source` | Upstream URL, ref, immutable commit, git owner/repo, subdirectory, archive digest, input-tree digest and license path |
| `license` | SPDX identifier; the fixture directory carries the license text |
| `providers` | Providers used by the module |
| `features` | Constructs used to derive expected warnings and coverage |
| `targets` | Targets this fixture is expected to emit |
| `support` | `full`, `partial` or `parse-only` |

There is no global `manifest.jsonl`. Each `fixture.json` is the source of truth;
the harness walks the tree and builds an index in memory. `corpus.lock.json` is
the only aggregate. CI recomputes it without writing and fails when it differs
from the committed file.

### Monorepo rules

| Rule | Reason |
|:--|:--|
| Commit text only: `.tf`, `.tf.json`, `LICENSE*`, `NOTICE*`, `fixture.json`, `NOTES.md` | Keeps clones and diffs reviewable |
| Prune binaries, `.terraform/`, state, lock files and examples/tests outside the selected module | Nothing else is used by the harness |
| Cap each fixture at 512 KiB and 200 files | Bounds worst-case input |
| Keep the whole `conformance/fixtures/` tree under 100 MB | Keeps the monorepo cloneable; CI fails above budget |
| Never reformat or patch upstream files | The fixture must remain what upstream published |
| Copy the license text into the fixture and record its SPDX identifier | Attribution travels with the code |
| Allow only GPL-3.0-compatible licenses | Avoids license contamination |
| Tier is metadata, not location; never move the files to re-tier | `verify` recomputes the tier and rejects a hand-edited value |
| No symlinks, binaries or executable content | Safe to parse, archive and review |
| Add a new directory for a new upstream version | History and provenance stay intact |

If the tree ever exceeds the budget, the fallback is Git LFS — still one
repository and one history — not a second corpus repo.

## Harness

```text
conformance/fixtures/
  └─ discover → verify → select tier
       ├─ L0 schema  ──▶ mapping names resolve
       ├─ L1 parse   ──▶ IR + warnings
       ├─ L2 emit    ──▶ tsc --noEmit
       ├─ L3 synth   ──▶ cdk synth / pulumi preview
       └─ L4 diff    ──▶ terraform plan -json vs target config
                          └─ report.json + per-fixture logs
```

### Runner layout

```text
cmd/openiac-conformance/          # CLI: collect, verify, lock
internal/conformance/
├── collect/                       # implemented: ingest, fixture writing, lock, verify
├── corpus/                        # planned: harness discovery, verify, classify
├── layer0/ … layer4/              # planned: one package per oracle
├── compare/                       # planned: normalization and diff engine
└── report/                        # planned: JSON and static dashboard
```

Generated output lives in `conformance/reports/<run-id>/` and is never
committed. The fixture tree itself is defined under [Corpus](#corpus).

### Container

One pinned image with:

| Tool | Purpose |
|:--|:--|
| Go | Build the CLI |
| Terraform | `validate`, schema dump, `plan -json` |
| Cached provider plugins | Avoid repeated downloads |
| Node + `tsc` | Compile emitted TypeScript |
| `aws-cdk-lib` + CDK CLI | Synth |
| Pulumi CLI + AWS plugin | Preview against local backend |

### Safety

- Never run `terraform apply`, `pulumi up` or `cdk deploy`.
- Never provide real credentials; use fake values and `--refresh=false`.
- Run with a restricted network and a read-only mount of the fixture.
- Reject executable constructs at ingest **and** block them at runtime.
- Secrets scan every input before it enters the corpus.

## Warning taxonomy (initial)

| Code | Meaning |
|:--|:--|
| `W_UNMAPPED_RESOURCE` | No mapping for the Terraform resource type |
| `W_UNMAPPED_PROPERTY` | Resource mapped, attribute not represented |
| `W_UNMAPPED_BLOCK` | Nested block not represented |
| `W_COUNT` | `count` present |
| `W_FOR_EACH` | `for_each` present |
| `W_LOCAL` | Local value not inlined |
| `W_DATA_SOURCE` | Data source not converted |
| `W_MODULE` | Module call not expanded |
| `W_PROVIDER` | Non-AWS or aliased provider not converted |
| `W_FUNCTION` | Unsupported HCL function |
| `W_HEREDOC` | Heredoc left unchanged |
| `W_PRESERVED_BLOCK` | A top-level block has no target mapping and was preserved (`moved`, `import`, `removed`, `check`, `ephemeral`, `action`, or an unknown type) |
| `W_TERRAFORM_BACKEND` | `terraform.backend` was not converted |
| `W_TERRAFORM_CLOUD` | `terraform.cloud` was not converted |
| `W_TERRAFORM_PROVIDER_META` | `terraform.provider_meta` was not converted |
| `W_CONDITION` | A `precondition` / `postcondition` was not converted |
| `W_VARIABLE_VALIDATION` | A variable `validation` block was not converted |
| `W_UNSUPPORTED_FIELD` | A typed node carries an attribute or nested block the IR does not model |

Warnings added from schema `0.2` onward are emitted as `CODE: message`, with
`file:line` taken from the IR's provenance so the text is portable. Codes are
grouped and sorted where the same construct repeats, and `Warnings` dedupes
identical text, so a fixture can assert an exact message.

`W_PRESERVED_BLOCK` is the code that keeps `silent_drops` at zero for the
Terraform 1.x language: a construct OpenIaC cannot convert always produces at
least one warning. `required_version` alone never warns — it is metadata about
the input, not a conversion gap.

Warning codes must be stable, documented and asserted by fixtures. Codes may be
added; changing or removing one is a breaking change.

## Gates and metrics

| Metric | Definition | Gate |
|:--|:--|:--|
| `parse_rate` | Valid fixtures parsed / valid fixtures | No regression |
| `mapping_coverage` | Distinct resource types mapped / types encountered | Report; target grows |
| `compile_rate` | Emitted files passing `tsc` / emitted files | 100% for supported fixtures |
| `synth_rate` | Target programs accepted / attempted | No regression beyond 1 pp |
| `semantic_match_rate` | Curated fixtures matching the plan / curated fixtures | No regression |
| `silent_drops` | Fixtures with unsupported constructs and no warning | **Must be zero** |
| `fixture_hygiene` | Vendored fixtures with SPDX license, provenance and size within budget / vendored fixtures | 100% |

| Cadence | Scope | Fails when |
|:--|:--|:--|
| Pull request | L0 + fixture hygiene + unit tests + `smoke` L1–L2 | Any L0/L1/L2 regression, any silent drop, any hygiene violation |
| Nightly | Full corpus L1–L3, curated L4 | Metric drops beyond budget, crash, silent drop |
| Weekly | Corpus refresh in its own PR | Never auto-merged; delta reviewed |
| Monthly | Warning taxonomy and coverage review | — |

Reports are published as `report.json` plus a static dashboard: parse rate,
mapping coverage, warning rate by code, compile rate, synth rate and semantic
match rate. Golden files are never updated automatically.

## LLM policy

| Allowed | Forbidden |
|:--|:--|
| Generate mutants and adversarial syntax | Writing expected IR or expected TypeScript |
| Suggest mapping fixes for human review | Acting as an oracle or replacing one |
| Summarise failures and cluster them | Counting synthetic cases in headline metrics |
| Fill feature gaps underrepresented online | Auto-accepting golden changes |

Synthetic fixtures live under `authored/synthetic/`, are labelled `synthetic` in
`fixture.json`, and are reported separately. A conformance run that trusts
generated expectations tests the generator's assumptions, not the compiler.

## Rollout

| Phase | Deliverable | Exit criteria |
|:--|:--|:--|
| **0 · Foundations** | Freeze `examples/` as smoke tier; fixture schema and monorepo hygiene gates; L0 schema validator; warning codes | L0 runs in CI and fails on a deliberately broken mapping |
| **1 · Corpus v1** | Fetcher, vendored fixture tree, 200-fixture snapshot, L1–L3 harness | Nightly report published; smoke tier runs on every PR |
| **2 · Semantics** | `terraform plan` oracle, normalizer, 50 curated L4 fixtures | Semantic report with per-fixture diffs |
| **3 · Sustain** | Weekly corpus refresh, dashboard, mutation tests | Two consecutive corpus refreshes reviewed without firefighting |

## Risks

| Risk | Mitigation |
|:--|:--|
| License contamination | SPDX allowlist, per-fixture license text, generated `THIRD_PARTY.md` |
| Corpus disappears or changes | Every input committed; upstream URL, ref and digests recorded in `fixture.json` |
| Monorepo growth | Text-only pruning, per-fixture and total size caps, immutable versions; Git LFS if the budget is exceeded |
| Network or plugin flakiness | Cache providers and Pulumi plugins in the image |
| Non-deterministic synth output | Fixed names, no asset staging, diff stable fields only |
| Untrusted code execution | Reject executable constructs, sandbox, no credentials, no apply |
| Secrets in collected code | Scan before ingest; exclude `.tfvars` and state |
| Golden files encode bugs | Human review; oracles outrank goldens |

## Related documents

- [Architecture](ARCHITECTURE.md) — IR, rewriting and emitter contract.
- [Roadmap](ROADMAP.md) — when conformance gates land.
- [Contributing](../CONTRIBUTING.md) — local checks before a pull request.
