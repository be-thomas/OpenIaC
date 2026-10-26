package hclparser_test

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/be-thomas/openiac/pkg/ir"
	hclparser "github.com/be-thomas/openiac/pkg/parsers/hcl"
)

// TestParseConformanceCorpus parses every collected fixture module. It is opt-in
// because it walks thousands of files:
//
//	OPENIAC_CORPUS=1 go test ./pkg/parsers/hcl -run TestParseConformanceCorpus -count=1
//
// The corpus is real-world configuration, so it is also the cheapest guard
// against a regression that only shows up outside the hand-written fixtures.
func TestParseConformanceCorpus(t *testing.T) {
	if os.Getenv("OPENIAC_CORPUS") == "" {
		t.Skip("set OPENIAC_CORPUS=1 to parse the conformance corpus")
	}

	root := filepath.Join("..", "..", "..", "conformance", "fixtures")
	var inputs []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && entry.Name() == "input" {
			inputs = append(inputs, path)
			return fs.SkipDir
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	sort.Strings(inputs)
	if len(inputs) == 0 {
		t.Fatalf("no fixture inputs found under %s", root)
	}

	var failures []string
	for _, dir := range inputs {
		graph, err := hclparser.ParseDir(dir)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", dir, err))
			continue
		}
		if graph.Version != ir.Version {
			failures = append(failures, fmt.Sprintf("%s: version %q, want %q", dir, graph.Version, ir.Version))
		}
		if graph.MinVersion == "" {
			failures = append(failures, fmt.Sprintf("%s: min_version is empty", dir))
		}
	}

	if len(failures) > 0 {
		t.Errorf("%d of %d fixture modules failed to parse:", len(failures), len(inputs))
		for _, failure := range failures {
			t.Error(failure)
		}
	}
	t.Logf("parsed %d fixture modules", len(inputs))
}
