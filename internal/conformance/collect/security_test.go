package collect

import (
	"strings"
	"testing"
)

func TestScanFiles(t *testing.T) {
	cases := []struct {
		name    string
		files   map[string][]byte
		want    string
		wantHit bool
	}{
		{
			name: "provisioner",
			files: map[string][]byte{"main.tf": []byte(`resource "aws_instance" "x" {
  provisioner "local-exec" {
    command = "echo hi"
  }
}`)},
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

func TestScanFilesVariableDefault(t *testing.T) {
	files := map[string][]byte{"variables.tf": []byte(`
variable "db_password" {
  default = "hunter2"
}
`)}
	got := ScanFiles(files)
	if len(got) == 0 || got[0].Kind != "secret" {
		t.Fatalf("violations = %v", got)
	}
	if got := ScanFiles(map[string][]byte{"variables.tf": []byte(`
variable "db_password" {
  default = "changeme"
}
`)}); len(got) != 0 {
		t.Fatalf("placeholder rejected: %v", got)
	}
}

func TestScanFilesPrivateKey(t *testing.T) {
	files := map[string][]byte{"main.tf": []byte(`locals {
  key = <<-EOT
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA
-----END RSA PRIVATE KEY-----
EOT
}`)}
	got := ScanFiles(files)
	if len(got) == 0 || got[0].Kind != "secret" {
		t.Fatalf("violations = %v", got)
	}
}

func TestScanFilesRejectsUnparseableInput(t *testing.T) {
	got := ScanFiles(map[string][]byte{"main.tf": []byte(`resource "aws_s3_bucket" "b" {`)})
	if len(got) != 1 || got[0].Kind != "parse" {
		t.Fatalf("violations = %v", got)
	}
}

func TestScanFilesJSONSecrets(t *testing.T) {
	cases := []struct {
		name  string
		files map[string][]byte
		want  string
	}{
		{
			name:  "nested object key",
			files: map[string][]byte{"main.tf.json": []byte(`{"variable": {"cfg": {"default": {"password": "hunter2"}}}}`)},
			want:  "credential-like literal for password",
		},
		{
			name:  "secret named variable default",
			files: map[string][]byte{"main.tf.json": []byte(`{"variable": {"db_password": {"default": "hunter2"}}}`)},
			want:  "credential-like default for variable db_password",
		},
		{
			name:  "nested list of objects",
			files: map[string][]byte{"main.tf.json": []byte(`{"resource": {"aws_instance": {"x": {"users": [{"api_key": "abc123"}]}}}}`)},
			want:  "credential-like literal for api_key",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a, err := AnalyzeFiles(tc.files)
			if err != nil {
				t.Fatal(err)
			}
			for _, v := range a.Violations {
				if v.Kind == "secret" && strings.Contains(v.Detail, tc.want) {
					return
				}
			}
			t.Fatalf("no violation containing %q: %v", tc.want, a.Violations)
		})
	}
}

func TestScanFilesJSONPlaceholder(t *testing.T) {
	a, err := AnalyzeFiles(map[string][]byte{"main.tf.json": []byte(`{"variable": {"cfg": {"default": {"password": "${var.input}"}}}}`)})
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range a.Violations {
		if v.Kind == "secret" {
			t.Fatalf("placeholder flagged: %v", v)
		}
	}
}

func TestSecretNameBoundaries(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want bool
	}{
		{"http_tokens enum", `resource "aws_eks_node_group" "x" { http_tokens = "required" }`, false},
		{"camelCase httpTokens", `resource "aws_eks_node_group" "x" { httpTokens = "required" }`, false},
		{"token type qualifier", `variable "github_token_type" { default = "pat" }`, false},
		{"credentials type qualifier", `variable "image_pull_credentials_type" { default = "SERVICE_ACCOUNT" }`, false},
		{"secret path qualifier", `variable "secret_path_format" { default = "/run/secrets" }`, false},
		{"private key algorithm", `variable "private_key_algorithm" { default = "RSA" }`, false},
		{"token validity unit", `variable "token_validity_units" { default = { access_token = "minutes" } }`, false},
		{"creation token phrase", `variable "creation_token" { default = "EFS-creation-token" }`, false},
		{"sample digit key", `variable "dd_api_key" { default = "1234567890" }`, false},
		{"angle bracket placeholder", `resource "aws_instance" "x" { api_token = "<TOKEN>" }`, false},
		{"api key", `resource "aws_instance" "x" { api_key = "abcdef123456" }`, true},
		{"secret key", `resource "aws_iam_access_key" "k" { secret = "abcdef123456" }`, true},
		{"access key id", `resource "aws_iam_access_key" "k" { aws_access_key_id = "abcdef123456" }`, true},
		{"snake password", `variable "db_password" { default = "hunter22" }`, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a, err := AnalyzeFiles(map[string][]byte{"main.tf": []byte(tc.src)})
			if err != nil {
				t.Fatal(err)
			}
			got := false
			for _, v := range a.Violations {
				if v.Kind == "secret" {
					got = true
				}
			}
			if got != tc.want {
				t.Errorf("secret violation = %v, want %v (%v)", got, tc.want, a.Violations)
			}
		})
	}
}

