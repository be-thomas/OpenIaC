package ir_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/be-thomas/openiac/pkg/ir"
)

func graphWith(ids ...string) *ir.Graph {
	graph := ir.New("test")
	for _, id := range ids {
		graph.Resources = append(graph.Resources, &ir.Resource{
			ID:   id,
			Kind: ir.KindResource,
			Type: id,
			Name: id,
		})
	}
	return graph
}

func TestAddEdgeDeduplicatesAndSkipsSelf(t *testing.T) {
	graph := graphWith("aws_vpc.main", "aws_subnet.public")
	graph.AddEdge("aws_subnet.public", "aws_vpc.main", ir.EdgeReference)
	graph.AddEdge("aws_subnet.public", "aws_vpc.main", ir.EdgeDependsOn)
	graph.AddEdge("aws_vpc.main", "aws_vpc.main", ir.EdgeReference)

	if len(graph.Edges) != 1 {
		t.Fatalf("edges = %d, want 1: %#v", len(graph.Edges), graph.Edges)
	}
}

func TestOrderRespectsDependencies(t *testing.T) {
	graph := graphWith("aws_instance.z", "aws_s3_bucket.a", "aws_vpc.main")
	graph.AddEdge("aws_instance.z", "aws_vpc.main", ir.EdgeReference)

	ordered, err := graph.Order()
	if err != nil {
		t.Fatalf("Order: %v", err)
	}
	got := make([]string, 0, len(ordered))
	for _, resource := range ordered {
		got = append(got, resource.ID)
	}
	want := []string{"aws_s3_bucket.a", "aws_vpc.main", "aws_instance.z"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("order = %v, want %v", got, want)
	}
}

func TestOrderIsDeterministic(t *testing.T) {
	build := func() *ir.Graph {
		graph := graphWith("c.c", "a.a", "b.b", "d.d")
		graph.AddEdge("d.d", "b.b", ir.EdgeReference)
		return graph
	}
	first, err := build().Order()
	if err != nil {
		t.Fatal(err)
	}
	second, err := build().Order()
	if err != nil {
		t.Fatal(err)
	}
	for i := range first {
		if first[i].ID != second[i].ID {
			t.Fatalf("order unstable at %d: %s vs %s", i, first[i].ID, second[i].ID)
		}
	}
}

func TestOrderDetectsCycle(t *testing.T) {
	graph := graphWith("a.a", "b.b", "c.c")
	graph.AddEdge("a.a", "b.b", ir.EdgeReference)
	graph.AddEdge("b.b", "c.c", ir.EdgeReference)
	graph.AddEdge("c.c", "a.a", ir.EdgeReference)

	_, err := graph.Order()
	if err == nil {
		t.Fatal("expected a cycle error")
	}
	if !strings.Contains(err.Error(), "dependency cycle") {
		t.Fatalf("error = %q, want dependency cycle", err)
	}
}

func TestDependencies(t *testing.T) {
	graph := graphWith("a.a", "b.b", "c.c")
	graph.AddEdge("a.a", "c.c", ir.EdgeReference)
	graph.AddEdge("a.a", "b.b", ir.EdgeDependsOn)

	got := graph.Dependencies("a.a")
	if len(got) != 2 || got[0] != "b.b" || got[1] != "c.c" {
		t.Fatalf("dependencies = %v, want [b.b c.c]", got)
	}
	if len(graph.Dependencies("b.b")) != 0 {
		t.Fatalf("b.b should have no dependencies")
	}
}

func TestHasNode(t *testing.T) {
	graph := graphWith("aws_vpc.main")
	graph.Variables = append(graph.Variables, &ir.Variable{Name: "region"})
	graph.Locals = append(graph.Locals, &ir.Local{Name: "name"})
	graph.Modules = append(graph.Modules, &ir.Module{Name: "vpc"})

	for _, id := range []string{"aws_vpc.main", "var.region", "local.name", "module.vpc"} {
		if !graph.HasNode(id) {
			t.Errorf("HasNode(%q) = false", id)
		}
	}
	if graph.HasNode("var.missing") {
		t.Error("HasNode(var.missing) = true")
	}
}

