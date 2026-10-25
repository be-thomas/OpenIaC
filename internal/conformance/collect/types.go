package collect

import "strings"

// RegistryModule is one module row from the Terraform Registry listing.
type RegistryModule struct {
	ID          string `json:"id"`
	Namespace   string `json:"namespace"`
	Name        string `json:"name"`
	Provider    string `json:"provider"`
	Version     string `json:"version"`
	Downloads   int    `json:"downloads"`
	Source      string `json:"source"`
	Description string `json:"description"`
	PublishedAt string `json:"published_at"`
}

// registryID is the version-independent module identifier. The public API's
// id field includes the version, so it cannot be used for de-duplication.
func registryID(m RegistryModule) string {
	return m.Namespace + "/" + m.Name + "/" + m.Provider
}

// RegistryPage is one page of /v1/modules.
type RegistryPage struct {
	Meta struct {
		Limit   int    `json:"limit"`
		Offset  int    `json:"offset"`
		Total   int    `json:"total"`
		NextURL string `json:"next_url"`
	} `json:"meta"`
	Modules []RegistryModule `json:"modules"`
}

// GitModule is a module resolved to an immutable git commit.
type GitModule struct {
	// Registry is set for registry modules.
	Registry *RegistryModule
	// Seed is set for configured GitHub seeds.
	Seed  *GitHubSeed
	Owner string
	Repo  string
	Ref   string
}

// SourceURL is the canonical, human-readable upstream location.
func (m GitModule) SourceURL() string {
	if m.Registry != nil {
		return "https://registry.terraform.io/modules/" + m.Registry.Namespace + "/" + m.Registry.Name + "/" + m.Registry.Provider + "/" + m.Registry.Version
	}
	return "https://github.com/" + m.Owner + "/" + m.Repo
}

// RefLabel is the version or commit recorded in fixture metadata.
func (m GitModule) RefLabel() string {
	if m.Registry != nil {
		return m.Registry.Version
	}
	return m.Ref
}

// Prefix is the fixture directory prefix for this upstream.
func (m GitModule) Prefix() string {
	if m.Registry != nil {
		return "registry/" + safeSegment(m.Registry.Namespace) + "/" + safeSegment(m.Registry.Name) + "/" + safeSegment(m.Registry.Version)
	}
	short := m.Ref
	if len(short) > 8 {
		short = short[:8]
	}
	return "github/" + safeSegment(m.Owner) + "/" + safeSegment(m.Repo) + "/" + safeSegment(short)
}

// safeSegment keeps a path component readable while preventing traversal and
// characters that are awkward on case-insensitive filesystems.
func safeSegment(s string) string {
	s = strings.TrimSpace(s)
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '.', r == '_', r == '-', r == '+', r == '@':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	out := strings.Trim(b.String(), ".-")
	if out == "" || out == "." || out == ".." {
		return "unnamed"
	}
	return out
}
