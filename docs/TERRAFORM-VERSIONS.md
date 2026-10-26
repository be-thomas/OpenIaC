# Terraform version support

OpenIaC reads Terraform **1.0 through 1.17**. That is a statement about
*parsing*, not about conversion: every valid 1.x configuration is accepted and
represented, while the set of constructs that produce CDK or Pulumi code is
smaller and grows over time.

Three rules follow from that split:

- **Nothing valid is rejected.** A construct the IR has no node for is preserved;
  an expression no emitter understands is kept as source text.
- **Nothing unconverted is silent.** Every construct that reaches an emitter
  without a target representation produces a coded warning. See
  [CONFORMANCE.md](CONFORMANCE.md#warning-taxonomy).
- **`min_version` is derived, not declared.** It reports the oldest Terraform
  release that accepts the input, whether or not the author wrote a
  `required_version` constraint.

Support is implemented in [package `tfversion`](../pkg/tfversion/), which is
pure: it depends only on `pkg/ir` and `golang.org/x/mod/semver`, and it never
fails a parse.

## Parse versus convert

| | Parse | Convert |
|:--|:--|:--|
| HCL2 syntax (0.12+) | ✅ | ✅ |
| `resource`, `data`, `variable`, `locals`, `output`, `module`, `provider` | ✅ | ✅ for the mapped subset |
| `terraform` settings (`required_version`, `required_providers`, backend, cloud, `provider_meta`) | ✅ | ❌ state and metadata stay with Terraform |
| `moved`, `import`, `removed`, `check`, `ephemeral`, `action` | ✅ preserved | ❌ no target equivalent |
| `precondition` / `postcondition`, `validation` | ✅ preserved | ❌ warned |
| Anything newer than 1.17 | ✅ preserved as an unknown block | ❌ warned |

## Release timeline

| Release | Date | Language additions tracked by OpenIaC |
|:--|:--|:--|
| 1.0.0 | 2021-06-08 | HCL2 configuration baseline: `terraform`, providers, resources, data, modules |
| 1.1.0 | 2021-12-08 | `moved` blocks, `cloud` block, variable `nullable` |
| 1.2.0 | 2022-05-18 | `precondition` / `postcondition`, `lifecycle.replace_triggered_by` |
| 1.3.0 | 2022-09-21 | `optional()` object attributes, `moved` across module boundaries |
| 1.4.0 | 2023-03-08 | built-in `terraform_data` resource |
| 1.5.0 | 2023-06-12 | `check` blocks, `import` blocks |
| 1.6.0 | 2023-10-04 | `import.id` as an expression |
| 1.7.0 | 2024-01-17 | `removed` blocks, `import` with `for_each` |
| 1.8.0 | 2024-04-10 | `moved` across resource types |
| 1.9.0 | 2024-06-26 | destroy-time provisioners in `removed` blocks |
| 1.10.0 | 2024-11-27 | `ephemeral` resource blocks, ephemeral variables and outputs |
| 1.11.0 | 2025-02-27 | write-only arguments (no syntactic marker) |
| 1.12.0 | 2025-05-14 | import `identity` |
| 1.13.0 | 2025-08-20 | (no tracked language feature) |
| 1.14.0 | 2025-11-19 | `action` blocks |
| 1.15.0 | 2026-04-29 | `deprecated` on variables and outputs, output `type` |
| 1.16.0 | 2026-08-26 | `lifecycle.destroy`, `terraform_data` `store`, action lifecycle events |
| 1.17.0 | 2026-10-21 | variables and locals in provider requirements |

The right-hand column lists what OpenIaC can identify from syntax alone, not
everything each release shipped. Tooling changes (`terraform test`, `terraform
query`, `terraform stacks`) have no configuration-language marker and are not
tracked.

## Feature registry

`pkg/tfversion` holds one row per feature. A row is:

```go
{
    name:        "check",          // stable identifier in Graph.Features
    since:       "1.5.0",          // canonical release that introduced it
    description: "check blocks",
    reliable:    true,             // may raise min_version
    detect:      func(s *scan) bool { … },
}
```

| Feature | Since | Detected from |
|:--|:--|:--|
| `action` | 1.14.0 | a preserved `action` block |
| `action_lifecycle_events` | 1.16.0 | `on_failure`, `before_destroy` or `after_destroy` inside an `action` block |
| `check` | 1.5.0 | a preserved `check` block |
| `cloud` | 1.1.0 | `Terraform.Cloud` |
| `conditions` | 1.2.0 | `precondition` / `postcondition` on a resource or output |
| `deprecated` | 1.15.0 | a non-empty `Variable.Deprecated` or `Output.Deprecated` |
| `ephemeral_resources` | 1.10.0 | a preserved `ephemeral` block |
| `ephemeral_values` | 1.10.0 | `Variable.Ephemeral` or `Output.Ephemeral` |
| `import` | 1.5.0 | a preserved `import` block |
| `import_expression` | 1.6.0 | an `import.id` that is not a literal string |
| `import_for_each` | 1.7.0 | an `import` block with `for_each` |
| `import_identity` | 1.12.0 | an `import` block with `identity` |
| `lifecycle_destroy` | 1.16.0 | `destroy` in a **resource's** `lifecycle` |
| `moved` | 1.1.0 | a preserved `moved` block |
| `moved_cross_type` | 1.8.0 | `from` / `to` in a `moved` block name different resource types |
| `moved_module` | 1.3.0 | `from` / `to` in a `moved` block name a `module.` segment |
| `nullable` | 1.1.0 | `Variable.Nullable` is set |
| `optional_attributes` | 1.3.0 | `optional(` in a variable's type constraint |
| `output_type` | 1.15.0 | a non-empty `Output.Type` |
| `provider_requirements_expressions` | 1.17.0 | a `required_providers` value that reads a variable or local |
| `removed` | 1.7.0 | a preserved `removed` block |
| `removed_provisioners` | 1.9.0 | a `provisioner` inside a `removed` block |
| `replace_triggered_by` | 1.2.0 | `replace_triggered_by` in a **resource's** `lifecycle` |
| `terraform_data` | 1.4.0 | a resource of type `terraform_data` |
| `terraform_data_store` | 1.16.0 | `store` on a `terraform_data` resource |

### Reliable versus unreliable

A detector can be syntactically correct and still report more than it should. If
it might fire on a configuration that works on an older release, it is marked
`reliable: false`:

| Feature | Why |
|:--|:--|
| `moved_cross_type` | `addressType` is an approximation: it reads the second-to-last segment of the address and cannot see provider-specific variations. |
| `action_lifecycle_events` | `on_failure` is also a provisioner argument; only the enclosing `action` block makes it a 1.16 construct, and action blocks themselves require 1.14. |

Unreliable features still appear in `features` — they are useful signal — but
they never raise `min_version`. Reporting an extra feature costs a reader a
minute; forcing a user onto a newer Terraform than they need can break their
pipeline.

The detectors deliberately look at narrow places:

- `lifecycle_destroy` and `replace_triggered_by` inspect only **resource**
  lifecycle blocks. A `lifecycle { destroy = false }` inside a `removed` block
  exists since 1.9 and must not be reported as the 1.16 resource argument.
- `conditions` inspects resources and outputs only. A `check` block's `assert`
  is not a precondition.
- `terraform_data_store` requires the resource type to be `terraform_data`.

### Deliberately not tracked

| Construct | Why |
|:--|:--|
| Write-only arguments (1.11) | Provider-schema dependent; no syntactic marker. |
| `issensitive`, `templatestring`, `convert`, `provider::` functions | Function availability is release-specific but expressions are preserved anyway; a false bump is worse than a conservative one. |
| Cross-object references in variable validation (1.9) | `validation` exists since 0.13; the reference form cannot be told apart reliably. |
| Provider-defined functions (1.8) | Same: preserved as expressions, no reliable version signal. |

If one of these becomes reliably detectable, add a row and a sample test.

## `required_version` and `min_version`

`Terraform.RequiredVersion` keeps the constraint **exactly as written**. It is
never validated at parse time: an unparseable constraint must not fail a
conversion.

`tfversion.ParseConstraint` understands the operators Terraform accepts in
`required_version`:

| Operator | Lower-bound contribution |
|:--|:--|
| bare `1.2.3`, `= 1.2.3` | `1.2.3` |
| `>= 1.2.3` | `1.2.3` |
| `~> 1.2.3`, `~> 1.2` | `1.2.3`, `1.2.0` |
| `> 1.2.3` | `1.2.4` |
| `<`, `<=`, `!=` | none |
| comma-separated list | the maximum of the clauses |

The result is `Graph.MinVersion`, the newest of:

1. the baseline `1.0.0`;
2. the constraint's lower bound (if it parses);
3. the `Since` value of every **reliable** detected feature.

`~>` is treated as a lower bound only: OpenIaC does not model upper bounds,
because `min_version` answers "what is the oldest release that works", not "what
is the highest release that works".

Examples:

| Input | `min_version` |
|:--|:--|
| `resource "aws_vpc" "main"` | `1.0.0` |
| `required_version = ">= 0.12"` | `1.0.0` (baseline wins) |
| `required_version = ">= 1.6.2"` | `1.6.2` |
| a `check` block, no constraint | `1.5.0` |
| a `check` block, `>= 1.6.0` | `1.6.0` |
| an `action` block and a `moved` block | `1.14.0` |
| an unreliable detector firing alone | unchanged |

## JSON configuration

`ParseDir` reads `*.tf` and `*.tf.json` in the same directory and may mix them.
The suffix chooses the syntax: `hclparse.ParseHCL` for `.tf`,
`hclparse.ParseJSON` for `.tf.json` (case-insensitive).

Two differences matter to consumers:

- **Top level is recovered structurally.** A top-level schema gives every block
  type its exact label count, which is how `.tf.json` blocks and labels are read
  back. A label mismatch is an error, matching Terraform.
- **Bodies below the top level are flat.** HCL exposes no block structure inside
  a JSON body, so an object property becomes a `map` value rather than a nested
  `Block`. `lifecycle`, `default_tags`, `provisioner` and friends therefore
  appear in `attributes`, not `blocks`. Every consumer must accept both shapes.

Expressions keep working: a JSON string containing `${…}` or `%{…}` becomes an
`expr` value with its references collected, and is normalised to a quoted HCL
template so the emitters render identical TypeScript for both syntaxes.

## Out of scope

OpenIaC reads **configuration** files for whole modules. It does not read:

| Path | Why |
|:--|:--|
| `.tftest.hcl` | Test definitions, not infrastructure |
| `.tfquery.hcl` | `terraform query` consultations |
| `.tfstack.hcl`, `.tfcomponent.hcl` | Stack and component definitions |
| `.tfvars`, `.tfvars.json` | Input values, not configuration |
| `.tfstate`, `.tfstate.backup` | State |
| `.terraform.lock.hcl`, `.terraform/` | Dependency material and provider plugins |
| Policy files (Sentinel, OPA) | Different language |

Directories are not searched recursively: a subdirectory is usually a module,
and a module call is a node, not an inlined subtree.

## Adding the next Terraform release

1. Read the changelog and list new **top-level block types**, **meta-arguments**
   and **nested blocks**.
2. For a new top-level block: add it to `topLevelSchema` and
   `topLevelLabelCount` in `pkg/parsers/hcl/schema.go` if its label count is
   fixed, so JSON documents are read correctly. Unknown types are already
   preserved — the schema entry is what makes labels work.
3. Add a `registry` row in `pkg/tfversion/features.go`: name, `since`,
   description, reliability and a detector. Prefer a narrow detector over a
   generous one; mark it unreliable if it can fire on older configuration.
4. Add one minimal sample to `versionCases` in
   `pkg/parsers/hcl/versions_test.go`. `TestEveryFeatureHasSample` fails if a
   registered feature has no sample, and each case also asserts `min_version`,
   so an over-reporting detector fails a test rather than a user.
5. Bump `tfversion.Latest`.
6. Update this file and, if the IR shape changes, [IR.md](IR.md) and
   `ir.Version`.
7. Run the full suite and the corpus:

```bash
go test ./...
OPENIAC_CORPUS=1 go test ./pkg/parsers/hcl -run TestParseConformanceCorpus -count=1
```

## Corpus validation

The conformance corpus contains 853 real-world modules. Parsing all of them is
the cheapest check that a change did not break syntax nobody wrote a unit test
for, so it is opt-in and runs on a schedule rather than on every edit:

```bash
OPENIAC_CORPUS=1 go test ./pkg/parsers/hcl -run TestParseConformanceCorpus -count=1
```

The test fails if any module fails to parse, or produces a document without a
schema version or `min_version`.
