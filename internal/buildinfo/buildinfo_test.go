package buildinfo_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/be-thomas/openiac/internal/buildinfo"
)

func TestVersion(t *testing.T) {
	if buildinfo.Version != strings.TrimSpace(buildinfo.Version) {
		t.Fatalf("Version = %q, want no surrounding whitespace", buildinfo.Version)
	}
	// MAJOR.MINOR.PATCH, optionally followed by a pre-release suffix.
	pattern := regexp.MustCompile(`^\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?$`)
	if !pattern.MatchString(buildinfo.Version) {
		t.Fatalf("Version = %q, want MAJOR.MINOR.PATCH", buildinfo.Version)
	}
}
