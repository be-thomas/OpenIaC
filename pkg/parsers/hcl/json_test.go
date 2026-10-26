package hclparser_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/be-thomas/openiac/pkg/ir"
	hclparser "github.com/be-thomas/openiac/pkg/parsers/hcl"
)

const jsonSample = `{
  "terraform": {
    "required_version": ">= 1.5.0",
    "required_providers": {
      "aws": {
        "source": "hashicorp/aws",
        "version": "~> 5.0"
      }
    },
    "backend": {
      "s3": {
        "bucket": "state-bucket",
        "key": "terraform.tfstate"
      }
    }
  },
  "variable": {
    "region": {
      "type": "string",
      "default": "us-east-1"
    },
    "environment": {
      "type": "string",
      "default": "dev"
    }
  },
  "locals": {
    "name_prefix": "app-${var.environment}"
  },
  "provider": {
    "aws": {
      "region": "${var.region}"
    }
  },
  "resource": {
    "aws_vpc": {
      "main": {
        "cidr_block": "10.0.0.0/16",
        "tags": {
          "Name": "app-${var.environment}",
          "Environment": "${var.environment}"
        }
      }
    },
    "aws_subnet": {
      "public": {
        "vpc_id": "${aws_vpc.main.id}",
        "cidr_block": "10.0.1.0/24"
      }
    },
    "aws_instance": {
      "web": {
        "ami": "ami-123",
        "subnet_id": "${aws_subnet.public.id}",
        "depends_on": ["aws_vpc.main"],
        "provider": "aws.west"
      }
    }
  },
  "data": {
    "aws_ami": {
      "ubuntu": {
        "most_recent": true
      }
    }
  },
  "module": {
    "vpc": {
      "source": "./modules/vpc",
      "version": "1.2.0"
    }
  },
  "output": {
    "bucket_arn": {
      "value": "${aws_s3_bucket.logs.arn}",
      "description": "Logs bucket"
    }
  }
}`

func parseJSONSample(t *testing.T) *ir.Graph {
	t.Helper()
	graph, err := hclparser.ParseString(jsonSample, "main.tf.json")
	if err != nil {
		t.Fatalf("ParseString(.tf.json): %v", err)
	}
	return graph
}

func TestJSONResources(t *testing.T) {
	graph := parseJSONSample(t)

	if got, want := len(graph.Resources), 4; got != want {
		t.Fatalf("resources = %d, want %d", got, want)
	}
	vpc := graph.ResourceByID("aws_vpc.main")
	if vpc == nil {
		t.Fatal("aws_vpc.main missing")
	}
	if vpc.Provider != "aws" {
		t.Errorf("provider = %q, want aws", vpc.Provider)
	}
	if cidr := vpc.Attributes["cidr_block"]; cidr.Kind != ir.KindString || cidr.Literal != "10.0.0.0/16" {
		t.Errorf("cidr_block = %#v", cidr)
	}

	ami := graph.ResourceByID("data.aws_ami.ubuntu")
	if ami == nil {
		t.Fatal("data.aws_ami.ubuntu missing")
	}
	if ami.Kind != ir.KindData {
		t.Errorf("data kind = %q", ami.Kind)
	}
}

func TestJSONInterpolationBecomesExpression(t *testing.T) {
	graph := parseJSONSample(t)
	vpc := graph.ResourceByID("aws_vpc.main")

	tags := vpc.Attributes["tags"]
	if tags.Kind != ir.KindMap {
		t.Fatalf("tags kind = %q, want map", tags.Kind)
	}
	env, ok := tags.Fields["Environment"]
	if !ok {
		t.Fatal("tags.Environment missing")
	}
	if env.Kind != ir.KindExpr {
		t.Fatalf("tags.Environment kind = %q, want expr", env.Kind)
	}
	if !contains(env.Refs, "var.environment") {
		t.Errorf("refs = %v, want var.environment", env.Refs)
	}
	// "${var.environment}" is one interpolation and nothing else, so Terraform
	// evaluates it to the expression's value and type. It must not become a
	// string template.
	if env.Expr != "var.environment" {
		t.Errorf("expr = %q, want the bare expression var.environment", env.Expr)
	}

	// Text around an interpolation is a real template and stays quoted.
	name := tags.Fields["Name"]
	if name.Kind != ir.KindExpr {
		t.Fatalf("tags.Name kind = %q, want expr", name.Kind)
	}
	if !strings.HasPrefix(name.Expr, `"`) || !strings.Contains(name.Expr, "var.environment") {
		t.Errorf("expr = %q, want an HCL quoted template", name.Expr)
	}
}

