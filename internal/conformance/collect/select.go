package collect

import (
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"
)

// Candidate is one self-contained Terraform module root.
type Candidate struct {
	Dir       string
	Files     map[string][]byte
	Features  []string
	Providers []string
	Tier      string
	Support   string
	Bytes     int64
}

// CandidateSkip records a directory that could not be used.
type CandidateSkip struct {
	Dir    string
	Reason string
}

var localSourceRE = regexp.MustCompile(`(?m)\bsource\s*=\s*"([^"]+)"`)

// BuildCandidates finds module roots in an archive. A candidate includes the
// local module directories it references, so the input tree is self-contained.
// Anything that references a path outside its own root is skipped because the
// fixture would not be a valid module on its own.
func BuildCandidates(a *Archive, maxFixtures int, limits Limits) ([]*Candidate, []CandidateSkip) {
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

func buildCandidate(root string, dirs map[string][]string, files map[string][]byte, limits Limits) (*Candidate, error) {
	included := map[string]bool{root: true}
	queue := []string{root}
	for len(queue) > 0 {
		dir := queue[0]
		queue = queue[1:]
		for _, file := range dirs[dir] {
			content, ok := files[file]
			if !ok {
				return nil, fmt.Errorf("unreadable file %s", file)
			}
			for _, m := range localSourceRE.FindAllSubmatch(content, -1) {
				src := string(m[1])
				if !strings.HasPrefix(src, "./") && !strings.HasPrefix(src, "../") {
					continue
				}
				resolved := path.Clean(path.Join(dir, src))
				if resolved == "." {
					resolved = ""
				}
				if resolved == dir {
					continue
				}
				if !within(root, resolved) {
					return nil, fmt.Errorf("relative source %q escapes module root", src)
				}
				if _, ok := dirs[resolved]; !ok {
					return nil, fmt.Errorf("local module %q not found", src)
				}
				if !included[resolved] {
					included[resolved] = true
					queue = append(queue, resolved)
				}
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
	if violations := ScanFiles(cand.Files); len(violations) > 0 {
		return nil, fmt.Errorf("policy: %s", violations[0].String())
	}

	cand.Features = DetectFeatures(cand.Files)
	cand.Providers = DetectProviders(cand.Files)
	if !stringSet(cand.Providers)["aws"] && !stringSet(cand.Features)["modules"] {
		return nil, fmt.Errorf("no aws resources or module calls")
	}
	cand.Tier = ClassifyTier(cand.Features)
	cand.Support = SupportLevel(cand.Features)
	return cand, nil
}

func within(root, p string) bool {
	if p == "" || p == "." || strings.HasPrefix(p, "../") {
		return false
	}
	if root == "" {
		return true
	}
	return p == root || strings.HasPrefix(p, root+"/")
}
