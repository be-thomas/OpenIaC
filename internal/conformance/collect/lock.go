package collect

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Lock is the deterministic aggregate committed at conformance/corpus.lock.json.
type Lock struct {
	Schema       int            `json:"schema"`
	FixtureCount int            `json:"fixtureCount"`
	FileCount    int            `json:"fileCount"`
	TotalBytes   int64          `json:"totalBytes"`
	TreeSHA256   string         `json:"treeSha256"`
	Tiers        map[string]int `json:"tiers"`
	Licenses     map[string]int `json:"licenses"`
	Origins      map[string]int `json:"origins"`
}

// WriteLock hashes the fixture tree and writes the aggregate lock.
func WriteLock(fixturesDir, lockPath string) (Lock, error) {
	lock, _, err := computeLock(fixturesDir)
	if err != nil {
		return lock, err
	}
	return lock, writeLockFile(lockPath, lock)
}

func writeLockFile(lockPath string, lock Lock) error {
	data, err := encodeLock(lock)
	if err != nil {
		return err
	}
	return writeFileAtomic(lockPath, data)
}

func encodeLock(lock Lock) ([]byte, error) {
	data, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

// writeFileAtomic replaces path through a sibling temporary file, so readers
// never observe a half-written lock or attribution document.
func writeFileAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// RefreshAggregates recomputes the lock and attribution from the fixture tree
// without touching it or the network. It returns an error when the tree is
// over the configured byte budget, before either file is written. Each file is
// replaced atomically; verify notices a run that failed between the two writes.
func RefreshAggregates(fixturesDir, lockPath, thirdPartyPath string, cfg Config) (Lock, error) {
	// Library callers that only care about the tree should not have to assemble
	// the limit set; a zero budget would reject every corpus.
	if cfg.Limits.MaxTotalBytes <= 0 {
		cfg.Limits = Defaults().Limits
	}
	lock, fixtures, err := computeLock(fixturesDir)
	if err != nil {
		return lock, err
	}
	if lock.TotalBytes > cfg.Limits.MaxTotalBytes {
		return lock, fmt.Errorf("fixture tree is %d bytes, over the %d byte budget", lock.TotalBytes, cfg.Limits.MaxTotalBytes)
	}
	lockData, err := encodeLock(lock)
	if err != nil {
		return lock, err
	}
	if err := writeFileAtomic(thirdPartyPath, RenderThirdParty(fixtures)); err != nil {
		return lock, err
	}
	return lock, writeFileAtomic(lockPath, lockData)
}

// computeLock walks the tree, refuses symlinks, and returns the aggregate plus
// every fixture it parsed.
func computeLock(fixturesDir string) (Lock, []Fixture, error) {
	lock := Lock{
		Schema:   1,
		Tiers:    make(map[string]int),
		Licenses: make(map[string]int),
		Origins:  make(map[string]int),
	}
	var fixtures []Fixture
	h := sha256.New()
	var paths []string
	err := filepath.WalkDir(fixturesDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if d.Type()&fs.ModeSymlink != 0 {
			return fmt.Errorf("symlink in corpus: %s", p)
		}
		rel, err := filepath.Rel(fixturesDir, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		paths = append(paths, rel)
		return nil
	})
	if err != nil {
		return lock, nil, err
	}
	sort.Strings(paths)
	for _, rel := range paths {
		full := filepath.Join(fixturesDir, filepath.FromSlash(rel))
		data, err := os.ReadFile(full)
		if err != nil {
			return lock, nil, err
		}
		sum := sha256.Sum256(data)
		h.Write([]byte(rel))
		h.Write([]byte{0})
		h.Write([]byte(hex.EncodeToString(sum[:])))
		h.Write([]byte{'\n'})
		lock.FileCount++
		lock.TotalBytes += int64(len(data))

		if filepath.Base(rel) != "fixture.json" {
			continue
		}
		var f Fixture
		if err := json.Unmarshal(data, &f); err != nil {
			return lock, nil, fmt.Errorf("parse %s: %w", rel, err)
		}
		fixtures = append(fixtures, f)
		lock.FixtureCount++
		lock.Tiers[f.Tier]++
		lock.Licenses[f.License]++
		lock.Origins[f.Origin]++
	}
	lock.TreeSHA256 = hex.EncodeToString(h.Sum(nil))
	sort.Slice(fixtures, func(i, j int) bool { return fixtures[i].ID < fixtures[j].ID })
	return lock, fixtures, nil
}

// WriteThirdParty generates the attribution table for vendored fixtures.
func WriteThirdParty(fixturesDir, outPath string) error {
	_, fixtures, err := computeLock(fixturesDir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(outPath, RenderThirdParty(fixtures), 0o644)
}

// RenderThirdParty returns the attribution document for a fixture set.
func RenderThirdParty(fixtures []Fixture) []byte {
	var b strings.Builder
	b.WriteString("# Third-party notices\n\n")
	b.WriteString("OpenIaC's own source is licensed under GPL-3.0. The Terraform inputs under\n")
	b.WriteString("`conformance/fixtures/` are vendored from public upstreams and remain under\n")
	b.WriteString("their original licenses. Each fixture directory contains the upstream\n")
	b.WriteString("license text as `LICENSE`, and the exact revision is recorded in its\n")
	b.WriteString("`fixture.json`.\n\n")
	if len(fixtures) == 0 {
		b.WriteString("_No fixtures are present._\n")
		return []byte(b.String())
	}
	b.WriteString("| Fixture | License | Upstream |\n")
	b.WriteString("|:--|:--|:--|\n")
	for _, f := range fixtures {
		source := escapeTable(f.Source.URL)
		if f.Source.Ref != "" {
			source += " @ `" + escapeTable(f.Source.Ref) + "`"
		}
		if f.Source.Commit != "" {
			source += " (`" + escapeTable(shortSHA(f.Source.Commit)) + "`)"
		}
		if f.Source.Path != "" && f.Source.Path != "." {
			source += " (`" + escapeTable(f.Source.Path) + "`)"
		}
		fmt.Fprintf(&b, "| `%s` | %s | %s |\n", escapeTable(f.ID), f.License, source)
	}
	return []byte(b.String())
}

func shortSHA(sha string) string {
	if len(sha) > 8 {
		return sha[:8]
	}
	return sha
}

func escapeTable(s string) string {
	return strings.ReplaceAll(s, "|", "\\|")
}
