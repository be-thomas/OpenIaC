package tfversion_test

import (
	"testing"

	"github.com/be-thomas/openiac/pkg/ir"
	"github.com/be-thomas/openiac/pkg/tfversion"
)

func TestFeaturesRegistry(t *testing.T) {
	features := tfversion.Features()
	if len(features) == 0 {
		t.Fatal("feature registry is empty")
	}

	seen := map[string]bool{}
	previous := ""
	for _, f := range features {
		if seen[f.Name] {
			t.Errorf("duplicate feature %q", f.Name)
		}
		seen[f.Name] = true
		if f.Name < previous {
			t.Errorf("features are not sorted: %q after %q", f.Name, previous)
		}
		previous = f.Name

		if f.Description == "" {
			t.Errorf("%s: empty description", f.Name)
		}
		if f.Detect == nil {
			t.Errorf("%s: nil detector", f.Name)
		}
		if _, ok := tfversion.Canonical(f.Since); !ok {
			t.Errorf("%s: invalid since %q", f.Name, f.Since)
		}
		if tfversion.Compare(f.Since, tfversion.Baseline) < 0 {
			t.Errorf("%s: since %q predates the baseline", f.Name, f.Since)
		}
		if tfversion.Compare(f.Since, tfversion.Latest) > 0 {
			t.Errorf("%s: since %q is newer than the latest known release", f.Name, f.Since)
		}
	}
}

func TestDetectNilGraph(t *testing.T) {
	if got := tfversion.Detect(nil); len(got) != 0 {
		t.Errorf("Detect(nil) = %v, want none", got)
	}
	features, min := tfversion.Analyze(nil)
	if len(features) != 0 || min != tfversion.Baseline {
		t.Errorf("Analyze(nil) = (%v, %q)", features, min)
	}
}

func TestMinVersion(t *testing.T) {
	cases := []struct {
		features []string
		required string
		want     string
	}{
		{nil, "", "1.0.0"},
		{nil, ">= 0.12", "1.0.0"},
		{nil, ">= 1.6.2", "1.6.2"},
		{nil, ">= 1.6.2, < 2.0", "1.6.2"},
		{nil, "nonsense", "1.0.0"},
		{[]string{"check"}, "", "1.5.0"},
		{[]string{"check"}, ">= 1.6.0", "1.6.0"},
		{[]string{"action", "moved"}, ">= 1.0", "1.14.0"},
		{[]string{"unknown_feature"}, "", "1.0.0"},
		// Unreliable detectors never move the minimum.
		{[]string{"moved_cross_type", "action_lifecycle_events"}, "", "1.0.0"},
	}
	for _, c := range cases {
		if got := tfversion.MinVersion(c.features, c.required); got != c.want {
			t.Errorf("MinVersion(%v, %q) = %q, want %q", c.features, c.required, got, c.want)
		}
	}
}

func TestDetectJSONFlatLifecycle(t *testing.T) {
	graph := &ir.Graph{
		Version: ir.Version,
		Resources: []*ir.Resource{{
			ID:   "aws_vpc.main",
			Kind: ir.KindResource,
			Type: "aws_vpc",
			Name: "main",
			// .tf.json spells lifecycle as an object property.
			Attributes: map[string]ir.Value{
				"lifecycle": {
					Kind: ir.KindMap,
					Fields: map[string]ir.Value{
						"destroy":              {Kind: ir.KindBool, Literal: false},
						"replace_triggered_by": {Kind: ir.KindList},
					},
				},
			},
		}},
	}
	features := tfversion.Detect(graph)
	if !contains(features, "lifecycle_destroy") {
		t.Errorf("features = %v, want lifecycle_destroy", features)
	}
	if !contains(features, "replace_triggered_by") {
		t.Errorf("features = %v, want replace_triggered_by", features)
	}
	if got := tfversion.MinVersion(features, ""); got != "1.16.0" {
		t.Errorf("min_version = %q, want 1.16.0", got)
	}
}

func TestDetectJSONFlatConditions(t *testing.T) {
	graph := &ir.Graph{
		Version: ir.Version,
		Outputs: []*ir.Output{{
			Name: "ip",
			Attributes: map[string]ir.Value{
				"precondition": {Kind: ir.KindMap, Fields: map[string]ir.Value{
					"condition": {Kind: ir.KindString, Literal: "true"},
				}},
			},
		}},
	}
	if features := tfversion.Detect(graph); !contains(features, "conditions") {
		t.Errorf("features = %v, want conditions", features)
	}
}

func TestDetectIgnoresRemovedLifecycleDestroy(t *testing.T) {
	// removed { lifecycle { destroy = false } } exists since 1.9, so it must not
	// be reported as the 1.16 resource-level lifecycle destroy.
	graph := &ir.Graph{
		Version: ir.Version,
		Preserved: []ir.Block{{
			Type: "removed",
			Blocks: []ir.Block{{
				Type:       "lifecycle",
				Attributes: map[string]ir.Value{"destroy": {Kind: ir.KindBool}},
			}},
		}},
	}
	features := tfversion.Detect(graph)
	if contains(features, "lifecycle_destroy") {
		t.Errorf("features = %v, lifecycle_destroy must not fire for removed blocks", features)
	}
	if !contains(features, "removed") {
		t.Errorf("features = %v, want removed", features)
	}
}

func TestDetectActionEventsDoNotUseProvisionerOnFailure(t *testing.T) {
	// on_failure is also a provisioner argument, so it must only count inside an
	// action block.
	graph := &ir.Graph{
		Version: ir.Version,
		Resources: []*ir.Resource{{
			ID:   "aws_instance.web",
			Kind: ir.KindResource,
			Type: "aws_instance",
			Name: "web",
			Blocks: []ir.Block{{Type: "provisioner", Labels: []string{"local-exec"}, Attributes: map[string]ir.Value{
				"on_failure": {Kind: ir.KindString, Literal: "continue"},
			}}},
		}},
	}
	if features := tfversion.Detect(graph); contains(features, "action_lifecycle_events") {
		t.Errorf("features = %v, action_lifecycle_events must not fire", features)
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