func TestStats(t *testing.T) {
	graph := graphWith("aws_vpc.main")
	graph.Resources = append(graph.Resources, &ir.Resource{ID: "data.aws_ami.ubuntu", Kind: ir.KindData})
	graph.Variables = append(graph.Variables, &ir.Variable{Name: "region"})
	graph.Outputs = append(graph.Outputs, &ir.Output{Name: "id"})
	graph.AddEdge("aws_vpc.main", "data.aws_ami.ubuntu", ir.EdgeReference)

	resources, data, variables, outputs, edges := graph.Stats()
	if resources != 1 || data != 1 || variables != 1 || outputs != 1 || edges != 1 {
		t.Fatalf("stats = %d/%d/%d/%d/%d, want 1/1/1/1/1", resources, data, variables, outputs, edges)
	}
}

func TestNewDocumentShape(t *testing.T) {
	if ir.Version != "0.2" {
		t.Fatalf("ir.Version = %q, want the 0.2 schema", ir.Version)
	}
	graph := ir.New("main.tf")
	if graph.Version != ir.Version {
		t.Fatalf("version = %q, want %q", graph.Version, ir.Version)
	}

	encoded, err := json.Marshal(graph)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	// An empty graph must still serialize the fields consumers index into, and
	// must not emit empty optional collections.
	want := `{"version":"0.2","source":"main.tf","resources":[],"edges":[]}`
	if string(encoded) != want {
		t.Fatalf("Marshal = %s, want %s", encoded, want)
	}
}

