package tfversion

import (
	"fmt"
	"sort"
	"strings"

	"github.com/be-thomas/openiac/pkg/ir"
)

// Feature describes one Terraform language feature that OpenIaC detects in an
// IR graph.
type Feature struct {
	// Name is the stable identifier reported in Graph.Features.
	Name string
	// Since is the canonical Terraform release that introduced the feature.
	Since string
	// Description is a one-line human summary.
	Description string
	// Reliable reports whether detection is strong enough to move MinVersion.
	// Unreliable features are still reported, but never raise the version.
	Reliable bool
	// Detect inspects a graph for the feature.
	Detect func(*ir.Graph) bool
}

type feature struct {
	name        string
	since       string
	description string
	reliable    bool
	detect      func(*scan) bool
}

// registry is the single source of truth for feature detection. Rows are added
// only for constructs whose presence is unambiguous in the IR; see
// docs/TERRAFORM-VERSIONS.md for the full rationale and the features that are
// deliberately not tracked.
var registry = []feature{
	{
		name: "action", since: "1.14.0", reliable: true,
		description: "Provider-defined action blocks",
		detect:      func(s *scan) bool { return len(s.preserved["action"]) > 0 },
	},
	{
		name: "action_lifecycle_events", since: "1.16.0", reliable: false,
		description: "Action failure modes and destroy-time events",
		detect: func(s *scan) bool {
			return blockSetsAny(s.preserved["action"], "on_failure", "before_destroy", "after_destroy")
		},
	},
	{
		name: "check", since: "1.5.0", reliable: true,
		description: "check blocks",
		detect:      func(s *scan) bool { return len(s.preserved["check"]) > 0 },
	},
	{
		name: "cloud", since: "1.1.0", reliable: true,
		description: "HCP Terraform / Terraform Cloud integration block",
		detect:      func(s *scan) bool { return s.cloud },
	},
	{
		name: "conditions", since: "1.2.0", reliable: true,
		description: "precondition and postcondition blocks",
		detect:      hasConditions,
	},
	{
		name: "deprecated", since: "1.15.0", reliable: true,
		description: "Deprecation messages on variables and outputs",
		detect: func(s *scan) bool {
			for _, v := range s.variables {
				if v.Deprecated != "" {
					return true
				}
			}
			for _, o := range s.outputs {
				if o.Deprecated != "" {
					return true
				}
			}
			return false
		},
	},
	{
		name: "ephemeral_resources", since: "1.10.0", reliable: true,
		description: "ephemeral resource blocks",
		detect:      func(s *scan) bool { return len(s.preserved["ephemeral"]) > 0 },
	},
	{
		name: "ephemeral_values", since: "1.10.0", reliable: true,
		description: "Ephemeral variables and outputs",
		detect: func(s *scan) bool {
			for _, v := range s.variables {
				if v.Ephemeral {
					return true
				}
			}
			for _, o := range s.outputs {
				if o.Ephemeral {
					return true
				}
			}
			return false
		},
	},
	{
		name: "import", since: "1.5.0", reliable: true,
		description: "import blocks",
		detect:      func(s *scan) bool { return len(s.preserved["import"]) > 0 },
	},
	{
		name: "import_expression", since: "1.6.0", reliable: true,
		description: "import id given as an expression",
		detect: func(s *scan) bool {
			for _, b := range s.preserved["import"] {
				if v, ok := b.Attributes["id"]; ok && v.Kind != ir.KindString {
					return true
				}
			}
			return false
		},
	},
	{
		name: "import_for_each", since: "1.7.0", reliable: true,
		description: "import blocks with for_each",
		detect: func(s *scan) bool {
			return anyPreserved(s.preserved["import"], func(b ir.Block) bool {
				_, ok := b.Attributes["for_each"]
				return ok
			})
		},
	},
	{
		name: "import_identity", since: "1.12.0", reliable: true,
		description: "import blocks using identity instead of id",
		detect: func(s *scan) bool {
			// Native HCL spells identity as a block, .tf.json as a property.
			return anyPreserved(s.preserved["import"], func(b ir.Block) bool {
				_, ok := b.Attributes["identity"]
				return ok || hasBlockType(b.Blocks, "identity")
			})
		},
	},
	{
		name: "lifecycle_destroy", since: "1.16.0", reliable: true,
		description: "lifecycle destroy argument",
		detect: func(s *scan) bool {
			return resourceLifecycleField(s.resources, "destroy")
		},
	},
	{
		name: "moved", since: "1.1.0", reliable: true,
		description: "moved blocks",
		detect:      func(s *scan) bool { return len(s.preserved["moved"]) > 0 },
	},
	{
		name: "moved_cross_type", since: "1.8.0", reliable: false,
		description: "moved blocks changing the resource type",
		detect: func(s *scan) bool {
			return anyPreserved(s.preserved["moved"], func(b ir.Block) bool {
				from := addressType(valueText(b.Attributes["from"]))
				to := addressType(valueText(b.Attributes["to"]))
				return from != "" && to != "" && from != to
			})
		},
	},
	{
		name: "moved_module", since: "1.3.0", reliable: true,
		description: "moved blocks that cross module boundaries",
		detect: func(s *scan) bool {
			return anyPreserved(s.preserved["moved"], func(b ir.Block) bool {
				return strings.Contains(valueText(b.Attributes["from"]), "module.") ||
					strings.Contains(valueText(b.Attributes["to"]), "module.")
			})
		},
	},
	{
		name: "nullable", since: "1.1.0", reliable: true,
		description: "nullable variable argument",
		detect: func(s *scan) bool {
			for _, v := range s.variables {
				if v.Nullable != nil {
					return true
				}
			}
			return false
		},
	},
	{
		name: "optional_attributes", since: "1.3.0", reliable: true,
		description: "optional object type attributes",
		detect: func(s *scan) bool {
			for _, v := range s.variables {
				if strings.Contains(strings.ReplaceAll(v.Type, " ", ""), "optional(") {
					return true
				}
			}
			return false
		},
	},
	{
		name: "output_type", since: "1.15.0", reliable: true,
		description: "output type constraints",
		detect: func(s *scan) bool {
			for _, o := range s.outputs {
				if o.Type != "" {
					return true
				}
			}
			return false
		},
	},
	{
		name: "provider_requirements_expressions", since: "1.17.0", reliable: true,
		description: "variables and locals in provider requirements",
		detect: func(s *scan) bool {
			for _, v := range s.requiredProviders {
				if hasVariableOrLocalRef(v) {
					return true
				}
			}
			return false
		},
	},
	{
		name: "removed", since: "1.7.0", reliable: true,
		description: "removed blocks",
		detect:      func(s *scan) bool { return len(s.preserved["removed"]) > 0 },
	},
	{
		name: "removed_provisioners", since: "1.9.0", reliable: true,
		description: "destroy-time provisioners in removed blocks",
		detect: func(s *scan) bool {
			return removedProvisioners(s.preserved["removed"])
		},
	},
	{
		name: "replace_triggered_by", since: "1.2.0", reliable: true,
		description: "lifecycle replace_triggered_by argument",
		detect: func(s *scan) bool {
			return resourceLifecycleField(s.resources, "replace_triggered_by")
		},
	},
	{
		name: "terraform_data", since: "1.4.0", reliable: true,
		description: "the built-in terraform_data resource",
		detect: func(s *scan) bool {
			for _, r := range s.resources {
				if r.Type == "terraform_data" {
					return true
				}
			}
			return false
		},
	},
	{
		name: "terraform_data_store", since: "1.16.0", reliable: true,
		description: "terraform_data store blocks",
		detect: func(s *scan) bool {
			for _, r := range s.resources {
				if r.Type != "terraform_data" {
					continue
				}
				if _, ok := r.Attributes["store"]; ok {
					return true
				}
				if hasBlockType(r.Blocks, "store") {
					return true
				}
			}
			return false
		},
	},
}

