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
	if got := report.Badge().Color; got != "red" {
		t.Errorf("badge color = %q, want red", got)
	}
}

func TestMinorRowsAggregateAndSort(t *testing.T) {
	report := status.Report{
		Parse: status.Parse{Modules: 6, Parsed: 6},
		MinVersions: map[string]int{
			"1.10.0": 2,
			"1.9.0":  1,
			"1.0.0":  3,
		},
	}
	rows := report.MinorRows()
	want := []status.VersionCount{{Version: "1.0", Fixtures: 3}, {Version: "1.9", Fixtures: 1}, {Version: "1.10", Fixtures: 2}}
	if len(rows) != len(want) {
		t.Fatalf("rows = %v, want %v", rows, want)
	}
	for i := range want {
		if rows[i] != want[i] {
			t.Fatalf("rows = %v, want %v", rows, want)
		}
	}

	chart := report.Mermaid()
	for _, token := range []string{`x-axis ["1.0", "1.9", "1.10"]`, "bar [3, 1, 2]", "y-axis \"Fixtures\" 0 --> 10"} {
		if !strings.Contains(chart, token) {
			t.Errorf("chart missing %q:\n%s", token, chart)
		}
	}
}

func TestSectionRoundTrip(t *testing.T) {
	report := status.Report{
		Parse:       status.Parse{Modules: 2, Parsed: 2},
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

	updated := status.Report{
		Parse:       status.Parse{Modules: 2, Parsed: 2},
		MinVersions: map[string]int{"1.0.0": 2},
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
