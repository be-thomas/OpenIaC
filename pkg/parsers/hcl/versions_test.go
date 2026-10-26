package hclparser_test

import (
	"testing"

	"github.com/be-thomas/openiac/pkg/ir"
	hclparser "github.com/be-thomas/openiac/pkg/parsers/hcl"
	"github.com/be-thomas/openiac/pkg/tfversion"
)

// versionCase is one minimal configuration per registered Terraform feature.
// min is the version the whole snippet needs, so it also checks that detectors
// do not over-report.
type versionCase struct {
	feature string
	min     string
	source  string
}

var versionCases = []versionCase{
	{
		feature: "action", min: "1.14.0",
		source: `
action "aws_lb_target_group" "register" {
  config {
    target_group_arn = "arn:aws:elasticloadbalancing:us-east-1:123:targetgroup/web/abc"
  }
}
`,
	},
	{
		feature: "action_lifecycle_events", min: "1.14.0",
		source: `
action "aws_cloudfront_invalidation" "purge" {
  config {
    paths = ["/*"]
  }

  on_failure = "continue"
}
`,
	},
	{
		feature: "check", min: "1.5.0",
		source: `
check "health" {
  assert {
    condition     = true
    error_message = "unhealthy"
  }
}
`,
	},
	{
		feature: "cloud", min: "1.1.0",
		source: `
terraform {
  cloud {
    organization = "example"
  }
}
`,
	},
	{
		feature: "conditions", min: "1.2.0",
		source: `
resource "aws_instance" "web" {
  ami = "ami-123"

  precondition {
    condition     = true
    error_message = "no instance"
  }
}
`,
	},
	{
		feature: "deprecated", min: "1.15.0",
		source: `
variable "legacy" {
  default    = "old"
  deprecated = "use the new variable instead"
}
`,
	},
	{
		feature: "ephemeral_resources", min: "1.10.0",
		source: `
ephemeral "random_password" "db" {
  length = 16
}
`,
	},
	{
		feature: "ephemeral_values", min: "1.10.0",
		source: `
variable "token" {
  type      = string
  ephemeral = true
}
`,
	},
	{
		feature: "import", min: "1.5.0",
		source: `
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}

import {
  to = aws_vpc.main
  id = "vpc-123"
}
`,
	},
	{
		feature: "import_expression", min: "1.6.0",
		source: `
resource "aws_vpc" "existing" {
  cidr_block = "10.0.0.0/16"
}

import {
  to = aws_vpc.existing
  id = aws_vpc.existing.id
}
`,
	},
	{
		feature: "import_for_each", min: "1.7.0",
		source: `
import {
  for_each = { a = "vpc-1" }
  to       = aws_vpc.main
  id       = each.value
}
`,
	},
	{
		feature: "import_identity", min: "1.12.0",
		source: `
import {
  to = aws_instance.example
  identity = {
    id     = "i-abcd1234"
    region = "us-east-1"
  }
}
`,
	},
	{
		feature: "lifecycle_destroy", min: "1.16.0",
		source: `
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"

  lifecycle {
    destroy = false
  }
}
`,
	},
	{
		feature: "moved", min: "1.1.0",
		source: `
moved {
  from = aws_vpc.old
  to   = aws_vpc.main
}
`,
	},
	{
		feature: "moved_cross_type", min: "1.1.0",
		source: `
moved {
  from = aws_vpc.old
  to   = aws_security_group.main
}
`,
	},
	{
		feature: "moved_module", min: "1.3.0",
		source: `
moved {
  from = module.old.aws_vpc.main
  to   = module.new.aws_vpc.main
}
`,
	},
	{
		feature: "nullable", min: "1.1.0",
		source: `
variable "name" {
  type     = string
  nullable = false
}
`,
	},
	{
		feature: "optional_attributes", min: "1.3.0",
		source: `
variable "config" {
  type = object({
    name    = string
    retries = optional(number, 3)
  })
}
`,
	},
	{
		feature: "output_type", min: "1.15.0",
		source: `
output "ip" {
  value = "10.0.0.1"
  type  = string
}
`,
	},
	{
		feature: "provider_requirements_expressions", min: "1.17.0",
		source: `
variable "aws_version" {
  default = "5.0"
}

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = var.aws_version
    }
  }
}
`,
	},
	{
		feature: "removed", min: "1.7.0",
		source: `
removed {
  from = aws_instance.legacy
}
`,
	},
	{
		feature: "removed_provisioners", min: "1.9.0",
		source: `
removed {
  from = aws_instance.legacy

  provisioner "local-exec" {
    when    = destroy
    command = "echo bye"
  }
}
`,
	},
	{
		feature: "replace_triggered_by", min: "1.2.0",
		source: `
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"

  lifecycle {
    replace_triggered_by = [terraform_data.redeploy]
  }
}
`,
	},
	{
		feature: "terraform_data", min: "1.4.0",
		source: `
resource "terraform_data" "redeploy" {
  input = "v1"
}
`,
	},
	{
		feature: "terraform_data_store", min: "1.16.0",
		source: `
resource "terraform_data" "cache" {
  store {
    key = "cache"
  }
}
`,
	},
}

