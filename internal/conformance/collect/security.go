package collect

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// PolicyViolation explains why a candidate cannot be committed.
type PolicyViolation struct {
	Kind   string
	Path   string
	Detail string
}

func (v PolicyViolation) String() string {
	return fmt.Sprintf("%s in %s: %s", v.Kind, v.Path, v.Detail)
}

var (
	executableRE = []*regexp.Regexp{
		regexp.MustCompile(`(?m)^\s*provisioner\s+"`),
		regexp.MustCompile(`(?m)^\s*data\s+"external"`),
		regexp.MustCompile(`(?i)\b(?:local-exec|remote-exec)\b`),
	}
	secretRE = []*regexp.Regexp{
		regexp.MustCompile(`\b(?:AKIA|ASIA)[0-9A-Z]{16}\b`),
		regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----`),
		regexp.MustCompile(`(?i)\b(?:aws_access_key_id|aws_secret_access_key|secret_access_key|access_key|password|passwd|token|secret)\s*=\s*"([^"\n]{8,})"`),
	}
)

// ScanFiles rejects executable Terraform and obvious credential material.
func ScanFiles(files map[string][]byte) []PolicyViolation {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)

	var out []PolicyViolation
	for _, name := range names {
		content := files[name]
		for _, re := range executableRE {
			if loc := re.FindIndex(content); loc != nil {
				out = append(out, PolicyViolation{
					Kind:   "executable",
					Path:   name,
					Detail: "provisioners and external data sources are not allowed",
				})
				break
			}
		}
		for _, re := range secretRE {
			for _, m := range re.FindAllSubmatch(content, -1) {
				if len(m) > 1 && looksLikePlaceholder(string(m[1])) {
					continue
				}
				if strings.Contains(strings.ToLower(string(m[0])), "example") {
					continue
				}
				out = append(out, PolicyViolation{
					Kind:   "secret",
					Path:   name,
					Detail: "credential-like literal",
				})
				break
			}
			if len(out) > 0 && out[len(out)-1].Path == name && out[len(out)-1].Kind == "secret" {
				break
			}
		}
	}
	return out
}

func looksLikePlaceholder(s string) bool {
	s = strings.ToLower(s)
	for _, marker := range []string{
		"example", "fake", "dummy", "test", "placeholder", "changeme",
		"change-me", "redacted", "xxxx", "****", "your-", "your_",
		"<", ">", "${", "%{", "var.", "local.", "data.", "module.",
		"aws_", "arn:", "none", "null",
	} {
		if strings.Contains(s, marker) {
			return true
		}
	}
	return false
}
