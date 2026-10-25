package collect

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFixturePreservesCuratedFiles(t *testing.T) {
	out := t.TempDir()
	cand := &Candidate{
		Dir: "",
		Files: map[string][]byte{
			"main.tf":               []byte(`resource "aws_s3_bucket" "b" {}`),
			"modules/child/main.tf": []byte(`resource "aws_sns_topic" "t" {}`),
		},
		Features:  []string{"modules", "resources"},
		Providers: []string{"aws"},
		Tier:      "modules",
		Support:   "partial",
		Bytes:     75,
	}
	m := GitModule{Registry: &RegistryModule{Namespace: "acme", Name: "mod", Provider: "aws", Version: "1.2.3"}, Owner: "acme", Repo: "mod", Ref: "abc123"}
	f := BuildFixture(m, cand, "archivehash", "Apache-2.0")
	if f.ID != "registry/acme/mod/1.2.3" {
		t.Fatalf("id = %q", f.ID)
	}
	dir := filepath.Join(out, filepath.FromSlash(f.ID))
	if err := os.MkdirAll(filepath.Join(dir, "expected"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "expected", "ir.json"), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := WriteFixture(out, f, cand, []byte("license"), []byte("notice")); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"input/main.tf", "input/modules/child/main.tf", "LICENSE", "NOTICE", "fixture.json", "expected/ir.json"} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(name))); err != nil {
			t.Errorf("missing %s: %v", name, err)
		}
	}
	data, err := os.ReadFile(filepath.Join(dir, "fixture.json"))
	if err != nil {
		t.Fatal(err)
	}
	var got Fixture
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Source.InputSHA256 == "" || got.Source.ArchiveSHA256 != "archivehash" {
		t.Errorf("source = %+v", got.Source)
	}
	if got.Source.Path != "." {
		t.Errorf("path = %q", got.Source.Path)
	}
}

func TestWriteFixtureRemovesStaleNotice(t *testing.T) {
	out := t.TempDir()
	cand := &Candidate{
		Files:     map[string][]byte{"main.tf": []byte(`resource "aws_s3_bucket" "b" {}`)},
		Features:  []string{"resources"},
		Providers: []string{"aws"},
		Tier:      "core",
		Support:   "full",
	}
	m := GitModule{Registry: &RegistryModule{Namespace: "acme", Name: "mod", Provider: "aws", Version: "1.0.0"}}
	f := BuildFixture(m, cand, "archivehash", "MIT")
	if _, err := WriteFixture(out, f, cand, []byte("license"), []byte("notice")); err != nil {
		t.Fatal(err)
	}
	if _, err := WriteFixture(out, f, cand, []byte("license"), nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(out, filepath.FromSlash(f.ID), "NOTICE")); !os.IsNotExist(err) {
		t.Errorf("stale NOTICE still present: %v", err)
	}
}

func TestInputHashIsOrderIndependent(t *testing.T) {
	a := map[string][]byte{"a.tf": []byte("a"), "b.tf": []byte("b")}
	b := map[string][]byte{"b.tf": []byte("b"), "a.tf": []byte("a")}
	if InputHash(a) != InputHash(b) {
		t.Error("input hash depends on map iteration")
	}
	if InputHash(a) == InputHash(map[string][]byte{"a.tf": []byte("a")}) {
		t.Error("input hash ignores file set")
	}
}

func TestSafeSegment(t *testing.T) {
	if got := safeSegment("../evil/name"); got == "..-evil-name" || got == "" {
		t.Errorf("unsafe segment = %q", got)
	}
	if got := safeSegment("v1.2.3"); got != "v1.2.3" {
		t.Errorf("version = %q", got)
	}
}
