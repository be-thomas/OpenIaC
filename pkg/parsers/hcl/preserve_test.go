package hclparser_test

import (
	"strings"
	"testing"

	"github.com/be-thomas/openiac/pkg/ir"
	hclparser "github.com/be-thomas/openiac/pkg/parsers/hcl"
)

const modernModule = `
terraform {
  required_version = ">= 1.6.2"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    bucket = "state"
    key    = "app/terraform.tfstate"
  }

  cloud {
    organization = "example"
  }

  provider_meta "aws" {
    module_name = "example"
  }

  experiments = [module_variable_optional_attrs]
}

moved {
  from = aws_vpc.old
  to   = aws_vpc.main
}

import {
  to = aws_vpc.main
  id = aws_vpc.existing.id
}

removed {
  from = aws_instance.legacy

  lifecycle {
    destroy = false
  }

  provisioner "local-exec" {
    when    = destroy
    command = "echo bye"
  }
}

check "health" {
  data "http" "ping" {
    url = "https://example.com"
  }

  assert {
    condition     = data.http.ping.status_code == 200
    error_message = "unhealthy"
  }
}

ephemeral "random_password" "db" {
  length = 16
}

action "aws_lb_target_group" "register" {
  config {
    target_group_arn = aws_lb_target_group.main.arn
  }
}

variable "password" {
  type        = string
  sensitive   = true
  nullable    = false
  ephemeral   = true
  deprecated  = "use db_password instead"
  description = "db password"

  validation {
    condition     = length(var.password) >= 8
    error_message = "too short"
  }
}

variable "config" {
  type = object({
    name    = string
    retries = optional(number, 3)
  })
}

output "instance_ip" {
  value       = aws_instance.web.private_ip
  type        = string
  ephemeral   = true
  deprecated  = "use instance_ips"
  description = "private ip"

  precondition {
    condition     = aws_instance.web.id != ""
    error_message = "no instance"
  }
}

provider "aws" {
  region = "us-east-1"

  default_tags {
    tags = { Environment = "dev" }
  }
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"

  lifecycle {
    replace_triggered_by = [terraform_data.redeploy]
    destroy              = false
  }
}

resource "terraform_data" "redeploy" {
  input = "v2"
}

resource "aws_subnet" "public" {
  vpc_id = aws_vpc.main.id
}

resource "aws_instance" "web" {
  ami           = "ami-123"
  instance_type = "t3.micro"
}
`

func parseModern(t *testing.T) *ir.Graph {
	t.Helper()
	graph, err := hclparser.ParseString(modernModule, "main.tf")
	if err != nil {
		t.Fatalf("ParseString: %v", err)
	}
	return graph
}

func preservedByType(graph *ir.Graph) map[string][]ir.Block {
	out := map[string][]ir.Block{}
	for _, block := range graph.Preserved {
		out[block.Type] = append(out[block.Type], block)
	}
	return out
}

func TestPreservedTopLevelBlocks(t *testing.T) {
	graph := parseModern(t)
	preserved := preservedByType(graph)

	cases := []struct {
		typ    string
		labels []string
	}{
		{"moved", nil},
		{"import", nil},
		{"removed", nil},
		{"check", []string{"health"}},
		{"ephemeral", []string{"random_password", "db"}},
		{"action", []string{"aws_lb_target_group", "register"}},
	}
	for _, c := range cases {
		blocks := preserved[c.typ]
		if len(blocks) != 1 {
			t.Errorf("%s preserved %d times, want 1", c.typ, len(blocks))
			continue
		}
		if len(blocks[0].Labels) != len(c.labels) {
			t.Errorf("%s labels = %v, want %v", c.typ, blocks[0].Labels, c.labels)
			continue
		}
		for i, label := range c.labels {
			if blocks[0].Labels[i] != label {
				t.Errorf("%s label %d = %q, want %q", c.typ, i, blocks[0].Labels[i], label)
			}
		}
		if blocks[0].File != "main.tf" || blocks[0].Line == 0 {
			t.Errorf("%s provenance = %q:%d", c.typ, blocks[0].File, blocks[0].Line)
		}
	}

	if len(graph.Resources) != 4 {
		t.Errorf("resources = %d, want 4", len(graph.Resources))
	}
	for _, res := range graph.Resources {
		switch res.Type {
		case "aws_vpc", "aws_subnet", "aws_instance", "terraform_data":
		default:
			t.Errorf("preserved type leaked into resources: %s", res.ID)
		}
	}

	ordered, err := graph.Order()
	if err != nil {
		t.Fatalf("Order: %v", err)
	}
	if len(ordered) != len(graph.Resources) {
		t.Errorf("ordered %d nodes, want %d", len(ordered), len(graph.Resources))
	}
}

