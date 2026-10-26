package collect

// PolicyViolation explains why a candidate cannot be committed.
type PolicyViolation struct {
	Kind   string
	Path   string
	Detail string
}

func (v PolicyViolation) String() string {
	if v.Path == "" {
		return v.Kind + ": " + v.Detail
	}
	return v.Kind + " in " + v.Path + ": " + v.Detail
}

// ScanFiles returns the policy violations for a set of Terraform files. It is
// a thin wrapper around AnalyzeFiles retained for focused tests.
func ScanFiles(files map[string][]byte) []PolicyViolation {
	a, err := AnalyzeFiles(files)
	if err != nil {
		return []PolicyViolation{{Kind: "parse", Detail: err.Error()}}
	}
	return a.Violations
}
