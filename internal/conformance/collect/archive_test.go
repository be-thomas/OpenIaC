package collect

import (
	"archive/tar"
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
	write("repo-abc/LICENSE", []byte("Apache License\nVersion 2.0, January 2004\n"))
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
