package collect

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

// VerifyResult reports the verified corpus size.
type VerifyResult struct {
	Fixtures int
	Files    int
	Bytes    int64
}

// Verify checks that the fixture tree matches its committed lock, that every
// fixture's metadata agrees with its input, and that the attribution document
// is current. It never writes. cfg supplies the license allowlist used by the
// collector, so the two commands cannot disagree about what may be vendored.
func Verify(fixturesDir, lockPath, thirdPartyPath string, cfg Config) (VerifyResult, error) {
	var res VerifyResult
	// A caller that only cares about the tree should not have to assemble a
	// full config: omitted limits and the empty allowlist fall back to Defaults.
	if cfg.Limits.MaxFixtureFiles <= 0 {
		cfg.Limits = Defaults().Limits
	}
	if len(cfg.Licenses) == 0 {
		cfg.Licenses = Defaults().Licenses
	}
	var problems []string
	add := func(format string, args ...any) {
		problems = append(problems, fmt.Sprintf(format, args...))
	}

	lock, fixtures, err := computeLock(fixturesDir)
	if err != nil {
		return res, err
	}
	res.Fixtures = lock.FixtureCount
	res.Files = lock.FileCount
	res.Bytes = lock.TotalBytes

	raw, err := os.ReadFile(lockPath)
	if err != nil {
		add("lock: %v", err)
	} else {
		var want Lock
		if err := json.Unmarshal(raw, &want); err != nil {
			add("lock %s: %v", lockPath, err)
		} else if !reflect.DeepEqual(lock, want) {
			add("%s is stale: run the collector and commit the regenerated lock", lockPath)
		}
	}

	wantThirdParty := RenderThirdParty(fixtures)
	gotThirdParty, err := os.ReadFile(thirdPartyPath)
	if err != nil {
		add("attribution: %v", err)
	} else if string(gotThirdParty) != string(wantThirdParty) {
		add("%s is stale: run the collector and commit the regenerated attribution", thirdPartyPath)
	}

	files, err := corpusFiles(fixturesDir)
	if err != nil {
		return res, err
	}
	owner := ownerIndex(fixtures, files)
	fixtureDirs := make(map[string]bool, len(fixtures))
	for _, f := range fixtures {
		fixtureDirs[trimDot(f.ID)] = true
	}
	sizes := fixtureInputSizes(fixturesDir, files, owner)
	smoke := expectedSmoke(fixtures, sizes)

	seenID := make(map[string]bool, len(fixtures))
	seenHash := make(map[string]bool, len(fixtures))
	for _, f := range fixtures {
		dir := trimDot(f.ID)
		if seenID[f.ID] {
			add("duplicate fixture id %s", f.ID)
		}
		seenID[f.ID] = true
		var owned []string
		for _, rel := range files {
			if owner[rel] == dir {
				owned = append(owned, rel)
			}
		}
		if err := verifyFixture(fixturesDir, dir, f, owned, cfg, smoke[dir]); err != nil {
			add("%s: %v", dir, err)
		}
		if f.Source.InputSHA256 != "" {
			if seenHash[f.Source.InputSHA256] {
				add("%s: duplicate input hash %s", dir, f.Source.InputSHA256)
			}
			seenHash[f.Source.InputSHA256] = true
		}
	}
	for _, rel := range files {
		if !ownedByFixture(rel, fixtureDirs) {
			add("stray file %s", rel)
		}
	}

	if len(problems) > 0 {
		sort.Strings(problems)
		const max = 40
		if len(problems) > max {
			problems = append(problems[:max], fmt.Sprintf("... and %d more", len(problems)-max))
		}
		return res, errors.New("corpus verification failed:\n  " + strings.Join(problems, "\n  "))
	}
	return res, nil
}

func trimDot(p string) string {
	if p == "." {
		return ""
	}
	return p
}

func corpusFiles(fixturesDir string) ([]string, error) {
	var out []string
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
		out = append(out, filepath.ToSlash(rel))
		return nil
	})
	sort.Strings(out)
	return out, err
}

// ownerIndex assigns every file to the deepest fixture directory that contains
// it, so a fixture nested under another fixture's directory does not look like
// a stray file.
func ownerIndex(fixtures []Fixture, files []string) map[string]string {
	dirs := make([]string, 0, len(fixtures))
	for _, f := range fixtures {
		if d := trimDot(f.ID); d != "" {
			dirs = append(dirs, d)
		}
	}
	sort.Slice(dirs, func(i, j int) bool { return len(dirs[i]) > len(dirs[j]) })
	out := make(map[string]string, len(files))
	for _, rel := range files {
		for _, d := range dirs {
			if rel == d || strings.HasPrefix(rel, d+"/") {
				out[rel] = d
				break
			}
		}
	}
	return out
}

func ownedByFixture(rel string, dirs map[string]bool) bool {
	parts := strings.Split(rel, "/")
	for i := len(parts) - 1; i >= 1; i-- {
		if dirs[strings.Join(parts[:i], "/")] {
			return true
		}
	}
	return dirs[""] && len(parts) > 0
}

