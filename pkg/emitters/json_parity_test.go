package emitters_test

import (
	"strings"
	"testing"

	"github.com/be-thomas/openiac/pkg/emitters"
	"github.com/be-thomas/openiac/pkg/emitters/awscdk"
	"github.com/be-thomas/openiac/pkg/emitters/pulumi"
	"github.com/be-thomas/openiac/pkg/ir"
	"github.com/be-thomas/openiac/pkg/mappings"
	hclparser "github.com/be-thomas/openiac/pkg/parsers/hcl"
)

// Terraform's JSON syntax cannot spell a nested block below the top level, so
// the parser leaves it as a map or array attribute. Both syntaxes must still
// produce the same TypeScript.
const (
	nativeNested = `
variable "port" {
  type = number
}

resource "aws_s3_bucket" "logs" {
  bucket = "logs"
  versioning {
    enabled = true
  }
}

resource "aws_security_group" "web" {
  name = "web"
  ingress {
    from_port   = var.port
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
`

	jsonNested = `{
  "variable": {
    "port": { "type": "number" }
  },
  "resource": {
    "aws_s3_bucket": {
      "logs": {
        "bucket": "logs",
        "versioning": { "enabled": true }
      }
    },
    "aws_security_group": {
      "web": {
        "name": "web",
        "ingress": [
          {
            "from_port": "${var.port}",
            "to_port": 80,
            "protocol": "tcp",
            "cidr_blocks": ["0.0.0.0/0"]
          }
        ]
      }
    }
  }
}`
)

func parseSource(t *testing.T, src, filename string) *ir.Graph {
	t.Helper()
	graph, err := hclparser.ParseString(src, filename)
	if err != nil {
		t.Fatalf("ParseString(%s): %v", filename, err)
	}
	return graph
}

func emitBoth(t *testing.T, graph *ir.Graph) map[string]*emitters.Result {
	t.Helper()
	cdk, err := awscdk.New().Emit(graph, emitters.Options{})
	if err != nil {
		t.Fatalf("cdk Emit: %v", err)
	}
	pulumiOut, err := pulumi.New().Emit(graph, emitters.Options{})
	if err != nil {
		t.Fatalf("pulumi Emit: %v", err)
	}
	return map[string]*emitters.Result{"cdk": cdk, "pulumi": pulumiOut}
}

func TestJSONNestedBlocksMatchNativeOutput(t *testing.T) {
	native := emitBoth(t, parseSource(t, nativeNested, "main.tf"))
	jsonSyntax := emitBoth(t, parseSource(t, jsonNested, "main.tf.json"))

	for target := range native {
		nativeFiles := native[target].Files
		jsonFiles := jsonSyntax[target].Files
		if len(nativeFiles) != len(jsonFiles) {
			t.Fatalf("%s: files = %d vs %d", target, len(nativeFiles), len(jsonFiles))
		}
		for i := range nativeFiles {
			want := nativeFiles[i].Content
			got := strings.ReplaceAll(jsonFiles[i].Content, "main.tf.json", "main.tf")
			if got != want {
				t.Errorf("%s: %s differs between native HCL and .tf.json\n--- native\n%s\n--- json\n%s",
					target, nativeFiles[i].Path, want, got)
			}
		}
	}
}

func TestJSONNestedBlocksDoNotWarn(t *testing.T) {
	for target, result := range emitBoth(t, parseSource(t, jsonNested, "main.tf.json")) {
		for _, warning := range result.Warnings {
			if strings.Contains(warning, "versioning") || strings.Contains(warning, "ingress") {
				t.Errorf("%s: nested block reported as unconverted: %s", target, warning)
			}
		}
	}
}

func TestUnflattenBlocksIgnoresScalars(t *testing.T) {
	blocks, consumed, conflicts := emitters.UnflattenBlocks(
		map[string]ir.Value{
			"versioning": {Kind: ir.KindBool, Literal: true},
			"bucket":     {Kind: ir.KindString, Literal: "logs"},
		},
		map[string]mappings.BlockMap{"versioning": {Property: "versioned"}},
		nil,
	)
	if len(blocks) != 0 || len(consumed) != 0 || len(conflicts) != 0 {
		t.Fatalf("blocks=%v consumed=%v conflicts=%v, want nothing", blocks, consumed, conflicts)
	}
}

func TestUnflattenBlocksConsumesEmptyArray(t *testing.T) {
	// "ingress": [] means there are no ingress blocks, which is what native HCL
	// with no ingress block means. It must not fall through as a property.
	blocks, consumed, _ := emitters.UnflattenBlocks(
		map[string]ir.Value{"ingress": {Kind: ir.KindList}},
		map[string]mappings.BlockMap{"ingress": {Property: "ingress", Shape: "array"}},
		nil,
	)
	if len(blocks) != 0 {
		t.Fatalf("blocks=%v, want none", blocks)
	}
	if !consumed["ingress"] {
		t.Fatalf("consumed=%v, want ingress consumed", consumed)
	}
}

