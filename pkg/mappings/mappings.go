// Package mappings loads the data-driven translation tables that describe how
// Terraform resources map onto CDK constructs and Pulumi classes.
//
// Mappings live in JSON so contributors can add a resource without touching
// emitter code.
package mappings

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed aws.json
var awsJSON []byte

// Registry is one provider's mapping table.
type Registry struct {
	Schema    string                      `json:"schema"`
	Provider  string                      `json:"provider"`
	Resources map[string]*ResourceMapping `json:"resources"`
}

// ResourceMapping is the target mapping for one Terraform resource type.
type ResourceMapping struct {
	CDK    *CDKMapping    `json:"cdk,omitempty"`
	Pulumi *PulumiMapping `json:"pulumi,omitempty"`
}

// CDKMapping targets an AWS CDK construct. Kind "l2" (default) emits an
// opinionated construct; kind "l1" emits a Cfn* construct with
// CloudFormation-style property names.
type CDKMapping struct {
	Kind       string                 `json:"kind,omitempty"`
	Module     string                 `json:"module"`
	Class      string                 `json:"class"`
	Properties map[string]PropertyMap `json:"properties,omitempty"`
	Blocks     map[string]BlockMap    `json:"blocks,omitempty"`
	Attributes map[string]string      `json:"attributes,omitempty"`
}

// PulumiMapping targets a class in the Pulumi AWS SDK. Pulumi does not need a
// property table by default: Terraform attribute names are camelCased, which
// matches the generated SDK.
type PulumiMapping struct {
	Package    string                 `json:"package,omitempty"`
	Module     string                 `json:"module"`
	Class      string                 `json:"class"`
	Properties map[string]PropertyMap `json:"properties,omitempty"`
	Blocks     map[string]BlockMap    `json:"blocks,omitempty"`
	Attributes map[string]string      `json:"attributes,omitempty"`
}

// PropertyMap renames a Terraform property and optionally transforms its value.
type PropertyMap struct {
	Name      string `json:"name"`
	Transform string `json:"transform,omitempty"`
}

// UnmarshalJSON accepts either "newName" or {"name": "newName"}.
func (p *PropertyMap) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err == nil {
		p.Name = name
		return nil
	}
	type alias PropertyMap
	var value alias
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*p = PropertyMap(value)
	return nil
}

// BlockMap describes how a nested Terraform block becomes a target property.
type BlockMap struct {
	Property   string                 `json:"property"`
	Shape      string                 `json:"shape,omitempty"` // array (default), object, presence
	Value      any                    `json:"value,omitempty"`
	Properties map[string]PropertyMap `json:"properties,omitempty"`
	Blocks     map[string]BlockMap    `json:"blocks,omitempty"`
}

var (
	once    sync.Once
	aws     *Registry
	loadErr error
)

// Load returns the embedded AWS mapping registry.
func Load() (*Registry, error) {
	once.Do(func() {
		aws = &Registry{}
		if err := json.Unmarshal(awsJSON, aws); err != nil {
			loadErr = fmt.Errorf("parse embedded aws mappings: %w", err)
			return
		}
		if aws.Resources == nil {
			loadErr = fmt.Errorf("embedded aws mappings contain no resources")
		}
	})
	return aws, loadErr
}

// Lookup returns the mapping for a Terraform resource type.
func Lookup(resourceType string) (*ResourceMapping, bool) {
	registry, err := Load()
	if err != nil {
		return nil, false
	}
	mapping, ok := registry.Resources[resourceType]
	return mapping, ok
}

// MustLoad is Load for tests and callers that treat a broken embedded table as
// a programming error.
func MustLoad() *Registry {
	registry, err := Load()
	if err != nil {
		panic(err)
	}
	return registry
}
