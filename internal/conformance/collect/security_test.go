package collect

import "testing"

func TestScanFiles(t *testing.T) {
	cases := []struct {
		name    string
		files   map[string][]byte
		want    string
		wantHit bool
	}{
		{
			name:    "provisioner",
			files:   map[string][]byte{"main.tf": []byte(`resource "aws_instance" "x" { provisioner "local-exec" { command = "echo hi" } }`)},
			want:    "executable",
			wantHit: true,
		},
		{
			name:    "external",
			files:   map[string][]byte{"main.tf": []byte(`data "external" "x" { program = ["sh", "-c", "id"] }`)},
			want:    "executable",
			wantHit: true,
		},
		{
			name:    "aws key",
			files:   map[string][]byte{"main.tf": []byte(`resource "aws_iam_access_key" "k" { secret = "AKIAABCDEFGHIJKLMNOP" }`)},
			want:    "secret",
			wantHit: true,
		},
		{
			name:    "placeholder key",
			files:   map[string][]byte{"main.tf": []byte(`resource "aws_iam_access_key" "k" { secret = "AKIAIOSFODNN7EXAMPLE" }`)},
			wantHit: false,
		},
		{
			name:    "password literal",
			files:   map[string][]byte{"main.tf": []byte(`resource "aws_db_instance" "db" { password = "supersecretvalue123" }`)},
			want:    "secret",
			wantHit: true,
		},
		{
			name:    "password interpolation",
			files:   map[string][]byte{"main.tf": []byte(`resource "aws_db_instance" "db" { password = "${var.db_password}" }`)},
			wantHit: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ScanFiles(tc.files)
			if tc.wantHit {
				if len(got) == 0 {
					t.Fatal("expected a violation")
				}
				if got[0].Kind != tc.want {
					t.Errorf("kind = %q, want %q", got[0].Kind, tc.want)
				}
				return
			}
			if len(got) != 0 {
				t.Fatalf("unexpected violations: %v", got)
			}
		})
	}
}
