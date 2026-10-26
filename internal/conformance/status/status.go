// Package status summarises how the parser and emitters handle the committed
// conformance corpus.
//
// The report is committed as conformance/status.json and rendered in the
// README, so a reader can see what the corpus actually proves without running
// the corpus test. It stays small so the playground can serve it unchanged.
package status

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/be-thomas/openiac/pkg/emitters"
	"github.com/be-thomas/openiac/pkg/engine"
	hclparser "github.com/be-thomas/openiac/pkg/parsers/hcl"
)

// Schema is the report format version. Version 2 added the emit counts.
const Schema = 2

// Report is the committed snapshot of corpus size, parse results and what each
// emitter produced.
type Report struct {
	Schema      int            `json:"schema"`
	Corpus      Corpus         `json:"corpus"`
	Parse       Parse          `json:"parse"`
	Emit        Emit           `json:"emit"`
	MinVersions map[string]int `json:"min_versions"`
}

// Corpus repeats the lock's aggregate sizes so a reader needs one file.
type Corpus struct {
	Fixtures   int    `json:"fixtures"`
	Files      int    `json:"files"`
	Bytes      int64  `json:"bytes"`
	TreeSHA256 string `json:"tree_sha256"`
}

// Parse counts fixture modules and how many of them parse cleanly.
type Parse struct {
	Modules  int      `json:"modules"`
	Parsed   int      `json:"parsed"`
	Failed   int      `json:"failed"`
	Failures []string `json:"failures,omitempty"`
}

// Emit counts what each emitter produced for the modules that parsed.
type Emit struct {
	CDK    EmitTarget `json:"cdk"`
	Pulumi EmitTarget `json:"pulumi"`
}

// EmitTarget is one emitter's result over the corpus. Emitted and Failed
// partition Modules, and Clean is a subset of Emitted: a module can compile but
// still need a human, which is what a warning means.
type EmitTarget struct {
	Modules  int      `json:"modules"`
	Emitted  int      `json:"emitted"`
	Failed   int      `json:"failed"`
	Clean    int      `json:"clean"`
	Warnings int      `json:"warnings"`
	Failures []string `json:"failures,omitempty"`
}

// Badge is the shields.io endpoint document rendered in the README.
type Badge struct {
	SchemaVersion int    `json:"schemaVersion"`
	Label         string `json:"label"`
	Message       string `json:"message"`
	Color         string `json:"color"`
}

// maxFailures caps the messages kept in the report so a broad regression still
// produces a committable file.
const maxFailures = 20

// lock mirrors the fields of conformance/corpus.lock.json that the report
// repeats. It is read here rather than imported so this package stays free of
// the collector's dependencies.
type lock struct {
	FixtureCount int    `json:"fixtureCount"`
	FileCount    int    `json:"fileCount"`
	TotalBytes   int64  `json:"totalBytes"`
	TreeSHA256   string `json:"treeSha256"`
}

