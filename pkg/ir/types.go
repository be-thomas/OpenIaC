// Package ir defines OpenIaC's universal intermediate representation.
//
// The IR is deliberately small: a flat graph of nodes plus dependency edges.
// Parsers boil their input language down to this shape, and emitters render it
// in a target language. Nothing in the IR represents loops, conditionals or
// other imperative logic — those are resolved by the parser before the graph is
// built.
//
// Schema 0.2 is additive over 0.1: no existing field changed type or meaning.
// Configuration the IR does not model is preserved rather than dropped, in
// Graph.Preserved, in Terraform, or in each node's Attributes and Blocks. The
// full contract is documented in docs/IR.md.
package ir

// Version is the schema version of the IR document.
const Version = "0.2"

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
	Terraform *Terraform  `json:"terraform,omitempty"`
	Resources []*Resource `json:"resources"`
	Variables []*Variable `json:"variables,omitempty"`
	Locals    []*Local    `json:"locals,omitempty"`
	Outputs   []*Output   `json:"outputs,omitempty"`
	Modules   []*Module   `json:"modules,omitempty"`
	Providers []*Provider `json:"providers,omitempty"`
	// Preserved holds top-level configuration the IR has no typed node for:
	// moved, import, removed, check, ephemeral, action, and any future or
	// unrecognised block type. Preserved blocks carry no node identity and never
	// appear in Resources, Order or Stats.
	Preserved []Block `json:"preserved,omitempty"`
	// Features lists the Terraform language features detected in the input, in
	// sorted order. See package tfversion.
	Features []string `json:"features,omitempty"`
	// MinVersion is the oldest Terraform release that accepts this
	// configuration, derived from Features and Terraform.RequiredVersion.
	MinVersion string `json:"min_version,omitempty"`
	Edges      []Edge `json:"edges"`
}

// Terraform is the merge of every terraform block in the module. Settings the
// IR models are typed; everything else is kept in Attributes and Blocks so
// nothing is lost.
type Terraform struct {
	RequiredVersion string `json:"required_version,omitempty"`
	// RequiredProviders maps a provider local name (for example "aws") to its
	// declaration: a map with source, version and configuration_aliases.
	RequiredProviders map[string]Value `json:"required_providers,omitempty"`
	Experiments       []string         `json:"experiments,omitempty"`
	// Backend is the legacy state backend; Labels[0] is the backend type.
	Backend *Block `json:"backend,omitempty"`
	// Cloud is the HCP Terraform / Terraform Cloud integration block.
	Cloud *Block `json:"cloud,omitempty"`
	// ProviderMeta holds provider_meta blocks; Labels[0] is the provider name.
	ProviderMeta []Block `json:"provider_meta,omitempty"`
	// Attributes keeps terraform settings the IR does not model.
	Attributes map[string]Value `json:"attributes,omitempty"`
	// Blocks keeps nested terraform blocks the IR does not model.
	Blocks []Block `json:"blocks,omitempty"`
	File   string  `json:"file,omitempty"`
	Line   int     `json:"line,omitempty"`
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
// rule), or a preserved top-level construct that has no typed node.
type Block struct {
	Type       string           `json:"type"`
	Labels     []string         `json:"labels,omitempty"`
	Attributes map[string]Value `json:"attributes,omitempty"`
	Blocks     []Block          `json:"blocks,omitempty"`
	// Body preserves a JSON value the parser could not decompose into
	// attributes and blocks. It is set only for unrecognised top-level
	// properties of a .tf.json document; native HCL always fills Attributes and
	// Blocks.
	Body *Value `json:"body,omitempty"`
	// File and Line locate the block. Top-level and native nested blocks carry
	// both. A block derived from a flat .tf.json body carries the enclosing
	// file and no line, because JSON has no per-block position below the top
	// level.
	File string `json:"file,omitempty"`
	Line int    `json:"line,omitempty"`
}

// Variable is a declared input variable.
type Variable struct {
	Name        string `json:"name"`
	Type        string `json:"type,omitempty"`
	Default     *Value `json:"default,omitempty"`
	Description string `json:"description,omitempty"`
	Sensitive   bool   `json:"sensitive,omitempty"`
	// Nullable is Terraform 1.1+. Nil means the argument was not set.
	Nullable *bool `json:"nullable,omitempty"`
	// Ephemeral is Terraform 1.10+.
	Ephemeral bool `json:"ephemeral,omitempty"`
	// Deprecated is the deprecation message, Terraform 1.15+.
	Deprecated string `json:"deprecated,omitempty"`
	// Attributes keeps variable settings the IR does not model.
	Attributes map[string]Value `json:"attributes,omitempty"`
	// Blocks keeps nested blocks such as validation.
	Blocks []Block `json:"blocks,omitempty"`
	File   string  `json:"file,omitempty"`
	Line   int     `json:"line,omitempty"`
}

// Local is a named local value.
type Local struct {
	Name  string `json:"name"`
	Value Value  `json:"value"`
	File  string `json:"file,omitempty"`
	Line  int    `json:"line,omitempty"`
}

// Output is a declared output value.
type Output struct {
	Name        string `json:"name"`
	Value       Value  `json:"value"`
	Description string `json:"description,omitempty"`
	Sensitive   bool   `json:"sensitive,omitempty"`
	// Type is the output type constraint, Terraform 1.15+.
	Type string `json:"type,omitempty"`
	// Ephemeral is Terraform 1.10+.
	Ephemeral bool `json:"ephemeral,omitempty"`
	// Deprecated is the deprecation message, Terraform 1.15+.
	Deprecated string `json:"deprecated,omitempty"`
	// Attributes keeps output settings the IR does not model.
	Attributes map[string]Value `json:"attributes,omitempty"`
	// Blocks keeps nested blocks such as precondition and postcondition.
	Blocks []Block `json:"blocks,omitempty"`
	File   string  `json:"file,omitempty"`
	Line   int     `json:"line,omitempty"`
}

// Module is a module call. OpenIaC does not yet expand modules; the call is
// preserved so emitters can report it rather than silently dropping it.
type Module struct {
	Name       string           `json:"name"`
	Source     string           `json:"source,omitempty"`
	Version    string           `json:"version,omitempty"`
	Attributes map[string]Value `json:"attributes,omitempty"`
	Blocks     []Block          `json:"blocks,omitempty"`
	File       string           `json:"file,omitempty"`
	Line       int              `json:"line,omitempty"`
}

// Provider is a provider configuration block.
type Provider struct {
	Name       string           `json:"name"`
	Alias      string           `json:"alias,omitempty"`
	Attributes map[string]Value `json:"attributes,omitempty"`
	Blocks     []Block          `json:"blocks,omitempty"`
	File       string           `json:"file,omitempty"`
	Line       int              `json:"line,omitempty"`
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
