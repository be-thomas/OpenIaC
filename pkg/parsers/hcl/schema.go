package hclparser

import "github.com/hashicorp/hcl/v2"

// topLevelSchema describes every top-level block Terraform accepts, so a
// .tf.json document can be decomposed into real blocks with real labels.
//
// hcl.Body hides block structure below the top level, but at the top level
// PartialContent recovers it exactly: hcl/v2's JSON body peels one label per
// nested object and hard-errors when the label count does not match. The schema
// must therefore mirror Terraform's own block headers precisely.
var topLevelSchema = &hcl.BodySchema{Blocks: []hcl.BlockHeaderSchema{
	{Type: "terraform"},
	{Type: "locals"},
	{Type: "moved"},
	{Type: "import"},
	{Type: "removed"},
	{Type: "resource", LabelNames: []string{"type", "name"}},
	{Type: "data", LabelNames: []string{"type", "name"}},
	{Type: "ephemeral", LabelNames: []string{"type", "name"}},
	{Type: "action", LabelNames: []string{"type", "name"}},
	{Type: "check", LabelNames: []string{"name"}},
	{Type: "variable", LabelNames: []string{"name"}},
	{Type: "output", LabelNames: []string{"name"}},
	{Type: "module", LabelNames: []string{"name"}},
	{Type: "provider", LabelNames: []string{"name"}},
}}

// terraformBodySchema describes the nested blocks of a terraform block.
//
// required_providers is a block in native HCL and an object property in JSON;
// listing it as a block makes PartialContent turn the JSON form into a block
// too, so both syntaxes converge on the same parsing path.
var terraformBodySchema = &hcl.BodySchema{Blocks: []hcl.BlockHeaderSchema{
	{Type: "required_providers"},
	{Type: "backend", LabelNames: []string{"type"}},
	{Type: "cloud"},
	{Type: "provider_meta", LabelNames: []string{"provider"}},
}}

// topLevelLabelCount is the number of labels Terraform requires for a
// recognised top-level block. Native HCL does not validate labels while
// parsing, so the parser checks them itself.
var topLevelLabelCount = map[string]int{
	"terraform": 0,
	"locals":    0,
	"moved":     0,
	"import":    0,
	"removed":   0,
	"resource":  2,
	"data":      2,
	"ephemeral": 2,
	"action":    2,
	"check":     1,
	"variable":  1,
	"output":    1,
	"module":    1,
	"provider":  1,
}
