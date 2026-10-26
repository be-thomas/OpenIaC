package collect

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"testing"
)

func TestExtractTarGzStripsRootAndKeepsText(t *testing.T) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	write := func(name string, data []byte) {
		t.Helper()
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(data)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	write("repo-abc/main.tf", []byte(`resource "aws_s3_bucket" "b" {}`))
	write("repo-abc/modules/child/main.tf", []byte(`resource "aws_sns_topic" "t" {}`))
	write("repo-abc/LICENSE", []byte("Apache License\nVersion 2.0, January 2004\nhttp://www.apache.org/licenses/LICENSE-2.0\n"))
	write("repo-abc/NOTICE", []byte("notices"))
	write("repo-abc/README.md", []byte("# readme"))
	write("repo-abc/logo.png", []byte{0x00, 0x01, 0xff})
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}

	a, err := ExtractArchive(buf.Bytes(), 1<<20, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	if len(a.Files) != 2 {
		t.Fatalf("files = %d, want 2: %v", len(a.Files), keys(a.Files))
	}
	if _, ok := a.Files["main.tf"]; !ok {
		t.Errorf("missing main.tf: %v", keys(a.Files))
	}
	if _, ok := a.Files["modules/child/main.tf"]; !ok {
		t.Errorf("missing child: %v", keys(a.Files))
	}
	if _, ok := a.Files["README.md"]; ok {
		t.Error("README.md should not be stored")
	}
	name, text, ok := a.PrimaryLicense()
	if !ok || name != "LICENSE" {
		t.Fatalf("license = %q, %v", name, ok)
	}
	if got := DetectLicense(text); got != "Apache-2.0" {
		t.Errorf("license = %q", got)
	}
	if _, _, ok := a.PrimaryNotice(); !ok {
		t.Error("notice missing")
	}
}

func TestExtractArchiveRejectsTextBomb(t *testing.T) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	data := bytes.Repeat([]byte("a"), 2048)
	if err := tw.WriteHeader(&tar.Header{Name: "repo-abc/main.tf", Mode: 0o644, Size: int64(len(data)), Typeflag: tar.TypeReg}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(data); err != nil {
		t.Fatal(err)
	}
	tw.Close()
	gz.Close()
	if _, err := ExtractArchive(buf.Bytes(), 1024, 4096); err == nil {
		t.Fatal("expected cap error")
	}
}

func keys(m map[string][]byte) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func TestExtractZipKeepsNestedLicense(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	write := func(name, data string) {
		t.Helper()
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(data)); err != nil {
			t.Fatal(err)
		}
	}
	write("repo-main/main.tf", `resource "aws_s3_bucket" "b" {}`)
	write("repo-main/modules/x/main.tf", `resource "aws_sns_topic" "t" {}`)
	write("repo-main/modules/x/LICENSE", "MIT License\nPermission is hereby granted, free of charge\nThe above copyright notice and this permission notice shall be included")
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	a, err := ExtractArchive(buf.Bytes(), 1<<20, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	if len(a.Files) != 2 {
		t.Fatalf("files = %v", keys(a.Files))
	}
	name, text, ok := a.NearestLicense("modules/x")
	if !ok || name != "modules/x/LICENSE" {
		t.Fatalf("license = %q, %v", name, ok)
	}
	if got := DetectLicense(text); got != "MIT" {
		t.Errorf("license = %q", got)
	}
	if _, _, ok := a.NearestLicense(""); ok {
		t.Error("root should not fall back to a nested license")
	}
}

func TestExtractArchiveRejectsOversizedFile(t *testing.T) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	data := bytes.Repeat([]byte("a"), 200)
	if err := tw.WriteHeader(&tar.Header{Name: "repo/main.tf", Mode: 0o644, Size: int64(len(data)), Typeflag: tar.TypeReg}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(data); err != nil {
		t.Fatal(err)
	}
	tw.Close()
	gz.Close()
	if _, err := ExtractArchive(buf.Bytes(), 1<<20, 100); err == nil {
		t.Fatal("expected oversize error, not a silent drop")
	}
}

func TestExtractArchiveSkipsSymlinks(t *testing.T) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{Name: "repo/main.tf", Mode: 0o644, Size: 31, Typeflag: tar.TypeReg}); err != nil {
		t.Fatal(err)
	}
	tw.Write([]byte(`resource "aws_s3_bucket" "b" {}`))
	if err := tw.WriteHeader(&tar.Header{Name: "repo/evil.tf", Linkname: "/etc/passwd", Typeflag: tar.TypeSymlink}); err != nil {
		t.Fatal(err)
	}
	tw.Close()
	gz.Close()
	a, err := ExtractArchive(buf.Bytes(), 1<<20, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := a.Files["evil.tf"]; ok {
		t.Error("symlink entry was extracted")
	}
}

func TestExtractArchiveBoundsSkippedExpansion(t *testing.T) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	body := bytes.Repeat([]byte("a"), 64<<10)
	h := &tar.Header{Name: "repo/.terraform/blob.bin", Mode: 0o644, Size: int64(len(body)), Typeflag: tar.TypeReg}
	if err := tw.WriteHeader(h); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(body); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	// The entry is skipped, but tar still has to decompress it. The 8x
	// expansion cap must stop it rather than silently reading 64 KiB.
	if _, err := ExtractArchive(buf.Bytes(), 1024, 1<<20); err == nil {
		t.Fatal("skipped expansion was not bounded")
	}
}
