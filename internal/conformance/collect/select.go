package collect

import (
	"fmt"
	"path"
	"sort"
	"strings"
)

// Candidate is one self-contained Terraform module root.
type Candidate struct {
	Dir         string
	Files       map[string][]byte
	Features    []string
	Providers   []string
	Targets     []string
	Tier        string
	Support     string
	Bytes       int64
	License     string
	LicenseAt   string
	LicenseText []byte
	Notice      []byte
	NoticeAt    string
}

// CandidateSkip records a directory that could not be used.
type CandidateSkip struct {
	Dir    string
	Reason string
}

// excludedSegments are directories whose module roots are test fixtures,
// examples of documentation, or private helpers rather than user-facing
// Terraform.
var excludedSegments = map[string]bool{
	"test":     true,
	"tests":    true,
	"spec":     true,
	"specs":    true,
	"docs":     true,
	"internal": true,
}

// BuildCandidates finds module roots in an archive. A candidate includes the
// local module directories it references, so the input tree is self-contained.
// Anything that references a path outside its own root is skipped because the
// fixture would not be a valid module on its own. When includes are given, only
// module roots under one of those path prefixes are considered; local modules
// the root references are still included.
func BuildCandidates(a *Archive, maxFixtures int, limits Limits, includes []string) ([]*Candidate, []CandidateSkip) {
	if maxFixtures <= 0 {
		maxFixtures = limits.MaxCandidatesPerRepo
	}
	dirs := groupByDir(a.Files)
	order := make([]string, 0, len(dirs))
	for dir := range dirs {
		order = append(order, dir)
	}
	sort.Slice(order, func(i, j int) bool {
		ri, rj := dirRank(order[i]), dirRank(order[j])
		if ri != rj {
			return ri < rj
		}
		return order[i] < order[j]
	})

	var out []*Candidate
	var skipped []CandidateSkip
	for _, dir := range order {
		if len(out) >= maxFixtures {
			break
		}
		if excludedDir(dir) {
			skipped = append(skipped, CandidateSkip{Dir: dir, Reason: "excluded directory"})
			continue
		}
		if !includedDir(dir, includes) {
			continue
		}
		cand, err := buildCandidate(dir, dirs, a.Files, limits)
		if err != nil {
			skipped = append(skipped, CandidateSkip{Dir: dir, Reason: err.Error()})
			continue
		}
		out = append(out, cand)
	}
	return out, skipped
}

func groupByDir(files map[string][]byte) map[string][]string {
	out := make(map[string][]string)
	for name := range files {
		dir := path.Dir(name)
		if dir == "." {
			dir = ""
		}
		out[dir] = append(out[dir], name)
	}
	for dir := range out {
		sort.Strings(out[dir])
	}
	return out
}

func dirRank(dir string) int {
	switch {
	case dir == "":
		return 0
	case dir == "modules" || strings.HasPrefix(dir, "modules/"):
		return 1
	case dir == "examples" || strings.HasPrefix(dir, "examples/"):
		return 2
	default:
		return 3
	}
}

func excludedDir(dir string) bool {
	for _, seg := range strings.Split(dir, "/") {
		if excludedSegments[strings.ToLower(seg)] {
			return true
		}
	}
	return false
}

func includedDir(dir string, includes []string) bool {
	if len(includes) == 0 {
		return true
	}
	for _, inc := range includes {
		inc = normalizeInclude(inc)
		if inc == "" {
			return true
		}
		if dir == inc || strings.HasPrefix(dir, inc+"/") {
			return true
		}
	}
	return false
}

func normalizeInclude(inc string) string {
	inc = strings.TrimSpace(inc)
	inc = strings.TrimSuffix(inc, "/**")
	inc = strings.TrimSuffix(inc, "/*")
	return strings.Trim(inc, "/")
}

func buildCandidate(root string, dirs map[string][]string, files map[string][]byte, limits Limits) (*Candidate, error) {
	included := map[string]bool{root: true}
	queue := []string{root}
	for len(queue) > 0 {
		dir := queue[0]
		queue = queue[1:]
		dirFiles := make(map[string][]byte, len(dirs[dir]))
		for _, file := range dirs[dir] {
			content, ok := files[file]
			if !ok {
				return nil, fmt.Errorf("unreadable file %s", file)
			}
			dirFiles[file] = content
		}
		analysis, err := AnalyzeFiles(dirFiles)
		if err != nil {
			return nil, err
		}
		for _, local := range analysis.LocalSources {
			resolved := path.Clean(path.Join(local.Dir, local.Source))
			if resolved == "." {
				resolved = ""
			}
			if resolved == dir {
				continue
			}
			if !within(root, resolved) {
				return nil, fmt.Errorf("relative source %q escapes module root", local.Source)
			}
			if _, ok := dirs[resolved]; !ok {
				return nil, fmt.Errorf("local module %q not found", local.Source)
			}
			if !included[resolved] {
				included[resolved] = true
				queue = append(queue, resolved)
			}
		}
	}

	cand := &Candidate{Dir: root, Files: make(map[string][]byte)}
	for dir := range included {
		for _, file := range dirs[dir] {
			rel := file
			if root != "" {
				rel = strings.TrimPrefix(file, root+"/")
			}
			content := files[file]
			cand.Files[rel] = content
			cand.Bytes += int64(len(content))
		}
	}
	if len(cand.Files) > limits.MaxFixtureFiles {
		return nil, fmt.Errorf("%d files exceeds fixture cap %d", len(cand.Files), limits.MaxFixtureFiles)
	}
	if cand.Bytes > limits.MaxFixtureBytes {
		return nil, fmt.Errorf("%d bytes exceeds fixture cap %d", cand.Bytes, limits.MaxFixtureBytes)
	}
	analysis, err := AnalyzeFiles(cand.Files)
	if err != nil {
		return nil, err
	}
	if len(analysis.Violations) > 0 {
		return nil, fmt.Errorf("policy: %s", analysis.Violations[0].String())
	}

	cand.Features = analysis.Features
	cand.Providers = analysis.Providers
	cand.Targets, err = targetsFor(analysis)
	if err != nil {
		return nil, err
	}
	cand.Tier = ClassifyTier(cand.Features)
	cand.Support = SupportLevel(cand.Features)
	return cand, nil
}

// targetsFor reports which output backends can consume a fixture. Only the AWS
// targets exist today, so a module that uses another provider without AWS has
// no meaningful output target.
func targetsFor(a *Analysis) ([]string, error) {
	hasModules := false
	for _, f := range a.Features {
		if f == "modules" {
			hasModules = true
			break
		}
	}
	hasAWS := false
	for _, p := range a.Providers {
		if p == "aws" {
			hasAWS = true
			break
		}
	}
	if hasAWS || (hasModules && len(a.Providers) == 0) {
		return []string{"cdk", "pulumi"}, nil
	}
	if len(a.Providers) == 0 {
		return nil, fmt.Errorf("no aws resources or module calls")
	}
	return nil, fmt.Errorf("no supported output target for providers %s", strings.Join(a.Providers, ", "))
}

func within(root, p string) bool {
	if p == "" || p == "." || strings.HasPrefix(p, "../") || path.IsAbs(p) {
		return false
	}
	if root == "" {
		return true
	}
	return p == root || strings.HasPrefix(p, root+"/")
}
