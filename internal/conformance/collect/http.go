package collect

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	registryBase   = "https://registry.terraform.io"
	githubAPI      = "https://api.github.com"
	githubCodeload = "https://codeload.github.com"
	userAgent      = "OpenIaC-conformance/0.1 (+https://github.com/be-thomas/OpenIaC)"
)

// Client fetches registry metadata and GitHub archives.
type Client struct {
	HTTP  *http.Client
	Token string
}

// NewClient returns a client with conservative timeouts and a descriptive
// user agent.
func NewClient(token string) *Client {
	return &Client{
		HTTP: &http.Client{
			Timeout: 120 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 5 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
		Token: token,
	}
}

func (c *Client) request(ctx context.Context, rawURL string, accept string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	if accept != "" {
		req.Header.Set("Accept", accept)
	}
	if c.Token != "" && strings.HasPrefix(rawURL, githubAPI) {
		req.Header.Set("Authorization", "Bearer "+c.Token)
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	}
	return c.do(req)
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt < 4; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(1<<uint(attempt-1)) * time.Second)
		}
		resp, err := c.HTTP.Do(req.Clone(req.Context()))
		if err != nil {
			lastErr = err
			continue
		}
		if resp.StatusCode == http.StatusTooManyRequests ||
			resp.StatusCode >= 500 ||
			(resp.StatusCode == http.StatusForbidden && resp.Header.Get("X-RateLimit-Remaining") == "0") {
			wait := retryAfter(resp)
			io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
			resp.Body.Close()
			lastErr = fmt.Errorf("%s: %s", req.URL, resp.Status)
			if wait > 0 {
				time.Sleep(wait)
			}
			continue
		}
		return resp, nil
	}
	return nil, lastErr
}

func retryAfter(resp *http.Response) time.Duration {
	if v := resp.Header.Get("Retry-After"); v != "" {
		if secs, err := strconv.Atoi(v); err == nil && secs >= 0 && secs <= 120 {
			return time.Duration(secs) * time.Second
		}
	}
	if reset := resp.Header.Get("X-RateLimit-Reset"); reset != "" {
		if unix, err := strconv.ParseInt(reset, 10, 64); err == nil {
			d := time.Until(time.Unix(unix, 0))
			if d > 0 && d < 2*time.Minute {
				return d
			}
		}
	}
	return 0
}

// RegistryModules pages through the registry listing. It stops after maxPages
// pages or when the registry reports no next page.
func (c *Client) RegistryModules(ctx context.Context, cfg RegistryConfig, logf func(string, ...any)) ([]RegistryModule, int, error) {
	next := fmt.Sprintf("%s/v1/modules?limit=100&offset=0&provider=%s", registryBase, url.QueryEscape(cfg.Provider))
	seen := make(map[string]bool)
	var out []RegistryModule
	pages := 0
	for next != "" && pages < cfg.MaxPages {
		pageURL := next
		resp, err := c.request(ctx, next, "application/json")
		if err != nil {
			return out, pages, fmt.Errorf("registry page %d: %w", pages+1, err)
		}
		var page RegistryPage
		err = json.NewDecoder(io.LimitReader(resp.Body, 32<<20)).Decode(&page)
		resp.Body.Close()
		if err != nil {
			return out, pages, fmt.Errorf("decode registry page %d: %w", pages+1, err)
		}
		pages++
		for _, m := range page.Modules {
			if m.ID == "" {
				m.ID = m.Namespace + "/" + m.Name + "/" + m.Provider
			}
			if m.Provider == "" {
				m.Provider = cfg.Provider
			}
			id := registryID(m)
			if seen[id] {
				continue
			}
			seen[id] = true
			out = append(out, m)
		}
		if logf != nil {
			logf("registry: page %d, %d modules (%d total)", pages, len(page.Modules), len(out))
		}
		next = page.Meta.NextURL
		if next != "" {
			u, err := url.Parse(next)
			if err == nil && !u.IsAbs() {
				base, _ := url.Parse(pageURL)
				next = base.ResolveReference(u).String()
			}
		}
	}
	return out, pages, nil
}

