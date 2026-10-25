<div align="center">

# OpenIaC

### Terraform in. Native TypeScript out.

**The open-source HCL → AWS CDK / Pulumi transpiler.**<br/>
One parser. One IR. Zero lock-in.

<br/>

[![License: GPL-3.0](https://img.shields.io/badge/License-GPL--3.0-22d3ee?style=for-the-badge&logo=gnu&logoColor=white)](#-license)
[![Go](https://img.shields.io/badge/Go-1.27-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev)
[![WebAssembly](https://img.shields.io/badge/WebAssembly-ready-654FF0?style=for-the-badge&logo=webassembly&logoColor=white)](playground/)
[![Status](https://img.shields.io/badge/status-alpha-fbbf24?style=for-the-badge)](#-timeline)
[![PRs](https://img.shields.io/badge/PRs-welcome-34d399?style=for-the-badge)](#-contributing)

[![Terraform](https://img.shields.io/badge/Terraform-HCL2-7B42BC?style=flat-square&logo=terraform&logoColor=white)](https://developer.hashicorp.com/terraform/language)
[![AWS CDK](https://img.shields.io/badge/AWS_CDK-TypeScript-FF9900?style=flat-square&logo=amazonaws&logoColor=white)](https://aws.amazon.com/cdk/)
[![Pulumi](https://img.shields.io/badge/Pulumi-TypeScript-8A3391?style=flat-square&logo=pulumi&logoColor=white)](https://www.pulumi.com/)
[![IR](https://img.shields.io/badge/IR-schema_0.1-818cf8?style=flat-square)](docs/ARCHITECTURE.md#the-ir-contract)
[![Made with Go](https://img.shields.io/badge/made_with-Go-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)

<br/>

**`terraform`** · **`hcl`** · **`hcl2`** · **`aws-cdk`** · **`pulumi`** · **`infrastructure-as-code`** · **`iac`** · **`transpiler`** · **`codemod`** · **`typescript`** · **`webassembly`** · **`devops`** · **`cloudformation`** · **`aws`** · **`golang`**

<br/>

[🚀 Playground](playground/) · [📦 Quick Start](#-quick-start) · [🗺️ Roadmap](docs/ROADMAP.md) · [🧠 How It Works](#-how-it-works) · [🧱 Architecture](docs/ARCHITECTURE.md) · [🧪 Conformance](docs/CONFORMANCE.md) · [🤝 Contribute](CONTRIBUTING.md)

</div>

---

## 🎯 The Aim

<table>
<tr>
<td width="33%" valign="top">

### 🔓 Break the lock-in

Your infrastructure is described in HCL. Your platform team writes TypeScript. OpenIaC translates one into the other — **structurally**, not with regex.

</td>
<td width="33%" valign="top">

### 🧠 Keep the meaning

Parsing happens with HashiCorp's own `hcl/v2` library. Dependencies become a real DAG. References are rewritten to the *target's* identifiers.

</td>
<td width="33%" valign="top">

### 🌍 Run anywhere

The same engine powers a CLI, a Go library, and a **zero-install browser playground** compiled to WebAssembly.

</td>
</tr>
</table>

> **One sentence:** migrate from Terraform to AWS CDK or Pulumi without hand-porting a thousand lines of HCL.

---

## 🤞 The Promises

| Promise | What it means |
|:--|:--|
| 🔒 **Nothing leaves your machine** | The playground runs the entire compiler in your browser tab. No account, no upload, no telemetry. |
| 🎯 **Deterministic output** | Same input → byte-identical TypeScript. Stable topological ordering with cycle detection. |
| 🧱 **Readable code, not spaghetti** | Named identifiers, typed variables, import blocks grouped by module — code a human would have written. |
| ⚠️ **Warnings over silence** | Unsupported constructs (`count`, `for_each`, modules) are flagged loudly. Never silently dropped. |
| 🧩 **Mappings are data** | Add a resource by editing [`aws.json`](pkg/mappings/aws.json) — no Go required. |
| 🪪 **Honest licensing** | GPL-3.0. Fork it, audit it, ship it. |

---

## ⚡ Quick Start

<table>
<tr><td>

**1 · Build the CLI**

```bash
git clone https://github.com/be-thomas/OpenIaC.git
cd OpenIaC
make build
```

</td><td>

**2 · Convert a module**

```bash
./bin/openiac convert ./examples/aws-web-app \
  --target both
```

</td><td>

**3 · Inspect the IR**

```bash
./bin/openiac ir ./examples/aws-web-app \
  | jq '.resources[].id'
```

</td></tr>
</table>

```text
openiac convert ./examples/aws-web-app --target both --stdout
```

<table>
<tr>
<td width="50%" valign="top">

**AWS CDK**

```ts
const mainVpc = new ec2.CfnVPC(this, "mainVpc", {
  cidrBlock: "10.0.0.0/16",
  enableDnsSupport: true,
  tags: [{ key: "Name", value: `openiac-${environment}` }],
});
```

</td>
<td width="50%" valign="top">

**Pulumi**

```ts
const mainVpc = new aws.ec2.Vpc("mainVpc", {
  cidrBlock: "10.0.0.0/16",
  enableDnsSupport: true,
  tags: { Name: `openiac-${environment}` },
});
```

</td>
</tr>
</table>

### 🖥️ Browser playground

```bash
make wasm                                        # builds main.wasm + wasm_exec.js
python3 -m http.server -d playground 8080        # serve it
```

Open `http://localhost:8080`, paste Terraform, get TypeScript. **No backend.**

---

## 🧠 How It Works

```text
┌──────────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  Terraform   │     │   Universal  │     │  Dependency  │     │   Native TS  │
│     HCL      │ ──▶ │      IR      │ ──▶ │     DAG      │ ──▶ │   emitters   │
│  .tf / .tfvars│     │  JSON-shaped │     │  topo-sorted │     │ CDK / Pulumi │
└──────────────┘     └──────────────┘     └──────────────┘     └──────────────┘
       ▲                    ▲                    ▲                     ▲
   hcl/v2 parser      resources+vars        Kahn's algorithm      JSON mappings
```

<table>
<tr>
<th align="left">Stage</th>
<th align="left">Package</th>
<th align="left">Job</th>
</tr>
<tr>
<td><b>Parse</b></td>
<td><a href="pkg/parsers/hcl/"><code>pkg/parsers/hcl</code></a></td>
<td>Reads <code>resource</code>, <code>data</code>, <code>variable</code>, <code>locals</code>, <code>output</code>, <code>module</code>, <code>provider</code> blocks into the IR.</td>
</tr>
<tr>
<td><b>Resolve</b></td>
<td><a href="pkg/ir/"><code>pkg/ir</code></a></td>
<td>Builds edges from traversals, sorts topologically, reports cycles with the offending path.</td>
</tr>
<tr>
<td><b>Map</b></td>
<td><a href="pkg/mappings/"><code>pkg/mappings</code></a></td>
<td>Looks up a resource's target class, property names, block shapes and output attributes.</td>
</tr>
<tr>
<td><b>Rewrite</b></td>
<td><a href="pkg/emitutil/"><code>pkg/emitutil</code></a></td>
<td>Turns <code>aws_vpc.main.id</code> into <code>mainVpc.ref</code> (CDK) or <code>mainVpc.id</code> (Pulumi) — including inside templates.</td>
</tr>
<tr>
<td><b>Emit</b></td>
<td><a href="pkg/emitters/"><code>pkg/emitters</code></a></td>
<td>Generates a complete, compilable TypeScript file per target.</td>
</tr>
</table>

---

## 📊 Feature Matrix

| Capability | AWS CDK | Pulumi | Status |
|:--|:--:|:--:|:--|
| Resources (L1 / L2 constructs) | ✅ | ✅ | **Stable** |
| Variables → context / config | ✅ | ✅ | **Stable** |
| Outputs | ✅ | ✅ | **Stable** |
| Interpolated strings & templates | ✅ | ✅ | **Stable** |
| Object & list literals | ✅ | ✅ | **Stable** |
| `depends_on` | ✅ | ✅ | **Stable** |
| Security group blocks | ✅ | ✅ | **Stable** |
| Provider alias | ✅ | ✅ | Partial |
| Data sources | ⚠️ | ⚠️ | Warned |
| `count` / `for_each` | ⚠️ | ⚠️ | Warned |
| `locals` | ⚠️ | ⚠️ | Warned |
| Modules | ⚠️ | ⚠️ | Warned |
| Non-AWS providers | ⚠️ | ⚠️ | Warned |

> Legend: ✅ emitted · ⚠️ converted with an explicit warning · ❌ not yet

---

## 🧩 Resource Coverage

<sub>14 AWS resources mapped out of the box · **add more in one JSON edit**</sub>

| Terraform | AWS CDK construct | Level | Pulumi class |
|:--|:--|:--:|:--|
| `aws_vpc` | `ec2.CfnVPC` | L1 | `ec2.Vpc` |
| `aws_subnet` | `ec2.CfnSubnet` | L1 | `ec2.Subnet` |
| `aws_security_group` | `ec2.CfnSecurityGroup` | L1 | `ec2.SecurityGroup` |
| `aws_internet_gateway` | `ec2.CfnInternetGateway` | L1 | `ec2.InternetGateway` |
| `aws_route_table` | `ec2.CfnRouteTable` | L1 | `ec2.RouteTable` |
| `aws_eip` | `ec2.CfnEIP` | L1 | `ec2.Eip` |
| `aws_instance` | `ec2.CfnInstance` | L1 | `ec2.Instance` |
| `aws_s3_bucket` | `s3.Bucket` | **L2** | `s3.Bucket` |
| `aws_sns_topic` | `sns.Topic` | **L2** | `sns.Topic` |
| `aws_sqs_queue` | `sqs.CfnQueue` | L1 | `sqs.Queue` |
| `aws_dynamodb_table` | `dynamodb.CfnTable` | L1 | `dynamodb.Table` |
| `aws_lambda_function` | `lambda.CfnFunction` | L1 | `lambda.Function` |
| `aws_iam_role` | `iam.CfnRole` | L1 | `iam.Role` |
| `aws_cloudwatch_log_group` | `logs.CfnLogGroup` | L1 | `cloudwatch.LogGroup` |

<details>
<summary><b>➕ Adding a resource (no Go required)</b></summary>

<br/>

```json
{
  "aws_kms_key": {
    "cdk": {
      "kind": "l1",
      "module": "aws-kms",
      "class": "CfnKey",
      "properties": { "description": "description" },
      "attributes": { "arn": "attrArn", "key_id": "attrKeyId" }
    },
    "pulumi": {
      "module": "kms",
      "class": "Key",
      "properties": { "description": "description" }
    }
  }
}
```

Add it to [`pkg/mappings/aws.json`](pkg/mappings/aws.json), rebuild, done.

</details>

---

## 🖥️ CLI Reference

| Command | Purpose |
|:--|:--|
| `openiac convert <file\|dir>` | Transpile Terraform into TypeScript |
| `openiac ir <file\|dir>` | Dump the universal JSON IR |
| `openiac version` | Print the version |

| Flag | Default | Meaning |
|:--|:--|:--|
| `--target` | `cdk` | `cdk`, `pulumi`, or `both` |
| `--out` | `out` | Output directory |
| `--stack` | `OpenIacStack` | Generated CDK stack class |
| `--stdout` | `false` | Print instead of write |
| `--json` | `false` | Machine-readable result |
| `-o` | — | IR output file (`ir` only) |

<details>
<summary><b>🧾 Example IR</b></summary>

<br/>

```json
{
  "version": "0.1",
  "resources": [
    {
      "id": "aws_vpc.main",
      "type": "aws_vpc",
      "name": "main",
      "attributes": { "cidr_block": { "kind": "string", "literal": "10.0.0.0/16" } }
    }
  ],
  "edges": [{ "from": "aws_vpc.main", "to": "aws_subnet.public", "kind": "reference" }]
}
```

</details>

---

## 🗺️ Timeline

| Milestone | Date | Highlights | State |
|:--|:--|:--|:--:|
| **Foundation** | Oct 2026 | HCL parser · IR · DAG · CDK + Pulumi emitters · WASM playground | ✅ |
| **v0.1 Public Alpha** | Nov 2026 | Docs · tests · CI · GitHub Pages playground | 🚧 |
| **v0.2 Depth** | Dec 2026 | Data sources · `locals` · richer L2 mappings | ⏳ |
| **v0.3 Scale** | Q1 2027 | Modules · `count` / `for_each` · multi-file imports | ⏳ |
| **v1.0 Stable** | Q2 2027 | CloudFormation + Terraform backends · plugin API | 🎯 |
| **v1.x Expansion** | 2027+ | Azure/Bicep · GCP · Kubernetes · Python & Go emitters | 🌅 |

```text
Oct 2026      Nov 2026      Dec 2026      Q1 2027       Q2 2027       2027+
  ✅ ─────────  🚧 ─────────  ⏳ ─────────  ⏳ ─────────  🎯 ─────────  🌅
 core      alpha release   data+locals    modules       v1.0       more clouds
```

---

## 🗂️ Project Layout

```text
OpenIaC/
├── cmd/openiac/          CLI entry point
├── pkg/
│   ├── ir/               Universal graph + topological sort
│   ├── parsers/hcl/      HCL2 → IR
│   ├── mappings/         Data-driven aws.json registry
│   ├── emitutil/         Naming, rewriting, rendering
│   ├── emitters/         aws-cdk/ + pulumi/
│   └── engine/           Orchestration
├── wasm/                 js/wasm entry point
├── playground/           Monaco + WASM UI
├── examples/             Sample Terraform modules
└── Makefile
```

---

## 🤝 Contributing

| Step | Command |
|:--|:--|
| 1 · Fork & clone | `git clone https://github.com/<you>/OpenIaC.git` |
| 2 · Build | `make build` |
| 3 · Test | `make test` |
| 4 · Lint | `make vet && gofmt -l .` |
| 5 · Rebuild WASM | `make wasm` |
| 6 · Open a PR | — |

**Good first issues:** new resource mappings · [conformance fixtures](docs/CONFORMANCE.md) · better warning messages · playground presets.

---

## ❓ FAQ

<details>
<summary><b>Is OpenIaC a Terraform replacement?</b></summary>
No. It reads existing Terraform and produces CDK or Pulumi code. State, providers and planning stay with the target tool.
</details>

<details>
<summary><b>Does the playground upload my code?</b></summary>
No. The parser and emitters are compiled to WebAssembly and run inside the tab. There is no server component.
</details>

<details>
<summary><b>Is the output production-ready?</b></summary>
Treat it as a migration accelerator: review the generated stack, then `cdk synth` / `pulumi preview`. Warnings tell you exactly what needs a human.
</details>

<details>
<summary><b>Can I use it as a Go library?</b></summary>
Yes — <code>engine.ConvertSource</code> and <code>engine.Convert</code> are the entry points. The IR is plain Go structs with JSON tags.
</details>

<details>
<summary><b>Why GPL-3.0?</b></summary>
The venture stays open: improvements to the compiler come back to the commons.
</details>

---

## 📄 License

[GNU General Public License v3.0](LICENSE) © OpenIaC contributors

<div align="center">

<sub>
OpenIaC is an independent project. Not affiliated with, endorsed by, or sponsored by<br/>
HashiCorp, Amazon Web Services, or Pulumi Corporation.<br/>
Terraform, AWS CDK and Pulumi are trademarks of their respective owners.
</sub>

<br/><br/>

**If this saves you a migration, star the repo ⭐**

</div>