func TestUnflattenBlocksReportsDuplicateAttribute(t *testing.T) {
	// An attribute that duplicates a native block is invalid Terraform. The
	// attribute must be consumed, or two blocks are rendered, and reported so
	// the emitter can warn about it.
	blocks, consumed, conflicts := emitters.UnflattenBlocks(
		map[string]ir.Value{"ingress": {Kind: ir.KindMap, Fields: map[string]ir.Value{
			"from_port": {Kind: ir.KindNumber, Literal: 80},
		}}},
		map[string]mappings.BlockMap{"ingress": {Property: "ingress", Shape: "array"}},
		[]ir.Block{{Type: "ingress"}},
	)
	if len(blocks) != 0 {
		t.Fatalf("blocks=%v, want none", blocks)
	}
	if !consumed["ingress"] {
		t.Fatalf("consumed=%v, want ingress consumed", consumed)
	}
	if len(conflicts) != 1 || conflicts[0] != "ingress" {
		t.Fatalf("conflicts=%v, want [ingress]", conflicts)
	}
}

// A body that carries both a block and an attribute of the same name is
// invalid Terraform, but it must never produce duplicate object keys or two
// blocks; the attribute is dropped and the conflict is reported.
func TestDuplicateBlockAndAttributeWarnsAndEmitsOnce(t *testing.T) {
	graph := parseSource(t, `
resource "aws_security_group" "web" {
  name = "web"
  ingress {
    from_port = 80
    to_port   = 80
    protocol  = "tcp"
  }
  ingress = [{
    from_port = 443
    to_port   = 443
    protocol  = "tcp"
  }]
}
`, "main.tf")

	for target, result := range emitBoth(t, graph) {
		content := result.Files[0].Content
		key := "ingress:"
		if target == "cdk" {
			key = "SecurityGroupIngress: ["
		}
		if got := strings.Count(content, key); got != 1 {
			t.Errorf("%s: %q appears %d times, want once\n%s", target, key, got, content)
		}
		if strings.Contains(content, "443") {
			t.Errorf("%s: duplicate attribute rendered\n%s", target, content)
		}
		warned := false
		for _, warning := range result.Warnings {
			if strings.Contains(warning, "both a block and an attribute") {
				warned = true
			}
		}
		if !warned {
			t.Errorf("%s: no conflict warning; warnings=%v", target, result.Warnings)
		}
	}
}

// An empty block list means the block is absent, whichever syntax spells it.
// The IR cannot tell a native attribute from a recovered JSON one, so the rule
// applies to both and native HCL with "ingress = []" matches native HCL with
// no ingress block.
func TestEmptyBlockArrayIsAbsentInNativeHCL(t *testing.T) {
	absent := emitBoth(t, parseSource(t, `
resource "aws_security_group" "web" {
  name = "web"
}
`, "main.tf"))
	empty := emitBoth(t, parseSource(t, `
resource "aws_security_group" "web" {
  name   = "web"
  ingress = []
}
`, "main.tf"))

	for target := range absent {
		want := absent[target].Files[0].Content
		got := empty[target].Files[0].Content
		if got != want {
			t.Errorf("%s: ingress = [] differs from no ingress block\n--- absent\n%s\n--- empty\n%s", target, want, got)
		}
	}
}

func TestJSONEmptyBlockArrayMatchesNativeOutput(t *testing.T) {
	native := emitBoth(t, parseSource(t, `
resource "aws_security_group" "web" {
  name = "web"
}
`, "main.tf"))
	jsonSyntax := emitBoth(t, parseSource(t, `{
  "resource": {
    "aws_security_group": {
      "web": {
        "name": "web",
        "ingress": []
      }
    }
  }
}`, "main.tf.json"))

	for target := range native {
		nativeFiles := native[target].Files
		jsonFiles := jsonSyntax[target].Files
		if len(nativeFiles) != len(jsonFiles) {
			t.Fatalf("%s: files = %d vs %d", target, len(nativeFiles), len(jsonFiles))
		}
		for i := range nativeFiles {
			want := nativeFiles[i].Content
			got := strings.ReplaceAll(jsonFiles[i].Content, "main.tf.json", "main.tf")
			if got != want {
				t.Errorf("%s: %s differs with an empty block array\n--- native\n%s\n--- json\n%s",
					target, nativeFiles[i].Path, want, got)
			}
		}
		for _, warning := range jsonSyntax[target].Warnings {
			if strings.Contains(warning, "ingress") {
				t.Errorf("%s: empty ingress array reported as unconverted: %s", target, warning)
			}
		}
	}
}