func verifyFixture(root, dir string, f Fixture, owned []string, cfg Config, wantSmoke bool) error {
	var problems []string
	add := func(format string, args ...any) {
		problems = append(problems, fmt.Sprintf(format, args...))
	}

	if f.Schema != 1 {
		add("unsupported schema %d", f.Schema)
	}
	if err := safeRelative(dir); err != nil {
		return fmt.Errorf("unsafe fixture directory: %w", err)
	}
	if f.Origin != "registry" && f.Origin != "github" {
		add("unknown origin %q", f.Origin)
	}
	if f.Source.URL == "" || f.Source.Ref == "" {
		add("source url and ref are required")
	}
	if !isHexSHA(f.Source.Commit) {
		add("source commit %q is not a git commit", f.Source.Commit)
	}
	if f.Source.GitOwner == "" || f.Source.GitRepo == "" {
		add("source git owner and repo are required")
	}
	if f.Source.Path == "" {
		add("source path is required")
	}
	if !isHexDigest(f.Source.ArchiveSHA256) {
		add("invalid archiveSha256")
	}
	if !isHexDigest(f.Source.InputSHA256) {
		add("invalid inputSha256")
	}
	if f.License == "" {
		add("license is required")
	} else if !cfg.LicenseAllowed(f.License) {
		add("license %q is not on the allowlist", f.License)
	}
	if err := checkProvenance(f); err != nil {
		add("%v", err)
	}
	switch f.Tier {
	case "smoke", "core", "advanced", "modules":
	default:
		add("unknown tier %q", f.Tier)
	}

	want := map[string]bool{
		"fixture.json": true,
		"LICENSE":      true,
		"NOTICE":       true,
		"NOTES.md":     true,
	}
	seenRequired := map[string]bool{}
	for _, rel := range owned {
		local := rel
		if dir != "" {
			local = strings.TrimPrefix(rel, dir+"/")
		}
		seenRequired[local] = true
		switch {
		case want[local]:
		case strings.HasPrefix(local, "input/") && isTerraformFile(local):
		case strings.HasPrefix(local, "expected/"):
		default:
			add("unexpected file %s", local)
		}
	}
	if !seenRequired["fixture.json"] {
		add("fixture.json is missing")
	}
	if !seenRequired["LICENSE"] {
		add("LICENSE is missing")
	}

	inputs := make(map[string][]byte)
	var inputBytes int64
	for _, rel := range owned {
		local := rel
		if dir != "" {
			local = strings.TrimPrefix(rel, dir+"/")
		}
		if !strings.HasPrefix(local, "input/") {
			if local == "LICENSE" {
				data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
				if err != nil {
					add("read LICENSE: %v", err)
					continue
				}
				if got := DetectLicense(data); got != f.License {
					add("LICENSE is %s, fixture.json says %s", displayLicense(got), f.License)
				}
			}
			continue
		}
		name := strings.TrimPrefix(local, "input/")
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			add("read %s: %v", local, err)
			continue
		}
		inputs[name] = data
		inputBytes += int64(len(data))
	}
	if len(inputs) == 0 {
		add("input tree is empty")
	}
	limits := cfg.Limits
	if len(inputs) > limits.MaxFixtureFiles {
		add("%d input files exceeds cap %d", len(inputs), limits.MaxFixtureFiles)
	}
	if inputBytes > limits.MaxFixtureBytes {
		add("%d input bytes exceeds cap %d", inputBytes, limits.MaxFixtureBytes)
	}
	if got := InputHash(inputs); got != f.Source.InputSHA256 {
		add("input hash is %s, fixture.json says %s", got, f.Source.InputSHA256)
	}

	analysis, err := AnalyzeFiles(inputs)
	if err != nil {
		add("parse: %v", err)
		return errors.New(strings.Join(problems, "; "))
	}
	for _, v := range analysis.Violations {
		add("policy: %s", v.String())
	}
	if !equalStrings(analysis.Features, f.Features) {
		add("features %v do not match parsed features %v", f.Features, analysis.Features)
	}
	if !equalStrings(analysis.Providers, f.Providers) {
		add("providers %v do not match parsed providers %v", f.Providers, analysis.Providers)
	}
	if got := SupportLevel(analysis.Features); got != f.Support {
		add("support is %s, fixture.json says %s", got, f.Support)
	}
	targets, err := targetsFor(analysis)
	if err != nil {
		add("targets: %v", err)
	} else if !equalStrings(targets, f.Targets) {
		add("targets %v do not match %v", f.Targets, targets)
	}
	if wantSmoke {
		if f.Tier != "smoke" {
			add("tier is %s, but this fixture is in the derived smoke set", f.Tier)
		}
	} else if f.Tier == "smoke" {
		add("tier is smoke, but this fixture is not in the derived smoke set")
	} else if got := ClassifyTier(analysis.Features); got != f.Tier {
		add("tier is %s, fixture.json says %s", got, f.Tier)
	}

	if len(problems) > 0 {
		return errors.New(strings.Join(problems, "; "))
	}
	return nil
}