func TestFeatureDetection(t *testing.T) {
	for _, c := range versionCases {
		c := c
		t.Run(c.feature, func(t *testing.T) {
			graph, err := hclparser.ParseString(c.source, "main.tf")
			if err != nil {
				t.Fatalf("ParseString: %v", err)
			}
			if !contains(graph.Features, c.feature) {
				t.Fatalf("features = %v, want %q", graph.Features, c.feature)
			}
			if graph.MinVersion != c.min {
				t.Errorf("min_version = %q, want %q (features %v)", graph.MinVersion, c.min, graph.Features)
			}
		})
	}
}

func TestEveryFeatureHasSample(t *testing.T) {
	covered := make(map[string]bool, len(versionCases))
	for _, c := range versionCases {
		covered[c.feature] = true
	}
	for _, feature := range tfversion.Features() {
		if !covered[feature.Name] {
			t.Errorf("feature %q has no sample configuration", feature.Name)
		}
	}
}

func TestVersionCasesUseOnlyKnownFeatures(t *testing.T) {
	known := make(map[string]bool)
	for _, feature := range tfversion.Features() {
		known[feature.Name] = true
	}
	for _, c := range versionCases {
		if !known[c.feature] {
			t.Errorf("%q is not in the feature registry", c.feature)
		}
		if _, ok := tfversion.Canonical(c.min); !ok {
			t.Errorf("%s: invalid min version %q", c.feature, c.min)
		}
	}
}

// TestUnreliableFeaturesDoNotRaiseMinVersion guards the conservative default:
// a detector that can misfire must not move min_version on its own.
func TestUnreliableFeaturesDoNotRaiseMinVersion(t *testing.T) {
	for _, feature := range tfversion.Features() {
		if feature.Reliable {
			continue
		}
		if got := tfversion.MinVersion([]string{feature.Name}, ""); got != tfversion.Baseline {
			t.Errorf("MinVersion([%q]) = %q, want %q", feature.Name, got, tfversion.Baseline)
		}
		if feature.Since == tfversion.Baseline {
			t.Errorf("%s: an unreliable feature cannot be part of the baseline", feature.Name)
		}
	}
}

// TestPreservedBlocksDoNotCreateNodes keeps moved/import/removed/check/ephemeral
// out of the resource graph, because both emitters treat every node as managed.
func TestPreservedBlocksDoNotCreateNodes(t *testing.T) {
	source := `
moved {
  from = aws_vpc.old
  to   = aws_vpc.main
}

import {
  to = aws_vpc.main
  id = "vpc-123"
}

removed {
  from = aws_instance.legacy
}

check "health" {
  assert {
    condition     = true
    error_message = "unhealthy"
  }
}

ephemeral "random_password" "db" {
  length = 16
}

action "aws_lb_target_group" "register" {
  config {
    target_group_arn = "arn"
  }
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}
`
	graph, err := hclparser.ParseString(source, "main.tf")
	if err != nil {
		t.Fatalf("ParseString: %v", err)
	}
	if len(graph.Resources) != 1 || graph.Resources[0].ID != "aws_vpc.main" {
		t.Fatalf("resources = %v, want only aws_vpc.main", resourceIDs(graph))
	}
	if len(graph.Edges) != 0 {
		t.Errorf("edges = %v, want none", graph.Edges)
	}

	ordered, err := graph.Order()
	if err != nil {
		t.Fatalf("Order: %v", err)
	}
	if len(ordered) != 1 || ordered[0].ID != "aws_vpc.main" {
		t.Errorf("order = %v, want only aws_vpc.main", ordered)
	}
	if resources, data, _, _, edges := graph.Stats(); resources != 1 || data != 0 || edges != 0 {
		t.Errorf("stats = (%d resources, %d data, %d edges)", resources, data, edges)
	}
	if graph.HasNode("moved") || graph.HasNode("random_password.db") {
		t.Error("a preserved block leaked into the node set")
	}
}

func resourceIDs(graph *ir.Graph) []string {
	out := make([]string, 0, len(graph.Resources))
	for _, res := range graph.Resources {
		out = append(out, res.ID)
	}
	return out
}
