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
	lock := Lock{
		Schema:   1,
		Tiers:    make(map[string]int),
		Licenses: make(map[string]int),
		Origins:  make(map[string]int),
	}
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
		return lock, err
	}
	sort.Strings(paths)
	for _, rel := range paths {
		full := filepath.Join(fixturesDir, filepath.FromSlash(rel))
		data, err := os.ReadFile(full)
		if err != nil {
			return lock, err
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
			return lock, fmt.Errorf("parse %s: %w", rel, err)
		}
		lock.FixtureCount++
		lock.Tiers[f.Tier]++
		lock.Licenses[f.License]++
		lock.Origins[f.Origin]++
	}
	lock.TreeSHA256 = hex.EncodeToString(h.Sum(nil))
	data, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return lock, err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return lock, err
	}
	return lock, os.WriteFile(lockPath, data, 0o644)
}

// WriteThirdParty generates the attribution table for vendored fixtures.
func WriteThirdParty(fixturesDir, outPath string) error {
	var fixtures []Fixture
	err := filepath.WalkDir(fixturesDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Base(p) != "fixture.json" {
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		var f Fixture
		if err := json.Unmarshal(data, &f); err != nil {
			return fmt.Errorf("parse %s: %w", p, err)
		}
		fixtures = append(fixtures, f)
		return nil
	})
	if err != nil {
		return err
	}
	sort.Slice(fixtures, func(i, j int) bool { return fixtures[i].ID < fixtures[j].ID })

	var b strings.Builder
	b.WriteString("# Third-party notices\n\n")
	b.WriteString("OpenIaC's own source is licensed under GPL-3.0. The Terraform inputs under\n")
	b.WriteString("`conformance/fixtures/` are vendored from public upstreams and remain under\n")
	b.WriteString("their original licenses. Each fixture directory contains the upstream\n")
	b.WriteString("license text as `LICENSE`, and the exact revision is recorded in its\n")
	b.WriteString("`fixture.json`.\n\n")
	if len(fixtures) == 0 {
		b.WriteString("_No fixtures are present._\n")
	} else {
		b.WriteString("| Fixture | License | Upstream |\n")
		b.WriteString("|:--|:--|:--|\n")
		for _, f := range fixtures {
			source := escapeTable(f.Source.URL)
			if f.Source.Ref != "" {
				source += " @ `" + escapeTable(f.Source.Ref) + "`"
			}
			if f.Source.Path != "" && f.Source.Path != "." {
				source += " (`" + escapeTable(f.Source.Path) + "`)"
			}
			fmt.Fprintf(&b, "| `%s` | %s | %s |\n", escapeTable(f.ID), f.License, source)
		}
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(outPath, []byte(b.String()), 0o644)
}

func escapeTable(s string) string {
	return strings.ReplaceAll(s, "|", "\\|")
}
