// Command openiac-conformance maintains the conformance corpus.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/be-thomas/openiac/internal/conformance/collect"
	"github.com/be-thomas/openiac/internal/conformance/status"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "collect":
		if err := runCollect(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "openiac-conformance:", err)
			os.Exit(1)
		}
	case "verify":
		if err := runVerify(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "openiac-conformance:", err)
			os.Exit(1)
		}
	case "lock":
		if err := runLock(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "openiac-conformance:", err)
			os.Exit(1)
		}
	case "status":
		if err := runStatus(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "openiac-conformance:", err)
			os.Exit(1)
		}
	case "-h", "--help", "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `usage: openiac-conformance <collect|verify|lock|status> [flags]

collect downloads upstream modules and rewrites the fixture tree.

  -config string        sources file (default conformance/collection/sources.json)
  -out string           fixture tree (default conformance/fixtures)
  -lock string          lock file (default conformance/corpus.lock.json)
  -third-party string   attribution file (default THIRD_PARTY.md)
  -report string        run report (default .cache/conformance/collect-report.json)
  -limit int            override registry module cap
  -pages int            override registry page cap
  -min-downloads int    override registry popularity floor
  -workers int          concurrent downloads (default 6)
  -v                    verbose progress

verify checks the committed tree against its lock, attribution and fixture
metadata. It never writes or uses the network.

  -config string        sources file (default conformance/collection/sources.json)
  -out string           fixture tree (default conformance/fixtures)
  -lock string          lock file (default conformance/corpus.lock.json)
  -third-party string   attribution file (default THIRD_PARTY.md)

lock rewrites the lock and attribution from the fixture tree. It never writes
to the fixture tree and never uses the network.

  -config string        sources file (default conformance/collection/sources.json)
  -out string           fixture tree (default conformance/fixtures)
  -lock string          lock file (default conformance/corpus.lock.json)
  -third-party string   attribution file (default THIRD_PARTY.md)

status parses every fixture and rewrites the committed summary the README
renders. It never uses the network.

  -out string           fixture tree (default conformance/fixtures)
  -lock string          lock file (default conformance/corpus.lock.json)
  -status string        summary file (default conformance/status.json)
  -badge string         shields.io endpoint file (default conformance/badge.json)
  -readme string        file holding the generated summary block (default README.md)
  -check                report a stale summary instead of rewriting it
