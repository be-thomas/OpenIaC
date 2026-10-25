package collect

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// Options controls one collection run.
type Options struct {
	OutDir         string
	LockPath       string
	ThirdPartyPath string
	ReportPath     string
	Workers        int
	Logf           func(string, ...any)
}

// Collector downloads, filters and writes the input corpus.
type Collector struct {
	cfg    Config
	client *Client
	opt    Options
	report Report
	mu     sync.Mutex
}

// Report is written to the gitignored cache for auditing a run.
type Report struct {
	Schema     int    `json:"schema"`
	StartedAt  string `json:"startedAt"`
	FinishedAt string `json:"finishedAt"`
	Registry   struct {
		Pages       int `json:"pages"`
		ModulesSeen int `json:"modulesSeen"`
		Selected    int `json:"selected"`
		Resolved    int `json:"resolved"`
	} `json:"registry"`
	GitHub struct {
		Seeds    int `json:"seeds"`
		Resolved int `json:"resolved"`
	} `json:"github"`
	Archives struct {
		Downloaded int `json:"downloaded"`
		Failed     int `json:"failed"`
	} `json:"archives"`
	Fixtures struct {
		Written    int   `json:"written"`
		Skipped    int   `json:"skipped"`
		Duplicates int   `json:"duplicates"`
		TotalBytes int64 `json:"totalBytes"`
	} `json:"fixtures"`
	Skips []Skip `json:"skips"`
}

// Skip is one rejected source or candidate.
type Skip struct {
	Source string `json:"source"`
	Path   string `json:"path,omitempty"`
	Reason string `json:"reason"`
}

// New returns a collector for a loaded config.
func New(cfg Config, client *Client, opt Options) *Collector {
	if opt.Workers <= 0 {
		opt.Workers = 6
	}
	if opt.Workers > 16 {
		// Unbounded concurrency turns a rate limit into a retry storm.
		opt.Workers = 16
	}
	if opt.Logf == nil {
		opt.Logf = func(string, ...any) {}
	}
	return &Collector{cfg: cfg, client: client, opt: opt}
}

