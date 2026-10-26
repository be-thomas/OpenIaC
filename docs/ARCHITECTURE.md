# Architecture

OpenIaC is a **compiler**, not a wrapper. It reads Terraform, builds a
target-agnostic graph, then renders that graph as TypeScript for AWS CDK or
Pulumi.

```text
.tf / .tf.json
   │
   ▼
┌────────────┐    ┌────────────┐    ┌────────────┐    ┌────────────┐
│  HCL2      │───▶│  Universal │───▶│ Dependency │───▶│  Emitters  │
│  parser    │    │    IR      │    │    DAG     │    │ CDK/Pulumi │
└────────────┘    └────────────┘    └────────────┘    └────────────┘
   hashicorp/        Go structs        Kahn sort        TS source
   hcl/v2            (+ JSON tags)     + cycle check
```

Everything between the parser and the emitters is shared. Adding a target means
writing one package, not a new frontend.

## Package map

| Package | Responsibility |
|:--|:--|
| [`pkg/parsers/hcl`](../pkg/parsers/hcl/) | Reads native HCL and JSON (`.tf` and `.tf.json`) into the IR. The only package that knows Terraform exists. |
| [`pkg/ir`](../pkg/ir/) | Graph types, edge bookkeeping, deterministic topological order, cycle reporting. Contract documented in [IR.md](IR.md). |
| [`pkg/tfversion`](../pkg/tfversion/) | Maps detected language features to `min_version`; parses `required_version` constraints. |
| [`pkg/mappings`](../pkg/mappings/) | Embedded JSON registry describing target constructs per resource type. |
| [`pkg/emitutil`](../pkg/emitutil/) | Naming, reference rewriting, value/template rendering shared by emitters. |
| [`pkg/emitters/awscdk`](../pkg/emitters/awscdk/) | `aws-cdk-lib` TypeScript, L1 and L2 constructs. |
| [`pkg/emitters/pulumi`](../pkg/emitters/pulumi/) | `@pulumi/aws` TypeScript. |
| [`pkg/engine`](../pkg/engine/) | Public API: parse, convert, marshal, target parsing. |
| [`wasm`](../wasm/) + [`playground`](../playground/) | Browser build and UI. |

## The IR contract

The IR is deliberately flat. It has no loops, conditionals or expressions the
emitter must evaluate — those are either resolved or reported as a warning.

Schema `0.2` is additive over `0.1`. The full reference — every field, value
kind, guarantee and the extension policy — is in [IR.md](IR.md).

```go
type Graph struct {
    Version    string
    Terraform  *Terraform  // merged terraform block settings
    Resources  []*Resource // managed resources and data sources
    Variables  []*Variable
    Locals     []*Local
    Outputs    []*Output
    Modules    []*Module
    Providers  []*Provider
    Preserved  []Block     // moved, import, removed, check, ephemeral, action, unknown
    Features   []string    // detected Terraform language features
    MinVersion string      // oldest Terraform release that accepts the input
    Edges      []Edge
}
```

`Preserved` is how the parser accepts the whole Terraform 1.x language without a
typed node per construct: anything OpenIaC cannot convert is kept in the graph so
an emitter can warn about it, and never enters `Resources` (both emitters treat
every entry there as a managed resource). See
[TERRAFORM-VERSIONS.md](TERRAFORM-VERSIONS.md).

### Nodes

A `Resource.ID` is the Terraform address:

| Terraform | `ID` |
|:--|:--|
| `resource "aws_vpc" "main"` | `aws_vpc.main` |
| `data "aws_ami" "ubuntu"` | `data.aws_ami.ubuntu` |
| `variable "region"` | `var.region` (edge target only) |

### Values

Every attribute becomes an `ir.Value`:

| `Kind` | Payload | Notes |
|:--|:--|:--|
| `string` | `Literal` | Plain string |
| `number` | `Literal` | `int64` when exact, else `float64` |
| `bool` | `Literal` | |
| `null` | — | Rendered as `undefined` in TypeScript |
| `list` / `map` | `Items` / `Fields` | Recursive, may mix literals and expressions |
| `expr` | `Raw`, `Expr`, `Refs` | Unresolved HCL source plus the addresses it reads |

