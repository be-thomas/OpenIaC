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
	cands, skipped := BuildCandidates(a, 10, Defaults().Limits, nil)
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
	if cands, _ := BuildCandidates(a, 10, limits, nil); len(cands) != 0 {
		t.Errorf("expected cap rejection, got %d", len(cands))
	}
}

func TestBuildCandidatesIgnoresCommentedSource(t *testing.T) {
	a := &Archive{Files: map[string][]byte{
		"main.tf": []byte("// source = \"../outside\"\nresource \"aws_s3_bucket\" \"b\" {}\n"),
	}}
	cands, skipped := BuildCandidates(a, 10, Defaults().Limits, nil)
	if len(cands) != 1 || cands[0].Dir != "" {
		t.Fatalf("candidates = %v, skips = %v", dirsOf(cands), skipped)
	}
}

func TestBuildCandidatesExcludesTestTrees(t *testing.T) {
	a := &Archive{Files: map[string][]byte{
		"tests/unit/main.tf": []byte(`resource "aws_s3_bucket" "b" {}`),
	}}
	cands, skipped := BuildCandidates(a, 10, Defaults().Limits, nil)
	if len(cands) != 0 {
		t.Fatalf("test tree became a candidate: %v", dirsOf(cands))
	}
	if len(skipped) != 1 || skipped[0].Reason != "excluded directory" {
		t.Fatalf("skips = %v", skipped)
	}
}

func TestBuildCandidatesIncludePrefix(t *testing.T) {
	a := &Archive{Files: map[string][]byte{
		"main.tf":           []byte(`resource "aws_s3_bucket" "b" {}`),
		"modules/a/main.tf": []byte(`resource "aws_sns_topic" "t" {}`),
	}}
	cands, _ := BuildCandidates(a, 10, Defaults().Limits, []string{"modules/**"})
	if len(cands) != 1 || cands[0].Dir != "modules/a" {
		t.Fatalf("candidates = %v", dirsOf(cands))
	}
}

func TestBuildCandidatesRejectsNonAWS(t *testing.T) {
	a := &Archive{Files: map[string][]byte{
		"main.tf": []byte(`resource "google_storage_bucket" "b" {}`),
	}}
	cands, skipped := BuildCandidates(a, 10, Defaults().Limits, nil)
	if len(cands) != 0 {
		t.Fatalf("non-AWS candidate kept: %v", dirsOf(cands))
	}
	if len(skipped) != 1 || !strings.Contains(skipped[0].Reason, "no supported output target") {
		t.Fatalf("skips = %v", skipped)
	}
}

func TestBuildCandidatesModuleWrapperKeepsTargets(t *testing.T) {
	a := &Archive{Files: map[string][]byte{
		"main.tf": []byte(`module "vpc" { source = "terraform-aws-modules/vpc/aws" }`),
	}}
	cands, skipped := BuildCandidates(a, 10, Defaults().Limits, nil)
	if len(cands) != 1 {
		t.Fatalf("skips = %v", skipped)
	}
	if len(cands[0].Targets) != 2 {
		t.Errorf("targets = %v", cands[0].Targets)
	}
	if cands[0].Tier != "modules" {
		t.Errorf("tier = %q", cands[0].Tier)
	}
}

func dirsOf(cands []*Candidate) []string {
	out := make([]string, 0, len(cands))
	for _, c := range cands {
		out = append(out, c.Dir)
	}
	return out
}
