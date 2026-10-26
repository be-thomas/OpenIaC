package collect

import "testing"

func TestRegistryNextURL(t *testing.T) {
	const page = "https://registry.terraform.io/v1/modules?limit=100&offset=0&provider=aws"
	cases := []struct {
		name string
		next string
		want string
	}{
		{"empty", "", ""},
		{"absolute", "https://registry.terraform.io/v1/modules?offset=100", "https://registry.terraform.io/v1/modules?offset=100"},
		{"relative", "/v1/modules?offset=200", "https://registry.terraform.io/v1/modules?offset=200"},
		{"protocol-relative", "//evil.example/v1/modules", ""},
		{"foreign origin", "https://evil.example/v1/modules", ""},
		{"foreign http", "http://registry.terraform.io/v1/modules", ""},
		{"userinfo trick", "https://registry.terraform.io@evil.example/v1/modules", ""},
		{"bad url", "://nope", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := registryNextURL(page, tc.next); got != tc.want {
				t.Errorf("registryNextURL(%q) = %q, want %q", tc.next, got, tc.want)
			}
		})
	}
}

func TestParseGitHubSource(t *testing.T) {
	cases := []struct {
		raw    string
		owner  string
		repo   string
		ref    string
		subdir string
		wantOK bool
	}{
		{"git::https://github.com/acme/mod?ref=abc123", "acme", "mod", "abc123", "", true},
		{"https://github.com/acme/mod.git?ref=main", "acme", "mod", "main", "", true},
		{"git@github.com:acme/mod.git?ref=v1", "acme", "mod", "v1", "", true},
		{"https://github.com/acme/mod/archive/v1.0.0.zip", "acme", "mod", "v1.0.0", "", true},
		{"https://github.com/acme/mod/archive/refs/tags/v2.1.0.tar.gz", "acme", "mod", "v2.1.0", "", true},
		{"https://github.com/acme/mod/archive/refs/heads/main.zip", "acme", "mod", "main", "", true},
		{"git::https://github.com/acme/mod.git?ref=abc123//modules/vpc", "acme", "mod", "abc123", "modules/vpc", true},
		{"git::http://github.com/acme/mod?ref=abc", "", "", "", "", false},
		{"git::https://gitlab.com/acme/mod?ref=abc", "", "", "", "", false},
		{"git::https://github.com/acme/mod", "", "", "", "", false},
		{"", "", "", "", "", false},
	}
	for _, tc := range cases {
		owner, repo, ref, subdir, ok := parseGitHubSource(tc.raw)
		if ok != tc.wantOK || owner != tc.owner || repo != tc.repo || ref != tc.ref || subdir != tc.subdir {
			t.Errorf("%q -> (%q,%q,%q,%q,%v), want (%q,%q,%q,%q,%v)",
				tc.raw, owner, repo, ref, subdir, ok, tc.owner, tc.repo, tc.ref, tc.subdir, tc.wantOK)
		}
	}
}

func TestRegistryIDIgnoresVersion(t *testing.T) {
	m := RegistryModule{ID: "acme/mod/aws/1.2.3", Namespace: "acme", Name: "mod", Provider: "aws", Version: "1.2.3"}
	if got := registryID(m); got != "acme/mod/aws" {
		t.Errorf("registryID = %q", got)
	}
}
