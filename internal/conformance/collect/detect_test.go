package collect

import (
	"reflect"
	"testing"
)

func TestDetectFeaturesProvidersAndTier(t *testing.T) {
	files := map[string][]byte{
		"main.tf": []byte(`
variable "name" { type = string }
locals { tags = { Name = var.name } }
module "child" { source = "./modules/child" }
resource "aws_s3_bucket" "b" { bucket = var.name }
data "aws_caller_identity" "current" {}
output "arn" { value = aws_s3_bucket.b.arn }
`),
		"modules/child/main.tf": []byte(`
resource "aws_iam_role" "r" {
  name = "x-${var.name}"
  count = 1
}
`),
	}
	features := DetectFeatures(files)
	wantFeatures := []string{
		"count", "data", "interpolation", "locals", "modules", "outputs",
		"resources", "variables",
	}
	if !reflect.DeepEqual(features, wantFeatures) {
		t.Errorf("features = %v, want %v", features, wantFeatures)
	}
	providers := DetectProviders(files)
	if !reflect.DeepEqual(providers, []string{"aws"}) {
		t.Errorf("providers = %v", providers)
	}
	if got := ClassifyTier(features); got != "modules" {
		t.Errorf("tier = %q", got)
	}
	if got := SupportLevel(features); got != "partial" {
		t.Errorf("support = %q", got)
	}
}

func TestClassifyBasic(t *testing.T) {
	features := []string{"outputs", "resources", "variables"}
	if got := ClassifyTier(features); got != "core" {
		t.Errorf("tier = %q", got)
	}
	if got := SupportLevel(features); got != "full" {
		t.Errorf("support = %q", got)
	}
	if got := ClassifyTier([]string{"data"}); got != "advanced" {
		t.Errorf("tier = %q", got)
	}
}

func TestJSONAnalysisMatchesHCL(t *testing.T) {
	hclSrc := `terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
    }
  }
  backend "s3" {}
}

provider "aws" {
  region = "us-east-1"
}

variable "name" {
  default = "x"
}

locals {
  a = 1
}

resource "aws_s3_bucket" "b" {
  count = 1
}

module "m" {
  source = "./m"
}

output "id" {
  value     = "x"
  sensitive = true
}
`
	jsonSrc := `{
  "terraform": {
    "required_providers": {
      "aws": {"source": "hashicorp/aws"}
    },
    "backend": {"s3": {}}
  },
  "provider": {"aws": {"region": "us-east-1"}},
  "variable": {"name": {"default": "x"}},
  "locals": {"a": 1},
  "resource": {
    "aws_s3_bucket": {
      "b": {"count": 1}
    }
  },
  "module": {"m": {"source": "./m"}},
  "output": {"id": {"value": "x", "sensitive": true}}
}`
	hclA, err := AnalyzeFiles(map[string][]byte{"main.tf": []byte(hclSrc)})
	if err != nil {
		t.Fatal(err)
	}
	jsonA, err := AnalyzeFiles(map[string][]byte{"main.tf.json": []byte(jsonSrc)})
	if err != nil {
		t.Fatal(err)
	}
	if !equalStrings(hclA.Features, jsonA.Features) {
		t.Errorf("features: hcl=%v json=%v", hclA.Features, jsonA.Features)
	}
	if !equalStrings(hclA.Providers, jsonA.Providers) {
		t.Errorf("providers: hcl=%v json=%v", hclA.Providers, jsonA.Providers)
	}
	if len(hclA.LocalSources) != len(jsonA.LocalSources) {
		t.Errorf("local sources: hcl=%v json=%v", hclA.LocalSources, jsonA.LocalSources)
	}
	if len(hclA.Violations) != 0 || len(jsonA.Violations) != 0 {
		t.Errorf("unexpected violations: hcl=%v json=%v", hclA.Violations, jsonA.Violations)
	}
}

func TestJSONAnalysisMatchesHCLExpressions(t *testing.T) {
	hclSrc := `locals {
  a = "${jsonencode({ x = 1 })}"
  b = var.flag ? "yes" : "no"
  c = "prefix-${var.name}"
}`
	jsonSrc := `{"locals": {
  "a": "${jsonencode({ x = 1 })}",
  "b": "${var.flag ? \"yes\" : \"no\"}",
  "c": "prefix-${var.name}"
}}`
	hclA, err := AnalyzeFiles(map[string][]byte{"main.tf": []byte(hclSrc)})
	if err != nil {
		t.Fatal(err)
	}
	jsonA, err := AnalyzeFiles(map[string][]byte{"main.tf.json": []byte(jsonSrc)})
	if err != nil {
		t.Fatal(err)
	}
	if !equalStrings(hclA.Features, jsonA.Features) {
		t.Errorf("features: hcl=%v json=%v", hclA.Features, jsonA.Features)
	}
	for _, want := range []string{"interpolation", "functions", "conditionals"} {
		found := false
		for _, f := range jsonA.Features {
			if f == want {
				found = true
			}
		}
		if !found {
			t.Errorf("json features missing %q: %v", want, jsonA.Features)
		}
	}
}

func TestJSONRequiredProvidersObjectShape(t *testing.T) {
	src := `{"terraform": {"required_providers": {"aws": {"source": "hashicorp/aws", "version": "~> 5.0"}}}}`
	a, err := AnalyzeFiles(map[string][]byte{"main.tf.json": []byte(src)})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range a.Features {
		if f == "required_providers" {
			found = true
		}
	}
	if !found {
		t.Errorf("features = %v", a.Features)
	}
}