func TestPreservedNestedStructure(t *testing.T) {
	graph := parseModern(t)
	preserved := preservedByType(graph)

	removed := preserved["removed"][0]
	if _, ok := removed.Attributes["from"]; !ok {
		t.Error("removed.from missing")
	}
	var lifecycle, provisioner bool
	for _, block := range removed.Blocks {
		switch block.Type {
		case "lifecycle":
			lifecycle = true
			if _, ok := block.Attributes["destroy"]; !ok {
				t.Error("removed.lifecycle.destroy missing")
			}
		case "provisioner":
			provisioner = true
			if len(block.Labels) != 1 || block.Labels[0] != "local-exec" {
				t.Errorf("provisioner labels = %v", block.Labels)
			}
		}
	}
	if !lifecycle || !provisioner {
		t.Errorf("removed blocks = %v, want lifecycle and provisioner", removed.Blocks)
	}

	check := preserved["check"][0]
	if len(check.Blocks) != 2 {
		t.Fatalf("check blocks = %d, want 2", len(check.Blocks))
	}

	action := preserved["action"][0]
	if len(action.Blocks) != 1 || action.Blocks[0].Type != "config" {
		t.Fatalf("action blocks = %v", action.Blocks)
	}
}

func TestVariableAndOutputPreservation(t *testing.T) {
	graph := parseModern(t)

	if len(graph.Variables) != 2 {
		t.Fatalf("variables = %d, want 2", len(graph.Variables))
	}
	password := graph.Variables[0]
	if password.Name != "password" {
		t.Fatalf("first variable = %q", password.Name)
	}
	if password.Nullable == nil || *password.Nullable {
		t.Errorf("nullable = %v, want false", password.Nullable)
	}
	if !password.Ephemeral {
		t.Error("ephemeral not parsed")
	}
	if password.Deprecated != "use db_password instead" {
		t.Errorf("deprecated = %q", password.Deprecated)
	}
	if len(password.Blocks) != 1 || password.Blocks[0].Type != "validation" {
		t.Fatalf("validation block missing: %v", password.Blocks)
	}
	if _, ok := password.Attributes["nullable"]; ok {
		t.Error("nullable should be typed, not duplicated in attributes")
	}

	config := graph.Variables[1]
	if !strings.Contains(config.Type, "optional(") {
		t.Errorf("config type = %q, want an optional() constraint", config.Type)
	}

	if len(graph.Outputs) != 1 {
		t.Fatalf("outputs = %d, want 1", len(graph.Outputs))
	}
	out := graph.Outputs[0]
	if out.Type != "string" {
		t.Errorf("output type = %q, want string", out.Type)
	}
	if !out.Ephemeral || out.Deprecated != "use instance_ips" {
		t.Errorf("output = %#v", out)
	}
	if len(out.Blocks) != 1 || out.Blocks[0].Type != "precondition" {
		t.Fatalf("precondition block missing: %v", out.Blocks)
	}
}

