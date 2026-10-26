package emitters_test

import (
	"strings"
	"testing"

	"github.com/be-thomas/openiac/pkg/emitters"
	"github.com/be-thomas/openiac/pkg/ir"
)

func warnGraph() *ir.Graph {
	graph := ir.New("main.tf")
	graph.Terraform = &ir.Terraform{
		RequiredVersion: ">= 1.6.2",
		Backend: &ir.Block{
			Type:   "backend",
			Labels: []string{"s3"},
			File:   "main.tf",
			Line:   2,
		},
		Cloud: &ir.Block{
			Type: "cloud",
			File: "main.tf",
			Line: 8,
		},
		ProviderMeta: []ir.Block{{
			Type:   "provider_meta",
			Labels: []string{"aws"},
			File:   "main.tf",
			Line:   12,
		}},
		Attributes: map[string]ir.Value{"future_setting": {Kind: ir.KindString, Literal: "x"}},
		Blocks:     []ir.Block{{Type: "future_block", File: "main.tf", Line: 2}},
		File:       "main.tf",
		Line:       1,
	}
	graph.Resources = []*ir.Resource{{
		ID:     "aws_instance.web",
		Kind:   ir.KindResource,
		Type:   "aws_instance",
		Name:   "web",
		Blocks: []ir.Block{{Type: "precondition", Line: 20}},
		File:   "main.tf",
		Line:   18,
	}}
	graph.Outputs = []*ir.Output{{
		Name:       "ip",
		Attributes: map[string]ir.Value{"precondition": {Kind: ir.KindMap}},
		File:       "main.tf",
		Line:       24,
	}}
	graph.Variables = []*ir.Variable{{
		Name:       "password",
		Attributes: map[string]ir.Value{"custom": {Kind: ir.KindString, Literal: "y"}},
		Blocks:     []ir.Block{{Type: "validation", Line: 30}},
		File:       "main.tf",
		Line:       28,
	}}
	graph.Preserved = []ir.Block{
		{Type: "moved", File: "main.tf", Line: 3},
		{Type: "import", File: "main.tf", Line: 10},
		{Type: "import", File: "main.tf", Line: 14},
	}
	return graph
}

func collect(graph *ir.Graph) []string {
	warnings := emitters.NewWarnings()
	emitters.WarnUnsupported(graph, warnings)
	return warnings.List()
}

func TestWarnUnsupportedCodes(t *testing.T) {
	got := collect(warnGraph())

	want := []string{
		`W_PRESERVED_BLOCK: 2 import blocks are preserved but not converted (first at main.tf:10)`,
		`W_PRESERVED_BLOCK: a moved block at main.tf:3 is preserved but not converted`,
		`W_TERRAFORM_BACKEND: terraform backend "s3" at main.tf:2 is not converted; configure state in the target stack instead`,
		`W_TERRAFORM_CLOUD: terraform cloud block at main.tf:8 is not converted`,
		`W_TERRAFORM_PROVIDER_META: provider_meta "aws" at main.tf:12 is not converted`,
		`W_CONDITION: aws_instance.web: precondition block at main.tf:20 is not converted`,
		`W_CONDITION: output.ip: precondition is not converted`,
		`W_VARIABLE_VALIDATION: var.password: validation block at main.tf:30 is not converted (return an error from the stack instead)`,
		`W_UNSUPPORTED_FIELD: terraform: attribute "future_setting" is not converted`,
		`W_UNSUPPORTED_FIELD: terraform: block "future_block" at main.tf:2 is not converted`,
		`W_UNSUPPORTED_FIELD: var.password: attribute "custom" is not converted`,
	}
	if len(got) != len(want) {
		t.Fatalf("warnings = %d, want %d:\n%s", len(got), len(want), strings.Join(got, "\n"))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("warning %d =\n  %q\nwant\n  %q", i, got[i], want[i])
		}
	}
}

func TestWarnUnsupportedArticle(t *testing.T) {
	graph := ir.New("main.tf")
	graph.Preserved = []ir.Block{
		{Type: "action", File: "main.tf", Line: 1},
		{Type: "ephemeral", File: "main.tf", Line: 5},
		{Type: "import", File: "main.tf", Line: 9},
		{Type: "moved", File: "main.tf", Line: 13},
	}

	want := []string{
		`W_PRESERVED_BLOCK: an action block at main.tf:1 is preserved but not converted`,
		`W_PRESERVED_BLOCK: an ephemeral block at main.tf:5 is preserved but not converted`,
		`W_PRESERVED_BLOCK: an import block at main.tf:9 is preserved but not converted`,
		`W_PRESERVED_BLOCK: a moved block at main.tf:13 is preserved but not converted`,
	}
	got := collect(graph)
	if len(got) != len(want) {
		t.Fatalf("warnings = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("warning %d =\n  %q\nwant\n  %q", i, got[i], want[i])
		}
	}
}

func TestWarnUnsupportedIsDeterministic(t *testing.T) {
	first := strings.Join(collect(warnGraph()), "\n")
	for i := 0; i < 5; i++ {
		if got := strings.Join(collect(warnGraph()), "\n"); got != first {
			t.Fatalf("warnings changed between runs:\n%s\n---\n%s", first, got)
		}
	}
}

func TestWarnUnsupportedDeduplicates(t *testing.T) {
	graph := ir.New("main.tf")
	graph.Resources = []*ir.Resource{{
		ID:     "aws_instance.web",
		Kind:   ir.KindResource,
		Type:   "aws_instance",
		Name:   "web",
		Blocks: []ir.Block{{Type: "precondition", Line: 9}, {Type: "precondition", Line: 9}},
		File:   "main.tf",
	}}
	warnings := collect(graph)
	if len(warnings) != 1 {
		t.Fatalf("warnings = %v, want one", warnings)
	}
}

func TestWarnUnsupportedBaselineIsSilent(t *testing.T) {
	graph := ir.New("main.tf")
	graph.Terraform = &ir.Terraform{RequiredVersion: ">= 1.6.2"}
	graph.Resources = []*ir.Resource{{
		ID:         "aws_vpc.main",
		Kind:       ir.KindResource,
		Type:       "aws_vpc",
		Name:       "main",
		Attributes: map[string]ir.Value{"cidr_block": {Kind: ir.KindString, Literal: "10.0.0.0/16"}},
	}}
	if warnings := collect(graph); len(warnings) != 0 {
		t.Fatalf("warnings = %v, want none", warnings)
	}
}

func TestWarnUnsupportedNil(t *testing.T) {
	// Neither argument may panic; a nil graph simply has nothing to report.
	emitters.WarnUnsupported(nil, emitters.NewWarnings())
	emitters.WarnUnsupported(ir.New("main.tf"), nil)
}

func TestWarnUnsupportedUnknownLocation(t *testing.T) {
	graph := ir.New("")
	graph.Preserved = []ir.Block{{Type: "moved"}}
	got := collect(graph)
	if len(got) != 1 || !strings.Contains(got[0], "at an unknown location") {
		t.Fatalf("warnings = %v", got)
	}
}