func TestJSONDependsOnAndProvider(t *testing.T) {
	graph := parseJSONSample(t)
	instance := graph.ResourceByID("aws_instance.web")
	if instance == nil {
		t.Fatal("aws_instance.web missing")
	}
	if len(instance.DependsOn) != 1 || instance.DependsOn[0] != "aws_vpc.main" {
		t.Errorf("DependsOn = %v, want [aws_vpc.main]", instance.DependsOn)
	}
	if instance.ProviderAlias != "aws.west" {
		t.Errorf("ProviderAlias = %q, want aws.west", instance.ProviderAlias)
	}
	if _, ok := instance.Attributes["depends_on"]; ok {
		t.Error("depends_on should not stay in attributes")
	}
	if !hasEdge(graph, "aws_instance.web", "aws_vpc.main") {
		t.Error("missing depends_on edge aws_instance.web -> aws_vpc.main")
	}
	if !hasEdge(graph, "aws_subnet.public", "aws_vpc.main") {
		t.Error("missing reference edge aws_subnet.public -> aws_vpc.main")
	}
}

func TestJSONTerraformMetadata(t *testing.T) {
	graph := parseJSONSample(t)
	tf := graph.Terraform
	if tf == nil {
		t.Fatal("terraform metadata missing")
	}
	if tf.RequiredVersion != ">= 1.5.0" {
		t.Errorf("required_version = %q", tf.RequiredVersion)
	}
	if tf.Backend == nil || len(tf.Backend.Labels) != 1 || tf.Backend.Labels[0] != "s3" {
		t.Fatalf("backend = %#v, want s3", tf.Backend)
	}
	if tf.Backend.File != "main.tf.json" {
		// JSON bodies are flat, so the enclosing terraform block supplies the file.
		t.Errorf("backend file = %q, want main.tf.json", tf.Backend.File)
	}
	if got := tf.Backend.Attributes["bucket"]; got.Kind != ir.KindString || got.Literal != "state-bucket" {
		t.Errorf("backend bucket = %#v", got)
	}
	aws, ok := tf.RequiredProviders["aws"]
	if !ok || aws.Kind != ir.KindMap {
		t.Fatalf("required_providers.aws = %#v", aws)
	}
	if source := aws.Fields["source"]; source.Kind != ir.KindString || source.Literal != "hashicorp/aws" {
		t.Errorf("aws source = %#v", source)
	}
}

func TestJSONVariablesAndModules(t *testing.T) {
	graph := parseJSONSample(t)

	if len(graph.Variables) != 2 {
		t.Fatalf("variables = %d, want 2", len(graph.Variables))
	}
	region := graph.Variables[0]
	if region.Name != "region" {
		t.Fatalf("first variable = %q, want region", region.Name)
	}
	if region.Type != "string" {
		t.Errorf("region type = %q, want string", region.Type)
	}
	if region.Default == nil || region.Default.Literal != "us-east-1" {
		t.Errorf("region default = %#v", region.Default)
	}
	if region.File != "main.tf.json" || region.Line == 0 {
		t.Errorf("region provenance = %q:%d", region.File, region.Line)
	}

	if len(graph.Modules) != 1 {
		t.Fatalf("modules = %d, want 1", len(graph.Modules))
	}
	module := graph.Modules[0]
	if module.Name != "vpc" || module.Source != "./modules/vpc" || module.Version != "1.2.0" {
		t.Errorf("module = %#v", module)
	}
	if _, ok := module.Attributes["source"]; ok {
		t.Error("source should be typed, not duplicated in attributes")
	}

	if len(graph.Providers) != 1 || graph.Providers[0].Name != "aws" {
		t.Fatalf("providers = %#v", graph.Providers)
	}
	regionAttr := graph.Providers[0].Attributes["region"]
	if regionAttr.Kind != ir.KindExpr || !contains(regionAttr.Refs, "var.region") {
		t.Errorf("provider region = %#v", regionAttr)
	}
}

func TestJSONLocalsAndOutputs(t *testing.T) {
	graph := parseJSONSample(t)

	if len(graph.Locals) != 1 {
		t.Fatalf("locals = %d, want 1", len(graph.Locals))
	}
	prefix := graph.Locals[0]
	if prefix.Name != "name_prefix" || prefix.Value.Kind != ir.KindExpr {
		t.Errorf("local = %#v", prefix)
	}
	if !contains(prefix.Value.Refs, "var.environment") {
		t.Errorf("local refs = %v", prefix.Value.Refs)
	}

	if len(graph.Outputs) != 1 {
		t.Fatalf("outputs = %d, want 1", len(graph.Outputs))
	}
	out := graph.Outputs[0]
	if out.Name != "bucket_arn" || out.Description != "Logs bucket" {
		t.Errorf("output = %#v", out)
	}
	if !contains(out.Value.Refs, "aws_s3_bucket.logs") {
		t.Errorf("output refs = %v, want aws_s3_bucket.logs", out.Value.Refs)
	}
}

