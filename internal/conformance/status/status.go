// Package status summarises how the parser handles the committed conformance
// corpus.
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
	"strconv"
	"strings"

	hclparser "github.com/be-thomas/openiac/pkg/parsers/hcl"
)

// Schema is the report format version.
const Schema = 1

// Report is the committed snapshot of corpus size and parse results.
type Report struct {
	Schema      int            `json:"schema"`
	Corpus      Corpus         `json:"corpus"`
	Parse       Parse          `json:"parse"`
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

// Compute parses every fixture input under fixturesDir. It reads the lock for
// the aggregate sizes but never writes anything.
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
	for _, dir := range inputs {
		graph, err := hclparser.ParseDir(dir)
		if err != nil {
			report.Parse.Failed++
			if len(report.Parse.Failures) < maxFailures {
				report.Parse.Failures = append(report.Parse.Failures, fmt.Sprintf("%s: %v", filepath.ToSlash(trim(fixturesDir, dir)), err))
			}
			continue
		}
		report.Parse.Parsed++
		report.MinVersions[graph.MinVersion]++
	}
	return report, nil
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
	if r.Parse.Failed > 0 {
		badge.Color = "red"
	}
	return badge
}

// VersionRows returns the minimum Terraform versions present, oldest first,
// with the fixture count for each.
func (r Report) VersionRows() []VersionCount {
	return sortedRows(r.MinVersions)
}

// MinorRows aggregates the distribution by MAJOR.MINOR, which is the
// resolution that reads well in a chart: patch releases come from individual
// modules' required_version pins, not from language differences.
func (r Report) MinorRows() []VersionCount {
	counts := make(map[string]int, len(r.MinVersions))
	for version, count := range r.MinVersions {
		counts[minor(version)] += count
	}
	return sortedRows(counts)
}

func sortedRows(counts map[string]int) []VersionCount {
	rows := make([]VersionCount, 0, len(counts))
	for version, count := range counts {
		rows = append(rows, VersionCount{Version: version, Fixtures: count})
	}
	sort.Slice(rows, func(i, j int) bool { return lessVersion(rows[i].Version, rows[j].Version) })
	return rows
}

// minor reduces "1.10.2" to "1.10".
func minor(version string) string {
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return version
	}
	return parts[0] + "." + parts[1]
}

// VersionCount is one row of the minimum-version distribution.
type VersionCount struct {
	Version  string
	Fixtures int
}

// lessVersion orders dotted version strings numerically, then as text so the
// order is total even for unexpected values.
func lessVersion(a, b string) bool {
	as := strings.Split(a, ".")
	bs := strings.Split(b, ".")
	for i := 0; i < len(as) && i < len(bs); i++ {
		an, aerr := strconv.Atoi(as[i])
		bn, berr := strconv.Atoi(bs[i])
		if aerr == nil && berr == nil {
			if an != bn {
				return an < bn
			}
			continue
		}
		if as[i] != bs[i] {
			return as[i] < bs[i]
		}
	}
	return len(as) < len(bs)
}

// Markdown renders the distribution as a table with inline bars, for the README
// and for `openiac-conformance status`.
func (r Report) Markdown() string {
	var b strings.Builder
	b.WriteString("| Minimum Terraform | Fixtures | Share |\n")
	b.WriteString("|:--|--:|:--|\n")
	total := r.Parse.Parsed
	for _, row := range r.MinorRows() {
		share := 0
		if total > 0 {
			share = row.Fixtures * 100 / total
		}
		bar := strings.Repeat("█", share/2)
		if bar == "" && row.Fixtures > 0 {
			bar = "▏"
		}
		fmt.Fprintf(&b, "| `%s` | %d | %s %d%% |\n", row.Version, row.Fixtures, bar, share)
	}
	return b.String()
}

// Mermaid renders the distribution as a GitHub-renderable bar chart.
func (r Report) Mermaid() string {
	rows := r.MinorRows()
	labels := make([]string, 0, len(rows))
	bars := make([]string, 0, len(rows))
	max := 0
	for _, row := range rows {
		labels = append(labels, fmt.Sprintf("%q", row.Version))
		bars = append(bars, fmt.Sprintf("%d", row.Fixtures))
		if row.Fixtures > max {
			max = row.Fixtures
		}
	}
	var b strings.Builder
	b.WriteString("```mermaid\nxychart-beta\n")
	b.WriteString("    title \"Fixtures by minimum Terraform version\"\n")
	fmt.Fprintf(&b, "    x-axis [%s]\n", strings.Join(labels, ", "))
	fmt.Fprintf(&b, "    y-axis \"Fixtures\" 0 --> %d\n", axisMax(max))
	fmt.Fprintf(&b, "    bar [%s]\n", strings.Join(bars, ", "))
	b.WriteString("```\n")
	return b.String()
}

// axisMax rounds the chart's y-axis up to a readable bound.
func axisMax(max int) int {
	if max <= 0 {
		return 10
	}
	step := 10
	for step*10 < max {
		step *= 10
	}
	return (max/step + 1) * step
}

// Section markers delimit the generated block in README.md.
const (
	SectionStart = "<!-- conformance:summary:start -->"
	SectionEnd   = "<!-- conformance:summary:end -->"
)

// Section renders the generated README block, markers included and without a
// trailing newline: the document keeps the one after the closing marker.
func (r Report) Section() string {
	return SectionStart + "\n" + r.Mermaid() + SectionEnd
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
