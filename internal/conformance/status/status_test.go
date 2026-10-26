package status_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/be-thomas/openiac/internal/conformance/status"
)

func TestComputeCountsParsedModules(t *testing.T) {
	root := t.TempDir()
	fixtures := filepath.Join(root, "fixtures")
	writeModule(t, fixtures, "registry/aws/mod/1.0.0/input", `resource "aws_s3_bucket" "b" {
  bucket = "example"
}
`)
	lock := writeLock(t, root, `{"fixtureCount":1,"fileCount":1,"totalBytes":42,"treeSha256":"abc"}`)

	report, err := status.Compute(fixtures, lock)
	if err != nil {
		t.Fatalf("Compute: %v", err)
	}
	if report.Parse.Modules != 1 || report.Parse.Parsed != 1 || report.Parse.Failed != 0 {
		t.Fatalf("parse = %+v, want 1/1 with no failures", report.Parse)
	}
	for name, target := range map[string]status.EmitTarget{"cdk": report.Emit.CDK, "pulumi": report.Emit.Pulumi} {
		if target.Modules != 1 || target.Emitted != 1 || target.Failed != 0 {
			t.Errorf("%s emit = %+v, want 1/1 with no failures", name, target)
		}
		if target.Clean != 1 || target.Warnings != 0 {
			t.Errorf("%s clean = %d with %d warnings, want a warning-free module", name, target.Clean, target.Warnings)
		}
	}
	if got := report.MinVersions["1.0.0"]; got != 1 {
		t.Errorf("min_versions[1.0.0] = %d, want 1 (%v)", got, report.MinVersions)
	}
	if report.Corpus != (status.Corpus{Fixtures: 1, Files: 1, Bytes: 42, TreeSHA256: "abc"}) {
		t.Errorf("corpus = %+v, want the lock's aggregates", report.Corpus)
	}
	if got := report.Badge(); got.Message != "1 / 1 modules" || got.Color != "brightgreen" {
		t.Errorf("badge = %+v, want 1 / 1 modules in brightgreen", got)
	}
}

// A module that emits with a warning still counts as emitted, but not as clean:
// the warning is what tells a reader a human is needed.
func TestComputeSeparatesEmittedFromClean(t *testing.T) {
	root := t.TempDir()
	fixtures := filepath.Join(root, "fixtures")
	writeModule(t, fixtures, "registry/aws/scaled/1.0.0/input", `resource "aws_instance" "web" {
  count         = 2
  ami           = "ami-0123456789abcdef0"
  instance_type = "t3.micro"
}
`)
	lock := writeLock(t, root, `{"fixtureCount":1,"fileCount":1,"totalBytes":1,"treeSha256":"abc"}`)

	report, err := status.Compute(fixtures, lock)
	if err != nil {
		t.Fatalf("Compute: %v", err)
	}
	for name, target := range map[string]status.EmitTarget{"cdk": report.Emit.CDK, "pulumi": report.Emit.Pulumi} {
		if target.Emitted != 1 {
			t.Fatalf("%s emitted = %d, want 1", name, target.Emitted)
		}
		if target.Clean != 0 || target.Warnings == 0 {
			t.Errorf("%s clean = %d with %d warnings, want a warned module", name, target.Clean, target.Warnings)
		}
	}
}

func TestComputeReportsFailures(t *testing.T) {
	root := t.TempDir()
	fixtures := filepath.Join(root, "fixtures")
	writeModule(t, fixtures, "registry/aws/broken/1.0.0/input", "resource \"aws_s3_bucket\" {\n")
	lock := writeLock(t, root, `{"fixtureCount":1,"fileCount":1,"totalBytes":1,"treeSha256":"abc"}`)

	report, err := status.Compute(fixtures, lock)
	if err != nil {
		t.Fatalf("Compute: %v", err)
	}
	if report.Parse.Failed != 1 || report.Parse.Parsed != 0 {
		t.Fatalf("parse = %+v, want one failure", report.Parse)
	}
	if len(report.Parse.Failures) != 1 || !strings.Contains(report.Parse.Failures[0], "broken/1.0.0/input") {
		t.Errorf("failures = %v, want the fixture path", report.Parse.Failures)
	}
	// A module that never parsed was never handed to an emitter.
	if report.Emit.CDK.Modules != 0 || report.Emit.Pulumi.Modules != 0 {
		t.Errorf("emit = %+v, want no attempts for an unparsed module", report.Emit)
	}
	if got := report.Badge().Color; got != "red" {
		t.Errorf("badge color = %q, want red", got)
	}
}

