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
resource "aws_s3_bucket" "logs" {
  bucket = "logs"
  versioning {
    enabled = true
  }
}

resource "aws_security_group" "web" {
  name = "web"
  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
`

	jsonNested = `{
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
            "from_port": 80,
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
	blocks, consumed := emitters.UnflattenBlocks(
		map[string]ir.Value{
			"versioning": {Kind: ir.KindBool, Literal: true},
			"bucket":     {Kind: ir.KindString, Literal: "logs"},
		},
		map[string]mappings.BlockMap{"versioning": {Property: "versioned"}},
	)
	if len(blocks) != 0 || len(consumed) != 0 {
		t.Fatalf("blocks=%v consumed=%v, want nothing", blocks, consumed)
	}
}