// Run executes the collection and writes fixtures plus aggregate files.
func (c *Collector) Run(ctx context.Context) error {
	c.report = Report{Schema: 1, StartedAt: time.Now().UTC().Format(time.RFC3339)}
	if err := os.MkdirAll(c.opt.OutDir, 0o755); err != nil {
		return err
	}

	mods, pages, err := c.client.RegistryModules(ctx, c.cfg.Registry, c.opt.Logf)
	if err != nil {
		return err
	}
	c.report.Registry.Pages = pages

	seen := make(map[string]bool, len(mods))
	for _, m := range mods {
		seen[registryID(m)] = true
	}
	for _, id := range c.cfg.Registry.Modules {
		if seen[id] {
			continue
		}
		m, err := c.client.RegistryModuleByID(ctx, id)
		if err != nil {
			c.addSkip(id, "", "configured module: "+err.Error())
			continue
		}
		mods = append(mods, m)
		seen[registryID(m)] = true
	}
	c.report.Registry.ModulesSeen = len(mods)

	selected := selectRegistryModules(mods, c.cfg, 0, 0)
	c.report.Registry.Selected = len(selected)
	c.opt.Logf("registry: selected %d module roots", len(selected))

	registryMods, skips := c.resolveRegistry(ctx, selected)
	for _, s := range skips {
		c.addSkip(s.Source, s.Path, s.Reason)
	}
	c.report.Registry.Resolved = len(registryMods)

	seeds, seedSkips := c.resolveSeeds(ctx)
	for _, s := range seedSkips {
		c.addSkip(s.Source, s.Path, s.Reason)
	}
	c.report.GitHub.Seeds = len(c.cfg.GitHub)
	c.report.GitHub.Resolved = len(seeds)

	all := append(registryMods, seeds...)
	all = dedupeModules(all)
	sort.Slice(all, func(i, j int) bool {
		ki, kj := moduleKey(all[i]), moduleKey(all[j])
		return ki < kj
	})
	c.opt.Logf("sources: %d unique repositories", len(all))

	results := c.processAll(ctx, all)
	// A cancelled run must not present its partial crawl as a finished corpus.
	if err := ctx.Err(); err != nil {
		return err
	}
	pending, totalBytes := c.collectFixtures(results)
	c.opt.Logf("fixtures: %d selected, %d bytes", len(pending), totalBytes)

	fallback := markSmoke(pending)
	if fallback > 0 {
		c.opt.Logf("smoke: %d of 20 tiers used the fallback because too few basic modules were available", fallback)
	}
	sort.Slice(pending, func(i, j int) bool { return pending[i].fixture.ID < pending[j].fixture.ID })

	for _, p := range pending {
		n, err := WriteFixture(c.opt.OutDir, p.fixture, p.cand, p.cand.LicenseText, p.cand.Notice)
		if err != nil {
			c.report.Fixtures.Skipped++
			c.addSkip(p.fixture.Source.URL, p.fixture.ID, "write: "+err.Error())
			continue
		}
		c.report.Fixtures.Written++
		c.report.Fixtures.TotalBytes += n
	}

	// RefreshAggregates checks the byte budget before writing either aggregate.
	// Each file is replaced atomically; verify reports a run that failed between
	// the two writes.
	lock, err := RefreshAggregates(c.opt.OutDir, c.opt.LockPath, c.opt.ThirdPartyPath, c.cfg)
	if err != nil {
		return err
	}
	c.opt.Logf("lock: %d fixtures, %d files, %d bytes, tree %s",
		lock.FixtureCount, lock.FileCount, lock.TotalBytes, lock.TreeSHA256)

	c.report.FinishedAt = time.Now().UTC().Format(time.RFC3339)
	if c.opt.ReportPath != "" {
		if err := c.writeReport(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Collector) resolveRegistry(ctx context.Context, mods []RegistryModule) ([]GitModule, []Skip) {
	type result struct {
		m   RegistryModule
		gm  GitModule
		err error
	}
	out := make([]result, len(mods))
	var done atomic.Int64
	jobs := make(chan int)
	var wg sync.WaitGroup
	workers := c.opt.Workers
	if workers > 4 {
		workers = 4
	}
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				gm, err := c.client.ResolveRegistryDownload(ctx, mods[idx])
				out[idx] = result{m: mods[idx], gm: gm, err: err}
				n := done.Add(1)
				if n%25 == 0 {
					c.opt.Logf("registry: resolved %d/%d", n, len(mods))
				}
			}
		}()
	}
	for i := range mods {
		jobs <- i
	}
	close(jobs)
	wg.Wait()

	var gm []GitModule
	var skips []Skip
	for _, r := range out {
		if r.err != nil {
			skips = append(skips, Skip{Source: r.m.ID, Reason: r.err.Error()})
			continue
		}
		gm = append(gm, r.gm)
	}
	sort.Slice(gm, func(i, j int) bool {
		ki, kj := moduleKey(gm[i]), moduleKey(gm[j])
		return ki < kj
	})
	return gm, skips
}

func (c *Collector) resolveSeeds(ctx context.Context) ([]GitModule, []Skip) {
	var out []GitModule
	var skips []Skip
	for _, seed := range c.cfg.GitHub {
		gm, err := c.client.ResolveSeed(ctx, seed)
		if err != nil {
			skips = append(skips, Skip{Source: seed.Repo, Reason: err.Error()})
			continue
		}
		out = append(out, gm)
	}
	return out, skips
}

func (c *Collector) processAll(ctx context.Context, items []GitModule) []archiveResult {
	out := make([]archiveResult, len(items))
	var done atomic.Int64
	jobs := make(chan int)
	var wg sync.WaitGroup
	for i := 0; i < c.opt.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				out[idx] = c.process(ctx, items[idx])
				n := done.Add(1)
				if n%20 == 0 {
					c.opt.Logf("archives: processed %d/%d", n, len(items))
				}
			}
		}()
	}
	for i := range items {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	return out
}

type archiveResult struct {
	module     GitModule
	archive    *Archive
	candidates []*Candidate
	skips      []Skip
	reason     string
}