`structuredValue` in the parser is what makes mixed collections work:
`tags = { Name = "web-${var.env}", Env = var.env }` keeps its shape instead of
becoming an opaque string.

### Edges

An edge means **`From` depends on `To`** — `To` must be emitted first.

| Kind | Source |
|:--|:--|
| `reference` | Any traversal in attributes or nested blocks |
| `depends_on` | Explicit `depends_on` meta-argument |

`Graph.Order` uses Kahn's algorithm. Ties are broken lexicographically, which
makes output byte-stable across runs. On a cycle it returns the path
(`a -> b -> c -> a`) rather than a generic error.

## Reference rewriting

The emitters never pattern-match Terraform themselves. They build a
`RefRewriter` from the address list:

| Terraform | CDK | Pulumi |
|:--|:--|:--|
| `aws_vpc.main.id` | `mainVpc.ref` | `mainVpc.id` |
| `aws_security_group.web.id` | `webSecurityGroup.attrGroupId` | `webSecurityGroup.id` |
| `data.aws_ami.ubuntu.id` | warned (unsupported) | warned |
| `var.environment` | `environment` | `environment` |

Rules the rewriter enforces:

- **Longest address wins** (`data.aws_ami.ubuntu` before `data.aws_ami`).
- **Identifier boundaries** — `aws_vpc.mainx` is not a reference.
- **Quoted templates** — only the interior of `${...}` and `%{...}` is rewritten.
- **Comments** — `#` and `//` tails are preserved verbatim.
- **Heredocs** — detected and left untouched by v0.

## Mapping registry

Mappings are data. `pkg/mappings/aws.json` is embedded with `go:embed`, so no
runtime files are needed.

```jsonc
{
  "aws_vpc": {
    "cdk": {
      "kind": "l1",              // "l1" → Cfn* construct, "l2" → opinionated
      "module": "aws-ec2",
      "class": "CfnVPC",
      "properties": {
        "cidr_block": "cidrBlock",
        "tags": { "name": "Tags", "transform": "cfnTags" }
      },
      "blocks": {},
      "attributes": { "id": "ref", "arn": "attrArn" }
    },
    "pulumi": {
      "module": "ec2",
      "class": "Vpc",
      "properties": { "cidr_block": "cidrBlock" }
    }
  }
}
```

| Field | Meaning |
|:--|:--|
| `kind` | `l1` uses CloudFormation-shaped props, `l2` uses the curated construct API |
| `properties` | Terraform attribute → target property; `{name, transform}` or a bare string |
| `blocks` | Nested block → property (see shapes below) |
| `attributes` | Terraform attribute suffix → target identifier used by references |
| `transform` | `cfnTags` (map → `[{Key,Value}]`), `json` (string → `JSON.parse`), `first` (list → first item) |

Block shapes:

| `shape` | Output |
|:--|:--|
| `array` (default) | `property: [{...}, {...}]` |
| `object` | `property: {...}` |
| `presence` | `property: value` (default `true`) |

## Emitter contract

```go
type Emitter interface {
    Name() string
    Emit(graph *ir.Graph, options Options) (*Result, error)
}
```

To add a target:

1. Create `pkg/emitters/<name>`.
2. Implement `Emit`: call `graph.Order()`, build identifiers, construct one
   `RefRewriter`, render each node.
3. Collect every unsupported construct in `emitters.Warnings` — never drop it.
   Call `emitters.WarnUnsupported(graph, warnings)` once, right after creating
   the collector, so preserved blocks, unconverted `terraform` settings and
   conditions are reported with their stable `W_*` codes.
4. Register the target in `engine.NewEmitter` and `engine.ParseTargets`.

## Determinism

Generated files must be reproducible. That means:

- sorted imports, modules and property keys;
- sorted topological ties;
- no timestamps, hostnames or random names in output.

CI should be able to diff two runs and get zero changes.

## Non-goals (v0)

- Evaluating Terraform functions, `count`, `for_each` or modules.
- Managing state, planning or applying infrastructure.
- Reading `.tfstate`.
- Emitting back to HCL.

These are tracked in [ROADMAP.md](ROADMAP.md).
