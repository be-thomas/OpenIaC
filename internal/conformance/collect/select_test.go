package collect

import (
	"strings"
	"testing"
)

func TestBuildCandidatesIsSelfContained(t *testing.T) {
	a := &Archive{Files: map[string][]byte{
		"main.tf": []byte(`
resource "aws_s3_bucket" "b" {}
module "child" { source = "./modules/child" }
`),
		"variables.tf":             []byte(`variable "x" {}`),
		"modules/child/main.tf":    []byte(`resource "aws_iam_role" "r" {}`),
		"examples/good/main.tf":    []byte(`resource "aws_sns_topic" "t" {}`),
		"examples/bad/main.tf":     []byte(`module "root" { source = "../../" }`),
		"examples/missing/main.tf": []byte(`module "x" { source = "./nope" }`),
	}}
	cands, skipped := BuildCandidates(a, 10, Defaults().Limits)
	byDir := make(map[string]*Candidate)
	for _, c := range cands {
		byDir[c.Dir] = c
	}
	root, ok := byDir[""]
	if !ok {
		t.Fatalf("root candidate missing: %v", skipped)
	}
	if _, ok := root.Files["modules/child/main.tf"]; !ok {
		t.Errorf("root did not include local module: %v", keys(root.Files))
	}
	if _, ok := root.Files["examples/good/main.tf"]; ok {
		t.Error("root should not include an unrelated example")
	}
	child, ok := byDir["modules/child"]
	if !ok {
		t.Fatalf("child candidate missing: %v", skipped)
	}
	if len(child.Files) != 1 || child.Files["main.tf"] == nil {
		t.Errorf("child files = %v", keys(child.Files))
	}
	if _, ok := byDir["examples/good"]; !ok {
		t.Errorf("self-contained example missing: %v", skipped)
	}
	if _, ok := byDir["examples/bad"]; ok {
		t.Error("escaping example should be rejected")
	}
	if _, ok := byDir["examples/missing"]; ok {
		t.Error("missing local module should be rejected")
	}
	var reasons []string
	for _, s := range skipped {
		reasons = append(reasons, s.Dir+": "+s.Reason)
	}
	joined := strings.Join(reasons, "\n")
	if !strings.Contains(joined, "escapes module root") {
		t.Errorf("missing escape reason:\n%s", joined)
	}
	if !strings.Contains(joined, "not found") {
		t.Errorf("missing not-found reason:\n%s", joined)
	}
}

func TestBuildCandidatesCaps(t *testing.T) {
	a := &Archive{Files: map[string][]byte{
		"main.tf": []byte(`resource "aws_s3_bucket" "b" {}`),
	}}
	limits := Defaults().Limits
	limits.MaxFixtureBytes = 5
	if cands, _ := BuildCandidates(a, 10, limits); len(cands) != 0 {
		t.Errorf("expected cap rejection, got %d", len(cands))
	}
}
