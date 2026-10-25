package collect

import (
	"regexp"
	"sort"
	"strings"
)

var featurePatterns = map[string]*regexp.Regexp{
	"backend":            regexp.MustCompile(`(?m)^\s*backend\s+"`),
	"check":              regexp.MustCompile(`(?m)^\s*check\s+"`),
	"conditionals":       regexp.MustCompile(`(?m)\b[a-z_][a-z0-9_]*\s*\?\s*[^:\n]+\s*:`),
	"count":              regexp.MustCompile(`(?m)\bcount\s*=`),
	"data":               regexp.MustCompile(`(?m)^\s*data\s+"`),
	"depends_on":         regexp.MustCompile(`(?m)\bdepends_on\s*=`),
	"dynamic":            regexp.MustCompile(`(?m)^\s*dynamic\s+"`),
	"for_each":           regexp.MustCompile(`(?m)\bfor_each\s*=`),
	"for_expressions":    regexp.MustCompile(`\bfor\s+[A-Za-z_][A-Za-z0-9_]*\s+in\b`),
	"functions":          regexp.MustCompile(`\b(?:can|cidrsubnet|coalesce|file|filebase64|flatten|format|join|jsondecode|jsonencode|lookup|lower|merge|replace|sha256|split|templatefile|tolist|tomap|toset|try|upper|yamlencode)\s*\(`),
	"heredoc":            regexp.MustCompile(`<<-?[A-Za-z]`),
	"import":             regexp.MustCompile(`(?m)^\s*import\s*\{`),
	"interpolation":      regexp.MustCompile(`\$\{`),
	"locals":             regexp.MustCompile(`(?m)^\s*locals\s*\{`),
	"modules":            regexp.MustCompile(`(?m)^\s*module\s+"`),
	"moved":              regexp.MustCompile(`(?m)^\s*moved\s*\{`),
	"outputs":            regexp.MustCompile(`(?m)^\s*output\s+"`),
	"required_providers": regexp.MustCompile(`(?m)^\s*required_providers\s*\{`),
	"resources":          regexp.MustCompile(`(?m)^\s*resource\s+"`),
	"sensitive":          regexp.MustCompile(`(?m)\bsensitive\s*=\s*true`),
	"terraform":          regexp.MustCompile(`(?m)^\s*terraform\s*\{`),
	"variables":          regexp.MustCompile(`(?m)^\s*variable\s+"`),
}

var (
	resourceTypeRE  = regexp.MustCompile(`(?m)^\s*(?:resource|data)\s+"([a-z0-9]+)_[^"]*"`)
	providerBlockRE = regexp.MustCompile(`(?m)^\s*provider\s+"([a-z0-9]+)"`)
)

// DetectFeatures returns the sorted Terraform features used by a set of files.
func DetectFeatures(files map[string][]byte) []string {
	joined := joinFiles(files)
	var out []string
	for name, re := range featurePatterns {
		if re.MatchString(joined) {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

// DetectProviders returns the sorted provider families used by resources,
// data sources and provider blocks.
func DetectProviders(files map[string][]byte) []string {
	joined := joinFiles(files)
	seen := make(map[string]bool)
	for _, m := range resourceTypeRE.FindAllStringSubmatch(joined, -1) {
		seen[m[1]] = true
	}
	for _, m := range providerBlockRE.FindAllStringSubmatch(joined, -1) {
		seen[m[1]] = true
	}
	out := make([]string, 0, len(seen))
	for p := range seen {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

func joinFiles(files map[string][]byte) string {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	var b strings.Builder
	for _, name := range names {
		b.Write(files[name])
		b.WriteByte('\n')
	}
	return b.String()
}

// ClassifyTier maps features to the corpus tier list.
func ClassifyTier(features []string) string {
	set := stringSet(features)
	if set["modules"] {
		return "modules"
	}
	for _, f := range []string{
		"backend", "check", "conditionals", "count", "data", "depends_on",
		"dynamic", "for_each", "for_expressions", "functions", "heredoc",
		"import", "interpolation", "locals", "moved", "sensitive",
	} {
		if set[f] {
			return "advanced"
		}
	}
	return "core"
}

// SupportLevel is the input-side expectation until output mappings exist.
func SupportLevel(features []string) string {
	basic := map[string]bool{
		"outputs": true, "providers": true, "required_providers": true,
		"resources": true, "terraform": true, "variables": true,
	}
	for _, f := range features {
		if !basic[f] {
			return "partial"
		}
	}
	return "full"
}

func stringSet(values []string) map[string]bool {
	out := make(map[string]bool, len(values))
	for _, v := range values {
		out[v] = true
	}
	return out
}