// Features returns the feature registry in alphabetical order. Each entry
// carries a Detect function that scans the given graph.
func Features() []Feature {
	out := make([]Feature, 0, len(registry))
	for _, f := range registry {
		f := f
		out = append(out, Feature{
			Name:        f.name,
			Since:       f.since,
			Description: f.description,
			Reliable:    f.reliable,
			Detect:      func(g *ir.Graph) bool { return f.detect(scanGraph(g)) },
		})
	}
	return out
}

// Detect returns the names of the registered features present in g, sorted.
func Detect(g *ir.Graph) []string {
	s := scanGraph(g)
	var names []string
	for _, f := range registry {
		if f.detect(s) {
			names = append(names, f.name)
		}
	}
	sort.Strings(names)
	return names
}

// Analyze returns the detected features and the resulting minimum Terraform
// version for g.
func Analyze(g *ir.Graph) (features []string, minVersion string) {
	features = Detect(g)
	required := ""
	if g != nil && g.Terraform != nil {
		required = g.Terraform.RequiredVersion
	}
	return features, MinVersion(features, required)
}

// MinVersion returns the oldest Terraform release that accepts a configuration
// with the given features and required_version constraint.
//
// The result is the newest of: the OpenIaC baseline, the constraint's lower
// bound, and the introduction version of every reliable detected feature.
func MinVersion(features []string, requiredVersion string) string {
	best := Baseline

	if c, err := ParseConstraint(requiredVersion); err == nil && c.LowerBound != "" {
		best = maxVersion(best, c.LowerBound)
	}

	byName := make(map[string]feature, len(registry))
	for _, f := range registry {
		byName[f.name] = f
	}
	for _, name := range features {
		f, ok := byName[name]
		if !ok || !f.reliable {
			continue
		}
		best = maxVersion(best, f.since)
	}
	return best
}