// RegistryModuleByID fetches the latest version of one registry module.
func (c *Client) RegistryModuleByID(ctx context.Context, id string) (RegistryModule, error) {
	raw := registryBase + "/v1/modules/" + strings.Trim(id, "/")
	resp, err := c.request(ctx, raw, "application/json")
	if err != nil {
		return RegistryModule{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
		return RegistryModule{}, fmt.Errorf("%s: %s", raw, resp.Status)
	}
	var m RegistryModule
	if err := json.NewDecoder(io.LimitReader(resp.Body, 16<<20)).Decode(&m); err != nil {
		return RegistryModule{}, err
	}
	if m.ID == "" {
		m.ID = m.Namespace + "/" + m.Name + "/" + m.Provider
	}
	return m, nil
}

// ResolveRegistryDownload follows the registry download endpoint to the exact
// git commit behind a module version.
func (c *Client) ResolveRegistryDownload(ctx context.Context, m RegistryModule) (GitModule, error) {
	raw := fmt.Sprintf("%s/v1/modules/%s/%s/%s/%s/download",
		registryBase, url.PathEscape(m.Namespace), url.PathEscape(m.Name), url.PathEscape(m.Provider), url.PathEscape(m.Version))
	resp, err := c.request(ctx, raw, "application/json")
	if err != nil {
		return GitModule{}, err
	}
	io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return GitModule{}, fmt.Errorf("registry download %s: %s", m.ID, resp.Status)
	}
	get := resp.Header.Get("X-Terraform-Get")
	owner, repo, ref, ok := parseGitHubSource(get)
	if !ok {
		return GitModule{}, fmt.Errorf("registry download %s: unsupported source %q", m.ID, get)
	}
	sha, err := c.resolveCommit(ctx, owner, repo, ref)
	if err != nil {
		return GitModule{}, err
	}
	return GitModule{Registry: &m, Owner: owner, Repo: repo, Ref: sha}, nil
}

// ResolveSeed turns a configured GitHub seed into an exact commit.
func (c *Client) ResolveSeed(ctx context.Context, seed GitHubSeed) (GitModule, error) {
	parts := strings.Split(strings.Trim(seed.Repo, "/"), "/")
	if len(parts) != 2 {
		return GitModule{}, fmt.Errorf("github seed %q: expected owner/repo", seed.Repo)
	}
	ref := seed.Ref
	if ref == "" {
		ref = "HEAD"
	}
	sha, err := c.resolveCommit(ctx, parts[0], parts[1], ref)
	if err != nil {
		return GitModule{}, err
	}
	s := seed
	return GitModule{Seed: &s, Owner: parts[0], Repo: parts[1], Ref: sha}, nil
}

func (c *Client) resolveCommit(ctx context.Context, owner, repo, ref string) (string, error) {
	if isHexSHA(ref) {
		return ref, nil
	}
	raw := fmt.Sprintf("%s/repos/%s/%s/commits/%s", githubAPI, url.PathEscape(owner), url.PathEscape(repo), url.PathEscape(ref))
	var body struct {
		SHA string `json:"sha"`
	}
	if err := c.getJSON(ctx, raw, &body); err != nil {
		return "", fmt.Errorf("resolve %s/%s@%s: %w", owner, repo, ref, err)
	}
	if body.SHA == "" {
		return "", fmt.Errorf("resolve %s/%s@%s: empty sha", owner, repo, ref)
	}
	return body.SHA, nil
}

func (c *Client) getJSON(ctx context.Context, rawURL string, dst any) error {
	resp, err := c.request(ctx, rawURL, "application/vnd.github+json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
		return fmt.Errorf("%s: %s", rawURL, resp.Status)
	}
	return json.NewDecoder(io.LimitReader(resp.Body, 16<<20)).Decode(dst)
}

// DownloadArchive fetches a GitHub tarball for an immutable ref.
func (c *Client) DownloadArchive(ctx context.Context, owner, repo, ref string, maxBytes int64) ([]byte, error) {
	raw := fmt.Sprintf("%s/%s/%s/tar.gz/%s", githubCodeload, url.PathEscape(owner), url.PathEscape(repo), url.PathEscape(ref))
	resp, err := c.request(ctx, raw, "application/octet-stream")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
		return nil, fmt.Errorf("archive %s/%s@%s: %s", owner, repo, ref, resp.Status)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("archive %s/%s@%s exceeds %d bytes", owner, repo, ref, maxBytes)
	}
	return data, nil
}

// parseGitHubSource recognizes the forms used by X-Terraform-Get.
func parseGitHubSource(raw string) (owner, repo, ref string, ok bool) {
	s := strings.TrimSpace(raw)
	s = strings.TrimPrefix(s, "git::")
	if strings.HasPrefix(s, "git@github.com:") {
		s = "https://github.com/" + strings.TrimPrefix(s, "git@github.com:")
	}
	u, err := url.Parse(s)
	if err != nil || (u.Host != "github.com" && u.Host != "www.github.com") {
		return "", "", "", false
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 {
		return "", "", "", false
	}
	owner = parts[0]
	repo = strings.TrimSuffix(parts[1], ".git")
	ref = u.Query().Get("ref")
	if repo == "" {
		return "", "", "", false
	}
	if ref == "" {
		// Older registry responses use /archive/<ref>.zip.
		if len(parts) >= 4 && parts[2] == "archive" {
			ref = strings.TrimSuffix(strings.TrimSuffix(parts[3], ".zip"), ".tar.gz")
		}
	}
	if ref == "" {
		return "", "", "", false
	}
	return owner, repo, ref, true
}

func isHexSHA(s string) bool {
	if len(s) != 40 {
		return false
	}
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}
