// Package ir defines OpenIaC's universal intermediate representation.
//
// The IR is deliberately small: a flat graph of nodes plus dependency edges.
// Parsers boil their input language down to this shape, and emitters render it
// in a target language. Nothing in the IR represents loops, conditionals or
// other imperative logic — those are resolved by the parser before the graph is
// built.
package ir

// Version is the schema version of the IR document.
const Version = "0.1"

// Kind identifies what a node represents.
type Kind string

const (
	KindResource Kind = "resource"
	KindData     Kind = "data"
	KindVariable Kind = "variable"
	KindLocal    Kind = "local"
	KindOutput   Kind = "output"
	KindModule   Kind = "module"
	KindProvider Kind = "provider"
)

// Graph is the complete, target-agnostic representation of one infrastructure
// project.
type Graph struct {
	Version   string      `json:"version"`
	Source    string      `json:"source,omitempty"`
	Resources []*Resource `json:"resources"`
	Variables []*Variable `json:"variables,omitempty"`
	Locals    []*Local    `json:"locals,omitempty"`
	Outputs   []*Output   `json:"outputs,omitempty"`
	Modules   []*Module   `json:"modules,omitempty"`
	Providers []*Provider `json:"providers,omitempty"`
	Edges     []Edge      `json:"edges"`
}

// Resource is a managed resource or data source.
type Resource struct {
	// ID is the canonical Terraform address, e.g. "aws_s3_bucket.logs" or
	// "data.aws_ami.ubuntu".
	ID         string           `json:"id"`
	Kind       Kind             `json:"kind"`
	Type       string           `json:"type"`
	Name       string           `json:"name"`
	Provider   string           `json:"provider,omitempty"`
	Attributes map[string]Value `json:"attributes,omitempty"`
	Blocks     []Block          `json:"blocks,omitempty"`
	// DependsOn holds explicit dependency addresses from a depends_on meta-argument.
	DependsOn []string `json:"depends_on,omitempty"`
	// ProviderAlias is the value of the provider meta-argument, if set.
	ProviderAlias string `json:"provider_alias,omitempty"`
	File          string `json:"file,omitempty"`
	Line          int    `json:"line,omitempty"`
}

// Block is a nested configuration block (for example a security group ingress
// rule).
type Block struct {
	Type       string           `json:"type"`
	Labels     []string         `json:"labels,omitempty"`
	Attributes map[string]Value `json:"attributes,omitempty"`
	Blocks     []Block          `json:"blocks,omitempty"`
	Line       int              `json:"line,omitempty"`
}

// Variable is a declared input variable.
type Variable struct {
	Name        string `json:"name"`
	Type        string `json:"type,omitempty"`
	Default     *Value `json:"default,omitempty"`
	Description string `json:"description,omitempty"`
	Sensitive   bool   `json:"sensitive,omitempty"`
}

// Local is a named local value.
type Local struct {
	Name  string `json:"name"`
	Value Value  `json:"value"`
}

// Output is a declared output value.
type Output struct {
	Name        string `json:"name"`
	Value       Value  `json:"value"`
	Description string `json:"description,omitempty"`
	Sensitive   bool   `json:"sensitive,omitempty"`
}

// Module is a module call. OpenIaC does not yet expand modules; the call is
// preserved so emitters can report it rather than silently dropping it.
type Module struct {
	Name       string           `json:"name"`
	Source     string           `json:"source,omitempty"`
	Version    string           `json:"version,omitempty"`
	Attributes map[string]Value `json:"attributes,omitempty"`
	Blocks     []Block          `json:"blocks,omitempty"`
}

// Provider is a provider configuration block.
type Provider struct {
	Name       string           `json:"name"`
	Alias      string           `json:"alias,omitempty"`
	Attributes map[string]Value `json:"attributes,omitempty"`
}

// ValueKind describes how a Value is stored.
type ValueKind string

const (
	KindString ValueKind = "string"
	KindNumber ValueKind = "number"
	KindBool   ValueKind = "bool"
	KindNull   ValueKind = "null"
	KindList   ValueKind = "list"
	KindMap    ValueKind = "map"
	// KindExpr is an expression the parser could not evaluate to a literal.
	// Raw holds the original source text and Refs holds the addresses it reads.
	KindExpr ValueKind = "expr"
)

// Value is a literal or an unresolved expression.
type Value struct {
	Kind    ValueKind        `json:"kind"`
	Literal any              `json:"literal,omitempty"`
	Expr    string           `json:"expr,omitempty"`
	Raw     string           `json:"raw,omitempty"`
	Refs    []string         `json:"refs,omitempty"`
	Items   []Value          `json:"items,omitempty"`
	Fields  map[string]Value `json:"fields,omitempty"`
}

// IsLiteral reports whether the value can be rendered without reference
// rewriting.
func (v Value) IsLiteral() bool {
	switch v.Kind {
	case KindString, KindNumber, KindBool, KindNull:
		return true
	case KindList:
		for _, item := range v.Items {
			if !item.IsLiteral() {
				return false
			}
		}
		return true
	case KindMap:
		for _, field := range v.Fields {
			if !field.IsLiteral() {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// EdgeKind describes why one node depends on another.
type EdgeKind string

const (
	EdgeReference EdgeKind = "reference"
	EdgeDependsOn EdgeKind = "depends_on"
)

// Edge means From depends on To: To must be declared first.
type Edge struct {
	From string   `json:"from"`
	To   string   `json:"to"`
	Kind EdgeKind `json:"kind"`
}
