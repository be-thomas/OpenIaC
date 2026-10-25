package ir

import (
	"fmt"
	"sort"
	"strings"
)

// New returns an empty graph ready to be populated.
func New(source string) *Graph {
	return &Graph{
		Version:   Version,
		Source:    source,
		Resources: []*Resource{},
		Edges:     []Edge{},
	}
}

// AddEdge records that from depends on to. Duplicate edges are ignored and
// self-edges are dropped.
func (g *Graph) AddEdge(from, to string, kind EdgeKind) {
	if from == "" || to == "" || from == to {
		return
	}
	for _, e := range g.Edges {
		if e.From == from && e.To == to {
			return
		}
	}
	g.Edges = append(g.Edges, Edge{From: from, To: to, Kind: kind})
}

// HasNode reports whether id names a node in the graph.
func (g *Graph) HasNode(id string) bool {
	if g.ResourceByID(id) != nil {
		return true
	}
	for _, v := range g.Variables {
		if id == "var."+v.Name {
			return true
		}
	}
	for _, l := range g.Locals {
		if id == "local."+l.Name {
			return true
		}
	}
	for _, m := range g.Modules {
		if id == "module."+m.Name {
			return true
		}
	}
	return false
}

// ResourceByID returns the resource or data source with the given address.
func (g *Graph) ResourceByID(id string) *Resource {
	for _, r := range g.Resources {
		if r.ID == id {
			return r
		}
	}
	return nil
}

// Dependencies returns the addresses that id directly depends on.
func (g *Graph) Dependencies(id string) []string {
	var deps []string
	for _, e := range g.Edges {
		if e.From == id {
			deps = append(deps, e.To)
		}
	}
	sort.Strings(deps)
	return deps
}

// Order returns resources in dependency order: a resource always appears after
// everything it references. The order is deterministic (ties are broken by
// address) so repeated runs produce identical output.
func (g *Graph) Order() ([]*Resource, error) {
	byID := make(map[string]*Resource, len(g.Resources))
	for _, r := range g.Resources {
		byID[r.ID] = r
	}

	indegree := make(map[string]int, len(byID))
	dependents := make(map[string][]string)
	for id := range byID {
		indegree[id] = 0
	}
	for _, e := range g.Edges {
		if _, ok := byID[e.From]; !ok {
			continue
		}
		if _, ok := byID[e.To]; !ok {
			continue
		}
		indegree[e.From]++
		dependents[e.To] = append(dependents[e.To], e.From)
	}

	ready := make([]string, 0, len(byID))
	for id, n := range indegree {
		if n == 0 {
			ready = append(ready, id)
		}
	}
	sort.Strings(ready)

	ordered := make([]*Resource, 0, len(byID))
	for len(ready) > 0 {
		id := ready[0]
		ready = ready[1:]
		ordered = append(ordered, byID[id])

		next := append([]string(nil), dependents[id]...)
		sort.Strings(next)
		for _, dep := range next {
			indegree[dep]--
			if indegree[dep] == 0 {
				ready = append(ready, dep)
			}
		}
		sort.Strings(ready)
	}

	if len(ordered) != len(byID) {
		cycle := findCycle(byID, g.Edges)
		if len(cycle) > 0 {
			return nil, fmt.Errorf("dependency cycle: %s", strings.Join(cycle, " -> "))
		}
		return nil, fmt.Errorf("dependency cycle detected")
	}
	return ordered, nil
}

// findCycle returns one cycle in the graph, if any, for error reporting.
func findCycle(byID map[string]*Resource, edges []Edge) []string {
	adj := make(map[string][]string)
	for _, e := range edges {
		if _, ok := byID[e.From]; !ok {
			continue
		}
		if _, ok := byID[e.To]; !ok {
			continue
		}
		adj[e.From] = append(adj[e.From], e.To)
	}

	const (
		white = 0
		gray  = 1
		black = 2
	)
	color := make(map[string]int, len(byID))
	stack := make([]string, 0, len(byID))
	var cycle []string

	var visit func(string) bool
	visit = func(id string) bool {
		color[id] = gray
		stack = append(stack, id)
		neighbors := append([]string(nil), adj[id]...)
		sort.Strings(neighbors)
		for _, next := range neighbors {
			switch color[next] {
			case gray:
				start := 0
				for i, s := range stack {
					if s == next {
						start = i
						break
					}
				}
				cycle = append(append([]string(nil), stack[start:]...), next)
				return true
			case white:
				if visit(next) {
					return true
				}
			}
		}
		stack = stack[:len(stack)-1]
		color[id] = black
		return false
	}

	ids := make([]string, 0, len(byID))
	for id := range byID {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		if color[id] == white && visit(id) {
			return cycle
		}
	}
	return nil
}

// Stats summarises the graph for CLI output.
func (g *Graph) Stats() (resources, data, variables, outputs, edges int) {
	for _, r := range g.Resources {
		if r.Kind == KindData {
			data++
		} else {
			resources++
		}
	}
	return resources, data, len(g.Variables), len(g.Outputs), len(g.Edges)
}
