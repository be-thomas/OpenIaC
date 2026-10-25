# Contributing to OpenIaC

Thanks for helping. This project is small on purpose: one parser, one IR, two
emitters, and a JSON registry that most contributions touch.

## Requirements

| Tool | Version |
|:--|:--|
| Go | 1.27+ |
| Make | any |
| Node (optional) | only for local playground tooling |

No Terraform installation is required — the parser reads HCL directly.

## Setup

```bash
git clone https://github.com/be-thomas/OpenIaC.git
cd OpenIaC
make hooks      # install the commit-msg hook
make build      # bin/openiac
make test       # go test ./...
make vet        # go vet ./...
make wasm       # playground/main.wasm + wasm_exec.js
```

Serve the playground locally:

```bash
python3 -m http.server -d playground 8080
```

## The fastest contribution: a resource mapping

You do **not** need Go to add AWS resource support.

1. Open [`pkg/mappings/aws.json`](pkg/mappings/aws.json).
2. Add an entry under `resources`.
3. Run `go test ./...` (the mapping tests validate the embedded JSON).
4. Add the resource to the coverage table in `README.md`.

```jsonc
"aws_kms_key": {
  "cdk": {
    "kind": "l1",
    "module": "aws-kms",
    "class": "CfnKey",
    "properties": {
      "description": "description",
      "enable_key_rotation": "enableKeyRotation",
      "tags": { "name": "Tags", "transform": "cfnTags" }
    },
    "attributes": { "arn": "attrArn", "key_id": "attrKeyId" }
  },
  "pulumi": {
    "module": "kms",
    "class": "Key",
    "properties": { "description": "description" }
  }
}
```

### Choosing L1 vs L2

| Use | When |
|:--|:--|
| `l1` (`Cfn*`) | Default. Property names mirror CloudFormation, so Terraform attributes map almost 1:1. |
| `l2` | The curated construct is clearly better and its props are known (e.g. `s3.Bucket`, `sns.Topic`). |

### Transforms

| Transform | Input | Output |
|:--|:--|:--|
| `cfnTags` | `{ Name = "x" }` | `[{ Key: "Name", Value: "x" }]` |
| `json` | `jsonencode(...)` string | `JSON.parse(...)` |
| `first` | `["a", "b"]` | `"a"` |

If a resource needs a new transform, add it to `transform()` in both emitters
and document it here.

## Changing the parser

- Keep output deterministic: sort maps before iterating.
- Never silently discard a construct. If it cannot be represented, add a
  warning through `emitters.Warnings`.
- Preserve `Raw` on expressions — the rewriter depends on the original text.
- Add a table test in [`pkg/parsers/hcl/parser_test.go`](pkg/parsers/hcl/parser_test.go).

## Adding an emitter

See [ARCHITECTURE.md](docs/ARCHITECTURE.md#emitter-contract). Register it in:

- `pkg/engine.NewEmitter`
- `pkg/engine.ParseTargets`
- the `--target` help text in `cmd/openiac/main.go`

## Tests

| Layer | What to cover |
|:--|:--|
| `pkg/parsers/hcl` | Literals, structured values, blocks, refs, depends_on, errors |
| `pkg/ir` | Ordering, determinism, cycles, stats |
| `pkg/emitutil` | Naming edge cases, boundary matching, templates, tags |
| `pkg/mappings` | Every mapping the emitters rely on |
| `pkg/engine` | End-to-end snapshots of CDK and Pulumi output |

Run everything before opening a PR:

```bash
make fmt && make vet && make test
```

`gofmt -l .` should print nothing.

## Code style

- Standard Go. No frameworks.
- Comments explain *why*, not *what*.
- Keep the IR free of target-specific fields. Target quirks belong in the
  emitter or the mapping JSON.
- Error messages: lowercase, contextual, actionable.

## Pull requests

- One concern per PR.
- Include a test that fails before the change.
- If output changes, include a before/after snippet.
- Commit messages: imperative mood, scoped (`parser: keep map shape when mixed`).
- Commit messages must not mention an AI assistant or model, and must not
  include `Co-Authored-By` or any other attribution trailer. This is a hard
  rule; `make hooks` enables the local check and CI enforces it too.

## Reporting bugs

Please include:

1. The smallest `.tf` file that reproduces it.
2. `openiac ir` output (it contains no secrets).
3. Expected vs actual TypeScript.
4. `openiac version`.

## License

Contributions are accepted under GPL-3.0.