func (c *Collector) process(ctx context.Context, gm GitModule) archiveResult {
	res := archiveResult{module: gm}
	data, err := c.client.DownloadArchive(ctx, gm.Owner, gm.Repo, gm.Ref, c.cfg.Limits.MaxArchiveBytes)
	if err != nil {
		res.reason = err.Error()
		return res
	}
	arch, err := ExtractArchive(data, c.cfg.Limits.MaxArchiveText, c.cfg.Limits.MaxFileBytes)
	if err != nil {
		res.reason = err.Error()
		return res
	}
	res.archive = arch

	maxFixtures := c.cfg.Registry.MaxFixturesPerRepo
	var includes []string
	if gm.Seed != nil {
		if gm.Seed.MaxFixtures > 0 {
			maxFixtures = gm.Seed.MaxFixtures
		}
		includes = gm.Seed.Include
	}
	if gm.Subdir != "" {
		// A registry module may declare its root inside a monorepo. Restrict
		// candidates to that directory so the plain fixture ID really denotes
		// the published module root.
		includes = []string{gm.Subdir}
	}
	cands, skipped := BuildCandidates(arch, maxFixtures, c.cfg.Limits, includes)
	for _, s := range skipped {
		res.skips = append(res.skips, Skip{
			Source: moduleKey(gm),
			Path:   s.Dir,
			Reason: s.Reason,
		})
	}

	// A candidate is governed by its own LICENSE or the nearest ancestor
	// license, so a permissive repository can still exclude a copyleft subtree.
	kept := cands[:0]
	for _, cand := range cands {
		licenseAt, licenseRaw, ok := arch.NearestLicense(cand.Dir)
		if !ok {
			res.skips = append(res.skips, Skip{Source: moduleKey(gm), Path: cand.Dir, Reason: "no license file"})
			continue
		}
		if int64(len(licenseRaw)) > c.cfg.Limits.MaxFileBytes {
			res.skips = append(res.skips, Skip{Source: moduleKey(gm), Path: cand.Dir, Reason: "license file too large"})
			continue
		}
		spdx := DetectLicense(licenseRaw)
		if spdx == "" {
			res.skips = append(res.skips, Skip{Source: moduleKey(gm), Path: cand.Dir, Reason: "unrecognized license"})
			continue
		}
		if !c.cfg.LicenseAllowed(spdx) {
			res.skips = append(res.skips, Skip{Source: moduleKey(gm), Path: cand.Dir, Reason: "license not allowed: " + spdx})
			continue
		}
		cand.License = spdx
		cand.LicenseAt = licenseAt
		cand.LicenseText = licenseRaw
		if noticeAt, notice, ok := arch.NearestNotice(cand.Dir); ok && int64(len(notice)) <= c.cfg.Limits.MaxFileBytes {
			cand.Notice = notice
			cand.NoticeAt = noticeAt
		}
		kept = append(kept, cand)
	}
	res.candidates = kept
	return res
}

type pendingFixture struct {
	fixture Fixture
	cand    *Candidate
}

func (c *Collector) collectFixtures(results []archiveResult) ([]pendingFixture, int64) {
	var pending []pendingFixture
	seenID := make(map[string]bool)
	seenHash := make(map[string]bool)
	var total int64
	for _, res := range results {
		// Per-directory skips are useful even when the archive as a whole is
		// rejected, so record them before the early exit.
		for _, s := range res.skips {
			c.addSkip(s.Source, s.Path, s.Reason)
		}
		if res.archive != nil {
			c.report.Archives.Downloaded++
		}
		if res.reason != "" {
			c.report.Archives.Failed++
			c.addSkip(moduleKey(res.module), "", res.reason)
			continue
		}
		for _, cand := range res.candidates {
			f := BuildFixture(res.module, cand, res.archive.SHA256)
			if len(f.ID) > 240 {
				c.report.Fixtures.Skipped++
				c.addSkip(moduleKey(res.module), f.ID, "fixture id too long")
				continue
			}
			if seenID[f.ID] {
				c.report.Fixtures.Duplicates++
				continue
			}
			if seenHash[f.Source.InputSHA256] {
				c.report.Fixtures.Duplicates++
				continue
			}
			estimate := cand.Bytes + int64(len(cand.LicenseText)+len(cand.Notice)) + 4096
			if total+estimate > c.cfg.Limits.MaxTotalBytes {
				c.report.Fixtures.Skipped++
				c.addSkip(moduleKey(res.module), f.ID, "total corpus budget reached")
				continue
			}
			seenID[f.ID] = true
			seenHash[f.Source.InputSHA256] = true
			total += estimate
			pending = append(pending, pendingFixture{fixture: f, cand: cand})
		}
	}
	return pending, total
}