// scan is a single pass over a graph, shared by every detector.
type scan struct {
	preserved         map[string][]ir.Block
	resources         []*ir.Resource
	variables         []*ir.Variable
	outputs           []*ir.Output
	requiredProviders map[string]ir.Value
	cloud             bool
}

func scanGraph(g *ir.Graph) *scan {
	s := &scan{preserved: map[string][]ir.Block{}}
	if g == nil {
		return s
	}

	for _, b := range g.Preserved {
		s.preserved[b.Type] = append(s.preserved[b.Type], b)
	}

	if tf := g.Terraform; tf != nil {
		s.cloud = tf.Cloud != nil
		s.requiredProviders = tf.RequiredProviders
	}

	s.resources = g.Resources
	s.variables = g.Variables
	s.outputs = g.Outputs

	return s
}

// resourceLifecycleField reports whether a resource's own lifecycle block sets
// the named argument.
//
// It looks at resources specifically: a lifecycle block inside a removed block
// also accepts destroy, but that form predates 1.16 and must not move
// min_version.
func resourceLifecycleField(resources []*ir.Resource, field string) bool {
	for _, r := range resources {
		for _, b := range r.Blocks {
			if b.Type != "lifecycle" {
				continue
			}
			if _, ok := b.Attributes[field]; ok {
				return true
			}
		}
		// JSON spells lifecycle as an object property.
		if v, ok := r.Attributes["lifecycle"]; ok && v.Kind == ir.KindMap {
			if _, ok := v.Fields[field]; ok {
				return true
			}
		}
	}
	return false
}

// hasConditions reports whether a resource or output declares a precondition
// or postcondition, as a nested block (native HCL) or as a flat property
// (.tf.json).
func hasConditions(s *scan) bool {
	for _, r := range s.resources {
		if conditionField(r.Attributes, r.Blocks) {
			return true
		}
	}
	for _, o := range s.outputs {
		if conditionField(o.Attributes, o.Blocks) {
			return true
		}
	}
	return false
}

func conditionField(attrs map[string]ir.Value, blocks []ir.Block) bool {
	for _, name := range []string{"precondition", "postcondition"} {
		if _, ok := attrs[name]; ok {
			return true
		}
		for _, b := range blocks {
			if b.Type == name {
				return true
			}
		}
	}
	return false
}

// removedProvisioners reports whether a removed block carries a provisioner.
func removedProvisioners(blocks []ir.Block) bool {
	for _, b := range blocks {
		if _, ok := b.Attributes["provisioner"]; ok {
			return true
		}
		if hasBlockType(b.Blocks, "provisioner") {
			return true
		}
	}
	return false
}

// blockSetsAny reports whether any block, at any depth, sets one of the named
// arguments.
func blockSetsAny(blocks []ir.Block, names ...string) bool {
	for _, b := range blocks {
		for _, name := range names {
			if _, ok := b.Attributes[name]; ok {
				return true
			}
		}
		if blockSetsAny(b.Blocks, names...) {
			return true
		}
	}
	return false
}

func anyPreserved(blocks []ir.Block, pred func(ir.Block) bool) bool {
	for _, b := range blocks {
		if pred(b) {
			return true
		}
	}
	return false
}

func hasBlockType(blocks []ir.Block, blockType string) bool {
	for _, b := range blocks {
		if b.Type == blockType || hasBlockType(b.Blocks, blockType) {
			return true
		}
	}
	return false
}

// hasVariableOrLocalRef reports whether a required_providers declaration reads
// a variable or a local, which is what Terraform 1.17 added. Other references
// are not a version signal: configuration_aliases writes a provider address
// such as aws.west, and that form has been valid since providers gained
// aliases.
func hasVariableOrLocalRef(v ir.Value) bool {
	for _, ref := range v.Refs {
		if strings.HasPrefix(ref, "var.") || strings.HasPrefix(ref, "local.") {
			return true
		}
	}
	for _, item := range v.Items {
		if hasVariableOrLocalRef(item) {
			return true
		}
	}
	for _, field := range v.Fields {
		if hasVariableOrLocalRef(field) {
			return true
		}
	}
	return false
}

// valueText renders a value the way Terraform addresses are written: literal
// strings as-is, expressions in their source form.
func valueText(v ir.Value) string {
	if v.Expr != "" {
		return v.Expr
	}
	if s, ok := v.Literal.(string); ok {
		return s
	}
	if v.Literal != nil {
		return fmt.Sprint(v.Literal)
	}
	return v.Raw
}

// addressType returns the resource type of a Terraform address, approximated by
// the second-to-last segment. It is only used by an unreliable detector.
func addressType(address string) string {
	address = strings.TrimSpace(address)
	parts := strings.Split(address, ".")
	if len(parts) < 2 {
		return ""
	}
	return parts[len(parts)-2]
}
