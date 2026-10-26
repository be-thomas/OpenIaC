package collect

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeValidCorpus(t *testing.T) (out, lockPath, thirdPartyPath string, cand *Candidate) {
	t.Helper()
	out = t.TempDir()
	lockPath = filepath.Join(t.TempDir(), "corpus.lock.json")
	thirdPartyPath = filepath.Join(t.TempDir(), "THIRD_PARTY.md")

	files := map[string][]byte{"main.tf": []byte(`resource "aws_s3_bucket" "b" {}`)}
	analysis, err := AnalyzeFiles(files)
	if err != nil {
		t.Fatal(err)
	}
	license := []byte("MIT License\nPermission is hereby granted, free of charge\nThe above copyright notice and this permission notice shall be included")
	cand = &Candidate{
		Files:       files,
		Features:    analysis.Features,
		Providers:   analysis.Providers,
		Targets:     []string{"cdk", "pulumi"},
		Tier:        ClassifyTier(analysis.Features),
		Support:     SupportLevel(analysis.Features),
		Bytes:       31,
		License:     "MIT",
		LicenseAt:   "LICENSE",
		LicenseText: license,
	}
	m := GitModule{
		Registry: &RegistryModule{Namespace: "acme", Name: "mod", Provider: "aws", Version: "1.0.0"},
		Owner:    "acme",
		Repo:     "mod",
		Ref:      strings.Repeat("a", 40),
	}
	f := BuildFixture(m, cand, strings.Repeat("b", 64))
	// A single basic fixture is always in the derived smoke set.
	f.Tier = "smoke"
	if _, err := WriteFixture(out, f, cand, license, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := WriteLock(out, lockPath); err != nil {
		t.Fatal(err)
	}
	if err := WriteThirdParty(out, thirdPartyPath); err != nil {
		t.Fatal(err)
	}
	return out, lockPath, thirdPartyPath, cand
}

func TestRefreshAggregatesUsesDefaultLimits(t *testing.T) {
	out, lockPath, thirdPartyPath, _ := writeValidCorpus(t)
	lock, err := RefreshAggregates(out, lockPath, thirdPartyPath, Config{})
	if err != nil {
		t.Fatalf("refresh with bare config: %v", err)
	}
	if lock.FixtureCount != 1 {
		t.Errorf("lock = %+v", lock)
	}
	if _, err := Verify(out, lockPath, thirdPartyPath, Defaults()); err != nil {
		t.Fatalf("verify after refresh: %v", err)
	}
	entries, err := os.ReadDir(filepath.Dir(lockPath))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.Contains(e.Name(), ".tmp-") {
			t.Errorf("temporary file left behind: %s", e.Name())
		}
	}
}

func TestVerifyAcceptsConsistentCorpus(t *testing.T) {
	out, lockPath, thirdPartyPath, _ := writeValidCorpus(t)
	res, err := Verify(out, lockPath, thirdPartyPath, Defaults())
	if err != nil {
		t.Fatal(err)
	}
	if res.Fixtures != 1 || res.Files != 3 {
		t.Errorf("result = %+v", res)
	}
}