func markSmoke(pending []pendingFixture) int {
	var preferred, fallback []int
	for i := range pending {
		if onlyBasicFeatures(pending[i].cand.Features) {
			preferred = append(preferred, i)
		} else {
			fallback = append(fallback, i)
		}
	}
	bySize := func(indices []int) {
		sort.Slice(indices, func(i, j int) bool {
			a, b := pending[indices[i]], pending[indices[j]]
			if a.cand.Bytes != b.cand.Bytes {
				return a.cand.Bytes < b.cand.Bytes
			}
			return a.fixture.ID < b.fixture.ID
		})
	}
	bySize(preferred)
	bySize(fallback)
	order := append(preferred, fallback...)
	usedFallback := 0
	for i, idx := range order {
		if i >= 20 {
			break
		}
		if !onlyBasicFeatures(pending[idx].cand.Features) {
			usedFallback++
		}
		pending[idx].fixture.Tier = "smoke"
	}
	return usedFallback
}

func onlyBasicFeatures(features []string) bool {
	basic := map[string]bool{
		"outputs": true, "providers": true, "required_providers": true,
		"resources": true, "terraform": true, "variables": true,
	}
	for _, f := range features {
		if !basic[f] {
			return false
		}
	}
	return true
}

func selectRegistryModules(mods []RegistryModule, cfg Config, limit, minDownloads int) []RegistryModule {
	if limit <= 0 {
		limit = cfg.Registry.MaxModules
	}
	if minDownloads <= 0 {
		minDownloads = cfg.Registry.MinDownloads
	}
	namespaces := make(map[string]bool, len(cfg.Registry.Namespaces))
	for _, n := range cfg.Registry.Namespaces {
		namespaces[n] = true
	}
	always := make(map[string]bool, len(cfg.Registry.Modules))
	for _, id := range cfg.Registry.Modules {
		always[id] = true
	}
	seen := make(map[string]bool)
	var out []RegistryModule
	for _, m := range mods {
		if m.Provider != cfg.Registry.Provider || m.Version == "" {
			continue
		}
		id := registryID(m)
		if seen[id] {
			continue
		}
		if m.Downloads < minDownloads && !namespaces[m.Namespace] && !always[id] {
			continue
		}
		seen[id] = true
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Downloads != out[j].Downloads {
			return out[i].Downloads > out[j].Downloads
		}
		return out[i].ID < out[j].ID
	})
	if limit > 0 && len(out) > limit {
		// Configured modules are pinned deliberately, so they keep their slot
		// even when popularity would place them below the cut.
		kept := make([]RegistryModule, 0, limit)
		var rest []RegistryModule
		for _, m := range out {
			if always[registryID(m)] && len(kept) < limit {
				kept = append(kept, m)
			} else {
				rest = append(rest, m)
			}
		}
		for _, m := range rest {
			if len(kept) >= limit {
				break
			}
			kept = append(kept, m)
		}
		sort.Slice(kept, func(i, j int) bool {
			if kept[i].Downloads != kept[j].Downloads {
				return kept[i].Downloads > kept[j].Downloads
			}
			return kept[i].ID < kept[j].ID
		})
		out = kept
	}
	return out
}

func dedupeModules(mods []GitModule) []GitModule {
	seen := make(map[string]bool)
	var out []GitModule
	for _, m := range mods {
		k := moduleKey(m)
		if seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, m)
	}
	return out
}

func moduleKey(m GitModule) string {
	if m.Registry != nil {
		return m.Registry.ID + "@" + m.Ref
	}
	return m.Owner + "/" + m.Repo + "@" + m.Ref
}

func (c *Collector) addSkip(source, path, reason string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.report.Skips) >= 5000 {
		return
	}
	c.report.Skips = append(c.report.Skips, Skip{Source: source, Path: path, Reason: reason})
}

// Report returns the counters from the last Run.
func (c *Collector) Report() Report {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.report
}

func (c *Collector) writeReport() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	data, err := json.MarshalIndent(c.report, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(c.opt.ReportPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(c.opt.ReportPath, data, 0o644)
}

func logf(format string, args ...any) { log.Printf(format, args...) }

// Logf is the default logger used by the command.
var Logf = logf

// Summary is a human-readable one-line result.
func Summary(r Report) string {
	return fmt.Sprintf("wrote %d fixtures (%d bytes); %d archives failed; %d skips",
		r.Fixtures.Written, r.Fixtures.TotalBytes, r.Archives.Failed, len(r.Skips))
}
