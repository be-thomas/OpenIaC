package engine_test

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/be-thomas/openiac/pkg/emitters"
	"github.com/be-thomas/openiac/pkg/engine"
)

// update rewrites the golden files instead of comparing against them:
//
//	go test ./pkg/engine -run TestGolden -update
var update = flag.Bool("update", false, "rewrite golden files")

// goldenTargets is the pair of emitters every snapshot covers.
var goldenTargets = []engine.Target{engine.TargetCDK, engine.TargetPulumi}

// TestGoldenAwsWebApp snapshots the documented example in both targets. It is
// the regression guard for emitter output: any change to generated TypeScript
// has to be visible here.
func TestGoldenAwsWebApp(t *testing.T) {
	snapshot(t, "aws-web-app", filepath.Join("examples", "aws-web-app"), false)
}

// TestGoldenWarned snapshots a module the targets cannot fully convert, so the
// warning section of the snapshots is exercised. Without it a warning-free
// fixture would leave the section empty in every golden file and changes to
// warning collection would go unnoticed.
func TestGoldenWarned(t *testing.T) {
	snapshot(t, "warned", filepath.Join("pkg", "engine", "testdata", "fixtures", "warned"), true)
}

// snapshot converts a fixture in both targets and compares each result with
// its committed golden file. wantWarnings fails the test unless every target
// reports at least one warning, so a fixture added for that purpose cannot
// decay into a copy of the warning-free case.
func snapshot(t *testing.T, name, dir string, wantWarnings bool) {
	t.Helper()

	// The generated "// Source:" comment carries the path the parser was given,
	// so the test runs from the repository root to keep the snapshot portable.
	t.Chdir(repoRoot(t))

	graph, err := engine.Parse(dir)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	results, err := engine.Convert(graph, goldenTargets, emitters.Options{
		StackName:   "OpenIacStack",
		ProjectName: "openiac",
	})
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	if len(results) != len(goldenTargets) {
		t.Fatalf("results = %d, want %d", len(results), len(goldenTargets))
	}

	for _, result := range results {
		if wantWarnings && len(result.Warnings) == 0 {
			t.Errorf("%s: warnings = 0, want at least one", result.Target)
		}
		path := filepath.Join("pkg", "engine", "testdata", "golden", name+"."+string(result.Target)+".txt")
		checkGolden(t, path, renderResult(result))
	}
}

// repoRoot finds the module root, so the snapshots do not depend on the
// directory the test binary was started from: `go test` starts in pkg/engine,
// but a prebuilt binary or a debugger can start anywhere.
func repoRoot(t *testing.T) string {
	t.Helper()
	if _, file, _, ok := runtime.Caller(0); ok {
		if root, ok := moduleRoot(filepath.Dir(file)); ok {
			return root
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("working directory: %v", err)
	}
	if root, ok := moduleRoot(wd); ok {
		return root
	}
	t.Fatalf("cannot find go.mod above %s", wd)
	return ""
}

// moduleRoot walks up from dir until it finds a go.mod.
func moduleRoot(dir string) (string, bool) {
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
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

func checkGolden(t *testing.T, path, got string) {
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
		t.Errorf("%s is stale; run `go test ./pkg/engine -run TestGolden -update`\n--- want\n%s\n--- got\n%s",
			path, want, got)
	}
}
