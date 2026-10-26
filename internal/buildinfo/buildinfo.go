// Package buildinfo reports the OpenIaC release version.
package buildinfo

import (
	_ "embed"
	"strings"
)

// raw holds the VERSION file that sits next to this package. Keeping the
// version in a file rather than in a constant means the CLI, the Makefile and
// a plain `go build` all read the same value.
//
//go:embed VERSION
var raw string

// Version is the release version, for example "0.1.0". A development build
// reports a "-dev" suffix.
var Version = strings.TrimSpace(raw)
