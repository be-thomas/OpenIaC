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
