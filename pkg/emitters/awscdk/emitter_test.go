package awscdk_test

import (
	"strings"
	"testing"

	"github.com/be-thomas/openiac/pkg/emitters"
	"github.com/be-thomas/openiac/pkg/emitters/awscdk"
	"github.com/be-thomas/openiac/pkg/ir"
)

func testGraph() *ir.Graph {
	graph := ir.New("main.tf")
	graph.Resources = []*ir.Resource{{
		ID:         "aws_s3_bucket.logs",
		Kind:       ir.KindResource,
		Type:       "aws_s3_bucket",
		Name:       "logs",
		Provider:   "aws",
		Attributes: map[string]ir.Value{"bucket": {Kind: ir.KindString, Literal: "logs"}},
		File:       "main.tf",
		Line:       1,
	}}
	return graph
}

func emit(t *testing.T, graph *ir.Graph) *emitters.Result {
	t.Helper()
	result, err := awscdk.New().Emit(graph, emitters.Options{})
	if err != nil {
		t.Fatalf("Emit: %v", err)
	}
	return result
}

func TestUnsupportedConstructsWarnWithoutChangingOutput(t *testing.T) {
	plain := emit(t, testGraph())

	graph := testGraph()
	graph.Terraform = &ir.Terraform{
		RequiredVersion: ">= 1.6.2",
		Backend:         &ir.Block{Type: "backend", Labels: []string{"s3"}, File: "main.tf", Line: 3},
		Cloud:           &ir.Block{Type: "cloud", File: "main.tf", Line: 9},
		ProviderMeta:    []ir.Block{{Type: "provider_meta", Labels: []string{"aws"}, File: "main.tf", Line: 13}},
		File:            "main.tf",
		Line:            2,
	}
	graph.Preserved = []ir.Block{
		{Type: "moved", File: "main.tf", Line: 20},
		{Type: "check", Labels: []string{"health"}, File: "main.tf", Line: 24},
	}
	graph.Resources[0].Blocks = []ir.Block{{Type: "precondition", Line: 6}}
	extended := emit(t, graph)

	if len(plain.Files) != len(extended.Files) {
		t.Fatalf("files = %d vs %d", len(plain.Files), len(extended.Files))
	}
	for i := range plain.Files {
		if plain.Files[i] != extended.Files[i] {
			t.Fatalf("generated %s changed when unsupported constructs were added:\n--- plain\n%s\n--- extended\n%s",
				plain.Files[i].Path, plain.Files[i].Content, extended.Files[i].Content)
		}
	}

	for _, code := range []string{
		"W_PRESERVED_BLOCK",
		"W_TERRAFORM_BACKEND",
		"W_TERRAFORM_CLOUD",
		"W_TERRAFORM_PROVIDER_META",
		"W_CONDITION",
	} {
		if !hasWarningCode(extended.Warnings, code) {
			t.Errorf("missing %s in warnings:\n%s", code, strings.Join(extended.Warnings, "\n"))
		}
	}
}

func TestBaselineGraphHasNoCodedWarnings(t *testing.T) {
	result := emit(t, testGraph())
	for _, warning := range result.Warnings {
		if strings.HasPrefix(warning, "W_") {
			t.Errorf("unexpected coded warning: %s", warning)
		}
	}
}

func hasWarningCode(warnings []string, code string) bool {
	for _, warning := range warnings {
		if strings.HasPrefix(warning, code+":") {
			return true
		}
	}
	return false
}