func TestComputeReportsEmitterFailures(t *testing.T) {
	root := t.TempDir()
	fixtures := filepath.Join(root, "fixtures")
	writeModule(t, fixtures, "registry/aws/cycle/1.0.0/input", `resource "aws_s3_bucket" "a" {
  bucket = aws_s3_bucket.b.id
}

resource "aws_s3_bucket" "b" {
  bucket = aws_s3_bucket.a.id
}
`)
	lock := writeLock(t, root, `{"fixtureCount":1,"fileCount":1,"totalBytes":1,"treeSha256":"abc"}`)

	report, err := status.Compute(fixtures, lock)
	if err != nil {
		t.Fatalf("Compute: %v", err)
	}
	// The cycle only appears when the graph is ordered for an emitter, so the
	// module parses and fails downstream: the emitter branch is reachable.
	if report.Parse.Parsed != 1 || report.Parse.Failed != 0 {
		t.Fatalf("parse = %+v, want the cyclic module to parse", report.Parse)
	}
	for name, target := range map[string]status.EmitTarget{"cdk": report.Emit.CDK, "pulumi": report.Emit.Pulumi} {
		if target.Modules != 1 || target.Emitted != 0 || target.Failed != 1 {
			t.Fatalf("%s emit = %+v, want one failed module", name, target)
		}
		if target.Emitted+target.Failed != target.Modules {
			t.Errorf("%s emitted + failed = %d, want the module count %d", name, target.Emitted+target.Failed, target.Modules)
		}
		if len(target.Failures) != 1 || !strings.Contains(target.Failures[0], "cycle") {
			t.Errorf("%s failures = %v, want the dependency cycle", name, target.Failures)
		}
	}
	if got := report.Badge().Color; got != "red" {
		t.Errorf("badge color = %q, want red", got)
	}
}

// The fixtures root can be absolute (-out is a path flag), but the committed
// report must not be: -check would then fail on a tree checked out elsewhere.
func TestComputeKeepsFailurePathsRelative(t *testing.T) {
	root := t.TempDir()
	fixtures := filepath.Join(root, "fixtures")
	writeModule(t, fixtures, "registry/aws/broken/1.0.0/input", "resource \"aws_s3_bucket\" {\n")
	lock := writeLock(t, root, `{"fixtureCount":1,"fileCount":1,"totalBytes":1,"treeSha256":"abc"}`)

	report, err := status.Compute(fixtures, lock)
	if err != nil {
		t.Fatalf("Compute: %v", err)
	}
	if len(report.Parse.Failures) != 1 {
		t.Fatalf("failures = %v, want exactly one", report.Parse.Failures)
	}
	failure := report.Parse.Failures[0]
	if strings.Contains(failure, root) {
		t.Errorf("failure quotes the absolute fixtures root: %s", failure)
	}
	if !strings.HasPrefix(failure, "registry/aws/broken/1.0.0/input: ") {
		t.Errorf("failure = %q, want it to start with the relative fixture path", failure)
	}
}

func TestBadgeIsRedWhenAnEmitterFails(t *testing.T) {
	ok := status.EmitTarget{Modules: 3, Emitted: 3}
	failing := status.EmitTarget{Modules: 3, Emitted: 2, Failed: 1}
	for name, emit := range map[string]status.Emit{
		"cdk":    {CDK: failing, Pulumi: ok},
		"pulumi": {CDK: ok, Pulumi: failing},
	} {
		report := status.Report{Parse: status.Parse{Modules: 3, Parsed: 3}, Emit: emit}
		if got := report.Badge().Color; got != "red" {
			t.Errorf("%s: badge color = %q, want red", name, got)
		}
	}
}

func TestOutcomesRenderPipelineOrder(t *testing.T) {
	report := status.Report{
		Parse: status.Parse{Modules: 100, Parsed: 100},
		Emit: status.Emit{
			CDK:    status.EmitTarget{Modules: 100, Emitted: 98, Failed: 2, Clean: 25, Warnings: 300},
			Pulumi: status.EmitTarget{Modules: 100, Emitted: 100, Clean: 2, Warnings: 120},
		},
	}
	want := []status.Outcome{
		{Label: "Parsed to the universal IR", Count: 100},
		{Label: "IR → AWS CDK TypeScript", Count: 98},
		{Label: "IR → Pulumi TypeScript", Count: 100},
		{Label: "IR → AWS CDK, no warnings", Count: 25},
		{Label: "IR → Pulumi, no warnings", Count: 2},
	}
	got := report.Outcomes()
	if len(got) != len(want) {
		t.Fatalf("outcomes = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("outcomes = %v, want %v", got, want)
		}
	}

	table := report.Markdown()
	for _, row := range []string{
		"| Parsed to the universal IR | 100 | ████████████████████ 100% |",
		"| IR → AWS CDK TypeScript | 98 | ███████████████████ 98% |",
		"| IR → AWS CDK, no warnings | 25 | █████ 25% |",
		"| IR → Pulumi, no warnings | 2 | ▏ 2% |",
	} {
		if !strings.Contains(table, row) {
			t.Errorf("table missing %q:\n%s", row, table)
		}
	}
	if strings.Contains(table, "Terraform") {
		t.Errorf("table still describes the version distribution:\n%s", table)
	}
}

