package engine_test

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/be-thomas/openiac/pkg/emitters"
	"github.com/be-thomas/openiac/pkg/engine"
)

// update rewrites the golden files instead of comparing against them:
//
//	go test ./pkg/engine -run TestGoldenAwsWebApp -update
var update = flag.Bool("update", false, "rewrite golden files")

// TestGoldenAwsWebApp snapshots the documented example in both targets. It is
// the regression guard for emitter output: any change to generated TypeScript
// has to be visible here, including the warnings a target reports.
func TestGoldenAwsWebApp(t *testing.T) {
	// The generated "// Source:" comment carries the path the parser was given,
	// so the test runs from the repository root to keep the snapshot portable.
	t.Chdir("../..")

	graph, err := engine.Parse(filepath.Join("examples", "aws-web-app"))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	results, err := engine.Convert(graph, []engine.Target{engine.TargetCDK, engine.TargetPulumi}, emitters.Options{
		StackName:   "OpenIacStack",
		ProjectName: "openiac",
	})
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("results = %d, want 2", len(results))
	}

	for _, result := range results {
		path := filepath.Join("pkg", "engine", "testdata", "golden", "aws-web-app."+string(result.Target)+".txt")
		golden(t, path, renderResult(result))
	}
}

// renderResult turns a target result into the text that is snapshotted: each
// generated file under a header, then every warning in order.
func renderResult(result engine.TargetResult) string {
	var b strings.Builder
	for i, file := range result.Files {
		if i > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "// ==== %s ====\n", file.Path)
		b.WriteString(file.Content)
		if !strings.HasSuffix(file.Content, "\n") {
			b.WriteString("\n")
		}
	}
	b.WriteString("\n// ==== warnings ====\n")
	for _, warning := range result.Warnings {
		fmt.Fprintf(&b, "// %s\n", warning)
	}
	return b.String()
}

func golden(t *testing.T, path, got string) {
	t.Helper()
	if *update {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("create golden directory: %v", err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
		return
	}

	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v (run with -update to create it)", path, err)
	}
	if string(want) != got {
		t.Errorf("%s is stale; run `go test ./pkg/engine -run TestGoldenAwsWebApp -update`\n--- want\n%s\n--- got\n%s",
			path, want, got)
	}
}
