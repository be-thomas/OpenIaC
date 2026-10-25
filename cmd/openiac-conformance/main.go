// Command openiac-conformance maintains the conformance corpus.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/be-thomas/openiac/internal/conformance/collect"
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
	case "-h", "--help", "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `usage: openiac-conformance collect [flags]

flags:
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
		MaxModules:     cfg.Registry.MaxModules,
		MaxPages:       cfg.Registry.MaxPages,
		MinDownloads:   cfg.Registry.MinDownloads,
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
