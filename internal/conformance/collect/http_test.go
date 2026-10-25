package collect

import "testing"

func TestParseGitHubSource(t *testing.T) {
	cases := []struct {
		raw    string
		owner  string
		repo   string
		ref    string
		wantOK bool
	}{
		{"git::https://github.com/acme/mod?ref=abc123", "acme", "mod", "abc123", true},
		{"https://github.com/acme/mod.git?ref=main", "acme", "mod", "main", true},
		{"git@github.com:acme/mod.git?ref=v1", "acme", "mod", "v1", true},
		{"https://github.com/acme/mod/archive/v1.0.0.zip", "acme", "mod", "v1.0.0", true},
		{"git::https://gitlab.com/acme/mod?ref=abc", "", "", "", false},
		{"git::https://github.com/acme/mod", "", "", "", false},
		{"", "", "", "", false},
	}
	for _, tc := range cases {
		owner, repo, ref, ok := parseGitHubSource(tc.raw)
		if ok != tc.wantOK || owner != tc.owner || repo != tc.repo || ref != tc.ref {
			t.Errorf("%q -> (%q,%q,%q,%v), want (%q,%q,%q,%v)",
				tc.raw, owner, repo, ref, ok, tc.owner, tc.repo, tc.ref, tc.wantOK)
		}
	}
}

func TestRegistryIDIgnoresVersion(t *testing.T) {
	m := RegistryModule{ID: "acme/mod/aws/1.2.3", Namespace: "acme", Name: "mod", Provider: "aws", Version: "1.2.3"}
	if got := registryID(m); got != "acme/mod/aws" {
		t.Errorf("registryID = %q", got)
	}
}