// Compute parses every fixture input under fixturesDir and runs both emitters
// over every graph that parsed. It reads the lock for the aggregate sizes but
// never writes anything.
func Compute(fixturesDir, lockPath string) (Report, error) {
	report := Report{Schema: Schema, MinVersions: map[string]int{}}

	var l lock
	raw, err := os.ReadFile(lockPath)
	if err != nil {
		return report, fmt.Errorf("read lock: %w", err)
	}
	if err := json.Unmarshal(raw, &l); err != nil {
		return report, fmt.Errorf("parse lock %s: %w", lockPath, err)
	}
	report.Corpus = Corpus{
		Fixtures:   l.FixtureCount,
		Files:      l.FileCount,
		Bytes:      l.TotalBytes,
		TreeSHA256: l.TreeSHA256,
	}

	inputs, err := fixtureInputs(fixturesDir)
	if err != nil {
		return report, err
	}
	report.Parse.Modules = len(inputs)

	targets := []struct {
		target engine.Target
		stats  *EmitTarget
	}{
		{engine.TargetCDK, &report.Emit.CDK},
		{engine.TargetPulumi, &report.Emit.Pulumi},
	}

	for _, dir := range inputs {
		graph, err := hclparser.ParseDir(dir)
		if err != nil {
			report.Parse.Failed++
			addFailure(&report.Parse.Failures, fixturesDir, dir, err)
			continue
		}
		report.Parse.Parsed++
		report.MinVersions[graph.MinVersion]++

		for _, t := range targets {
			t.stats.Modules++
			results, err := engine.Convert(graph, []engine.Target{t.target}, emitters.Options{})
			if err == nil && len(results) == 0 {
				err = fmt.Errorf("%s emitter returned no result", t.target)
			}
			if err == nil && !hasContent(results[0].Files) {
				err = fmt.Errorf("%s emitter produced no files", t.target)
			}
			if err != nil {
				t.stats.Failed++
				addFailure(&t.stats.Failures, fixturesDir, dir, err)
				continue
			}
			t.stats.Emitted++
			warnings := results[0].Warnings
			t.stats.Warnings += len(warnings)
			if len(warnings) == 0 {
				t.stats.Clean++
			}
		}
	}
	return report, nil
}

// addFailure records a capped failure message with the fixture path relative to
// the fixtures root. The error itself usually quotes the paths it was given, so
// they are trimmed too: an absolute -out would otherwise put a checkout-specific
// path in the committed report and make -check depend on where the tree lives.
func addFailure(failures *[]string, root, dir string, err error) {
	if len(*failures) >= maxFailures {
		return
	}
	*failures = append(*failures, fmt.Sprintf("%s: %s", filepath.ToSlash(trim(root, dir)), relativize(err.Error(), root)))
}

// relativize drops the fixtures root from a message. Both separator forms are
// handled so the report stays portable across platforms.
func relativize(message, root string) string {
	if root == "" {
		return message
	}
	for _, prefix := range []string{filepath.ToSlash(root) + "/", root + string(filepath.Separator)} {
		message = strings.ReplaceAll(message, prefix, "")
	}
	return message
}

// hasContent reports whether an emitter produced any non-empty file. An emitter
// that "succeeds" with nothing to show has not converted the module.
func hasContent(files []emitters.File) bool {
	for _, file := range files {
		if len(file.Content) > 0 {
			return true
		}
	}
	return false
}

// fixtureInputs lists every fixture module root in deterministic order.
func fixtureInputs(root string) ([]string, error) {
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
		return nil, fmt.Errorf("walk %s: %w", root, err)
	}
	sort.Strings(inputs)
	if len(inputs) == 0 {
		return nil, fmt.Errorf("no fixture inputs found under %s", root)
	}
	return inputs, nil
}

func trim(root, path string) string {
	if rel, err := filepath.Rel(root, path); err == nil {
		return rel
	}
	return path
}

// Badge renders the shields.io endpoint document.
func (r Report) Badge() Badge {
	badge := Badge{
		SchemaVersion: 1,
		Label:         "conformance",
		Message:       fmt.Sprintf("%d / %d modules", r.Parse.Parsed, r.Parse.Modules),
		Color:         "brightgreen",
	}
	if r.Parse.Failed > 0 || r.Emit.CDK.Failed > 0 || r.Emit.Pulumi.Failed > 0 {
		badge.Color = "red"
	}
	return badge
}

// Total is the denominator for every outcome: a module that fails before an
// emitter runs still counts against it.
func (r Report) Total() int {
	return r.Parse.Modules
}

// Outcome is one row of the generated summary.
type Outcome struct {
	Label string
	Count int
}