func TestVerifyRejectsStaleLock(t *testing.T) {
	out, lockPath, thirdPartyPath, _ := writeValidCorpus(t)
	fixtureDir := filepath.Join(out, "registry", "acme", "mod", "1.0.0")
	if err := os.WriteFile(filepath.Join(fixtureDir, "NOTES.md"), []byte("curated\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Verify(out, lockPath, thirdPartyPath, Defaults())
	if err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("err = %v", err)
	}
}

func TestVerifyRejectsTamperedInput(t *testing.T) {
	out, lockPath, thirdPartyPath, _ := writeValidCorpus(t)
	input := filepath.Join(out, "registry", "acme", "mod", "1.0.0", "input", "main.tf")
	if err := os.WriteFile(input, []byte(`resource "aws_s3_bucket" "b" { bucket = "x" }`), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Verify(out, lockPath, thirdPartyPath, Defaults())
	if err == nil || !strings.Contains(err.Error(), "input hash") {
		t.Fatalf("err = %v", err)
	}
}

func TestVerifyRejectsStrayFile(t *testing.T) {
	out, lockPath, thirdPartyPath, _ := writeValidCorpus(t)
	if err := os.WriteFile(filepath.Join(out, "stray.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Verify(out, lockPath, thirdPartyPath, Defaults())
	if err == nil || !strings.Contains(err.Error(), "stray file") {
		t.Fatalf("err = %v", err)
	}
}

func TestVerifyRejectsSymlink(t *testing.T) {
	out, lockPath, thirdPartyPath, _ := writeValidCorpus(t)
	link := filepath.Join(out, "registry", "acme", "mod", "1.0.0", "input", "link.tf")
	if err := os.Symlink("/etc/passwd", link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if _, err := Verify(out, lockPath, thirdPartyPath, Defaults()); err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("err = %v", err)
	}
}

func TestVerifyRejectsWrongLicense(t *testing.T) {
	out, lockPath, thirdPartyPath, _ := writeValidCorpus(t)
	license := filepath.Join(out, "registry", "acme", "mod", "1.0.0", "LICENSE")
	text := "GNU GENERAL PUBLIC LICENSE Version 3"
	if err := os.WriteFile(license, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Verify(out, lockPath, thirdPartyPath, Defaults())
	if err == nil || !strings.Contains(err.Error(), "LICENSE") {
		t.Fatalf("err = %v", err)
	}
}

func TestVerifyHonorsConfigLicenseAllowlist(t *testing.T) {
	out, lockPath, thirdPartyPath, _ := writeValidCorpus(t)
	cfg := Defaults()
	cfg.Licenses = []string{"Apache-2.0"}
	_, err := Verify(out, lockPath, thirdPartyPath, cfg)
	if err == nil || !strings.Contains(err.Error(), "not on the allowlist") {
		t.Fatalf("err = %v", err)
	}
}

func TestVerifyRejectsTierOutsideDerivedSmokeSet(t *testing.T) {
	out, lockPath, thirdPartyPath, _ := writeValidCorpus(t)
	path := filepath.Join(out, "registry", "acme", "mod", "1.0.0", "fixture.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	data = []byte(strings.Replace(string(data), `"tier": "smoke"`, `"tier": "advanced"`, 1))
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = Verify(out, lockPath, thirdPartyPath, Defaults())
	if err == nil || !strings.Contains(err.Error(), "derived smoke set") {
		t.Fatalf("err = %v", err)
	}
}

func TestVerifyRejectsProvenanceMismatch(t *testing.T) {
	out, lockPath, thirdPartyPath, _ := writeValidCorpus(t)
	path := filepath.Join(out, "registry", "acme", "mod", "1.0.0", "fixture.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	data = []byte(strings.Replace(string(data),
		`"url": "https://registry.terraform.io/modules/acme/mod/aws/1.0.0"`,
		`"url": "https://example.com/acme/mod"`, 1))
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = Verify(out, lockPath, thirdPartyPath, Defaults())
	if err == nil || !strings.Contains(err.Error(), "non-registry source url") {
		t.Fatalf("err = %v", err)
	}
}

func TestRenderThirdPartyIsDeterministic(t *testing.T) {
	fixtures := []Fixture{
		{ID: "b", License: "MIT", Source: Source{URL: "https://example.com/b", Ref: "v1", Commit: strings.Repeat("c", 40)}},
		{ID: "a", License: "Apache-2.0", Source: Source{URL: "https://example.com/a", Ref: "v2", Commit: strings.Repeat("d", 40)}},
	}
	first := string(RenderThirdParty(fixtures))
	second := string(RenderThirdParty(fixtures))
	if first != second {
		t.Fatal("render is not deterministic")
	}
	if !strings.Contains(first, "| `a` | Apache-2.0 |") {
		t.Errorf("unexpected table:\n%s", first)
	}
}
