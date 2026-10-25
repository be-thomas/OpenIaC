package ir_test

import (
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