func displayLicense(spdx string) string {
	if spdx == "" {
		return "unrecognized"
	}
	return spdx
}

// checkProvenance ties the metadata to its directory layout: the origin must
// match the ID prefix, the registry coordinates (or the git repository and
// short commit) must match source.url/source.commit, and the ID suffix must be
// the recorded source path.
func checkProvenance(f Fixture) error {
	id := trimDot(f.ID)
	var prefix string
	switch f.Origin {
	case "registry":
		const base = "https://registry.terraform.io/modules/"
		if !strings.HasPrefix(f.Source.URL, base) {
			return fmt.Errorf("registry fixture has non-registry source url %q", f.Source.URL)
		}
		parts := strings.Split(strings.TrimPrefix(f.Source.URL, base), "/")
		if len(parts) != 4 || parts[0] == "" || parts[1] == "" || parts[2] == "" || parts[3] == "" {
			return fmt.Errorf("registry source url %q is not modules/<namespace>/<name>/<provider>/<version>", f.Source.URL)
		}
		if f.Source.Ref != parts[3] {
			return fmt.Errorf("source ref %q does not match registry version %q", f.Source.Ref, parts[3])
		}
		prefix = "registry/" + safeSegment(parts[0]) + "/" + safeSegment(parts[1]) + "/" + safeSegment(parts[3])
	case "github":
		const base = "https://github.com/"
		if !strings.HasPrefix(f.Source.URL, base) {
			return fmt.Errorf("github fixture has non-github source url %q", f.Source.URL)
		}
		parts := strings.Split(strings.TrimPrefix(f.Source.URL, base), "/")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return fmt.Errorf("github source url %q is not github.com/<owner>/<repo>", f.Source.URL)
		}
		if safeSegment(parts[0]) != safeSegment(f.Source.GitOwner) || safeSegment(parts[1]) != safeSegment(f.Source.GitRepo) {
			return fmt.Errorf("source url %q disagrees with gitOwner/gitRepo %q/%q", f.Source.URL, f.Source.GitOwner, f.Source.GitRepo)
		}
		if len(f.Source.Commit) < 8 {
			return fmt.Errorf("source commit %q is too short for its id prefix", f.Source.Commit)
		}
		prefix = "github/" + safeSegment(f.Source.GitOwner) + "/" + safeSegment(f.Source.GitRepo) + "/" + safeSegment(f.Source.Commit[:8])
	default:
		return nil
	}
	if id != prefix && !strings.HasPrefix(id, prefix+"/") {
		return fmt.Errorf("id %q is not under %q", f.ID, prefix)
	}
	suffix, _ := idSuffix(id, prefix)
	if want := trimDot(f.Source.Path); suffix != want {
		return fmt.Errorf("id suffix %q does not match source.path %q", suffix, f.Source.Path)
	}
	return nil
}

func idSuffix(id, prefix string) (string, bool) {
	if id == prefix {
		return "", true
	}
	if strings.HasPrefix(id, prefix+"/") {
		return id[len(prefix)+1:], true
	}
	return "", false
}

// fixtureInputSizes totals the input files owned by each fixture. The smoke
// selection is size-based, so verify has to derive the same numbers.
func fixtureInputSizes(fixturesDir string, files []string, owner map[string]string) map[string]int64 {
	sizes := make(map[string]int64, len(files))
	for _, rel := range files {
		dir := owner[rel]
		local := rel
		if dir != "" {
			local = strings.TrimPrefix(rel, dir+"/")
		}
		if !strings.HasPrefix(local, "input/") {
			continue
		}
		info, err := os.Stat(filepath.Join(fixturesDir, filepath.FromSlash(rel)))
		if err != nil {
			continue
		}
		sizes[dir] += info.Size()
	}
	return sizes
}

// smokeSize is the number of smoke fixtures the collector aims for.
const smokeSize = 20

// expectedSmoke re-derives the smoke set with the same rule the collector
// uses: prefer feature-basic fixtures, then fill from the rest, smallest
// input first with the ID as a deterministic tiebreaker.
func expectedSmoke(fixtures []Fixture, sizes map[string]int64) map[string]bool {
	var preferred, fallback []Fixture
	for _, f := range fixtures {
		if onlyBasicFeatures(f.Features) {
			preferred = append(preferred, f)
		} else {
			fallback = append(fallback, f)
		}
	}
	bySize := func(fs []Fixture) {
		sort.Slice(fs, func(i, j int) bool {
			si, sj := sizes[trimDot(fs[i].ID)], sizes[trimDot(fs[j].ID)]
			if si != sj {
				return si < sj
			}
			return fs[i].ID < fs[j].ID
		})
	}
	bySize(preferred)
	bySize(fallback)
	out := make(map[string]bool, smokeSize)
	for _, group := range [][]Fixture{preferred, fallback} {
		for _, f := range group {
			if len(out) >= smokeSize {
				return out
			}
			out[trimDot(f.ID)] = true
		}
	}
	return out
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func isHexDigest(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
			return false
		}
	}
	return true
}
