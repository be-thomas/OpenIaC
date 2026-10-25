// Package collect builds the conformance input corpus from public Terraform
// module sources.
package collect

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config describes where inputs come from and how much of them to keep.
type Config struct {
	Schema   int            `json:"schema"`
	Registry RegistryConfig `json:"registry"`
	GitHub   []GitHubSeed   `json:"github"`
	Limits   Limits         `json:"limits"`
	Licenses []string       `json:"licenses"`
}

// RegistryConfig controls the Terraform Registry crawl.
type RegistryConfig struct {
	Provider string `json:"provider"`
	// MaxPages bounds the registry listing crawl. Each page holds up to 100
	// modules.
	MaxPages int `json:"maxPages"`
	// MinDownloads keeps modules above a popularity floor.
	MinDownloads int `json:"minDownloads"`
	// MaxModules is the number of module roots downloaded after filtering.
	MaxModules int `json:"maxModules"`
	// Namespaces are always kept even below MinDownloads.
	Namespaces []string `json:"namespaces"`
	// Modules are always kept by full registry ID (namespace/name/provider).
	Modules []string `json:"modules"`
	// MaxFixturesPerRepo bounds submodules and examples taken from one repo.
	MaxFixturesPerRepo int `json:"maxFixturesPerRepo"`
}

// GitHubSeed adds a repository that is not reached through the registry.
type GitHubSeed struct {
	Repo string `json:"repo"`
	Ref  string `json:"ref"`
	// Include is a list of path prefixes. Empty means every directory is a
	// candidate.
	Include []string `json:"include"`
	// MaxFixtures overrides RegistryConfig.MaxFixturesPerRepo for this seed.
	MaxFixtures int `json:"maxFixtures"`
}

// Limits are hard caps for the committed corpus.
type Limits struct {
	MaxArchiveBytes      int64 `json:"maxArchiveBytes"`
	MaxArchiveText       int64 `json:"maxArchiveTextBytes"`
	MaxFixtureBytes      int64 `json:"maxFixtureBytes"`
	MaxFixtureFiles      int   `json:"maxFixtureFiles"`
	MaxTotalBytes        int64 `json:"maxTotalBytes"`
	MaxFileBytes         int64 `json:"maxFileBytes"`
	MaxCandidatesPerRepo int   `json:"maxCandidatesPerRepo"`
}

// Defaults returns the caps used when a config omits a field.
func Defaults() Config {
	return Config{
		Schema: 1,
		Registry: RegistryConfig{
			Provider:           "aws",
			MaxPages:           100,
			MinDownloads:       1000,
			MaxModules:         400,
			MaxFixturesPerRepo: 50,
		},
		Limits: Limits{
			MaxArchiveBytes:      40 << 20,
			MaxArchiveText:       80 << 20,
			MaxFixtureBytes:      512 << 10,
			MaxFixtureFiles:      200,
			MaxTotalBytes:        100 << 20,
			MaxFileBytes:         1 << 20,
			MaxCandidatesPerRepo: 50,
		},
		Licenses: []string{
			"Apache-2.0",
			"MIT",
			"BSD-2-Clause",
			"BSD-3-Clause",
			"ISC",
			"MPL-2.0",
			"Unlicense",
			"CC0-1.0",
		},
	}
}

// LoadConfig reads a sources.json file and fills defaults for omitted caps.
func LoadConfig(path string) (Config, error) {
	cfg := Defaults()
	raw, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config %s: %w", path, err)
	}
	if cfg.Schema != 1 {
		return cfg, fmt.Errorf("unsupported config schema %d", cfg.Schema)
	}
	if cfg.Registry.Provider == "" {
		cfg.Registry.Provider = "aws"
	}
	if cfg.Registry.MaxPages <= 0 {
		cfg.Registry.MaxPages = 100
	}
	if cfg.Registry.MaxModules <= 0 {
		cfg.Registry.MaxModules = 400
	}
	if cfg.Registry.MaxFixturesPerRepo <= 0 {
		cfg.Registry.MaxFixturesPerRepo = 50
	}
	if cfg.Limits.MaxArchiveBytes <= 0 {
		cfg.Limits.MaxArchiveBytes = 40 << 20
	}
	if cfg.Limits.MaxArchiveText <= 0 {
		cfg.Limits.MaxArchiveText = 80 << 20
	}
	if cfg.Limits.MaxFixtureBytes <= 0 {
		cfg.Limits.MaxFixtureBytes = 512 << 10
	}
	if cfg.Limits.MaxFixtureFiles <= 0 {
		cfg.Limits.MaxFixtureFiles = 200
	}
	if cfg.Limits.MaxTotalBytes <= 0 {
		cfg.Limits.MaxTotalBytes = 100 << 20
	}
	if cfg.Limits.MaxFileBytes <= 0 {
		cfg.Limits.MaxFileBytes = 1 << 20
	}
	if cfg.Limits.MaxCandidatesPerRepo <= 0 {
		cfg.Limits.MaxCandidatesPerRepo = 50
	}
	if len(cfg.Licenses) == 0 {
		cfg.Licenses = Defaults().Licenses
	}
	return cfg, nil
}

// LicenseAllowed reports whether the SPDX identifier may be vendored.
func (c Config) LicenseAllowed(spdx string) bool {
	for _, l := range c.Licenses {
		if l == spdx {
			return true
		}
	}
	return false
}