`)
}

func runCollect(args []string) error {
	fs := flag.NewFlagSet("collect", flag.ExitOnError)
	configPath := fs.String("config", "conformance/collection/sources.json", "sources file")
	out := fs.String("out", "conformance/fixtures", "fixture tree")
	lock := fs.String("lock", "conformance/corpus.lock.json", "lock file")
	thirdParty := fs.String("third-party", "THIRD_PARTY.md", "attribution file")
	report := fs.String("report", ".cache/conformance/collect-report.json", "run report")
	limit := fs.Int("limit", 0, "override registry module cap")
	pages := fs.Int("pages", 0, "override registry page cap")
	minDownloads := fs.Int("min-downloads", 0, "override registry popularity floor")
	workers := fs.Int("workers", 6, "concurrent downloads")
	verbose := fs.Bool("v", false, "verbose progress")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := collect.LoadConfig(*configPath)
	if err != nil {
		return err
	}
	if *limit > 0 {
		cfg.Registry.MaxModules = *limit
	}
	if *pages > 0 {
		cfg.Registry.MaxPages = *pages
	}
	if *minDownloads > 0 {
		cfg.Registry.MinDownloads = *minDownloads
	}

	logf := func(string, ...any) {}
	if *verbose {
		logf = collect.Logf
	}
	client := collect.NewClient(githubToken())
	c := collect.New(cfg, client, collect.Options{
		OutDir:         *out,
		LockPath:       *lock,
		ThirdPartyPath: *thirdParty,
		ReportPath:     *report,
		Workers:        *workers,
		Logf:           logf,
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := c.Run(ctx); err != nil {
		return err
	}
	fmt.Println(collect.Summary(c.Report()))
	return nil
}

func runVerify(args []string) error {
	fs := flag.NewFlagSet("verify", flag.ExitOnError)
	configPath := fs.String("config", "conformance/collection/sources.json", "sources file")
	out := fs.String("out", "conformance/fixtures", "fixture tree")
	lock := fs.String("lock", "conformance/corpus.lock.json", "lock file")
	thirdParty := fs.String("third-party", "THIRD_PARTY.md", "attribution file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	cfg, err := collect.LoadConfig(*configPath)
	if err != nil {
		return err
	}
	res, err := collect.Verify(*out, *lock, *thirdParty, cfg)
	if err != nil {
		return err
	}
	fmt.Printf("verified %d fixtures, %d files, %d bytes\n", res.Fixtures, res.Files, res.Bytes)
	return nil
}

func runLock(args []string) error {
	fs := flag.NewFlagSet("lock", flag.ExitOnError)
	configPath := fs.String("config", "conformance/collection/sources.json", "sources file")
	out := fs.String("out", "conformance/fixtures", "fixture tree")
	lock := fs.String("lock", "conformance/corpus.lock.json", "lock file")
	thirdParty := fs.String("third-party", "THIRD_PARTY.md", "attribution file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	cfg, err := collect.LoadConfig(*configPath)
	if err != nil {
		return err
	}
	l, err := collect.RefreshAggregates(*out, *lock, *thirdParty, cfg)
	if err != nil {
		return err
	}
	fmt.Printf("lock: %d fixtures, %d files, %d bytes, tree %s\n",
		l.FixtureCount, l.FileCount, l.TotalBytes, l.TreeSHA256)
	return nil
}

func runStatus(args []string) error {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	out := fs.String("out", "conformance/fixtures", "fixture tree")
	lockPath := fs.String("lock", "conformance/corpus.lock.json", "lock file")
	statusPath := fs.String("status", "conformance/status.json", "summary file")
	badgePath := fs.String("badge", "conformance/badge.json", "shields.io endpoint file")
	readmePath := fs.String("readme", "README.md", "file holding the generated summary block")
	check := fs.Bool("check", false, "report a stale summary instead of rewriting it")
	if err := fs.Parse(args); err != nil {
		return err
	}

	report, err := status.Compute(*out, *lockPath)
	if err != nil {
		return err
	}
	statusJSON, err := marshalIndent(report)
	if err != nil {
		return err
	}
	badgeJSON, err := marshalIndent(report.Badge())
	if err != nil {
		return err
	}
	readme, err := os.ReadFile(*readmePath)
	if err != nil {
		return err
	}

	if *check {
		var stale []string
		for _, file := range []struct {
			path string
			want []byte
		}{
			{*statusPath, statusJSON},
			{*badgePath, badgeJSON},
		} {
			got, err := os.ReadFile(file.path)
			if err != nil || normalizeNewlines(string(got)) != normalizeNewlines(string(file.want)) {
				stale = append(stale, file.path)
			}
		}
		section, err := status.ExtractSection(string(readme))
		if err != nil {
			return fmt.Errorf("%s: %w", *readmePath, err)
		}
		if normalizeNewlines(section) != normalizeNewlines(report.Section()) {
			stale = append(stale, *readmePath)
		}
		if len(stale) > 0 {
			return fmt.Errorf("%s is stale: run `make conformance-status` and commit the result", strings.Join(stale, ", "))
		}
		fmt.Printf("status: %d/%d modules parse, %d/%d emit CDK, %d/%d emit Pulumi, summary current\n",
			report.Parse.Parsed, report.Parse.Modules, report.Emit.CDK.Emitted, report.Emit.CDK.Modules, report.Emit.Pulumi.Emitted, report.Emit.Pulumi.Modules)
		return nil
	}

	// Render the README block before touching any file, so a missing marker
	// fails the run instead of leaving half the summary rewritten.
	updated, err := status.ReplaceSection(string(readme), report.Section())
	if err != nil {
		return fmt.Errorf("%s: %w", *readmePath, err)
	}
	if err := os.WriteFile(*statusPath, statusJSON, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(*badgePath, badgeJSON, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(*readmePath, []byte(updated), 0o644); err != nil {
		return err
	}
	fmt.Printf("status: %d/%d modules parse, %d/%d emit CDK, %d/%d emit Pulumi\n",
		report.Parse.Parsed, report.Parse.Modules, report.Emit.CDK.Emitted, report.Emit.CDK.Modules, report.Emit.Pulumi.Emitted, report.Emit.Pulumi.Modules)
	fmt.Println(report.Markdown())
	return nil
}

// normalizeNewlines keeps -check independent of the checkout's line endings:
// Git for Windows hands out CRLF files for content stored as LF.
func normalizeNewlines(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}

// marshalIndent encodes a report the way the rest of the repository encodes
// committed JSON: two-space indent and a trailing newline.
func marshalIndent(v any) ([]byte, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func githubToken() string {
	for _, key := range []string{"GITHUB_TOKEN", "GH_TOKEN"} {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return v
		}
	}
	out, err := exec.Command("gh", "auth", "token").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