// Outcomes lists what the corpus proves in pipeline order: the module reached
// the IR, then each emitter, then whether that output still needed a human. The
// same module counts in several rows, so the rows are not a partition.
func (r Report) Outcomes() []Outcome {
	return []Outcome{
		{"Parsed to the universal IR", r.Parse.Parsed},
		{"IR → AWS CDK TypeScript", r.Emit.CDK.Emitted},
		{"IR → Pulumi TypeScript", r.Emit.Pulumi.Emitted},
		{"IR → AWS CDK, no warnings", r.Emit.CDK.Clean},
		{"IR → Pulumi, no warnings", r.Emit.Pulumi.Clean},
	}
}

// Markdown renders the outcomes as a table with inline bars, for the README and
// for `openiac-conformance status`.
func (r Report) Markdown() string {
	total := r.Total()
	var b strings.Builder
	b.WriteString("| Outcome | Modules | Share |\n")
	b.WriteString("|:--|--:|:--|\n")
	for _, row := range r.Outcomes() {
		fmt.Fprintf(&b, "| %s | %d | %s |\n", row.Label, row.Count, share(row.Count, total))
	}
	return b.String()
}

// share is the percentage with a bar whose length is the percentage: 20 cells
// for 100%, at least one for any non-zero count.
func share(count, total int) string {
	if count <= 0 || total <= 0 {
		return "0%"
	}
	pct := percent(count, total)
	cells := pct / 5
	if cells == 0 {
		return fmt.Sprintf("▏ %d%%", pct)
	}
	return fmt.Sprintf("%s %d%%", strings.Repeat("█", cells), pct)
}

// percent rounds to the nearest whole percentage so 27/853 reads as 3%, not 0%.
// A partial result is clamped to 1–99: rounding 852/853 to 100% would print a
// full bar beside a failed module, and 1/853 must not read as nothing at all.
func percent(count, total int) int {
	if count <= 0 || total <= 0 {
		return 0
	}
	if count >= total {
		return 100
	}
	pct := (count*100 + total/2) / total
	if pct < 1 {
		return 1
	}
	if pct > 99 {
		return 99
	}
	return pct
}

// headline states the result in one sentence above the table.
func (r Report) headline() string {
	total := r.Total()
	if total == 0 {
		return "_The corpus is empty._"
	}
	if r.Parse.Parsed == total && r.Emit.CDK.Emitted == total && r.Emit.Pulumi.Emitted == total {
		return fmt.Sprintf("**%d / %d modules — 100%% — reach the IR and emit both AWS CDK and Pulumi TypeScript.**", total, total)
	}
	return fmt.Sprintf("**%d / %d modules reach the IR, %d / %d reach AWS CDK and %d / %d reach Pulumi.**",
		r.Parse.Parsed, total, r.Emit.CDK.Emitted, total, r.Emit.Pulumi.Emitted, total)
}

// Section markers delimit the generated block in README.md.
const (
	SectionStart = "<!-- conformance:summary:start -->"
	SectionEnd   = "<!-- conformance:summary:end -->"
)

// Section renders the generated README block, markers included and without a
// trailing newline: the document keeps the one after the closing marker.
func (r Report) Section() string {
	return SectionStart + "\n" + r.headline() + "\n\n" + r.Markdown() + SectionEnd
}

// ExtractSection returns the marked block of document.
func ExtractSection(document string) (string, error) {
	start, end, err := sectionBounds(document)
	if err != nil {
		return "", err
	}
	return document[start:end], nil
}

// ReplaceSection returns document with its marked block replaced by section.
func ReplaceSection(document, section string) (string, error) {
	start, end, err := sectionBounds(document)
	if err != nil {
		return "", err
	}
	return document[:start] + section + document[end:], nil
}

// sectionBounds locates the marked block, markers included.
func sectionBounds(document string) (int, int, error) {
	start := strings.Index(document, SectionStart)
	if start < 0 {
		return 0, 0, fmt.Errorf("missing marker %s", SectionStart)
	}
	end := strings.Index(document[start:], SectionEnd)
	if end < 0 {
		return 0, 0, fmt.Errorf("missing marker %s after %s", SectionEnd, SectionStart)
	}
	return start, start + end + len(SectionEnd), nil
}