func TestSecretDetectionWalksWrappers(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{"function call", `locals { cfg = jsonencode({ password = "secretpassword" }) }`},
		{"conditional", `variable "cfg" { default = var.enabled ? { project_token = "phc_xwn7HLdbaLxDAq7gbiKTbbJt5oMtXHkRRmaN8aeWeUnk" } : null }`},
		{"parentheses", `locals { cfg = ({ api_key = "abc123" }) }`},
		{"template", `locals { cfg = "${jsonencode({ secret = "hunter22" })}" }`},
		{"for expression", `locals { cfg = [for s in var.list : { token = "abc123" }] }`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a, err := AnalyzeFiles(map[string][]byte{"main.tf": []byte(tc.src)})
			if err != nil {
				t.Fatal(err)
			}
			for _, v := range a.Violations {
				if v.Kind == "secret" {
					return
				}
			}
			t.Fatalf("no secret violation: %v", a.Violations)
		})
	}
}

func TestSecretNameVariants(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want bool
	}{
		{"camelCase password", `variable "x" { dbPassword = "hunter2" }`, true},
		{"secret string", `variable "x" { secret_string = "correcthorsebatterystaple" }`, true},
		{"private key pem", `variable "x" { private_key_pem = "abc" }`, true},
		{"password value", `variable "x" { password_value = "notthenormalform" }`, true},
		{"token type", `variable "x" { token_type = "required" }`, false},
		{"secret path", `variable "x" { secret_path_format = "/run/secrets" }`, false},
		{"credentials type", `variable "x" { image_pull_credentials_type = "SERVICE_ACCOUNT" }`, false},
		{"secret manager key name", `variable "x" { aws_secret_manager_git_private_ssh_key_name = "github-blueprint-ssh-key" }`, false},
		{"password group", `variable "x" { iam_change_password_group = "CCS_IAM_User_Change_Password" }`, false},
		{"kms alias", `variable "x" { repository_credentials_kms_key = "alias/aws/secretsmanager" }`, false},
		{"secret extensions", `variable "x" { secret_extensions = { private_key = "key" } }`, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a, err := AnalyzeFiles(map[string][]byte{"main.tf": []byte(tc.src)})
			if err != nil {
				t.Fatal(err)
			}
			got := false
			for _, v := range a.Violations {
				if v.Kind == "secret" {
					got = true
				}
			}
			if got != tc.want {
				t.Errorf("secret violation = %v, want %v (%v)", got, tc.want, a.Violations)
			}
		})
	}
}

func TestSecretViolationsAreNotDuplicated(t *testing.T) {
	src := `variable "cfg" {
  default = {
    password = "hunter2"
  }
}`
	a, err := AnalyzeFiles(map[string][]byte{"main.tf": []byte(src)})
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, v := range a.Violations {
		if v.Kind == "secret" && strings.Contains(v.Detail, "password") {
			count++
		}
	}
	if count != 1 {
		t.Errorf("password violations = %d, want 1: %v", count, a.Violations)
	}
}