func TestJSONRoundTrip(t *testing.T) {
	nullable := false
	graph := ir.New("main.tf")
	graph.Terraform = &ir.Terraform{
		RequiredVersion: ">= 1.6.2",
		RequiredProviders: map[string]ir.Value{
			"aws": {Kind: ir.KindMap, Fields: map[string]ir.Value{
				"source":  {Kind: ir.KindString, Literal: "hashicorp/aws"},
				"version": {Kind: ir.KindString, Literal: "~> 5.0"},
			}},
		},
		Experiments: []string{"module_variable_optional_attrs"},
		Backend:     &ir.Block{Type: "backend", Labels: []string{"s3"}, Attributes: map[string]ir.Value{"bucket": {Kind: ir.KindString, Literal: "state"}}},
		Cloud:       &ir.Block{Type: "cloud", Attributes: map[string]ir.Value{"organization": {Kind: ir.KindString, Literal: "example"}}},
		ProviderMeta: []ir.Block{{
			Type:   "provider_meta",
			Labels: []string{"aws"},
		}},
		Attributes: map[string]ir.Value{"future": {Kind: ir.KindString, Literal: "x"}},
		File:       "main.tf",
		Line:       1,
	}
	graph.Resources = []*ir.Resource{{
		ID:            "aws_instance.web",
		Kind:          ir.KindResource,
		Type:          "aws_instance",
		Name:          "web",
		Provider:      "aws",
		ProviderAlias: "aws.west",
		DependsOn:     []string{"aws_vpc.main"},
		Attributes: map[string]ir.Value{
			"ami":            {Kind: ir.KindString, Literal: "ami-123"},
			"count":          {Kind: ir.KindNumber, Literal: int64(2)},
			"enabled":        {Kind: ir.KindBool, Literal: true},
			"missing":        {Kind: ir.KindNull},
			"subnets":        {Kind: ir.KindList, Items: []ir.Value{{Kind: ir.KindString, Literal: "a"}}},
			"vpc_id":         {Kind: ir.KindExpr, Expr: "aws_vpc.main.id", Raw: "aws_vpc.main.id", Refs: []string{"aws_vpc.main"}},
			"tags":           {Kind: ir.KindMap, Fields: map[string]ir.Value{"env": {Kind: ir.KindString, Literal: "dev"}}},
			"nested_empty":   {Kind: ir.KindMap, Fields: map[string]ir.Value{}},
			"empty_list":     {Kind: ir.KindList},
			"unset_optional": {},
		},
		Blocks: []ir.Block{{Type: "lifecycle", Attributes: map[string]ir.Value{"destroy": {Kind: ir.KindBool, Literal: false}}}},
		File:   "main.tf",
		Line:   10,
	}}
	graph.Variables = []*ir.Variable{{
		Name:        "config",
		Type:        "object({ retries = optional(number, 3) })",
		Default:     &ir.Value{Kind: ir.KindString, Literal: "x"},
		Description: "config",
		Sensitive:   true,
		Nullable:    &nullable,
		Ephemeral:   true,
		Deprecated:  "use config2",
		Attributes:  map[string]ir.Value{"custom": {Kind: ir.KindString, Literal: "y"}},
		Blocks:      []ir.Block{{Type: "validation", Attributes: map[string]ir.Value{"condition": {Kind: ir.KindString, Literal: "true"}}}},
		File:        "main.tf",
		Line:        30,
	}}
	graph.Locals = []*ir.Local{{Name: "name", Value: ir.Value{Kind: ir.KindString, Literal: "app"}, File: "main.tf", Line: 40}}
	graph.Outputs = []*ir.Output{{
		Name:        "ip",
		Value:       ir.Value{Kind: ir.KindExpr, Expr: "aws_instance.web.private_ip", Raw: "aws_instance.web.private_ip"},
		Description: "private ip",
		Sensitive:   true,
		Type:        "string",
		Ephemeral:   true,
		Deprecated:  "use ips",
		Attributes:  map[string]ir.Value{"custom": {Kind: ir.KindString, Literal: "z"}},
		Blocks:      []ir.Block{{Type: "precondition"}},
		File:        "main.tf",
		Line:        50,
	}}
	graph.Modules = []*ir.Module{{
		Name:       "vpc",
		Source:     "terraform-aws-modules/vpc/aws",
		Version:    "5.0.0",
		Attributes: map[string]ir.Value{"cidr": {Kind: ir.KindString, Literal: "10.0.0.0/16"}},
		File:       "main.tf",
		Line:       60,
	}}
	graph.Providers = []*ir.Provider{{
		Name:       "aws",
		Alias:      "west",
		Attributes: map[string]ir.Value{"region": {Kind: ir.KindString, Literal: "us-west-2"}},
		Blocks:     []ir.Block{{Type: "default_tags"}},
		File:       "main.tf",
		Line:       70,
	}}
	graph.Preserved = []ir.Block{{
		Type: "moved",
		File: "main.tf",
		Line: 80,
		Attributes: map[string]ir.Value{
			"from": {Kind: ir.KindExpr, Expr: "aws_vpc.old", Raw: "aws_vpc.old"},
			"to":   {Kind: ir.KindExpr, Expr: "aws_vpc.main", Raw: "aws_vpc.main"},
		},
	}, {
		Type: "stray_argument",
		Body: &ir.Value{Kind: ir.KindNumber, Literal: int64(42)},
		File: "main.tf",
		Line: 90,
	}}
	graph.Features = []string{"moved", "terraform_data"}
	graph.MinVersion = "1.4.0"
	graph.AddEdge("aws_instance.web", "aws_vpc.main", ir.EdgeDependsOn)

	first, err := json.Marshal(graph)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded ir.Graph
	if err := json.Unmarshal(first, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	// Compare the encoding rather than the Go values: JSON numbers decode to
	// float64, so a literal int64 would never be DeepEqual even though the
	// document is identical.
	second, err := json.Marshal(&decoded)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(first) != string(second) {
		t.Fatalf("round trip changed the document:\nfirst:  %s\nsecond: %s", first, second)
	}
}

func TestDecodeLegacyDocument(t *testing.T) {
	// Schema 0.1 had no terraform, preserved, features or min_version fields.
	legacy := `{
  "version": "0.1",
  "source": "main.tf",
  "resources": [{"id": "aws_vpc.main", "kind": "resource", "type": "aws_vpc", "name": "main"}],
  "edges": []
}`
	var graph ir.Graph
	if err := json.Unmarshal([]byte(legacy), &graph); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if graph.Version != "0.1" {
		t.Errorf("version = %q", graph.Version)
	}
	if len(graph.Resources) != 1 || graph.Resources[0].ID != "aws_vpc.main" {
		t.Fatalf("resources = %v", graph.Resources)
	}
	if graph.Terraform != nil || len(graph.Preserved) != 0 || graph.MinVersion != "" {
		t.Errorf("0.2 fields should stay unset: %#v", graph)
	}
	if !graph.HasNode("aws_vpc.main") {
		t.Error("HasNode(aws_vpc.main) = false")
	}
}

func TestOrderAndStatsIgnorePreserved(t *testing.T) {
	graph := graphWith("aws_vpc.main", "aws_subnet.public")
	graph.AddEdge("aws_subnet.public", "aws_vpc.main", ir.EdgeReference)
	graph.Preserved = []ir.Block{
		{Type: "moved"},
		{Type: "import"},
		{Type: "removed"},
		{Type: "check", Labels: []string{"health"}},
		{Type: "ephemeral", Labels: []string{"random_password", "db"}},
		{Type: "action", Labels: []string{"aws_lb_target_group", "register"}},
	}

	ordered, err := graph.Order()
	if err != nil {
		t.Fatalf("Order: %v", err)
	}
	if len(ordered) != 2 || ordered[0].ID != "aws_vpc.main" || ordered[1].ID != "aws_subnet.public" {
		t.Fatalf("order = %v", ordered)
	}

	resources, data, _, _, edges := graph.Stats()
	if resources != 2 || data != 0 || edges != 1 {
		t.Fatalf("stats = %d/%d/%d", resources, data, edges)
	}

	for _, id := range []string{"moved", "check.health", "random_password.db", "aws_lb_target_group.register"} {
		if graph.HasNode(id) {
			t.Errorf("HasNode(%q) = true, preserved blocks are not nodes", id)
		}
	}
}

func TestValueIsLiteral(t *testing.T) {
	cases := []struct {
		name  string
		value ir.Value
		want  bool
	}{
		{"string", ir.Value{Kind: ir.KindString, Literal: "a"}, true},
		{"number", ir.Value{Kind: ir.KindNumber, Literal: 1.5}, true},
		{"bool", ir.Value{Kind: ir.KindBool, Literal: true}, true},
		{"null", ir.Value{Kind: ir.KindNull}, true},
		{"expr", ir.Value{Kind: ir.KindExpr, Expr: "var.x"}, false},
		{"empty list", ir.Value{Kind: ir.KindList}, true},
		{"empty map", ir.Value{Kind: ir.KindMap}, true},
		{"literal list", ir.Value{Kind: ir.KindList, Items: []ir.Value{{Kind: ir.KindString, Literal: "a"}}}, true},
		{"mixed list", ir.Value{Kind: ir.KindList, Items: []ir.Value{{Kind: ir.KindExpr, Expr: "var.x"}}}, false},
		{"literal map", ir.Value{Kind: ir.KindMap, Fields: map[string]ir.Value{"a": {Kind: ir.KindNumber, Literal: 1}}}, true},
		{"mixed map", ir.Value{Kind: ir.KindMap, Fields: map[string]ir.Value{"a": {Kind: ir.KindExpr, Expr: "var.x"}}}, false},
		{"zero value", ir.Value{}, false},
	}
	for _, c := range cases {
		if got := c.value.IsLiteral(); got != c.want {
			t.Errorf("%s: IsLiteral() = %v, want %v", c.name, got, c.want)
		}
	}
}
