# Roadmap

OpenIaC aims to be the most predictable way to leave HCL without rewriting your
infrastructure by hand.

**Guiding rule:** never trade correctness for coverage. An unsupported construct
gets a warning; it does not get guessed.

| State | Meaning |
|:--:|:--|
| ✅ | Shipped |
| 🚧 | In progress |
| ⏳ | Planned |
| 🌅 | Exploring |

---

## v0.1 — Foundation · Oct–Nov 2026 ✅

The compiler exists and does something real.

| Scope | Detail |
|:--|:--|
| HCL parser | Native HCL and `.tf.json`; every Terraform 1.0–1.17 top-level block is parsed or preserved |
| IR | Flat graph, typed values, preserved blocks, `features` / `min_version`, JSON schema `0.2` ([reference](IR.md)) |
| Ordering | Deterministic topological sort with cycle reporting |
| Rewriting | References, templates, comments; boundary-safe |
| Emitters | AWS CDK (L1/L2) and Pulumi (aws SDK) |
| Mappings | 14 AWS resource types in `aws.json` |
| CLI | `convert`, `ir`, `version` |
| Playground | Monaco + Go/WASM, no backend |

**Exit criteria:** `make test` green; example module converts to both targets;
playground produces the same output as the CLI.

---

## v0.1.1 — Release hardening · Nov 2026 🚧

| Task | Exit criteria |
|:--|:--|
| CI | Format, vet, test, native build, WASM build on every PR |
| Pages | Playground deployed from `main` on every push |
| Golden tests | Snapshot output for `examples/aws-web-app` in both targets |
| Docs | Architecture, IR reference, Terraform version support, contributing, roadmap, README |
| Versioning | `openiac version` reads a single source of truth |

---

## v0.2 — Depth · Dec 2026 ⏳

Make real modules convert without hand-editing.

| Feature | Notes | Exit criteria |
|:--|:--|:--|
| Data sources | `data "aws_ami"`, `data "aws_availability_zones"` | CDK emits `Vpc.fromLookup`-style helpers or warns precisely; Pulumi emits `getAmi` |
| `locals` | Inline when static, warn when computed | No unresolved `local.*` in output |
| Functions | `jsonencode`, `file`, `cidrsubnet`, `lookup` | Supported subset documented |
| More L2 mappings | `aws_iam_role` → `iam.Role`, `aws_lambda_function` → `lambda.Function` | Snapshot tests per construct |
| Better warnings | Source file + line + suggested action | Warnings carry `file:line` and a stable `W_*` code |

---

## v0.3 — Scale · Q1 2027 ⏳

Handle the modules people actually have.

| Feature | Notes | Exit criteria |
|:--|:--|:--|
| Modules | Expand local paths; remote modules behind a flag | Nested module in a fixture converts |
| `count` | Constant counts → loop or repeated resources | Warn on computed counts |
| `for_each` | Static maps/sets → loops | Warn on computed collections |
| Multi-file layout | One output file per logical group | Imports resolve without cycles |
| Provider config | Region, profile, assume-role → target config | Pulumi `aws.Provider`, CDK env |

---

## v1.0 — Stable · Q2 2027 🎯

| Goal | Exit criteria |
|:--|:--|
| Stable IR schema | `1.0` with a documented migration policy |
| Plugin API | Third-party emitters without forking core |
| Terraform backend | HCL as a target, round-trip tested |
| CloudFormation target | YAML/JSON emitter sharing the same mappings |
| Coverage | Top 50 AWS resource types mapped |
| Quality | `go vet`, race tests, fuzz tests on the parser, ≥80% coverage |

---

## Beyond · 2027+ 🌅

| Direction | Why it is interesting |
|:--|:--|
| Azure / Bicep | Same IR, new mapping registry |
| GCP / Config Connector | Same shape as AWS mappings |
| Kubernetes / Helm | Structural conversion, not templating |
| Python emitters | CDK Python and Pulumi Python |
| Go emitters | Pulumi Go, CDK for Go via jsii |
| Drift view | Compare IR against deployed state (read-only) |

---

## Explicit non-goals

| Not planned | Reason |
|:--|:--|
| `terraform plan` / `apply` | Belongs to the target tool |
| State file management | Out of scope; migration is a separate concern |
| Reading `.terraform/` | Provider schemas are not a dependency |
| Executing HCL expressions | Static analysis only; no arbitrary evaluation |
| Silent best-effort output | Warnings are a contract |

---

## How priorities are chosen

1. **Correctness bugs** first.
2. **Resources people ask for** — open an issue with a `.tf` snippet.
3. **Coverage of common patterns** (`count`, modules, data sources).
4. **New targets** only after the IR is stable.

Want to move something up? Open an issue titled `roadmap: <item>` with a real
Terraform snippet and the TypeScript you would expect.
