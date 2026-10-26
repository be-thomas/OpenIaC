package collect

import "sort"

// DetectFeatures returns the sorted Terraform features used by a set of files.
// It parses the files; unparseable input has no reliable feature set.
func DetectFeatures(files map[string][]byte) []string {
	a, err := AnalyzeFiles(files)
	if err != nil {
		return nil
	}
	return a.Features
}

// DetectProviders returns the sorted provider families used by resources,
// data sources and provider blocks.
func DetectProviders(files map[string][]byte) []string {
	a, err := AnalyzeFiles(files)
	if err != nil {
		return nil
	}
	return a.Providers
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

func sortedStrings(values []string) []string {
	out := append([]string(nil), values...)
	sort.Strings(out)
	return out
}
