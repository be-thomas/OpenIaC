// Package tfversion maps the Terraform language features found in an IR graph
// to the oldest Terraform release that accepts the configuration.
//
// It is a pure package: it depends only on pkg/ir and golang.org/x/mod/semver,
// and it never fails the parse. A configuration whose required_version cannot
// be understood still parses; only MinVersion falls back to the detected
// feature set.
package tfversion

import (
	"strings"

	"golang.org/x/mod/semver"
)

const (
	// Baseline is the oldest Terraform release OpenIaC targets. A configuration
	// that uses only Terraform 1.0 constructs reports this version even if it
	// would also parse under 0.12.
	Baseline = "1.0.0"

	// Latest is the newest Terraform release OpenIaC understands. Features
	// introduced by later releases are preserved generically and do not move
	// MinVersion past this value.
	Latest = "1.17.0"
)

// Canonical returns the canonical "MAJOR.MINOR.PATCH" form of a Terraform
// version string. It accepts one, two or three numeric components with the
// optional leading "v" that semver requires. Pre-release and build forms are
// rejected: they never appear in Terraform's own version numbers.
func Canonical(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", false
	}
	v := s
	if !strings.HasPrefix(v, "v") {
		v = "v" + s
	}
	// Reject pre-release and build metadata before semver normalizes them away.
	if strings.ContainsAny(v, "-+") {
		return "", false
	}
	if !semver.IsValid(v) {
		return "", false
	}
	c := semver.Canonical(v)
	if c == "" {
		return "", false
	}
	return strings.TrimPrefix(c, "v"), true
}

// Compare returns -1, 0 or 1 as a is older than, equal to or newer than b.
// Unparseable versions compare as older than any parseable one.
func Compare(a, b string) int {
	ca, oka := Canonical(a)
	cb, okb := Canonical(b)
	switch {
	case !oka && !okb:
		return 0
	case !oka:
		return -1
	case !okb:
		return 1
	}
	return semver.Compare("v"+ca, "v"+cb)
}

// maxVersion returns the newer of two canonical versions.
func maxVersion(a, b string) string {
	if Compare(a, b) >= 0 {
		return a
	}
	return b
}

// bumpPatch returns the smallest version strictly greater than a canonical
// version, which is what a strict ">" constraint lower-bounds.
func bumpPatch(canonical string) (string, bool) {
	c, ok := Canonical(canonical)
	if !ok {
		return "", false
	}
	parts := strings.Split(c, ".")
	if len(parts) != 3 {
		return "", false
	}
	patch, ok := parseUint(parts[2])
	if !ok {
		return "", false
	}
	return parts[0] + "." + parts[1] + "." + itoa(patch+1), true
}

func parseUint(s string) (int, bool) {
	if s == "" {
		return 0, false
	}
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, false
		}
		n = n*10 + int(r-'0')
	}
	return n, true
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
