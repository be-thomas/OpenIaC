package collect

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// Fixture is the metadata committed next to each input tree.
type Fixture struct {
	Schema    int      `json:"schema"`
	ID        string   `json:"id"`
	Tier      string   `json:"tier"`
	Origin    string   `json:"origin"`
	Source    Source   `json:"source"`
	License   string   `json:"license"`
	Providers []string `json:"providers"`
	Features  []string `json:"features"`
	Targets   []string `json:"targets"`
	Support   string   `json:"support"`
}

// Source records provenance. Ref is a registry version or an immutable commit.
type Source struct {
	URL           string `json:"url"`
	Ref           string `json:"ref"`
	Path          string `json:"path"`
	ArchiveSHA256 string `json:"archiveSha256"`
	InputSHA256   string `json:"inputSha256"`
}

// BuildFixture assembles metadata for one candidate.
func BuildFixture(m GitModule, cand *Candidate, archiveSHA, license string) Fixture {
	origin := "github"
	if m.Registry != nil {
		origin = "registry"
	}
	sub := cand.Dir
	if sub == "" {
		sub = "."
	}
	return Fixture{
		Schema: 1,
		ID:     fixtureID(m, cand.Dir),
		Tier:   cand.Tier,
		Origin: origin,
		Source: Source{
			URL:           m.SourceURL(),
			Ref:           m.RefLabel(),
			Path:          sub,
			ArchiveSHA256: archiveSHA,
			InputSHA256:   InputHash(cand.Files),
		},
		License:   license,
		Providers: cand.Providers,
		Features:  cand.Features,
		Targets:   []string{"cdk", "pulumi"},
		Support:   cand.Support,
	}
}

func fixtureID(m GitModule, dir string) string {
	if dir == "" {
		return m.Prefix()
	}
	return path.Join(m.Prefix(), dir)
}

// InputHash is a deterministic digest of the input tree.
func InputHash(files map[string][]byte) string {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	h := sha256.New()
	for _, name := range names {
		sum := sha256.Sum256(files[name])
		h.Write([]byte(name))
		h.Write([]byte{0})
		h.Write([]byte(hex.EncodeToString(sum[:])))
		h.Write([]byte{'\n'})
	}
	return hex.EncodeToString(h.Sum(nil))
}

// WriteFixture writes input/, LICENSE, NOTICE and fixture.json. It only
// replaces the input tree, so a curated expected/ or NOTES.md is preserved.
func WriteFixture(outDir string, f Fixture, cand *Candidate, license, notice []byte) (int64, error) {
	if err := safeRelative(f.ID); err != nil {
		return 0, err
	}
	dir := filepath.Join(outDir, filepath.FromSlash(f.ID))
	inputDir := filepath.Join(dir, "input")
	if err := os.RemoveAll(inputDir); err != nil {
		return 0, err
	}
	if err := os.MkdirAll(inputDir, 0o755); err != nil {
		return 0, err
	}
	var written int64
	names := make([]string, 0, len(cand.Files))
	for name := range cand.Files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if err := safeRelative(name); err != nil {
			return written, fmt.Errorf("input %s: %w", name, err)
		}
		dest := filepath.Join(inputDir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return written, err
		}
		if err := os.WriteFile(dest, cand.Files[name], 0o644); err != nil {
			return written, err
		}
		written += int64(len(cand.Files[name]))
	}
	if len(license) > 0 {
		dest := filepath.Join(dir, "LICENSE")
		if err := os.WriteFile(dest, license, 0o644); err != nil {
			return written, err
		}
		written += int64(len(license))
	}
	// NOTICE is generated from upstream, so a stale copy must not survive a
	// refresh that no longer has one.
	noticeDest := filepath.Join(dir, "NOTICE")
	if err := os.Remove(noticeDest); err != nil && !os.IsNotExist(err) {
		return written, err
	}
	if len(notice) > 0 {
		if err := os.WriteFile(noticeDest, notice, 0o644); err != nil {
			return written, err
		}
		written += int64(len(notice))
	}
	meta, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return written, err
	}
	meta = append(meta, '\n')
	if err := os.WriteFile(filepath.Join(dir, "fixture.json"), meta, 0o644); err != nil {
		return written, err
	}
	written += int64(len(meta))
	return written, nil
}

func safeRelative(name string) error {
	if name == "" || name == "." {
		return fmt.Errorf("empty path")
	}
	if path.IsAbs(name) {
		return fmt.Errorf("absolute path")
	}
	clean := path.Clean(name)
	if clean != name {
		return fmt.Errorf("path is not clean: %q", name)
	}
	for _, seg := range strings.Split(clean, "/") {
		if seg == "" || seg == "." || seg == ".." {
			return fmt.Errorf("unsafe path segment %q", seg)
		}
	}
	return nil
}
