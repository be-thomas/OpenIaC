# The OpenIaC IR

The intermediate representation (IR) is the contract between parsers and
emitters. A parser turns one input language into an IR document; an emitter
renders that document in one target language. Everything the project does
happens on one side of that contract or the other, so the IR's shape is the
single most important design decision in the codebase.

This document is the reference for that shape. It describes the JSON document,
the Go types in [`pkg/ir`](../pkg/ir/), and the guarantees a consumer may rely
on.

- [Stability promise](#stability-promise)
- [Document envelope](#document-envelope)
- [Graph fields](#graph-fields)
- [Terraform metadata](#terraform-metadata)
- [Nodes](#nodes)
- [Values](#values)
- [Edges](#edges)
- [Features and minimum version](#features-and-minimum-version)
- [Preserved configuration](#preserved-configuration)
- [Determinism](#determinism)
- [Go API](#go-api)
- [Extension policy](#extension-policy)
- [Limitations and non-goals](#limitations-and-non-goals)
- [Complete example](#complete-example)

## Stability promise

The current schema is **`0.2`**, exported as `ir.Version`. While the major
version is `0`, a minor bump means the schema changed. The rules:

| Change | Version effect |
|:--|:--|
| New optional field, new `Kind`, new `EdgeKind` | Minor bump (`0.1` → `0.2`) |
| Field renamed, removed, or changed in meaning | Major bump (`0.2` → `1.0`) |
| Behaviour of an existing field changes | Treated as a meaning change |

Everything added since `0.1` is additive and `omitempty`: a `0.1` document
decodes into today's structs with the new fields left at their zero values, and
a `0.2` document that uses none of them serializes the same as before.

Consumers should:

- **Ignore unknown fields.** Do not reject a document because a newer producer
  added a key.
- **Check `version` before trusting a field.** Read the major version; refuse a
  document from a newer major version.
- **Never assume a field is present.** Every optional field is either
  `omitempty` or documented as always present.

## Document envelope

A document is one JSON object:

```jsonc
{
  "version": "0.2",           // always present
  "source": "main.tf",        // optional: how the input was named
  "terraform": { … },         // optional
  "resources": [ … ],         // always present, may be empty
  "variables": [ … ],         // optional
  "locals": [ … ],            // optional
  "outputs": [ … ],           // optional
  "modules": [ … ],           // optional
  "providers": [ … ],         // optional
  "preserved": [ … ],         // optional
  "features": [ … ],          // optional
  "min_version": "1.6.0",     // optional
  "edges": [ … ]              // always present, may be empty
}
```

`resources` and `edges` are the only collections without `omitempty`, because
consumers index into them and an explicit `[]` is friendlier than a missing
key. `ir.New` initialises both so a graph never serializes `null`.

## Graph fields

| Field | Go type | Meaning |
|:--|:--|:--|
| `version` | `string` | Schema version of this document. |
| `source` | `string` | The path or name the input came from. Parsers pass a comma-joined path list for multi-file modules. Informational only. |
| `terraform` | `*Terraform` | Merged `terraform` block settings. |
| `resources` | `[]*Resource` | Managed resources and data sources. |
| `variables` | `[]*Variable` | Declared input variables. |
| `locals` | `[]*Local` | Local values, one entry per attribute. |
| `outputs` | `[]*Output` | Declared outputs. |
| `modules` | `[]*Module` | Module calls. |
| `providers` | `[]*Provider` | Provider configuration blocks. |
| `preserved` | `[]Block` | Top-level configuration the IR has no typed node for. |
| `features` | `[]string` | Terraform language features detected in the input, sorted. |
| `min_version` | `string` | Oldest Terraform release that accepts the input. |
| `edges` | `[]Edge` | Dependency edges between nodes. |

## Terraform metadata

`terraform { … }` is the one block that is not a node: it is module-wide
settings, and two of them (`required_version`, `required_providers`) describe
what Terraform must be able to run the module at all. Every `terraform` block in
the module is merged into one object.

```go
type Terraform struct {
    RequiredVersion   string           // verbatim, e.g. ">= 1.6.2"
    RequiredProviders map[string]Value // local name -> {source, version, …}
    Experiments       []string
    Backend           *Block           // Labels[0] is the backend type
    Cloud             *Block           // HCP Terraform / Terraform Cloud
    ProviderMeta      []Block          // Labels[0] is the provider name
    Attributes        map[string]Value // settings the IR does not model
    Blocks            []Block          // nested blocks the IR does not model
    File, Line
}
```

`required_version` is kept exactly as written. It is **not** validated at parse
time: an unparseable constraint must not fail a conversion. Use
`tfversion.ParseConstraint` if you need the lower bound.

`RequiredProviders` normalises both syntaxes to one shape. Native HCL writes it
as a block with one attribute per provider; `.tf.json` writes it as an object
property. Both arrive as a map from local name to a `map` value holding
`source`, `version` and `configuration_aliases`.

`Backend` is the legacy partial-configuration backend. `Cloud` is the HCP
Terraform alternative. Both are preserved raw; no emitter converts state
management.

## Nodes

A node is anything with an address that another node can reference.

| Terraform block | Go type | Address |
|:--|:--|:--|
| `resource "aws_vpc" "main"` | `*Resource`, `Kind == "resource"` | `aws_vpc.main` |
| `data "aws_ami" "ubuntu"` | `*Resource`, `Kind == "data"` | `data.aws_ami.ubuntu` |
| `variable "region"` | `*Variable` | `var.region` |
| `locals { name = … }` | `*Local` | `local.name` |
| `output "id"` | `*Output` | `output.id` |
| `module "vpc"` | `*Module` | `module.vpc` |
| `provider "aws"` | `*Provider` | `provider.aws` |

Only nodes in `resources` participate in `Order` and `Stats`. Variables, locals,
modules and outputs are listed so emitters can declare or reference them, and so
`HasNode` can tell a real reference from a bare identifier.

### Typed fields versus everything else

Every node has a small set of typed fields — the ones emitters need by name —
plus two catch-alls:

- `attributes map[string]Value` — settings the IR does not model
- `blocks []Block` — nested blocks the IR does not model

The rule is: **typed fields are conveniences; `attributes` and `blocks` keep
everything else**, and a key promoted to a typed field is removed from
`attributes`, so nothing appears twice. A consumer can therefore always treat
the typed field and the catch-alls together as the complete configuration.

Examples of typed fields: `Variable.Type`, `Variable.Nullable`,
`Variable.Ephemeral`, `Variable.Deprecated`, `Output.Type`, `Output.Value`,
`Resource.DependsOn`, `Resource.ProviderAlias`, `Module.Source`.

### Provenance

Top-level nodes and preserved blocks carry `file` and `line`; native nested
blocks carry both too. Blocks derived from a flat `.tf.json` body cannot be
located inside the file, so they carry the enclosing `file` and no `line`, and
warnings fall back to the file name alone. Locations exist so warnings can name
a position without the emitter re-reading the source. They are informational: a
document that moved between machines may have different paths.

`Provider` and `Module` also carry `file`/`line`; both were added in `0.2`.

### Blocks

```go
type Block struct {
    Type       string
    Labels     []string
    Attributes map[string]Value
    Blocks     []Block
    Body       *Value  // JSON escape hatch, see below
    File       string
    Line       int
}
```

`Block.Body` is set **only** for an unrecognised top-level property in a
`.tf.json` document whose value is not an object or an array of objects — for
example `"stray": 42`. Native HCL always has a decomposable shape, so it always
fills `Attributes`/`Blocks` instead. A consumer that understands `Attributes`
and `Blocks` can ignore `Body`, but ignoring it silently drops a value, so
prefer to warn when it is set.

## Values

Every attribute is an `ir.Value`:

```go
type Value struct {
    Kind    ValueKind
    Literal any
    Expr    string
    Raw     string
    Refs    []string
    Items   []Value
    Fields  map[string]Value
}
```

| `Kind` | Payload | Notes |
|:--|:--|:--|
| `string` | `Literal` (`string`) | Plain string, templates already resolved if they were interpolation-only |
| `number` | `Literal` | `int64` when the value is an exact integer, else `float64` |
| `bool` | `Literal` (`bool`) | |
| `null` | — | `Literal` is nil |
| `list` | `Items` | Recursive; may mix literals and expressions |
| `map` | `Fields` | Recursive; may mix literals and expressions |
| `expr` | `Expr`, `Raw`, `Refs` | An expression the parser could not evaluate |

### Numbers

Numbers are stored as Go values rather than strings so they can be rendered
directly. Exact integers become `int64`; everything else becomes `float64`.
After a JSON round trip both arrive as `float64`, because that is how
`encoding/json` decodes into `any` — consumers that need integer semantics
should check for an exact value rather than a Go type.

### Expressions

An `expr` value means "this is code, not data". The fields:

| Field | Meaning |
|:--|:--|
| `Expr` | The expression as it should be rendered. For native HCL, the source slice. For `.tf.json`, a template string is normalised to HCL syntax. |
| `Raw` | The original source text the expression came from. |
| `Refs` | Canonical addresses the expression reads, sorted and deduplicated. |

For native HCL `Expr` and `Raw` are the same string. They are separate so a
future producer can normalise one without losing the other.

`Refs` is how edges and reference rewriting work. A ref is the address of a
**node**, not of an attribute: `aws_vpc.main.id` contributes `aws_vpc.main`,
`var.region` contributes `var.region`, `data.aws_ami.ubuntu.id` contributes
`data.aws_ami.ubuntu`. Traversals rooted at `each`, `count`, `self`, `path` or
`terraform` contribute nothing, because they name no static node.

An expression is not failure. `var.environment` cannot be resolved without
Terraform's evaluation context, but it is perfectly convertible: the emitter
rewrites the reference. Only values that no emitter can express produce a
warning.

### Collection refs

A `list` or `map` may also carry `Refs`, aggregating every reference found
inside its members. This is a convenience for consumers that only want to know
"does this value depend on anything", without walking it. Do not treat a
collection's `Refs` as an alternative to its members: the members remain the
authoritative structure.

### Mixed collections

A collection that mixes literals and expressions keeps its shape:

```jsonc
"tags": {
  "kind": "map",
  "refs": ["var.environment"],
  "fields": {
    "Name": { "kind": "expr", "expr": "\"openiac-${var.environment}\"", "refs": ["var.environment"] }
  }
}
```

Evaluating that object fails, but the shape is static, so it is represented
rather than degraded to one opaque string. The template stays a quoted HCL
template, which is what lets the emitters render it as a TypeScript template
literal.

### `IsLiteral`

`Value.IsLiteral()` reports whether a value can be rendered without reference
rewriting. It is true for `string`, `number`, `bool`, `null`, and for `list` /
`map` whose members are all literal. It is the check an emitter uses to decide
whether it can emit a constant.

## Edges

An edge means **`From` depends on `To`**, so `To` must be declared first.

```go
type Edge struct {
    From string
    To   string
    Kind EdgeKind // "reference" or "depends_on"
}
```

| `Kind` | Created from |
|:--|:--|
| `reference` | Any traversal inside a resource's attributes or nested blocks |
| `depends_on` | The explicit `depends_on` meta-argument |

Rules:

- **Only resources and data sources produce edges.** An output reading
  `aws_vpc.main.id` does not create an edge, because outputs are not ordered.
- **An edge is only created when both ends name a node.** `HasNode` recognises
  resource and data source addresses and the `var.`, `local.` and `module.`
  forms, so those edges are kept; a reference such as `aws.west` or
  `path.module` is dropped.
- **Duplicates and self-edges are dropped.** `AddEdge` is idempotent; add them in
  any order.
- **Preserved blocks never create edges.** A `moved` or `import` block mentions
  addresses, but it is not a dependency.

`Graph.Order` runs Kahn's algorithm over resources. Ties are broken by address,
so the order is byte-stable across runs and machines. On a cycle it returns the
path:

```text
dependency cycle: a.a -> b.b -> c.c -> a.a
```

## Features and minimum version

`features` and `min_version` describe which Terraform release can run the input.
They are computed by [`pkg/tfversion`](../pkg/tfversion/) at parse time, so an
emitter never has to reason about versions.

- `features` is a sorted list of stable snake_case names, one per detected
  language feature (`moved`, `check`, `import_identity`, …). The registry and its
  detection rules are documented in
  [TERRAFORM-VERSIONS.md](TERRAFORM-VERSIONS.md).
- `min_version` is the newest of: the OpenIaC baseline `1.0.0`, the lower bound
  of `required_version`, and the introduction version of every **reliable**
  detected feature.

A detector that can misfire (for example "moved across resource types", which is
hard to tell apart syntactically) is marked unreliable: it still appears in
`features`, but never raises `min_version`. Reporting an extra feature is
harmless; forcing a user onto a newer Terraform than they need is not.

`required_version` is never rewritten or rejected. `min_version` is derived, not
declared: it is what the input needs, whether or not the author wrote a
constraint.

## Preserved configuration

Terraform 1.x has top-level constructs the IR has no typed node for. Rather than
drop them, the parser records them in `preserved`:

| Block | Since | Why it is preserved |
|:--|:--|:--|
| `moved` | 1.1 | State refactoring; no runtime object |
| `import` | 1.5 | Imports existing infrastructure |
| `removed` | 1.7 | Removes an object from state |
| `check` | 1.5 | Continuous validation, not infrastructure |
| `ephemeral` | 1.10 | Ephemeral resource lifecycle |
| `action` | 1.14 | Provider-defined invocable actions |
| anything unknown | — | Forward compatibility |

Preserved blocks are stored as ordinary `Block` values with their labels,
attributes, nested blocks, `file` and `line`. They are ordered by document
position, files in input order.

**Preserved blocks are not nodes.** They never appear in `resources`, `Order`,
`Stats` or `HasNode`, because both emitters treat every entry in `resources` as
a managed resource. A consumer that converts one must do so explicitly.

Emitters call `emitters.WarnUnsupported` once per run, which turns each
preserved block type, each unconverted terraform setting, each condition and each
variable validation into one coded warning. That is how the project keeps its
"warnings over silence" promise: nothing disappears without a line of output.

## Determinism

Two conversions of the same input must produce byte-identical output. The IR
carries no timestamps, hostnames, random identifiers or map iteration order:

- Every list is emitted in a stable order: nodes in source order, features
  sorted, edges in creation order, refs sorted, preserved blocks in document
  order.
- Maps are unordered in Go but `encoding/json` sorts their keys, so documents
  are stable on the wire.
- `Order` breaks ties by address.
- `source` is echoed exactly as given, so it can differ between machines; it is
  informational and not part of any generated code.

## Go API

```go
graph := ir.New("main.tf")             // version, source, empty resources/edges
graph.AddEdge("aws_subnet.public", "aws_vpc.main", ir.EdgeReference)

graph.HasNode("aws_vpc.main")          // true for resources, var.x, local.x, module.x
graph.ResourceByID("aws_vpc.main")     // *Resource or nil
graph.Dependencies("aws_subnet.public")// sorted addresses it depends on
graph.Order()                          // ([]*Resource, error); cycle error names the path
graph.Stats()                          // (resources, data, variables, outputs, edges int)

value.IsLiteral()                      // renderable without rewriting?
```

`pkg/engine` wraps these for whole conversions: `engine.Parse`,
`engine.ParseSource`, `engine.Convert`, `engine.ConvertSource` and
`engine.MarshalIR`.

## Extension policy

Adding a field is a minor version bump. To add one:

1. Give it a JSON tag with `omitempty` unless consumers must always see it.
2. Keep the typed-versus-catch-all rule: if a parser starts extracting a key into
   the new field, delete it from `attributes` so it is not duplicated.
3. If old documents must still work, document the zero value's meaning.
4. Update this file and the schema version in `pkg/ir/types.go`.
5. Add a round-trip test in `pkg/ir/graph_test.go`.

New `Kind` and `EdgeKind` values are additive: a consumer that does not know a
kind must treat it as "cannot convert" and warn, never as a literal.

## Limitations and non-goals

- **No evaluation.** The IR does not carry the result of `count`, `for_each`,
  functions or module expansion. Values that need them stay expressions.
- **No nested module graphs.** Modules are calls, not expanded subtrees.
- **No state.** The IR describes configuration, not resources that exist.
- **No imperative logic.** There are no conditionals or loops to render; a
  parser either resolves them or reports an expression.
- **JSON bodies are flat below the top level.** HCL exposes no block structure in
  a `.tf.json` body, so a nested block arrives as a `map` value. An emitter that
  knows its block schema recovers the native shape with
  `emitters.UnflattenBlocks`; otherwise the value stays flat. This is documented
  rather than guessed.
- **`line` is informational.** It is not stable across reformatting and must not
  be used for identity.

## Complete example

Input:

```hcl
terraform {
  required_version = ">= 1.6.2"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    bucket = "openiac-state"
    key    = "app/terraform.tfstate"
  }
}

moved {
  from = aws_vpc.legacy
  to   = aws_vpc.main
}

variable "environment" {
  type    = string
  default = "dev"
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "openiac-${var.environment}"
  }

  lifecycle {
    destroy = false
  }
}

output "vpc_id" {
  value = aws_vpc.main.id
}
```

Document (`openiac ir`):

```json
{
  "version": "0.2",
  "source": "main.tf",
  "terraform": {
    "required_version": ">= 1.6.2",
    "required_providers": {
      "aws": {
        "kind": "map",
        "fields": {
          "source": { "kind": "string", "literal": "hashicorp/aws" },
          "version": { "kind": "string", "literal": "~> 5.0" }
        }
      }
    },
    "backend": {
      "type": "backend",
      "labels": ["s3"],
      "attributes": {
        "bucket": { "kind": "string", "literal": "openiac-state" },
        "key": { "kind": "string", "literal": "app/terraform.tfstate" }
      },
      "line": 11
    },
    "file": "main.tf",
    "line": 1
  },
  "resources": [
    {
      "id": "aws_vpc.main",
      "kind": "resource",
      "type": "aws_vpc",
      "name": "main",
      "provider": "aws",
      "attributes": {
        "cidr_block": { "kind": "string", "literal": "10.0.0.0/16" },
        "tags": {
          "kind": "map",
          "refs": ["var.environment"],
          "fields": {
            "Name": {
              "kind": "expr",
              "expr": "\"openiac-${var.environment}\"",
              "raw": "\"openiac-${var.environment}\"",
              "refs": ["var.environment"]
            }
          }
        }
      },
      "blocks": [
        {
          "type": "lifecycle",
          "attributes": { "destroy": { "kind": "bool", "literal": false } },
          "line": 34
        }
      ],
      "file": "main.tf",
      "line": 27
    }
  ],
  "variables": [
    {
      "name": "environment",
      "type": "string",
      "default": { "kind": "string", "literal": "dev" },
      "file": "main.tf",
      "line": 22
    }
  ],
  "outputs": [
    {
      "name": "vpc_id",
      "value": {
        "kind": "expr",
        "expr": "aws_vpc.main.id",
        "raw": "aws_vpc.main.id",
        "refs": ["aws_vpc.main"]
      },
      "file": "main.tf",
      "line": 39
    }
  ],
  "preserved": [
    {
      "type": "moved",
      "attributes": {
        "from": {
          "kind": "expr",
          "expr": "aws_vpc.legacy",
          "raw": "aws_vpc.legacy",
          "refs": ["aws_vpc.legacy"]
        },
        "to": {
          "kind": "expr",
          "expr": "aws_vpc.main",
          "raw": "aws_vpc.main",
          "refs": ["aws_vpc.main"]
        }
      },
      "file": "main.tf",
      "line": 17
    }
  ],
  "features": ["lifecycle_destroy", "moved"],
  "min_version": "1.16.0",
  "edges": [
    { "from": "aws_vpc.main", "to": "var.environment", "kind": "reference" }
  ]
}
```

Notes on this document:

- The `moved` block is in `preserved`, not `resources`, so the graph has exactly
  one node to emit.
- `min_version` is `1.16.0` because the resource lifecycle sets `destroy`, which
  Terraform 1.16 introduced — not because of `required_version`, which allows an
  older release.
- The output references `aws_vpc.main` but creates no edge: outputs are not
  ordered against resources.
- The `tags` map keeps its shape and its reference instead of collapsing into one
  opaque string.