func TestTerraformMetadataNative(t *testing.T) {
	graph := parseModern(t)
	tf := graph.Terraform
	if tf == nil {
		t.Fatal("terraform metadata missing")
	}
	if tf.RequiredVersion != ">= 1.6.2" {
		t.Errorf("required_version = %q", tf.RequiredVersion)
	}
	if tf.Backend == nil || tf.Backend.Labels[0] != "s3" {
		t.Fatalf("backend = %#v", tf.Backend)
	}
	if tf.Backend.File != "main.tf" || tf.Backend.Line == 0 {
		t.Errorf("backend provenance = %q:%d", tf.Backend.File, tf.Backend.Line)
	}
	if tf.Cloud == nil {
		t.Fatal("cloud missing")
	}
	if got := tf.Cloud.Attributes["organization"]; got.Literal != "example" {
		t.Errorf("cloud organization = %#v", got)
	}
	if len(tf.ProviderMeta) != 1 || tf.ProviderMeta[0].Labels[0] != "aws" {
		t.Fatalf("provider_meta = %#v", tf.ProviderMeta)
	}
	if aws, ok := tf.RequiredProviders["aws"]; !ok || aws.Kind != ir.KindMap {
		t.Fatalf("required_providers.aws = %#v", aws)
	}
	if len(tf.Experiments) != 1 || tf.Experiments[0] != "module_variable_optional_attrs" {
		t.Errorf("experiments = %v", tf.Experiments)
	}
	if tf.File != "main.tf" || tf.Line == 0 {
		t.Errorf("terraform provenance = %q:%d", tf.File, tf.Line)
	}
}

func TestProviderBlocksPreserved(t *testing.T) {
	graph := parseModern(t)
	if len(graph.Providers) != 1 {
		t.Fatalf("providers = %d, want 1", len(graph.Providers))
	}
	provider := graph.Providers[0]
	if len(provider.Blocks) != 1 || provider.Blocks[0].Type != "default_tags" {
		t.Fatalf("provider blocks = %v", provider.Blocks)
	}
	if provider.File != "main.tf" || provider.Line == 0 {
		t.Errorf("provider provenance = %q:%d", provider.File, provider.Line)
	}
}

func TestFeaturesAndMinVersion(t *testing.T) {
	graph := parseModern(t)

	want := []string{
		"action",
		"check",
		"cloud",
		"conditions",
		"deprecated",
		"ephemeral_resources",
		"ephemeral_values",
		"import",
		"import_expression",
		"lifecycle_destroy",
		"moved",
		"nullable",
		"optional_attributes",
		"output_type",
		"removed",
		"removed_provisioners",
		"replace_triggered_by",
		"terraform_data",
	}
	for _, name := range want {
		if !contains(graph.Features, name) {
			t.Errorf("feature %q not detected; got %v", name, graph.Features)
		}
	}
	if graph.MinVersion != "1.16.0" {
		t.Errorf("min_version = %q, want 1.16.0", graph.MinVersion)
	}
}

func TestMinVersionBaseline(t *testing.T) {
	graph, err := hclparser.ParseString(`resource "aws_vpc" "main" { cidr_block = "10.0.0.0/16" }`, "main.tf")
	if err != nil {
		t.Fatal(err)
	}
	if graph.MinVersion != "1.0.0" {
		t.Errorf("min_version = %q, want 1.0.0", graph.MinVersion)
	}
	if len(graph.Features) != 0 {
		t.Errorf("features = %v, want none", graph.Features)
	}
}

func TestUnknownTopLevelBlockPreserved(t *testing.T) {
	graph, err := hclparser.ParseString(`
future_block "a" "b" {
  setting = "x"
}

stray_argument = "y"
`, "main.tf")
	if err != nil {
		t.Fatalf("ParseString: %v", err)
	}

	preserved := preservedByType(graph)
	future := preserved["future_block"]
	if len(future) != 1 || len(future[0].Labels) != 2 {
		t.Fatalf("future_block = %#v", future)
	}
	if got := future[0].Attributes["setting"]; got.Literal != "x" {
		t.Errorf("future_block.setting = %#v", got)
	}

	stray := preserved["stray_argument"]
	if len(stray) != 1 || stray[0].Body == nil || stray[0].Body.Literal != "y" {
		t.Fatalf("stray_argument = %#v", stray)
	}
}

func TestLabelValidation(t *testing.T) {
	cases := []string{
		`resource "aws_vpc" {}`,
		`variable {}`,
		`module "a" "b" {}`,
	}
	for _, source := range cases {
		if _, err := hclparser.ParseString(source, "main.tf"); err == nil {
			t.Errorf("expected a label error for %q", source)
		}
	}
}

func TestParseDirRequiresTerraformFiles(t *testing.T) {
	if _, err := hclparser.ParseDir(t.TempDir()); err == nil {
		t.Fatal("expected an error for an empty directory")
	} else if !strings.Contains(err.Error(), "*.tf") {
		t.Errorf("error = %v", err)
	}
}