func TestMarkdownNeverRoundsPastAPartialResult(t *testing.T) {
	report := status.Report{
		Parse: status.Parse{Modules: 853, Parsed: 853},
		Emit: status.Emit{
			CDK:    status.EmitTarget{Modules: 853, Emitted: 852, Failed: 1, Clean: 1},
			Pulumi: status.EmitTarget{Modules: 853, Emitted: 853, Clean: 0},
		},
	}
	table := report.Markdown()
	for _, row := range []string{
		// 852 of 853 must not read as a finished stage with a full bar.
		"| IR → AWS CDK TypeScript | 852 | ███████████████████ 99% |",
		// One module is a real result, not nothing.
		"| IR → AWS CDK, no warnings | 1 | ▏ 1% |",
		"| IR → Pulumi, no warnings | 0 | 0% |",
	} {
		if !strings.Contains(table, row) {
			t.Errorf("table missing %q:\n%s", row, table)
		}
	}
	if strings.Contains(table, "852 | ████████████████████ 100%") {
		t.Errorf("a partial corpus still renders as complete:\n%s", table)
	}
}

func TestSectionHeadline(t *testing.T) {
	full := status.Report{
		Parse: status.Parse{Modules: 2, Parsed: 2},
		Emit: status.Emit{
			CDK:    status.EmitTarget{Modules: 2, Emitted: 2},
			Pulumi: status.EmitTarget{Modules: 2, Emitted: 2},
		},
	}
	if want := "**2 / 2 modules — 100% — reach the IR and emit both AWS CDK and Pulumi TypeScript.**"; !strings.Contains(full.Section(), want) {
		t.Errorf("section missing %q:\n%s", want, full.Section())
	}

	partial := status.Report{
		Parse: status.Parse{Modules: 4, Parsed: 3},
		Emit: status.Emit{
			CDK:    status.EmitTarget{Modules: 3, Emitted: 2},
			Pulumi: status.EmitTarget{Modules: 3, Emitted: 3},
		},
	}
	if want := "**3 / 4 modules reach the IR, 2 / 4 reach AWS CDK and 3 / 4 reach Pulumi.**"; !strings.Contains(partial.Section(), want) {
		t.Errorf("section missing %q:\n%s", want, partial.Section())
	}
}

func TestSectionRoundTrip(t *testing.T) {
	report := status.Report{
		Parse:       status.Parse{Modules: 2, Parsed: 2},
		Emit:        status.Emit{CDK: status.EmitTarget{Modules: 2, Emitted: 2}, Pulumi: status.EmitTarget{Modules: 2, Emitted: 2}},
		MinVersions: map[string]int{"1.0.0": 1, "1.5.0": 1},
	}
	header := "# Title\n\nIntro.\n\n"
	footer := "\n\nTrailing prose.\n"
	document := header + report.Section() + footer

	section, err := status.ExtractSection(document)
	if err != nil {
		t.Fatalf("ExtractSection: %v", err)
	}
	if section != report.Section() {
		t.Fatalf("ExtractSection = %q, want the generated section", section)
	}
	if !strings.HasSuffix(section, status.SectionEnd) {
		t.Errorf("section does not end with the closing marker:\n%s", section)
	}

	updated := status.Report{
		Parse: status.Parse{Modules: 2, Parsed: 2},
		Emit:  status.Emit{CDK: status.EmitTarget{Modules: 2, Emitted: 1}, Pulumi: status.EmitTarget{Modules: 2, Emitted: 2}},
	}
	replaced, err := status.ReplaceSection(document, updated.Section())
	if err != nil {
		t.Fatalf("ReplaceSection: %v", err)
	}
	if want := header + updated.Section() + footer; replaced != want {
		t.Fatalf("ReplaceSection = %q, want %q", replaced, want)
	}

	again, err := status.ReplaceSection(replaced, updated.Section())
	if err != nil {
		t.Fatalf("ReplaceSection (second pass): %v", err)
	}
	if again != replaced {
		t.Fatalf("ReplaceSection is not idempotent:\nfirst  %q\nsecond %q", replaced, again)
	}
}

func TestSectionRejectsUnmarkedDocument(t *testing.T) {
	if _, err := status.ExtractSection("# no markers here\n"); err == nil {
		t.Error("ExtractSection accepted a document without markers")
	}
	if _, err := status.ReplaceSection(status.SectionStart+"\n", "replacement"); err == nil {
		t.Error("ReplaceSection accepted a document with an unclosed block")
	}
}

func writeModule(t *testing.T, fixtures, rel, content string) {
	t.Helper()
	dir := filepath.Join(fixtures, rel)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.tf"), []byte(content), 0o644); err != nil {
		t.Fatalf("write module: %v", err)
	}
}

func writeLock(t *testing.T, root, content string) string {
	t.Helper()
	path := filepath.Join(root, "corpus.lock.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write lock: %v", err)
	}
	return path
}