func TestJSONEscapesAndEmptyCollections(t *testing.T) {
	graph, err := hclparser.ParseString(`{
  "resource": {
    "aws_ssm_parameter": {
      "quoted": {
        "value": "a \"quoted\" \\ value\nsecond line ${var.suffix}",
        "empty_object": {},
        "empty_array": [],
        "numbers": [1, 2.5, true, null]
      }
    }
  }
}`, "main.tf.json")
	if err != nil {
		t.Fatalf("ParseString: %v", err)
	}

	param := graph.ResourceByID("aws_ssm_parameter.quoted")
	if param == nil {
		t.Fatal("resource missing")
	}

	value := param.Attributes["value"]
	if value.Kind != ir.KindExpr {
		t.Fatalf("value kind = %q, want expr", value.Kind)
	}
	if !strings.Contains(value.Expr, `\"quoted\"`) {
		t.Errorf("value expr lost the escaped quote: %q", value.Expr)
	}
	if !strings.Contains(value.Expr, `\n`) {
		t.Errorf("value expr lost the escaped newline: %q", value.Expr)
	}
	if !contains(value.Refs, "var.suffix") {
		t.Errorf("value refs = %v, want var.suffix", value.Refs)
	}

	if got := param.Attributes["empty_object"]; got.Kind != ir.KindMap {
		t.Errorf("empty_object kind = %q, want map", got.Kind)
	}
	if got := param.Attributes["empty_array"]; got.Kind != ir.KindList {
		t.Errorf("empty_array kind = %q, want list", got.Kind)
	}

	numbers := param.Attributes["numbers"]
	if numbers.Kind != ir.KindList || len(numbers.Items) != 4 {
		t.Fatalf("numbers = %#v", numbers)
	}
	want := []ir.ValueKind{ir.KindNumber, ir.KindNumber, ir.KindBool, ir.KindNull}
	for i, kind := range want {
		if numbers.Items[i].Kind != kind {
			t.Errorf("numbers[%d] kind = %q, want %q", i, numbers.Items[i].Kind, kind)
		}
	}
}

func TestJSONUnknownPropertyIsPreserved(t *testing.T) {
	graph, err := hclparser.ParseString(`{
  "future_block": {
    "from": "a",
    "to": "b"
  },
  "repeated": [{"n": 1}, {"n": 2}],
  "scalar": 42,
  "resource": {"aws_vpc": {"main": {}}}
}`, "main.tf.json")
	if err != nil {
		t.Fatalf("ParseString: %v", err)
	}

	byType := map[string][]ir.Block{}
	for _, b := range graph.Preserved {
		byType[b.Type] = append(byType[b.Type], b)
	}
	if got := len(byType["future_block"]); got != 1 {
		t.Fatalf("future_block preserved %d times, want 1", got)
	}
	if _, ok := byType["future_block"][0].Attributes["from"]; !ok {
		t.Error("future_block attributes missing")
	}
	if got := len(byType["repeated"]); got != 2 {
		t.Errorf("repeated preserved %d times, want 2", got)
	}
	scalar := byType["scalar"]
	if len(scalar) != 1 || scalar[0].Body == nil || scalar[0].Body.Literal != int64(42) {
		t.Errorf("scalar preserved = %#v", scalar)
	}
	if len(graph.Resources) != 1 {
		t.Errorf("resources = %d, want 1", len(graph.Resources))
	}
}

func TestJSONMalformed(t *testing.T) {
	if _, err := hclparser.ParseString(`{"resource": `, "broken.tf.json"); err == nil {
		t.Fatal("expected a parse error")
	}
}

func TestParseDirMixedSyntax(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"vpc.tf": `resource "aws_vpc" "main" { cidr_block = "10.0.0.0/16" }`,
		"subnet.tf.json": `{
  "resource": {
    "aws_subnet": {
      "public": {"vpc_id": "${aws_vpc.main.id}"}
    }
  }
}`,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	graph, err := hclparser.ParseDir(dir)
	if err != nil {
		t.Fatalf("ParseDir: %v", err)
	}
	if len(graph.Resources) != 2 {
		t.Fatalf("resources = %d, want 2", len(graph.Resources))
	}
	if !hasEdge(graph, "aws_subnet.public", "aws_vpc.main") {
		t.Error("cross-syntax edge missing")
	}
}
