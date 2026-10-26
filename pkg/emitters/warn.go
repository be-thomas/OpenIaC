package emitters

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/be-thomas/openiac/pkg/ir"
)

// Stable warning codes. They are documented in docs/CONFORMANCE.md and may be
// added to, but changing or removing one is a breaking change.
const (
	codePreservedBlock     = "W_PRESERVED_BLOCK"
	codeTerraformBackend   = "W_TERRAFORM_BACKEND"
	codeTerraformCloud     = "W_TERRAFORM_CLOUD"
	codeTerraformMeta      = "W_TERRAFORM_PROVIDER_META"
	codeCondition          = "W_CONDITION"
	codeVariableValidation = "W_VARIABLE_VALIDATION"
	codeUnsupportedField   = "W_UNSUPPORTED_FIELD"
)

// WarnUnsupported records one coded warning for every piece of the graph the
// target cannot convert. Parsers preserve Terraform 1.x constructs the IR has
// no node for; without this an unsupported construct would disappear silently.
//
// Emitters call it once, right after they create their warnings collector.
// Output is deterministic: preserved block types and attribute names are
// sorted, and locations use the base file name so paths stay portable.
func WarnUnsupported(graph *ir.Graph, w *Warnings) {
	if graph == nil || w == nil {
		return
	}

	warnPreservedBlocks(graph.Preserved, w)
	if tf := graph.Terraform; tf != nil {
		warnTerraform(tf, w)
	}
	warnConditions(graph, w)
	warnVariableValidations(graph.Variables, w)
	warnUnmodelledFields(graph, w)
}

func warnPreservedBlocks(blocks []ir.Block, w *Warnings) {
	if len(blocks) == 0 {
		return
	}

	order := make([]string, 0, len(blocks))
	byType := make(map[string][]ir.Block, len(blocks))
	for _, block := range blocks {
		if _, ok := byType[block.Type]; !ok {
			order = append(order, block.Type)
		}
		byType[block.Type] = append(byType[block.Type], block)
	}
	sort.Strings(order)

	for _, blockType := range order {
		items := byType[blockType]
		first := location(items[0].File, items[0].Line)
		if len(items) == 1 {
			w.Addf("%s: %s %s block at %s is preserved but not converted", codePreservedBlock, article(blockType), blockType, first)
			continue
		}
		w.Addf("%s: %d %s blocks are preserved but not converted (first at %s)", codePreservedBlock, len(items), blockType, first)
	}
}

func warnTerraform(tf *ir.Terraform, w *Warnings) {
	if tf.Backend != nil {
		w.Addf("%s: terraform backend%s at %s is not converted; configure state in the target stack instead",
			codeTerraformBackend, label(tf.Backend.Labels), location(tf.Backend.File, tf.Backend.Line))
	}
	if tf.Cloud != nil {
		w.Addf("%s: terraform cloud block at %s is not converted",
			codeTerraformCloud, location(tf.Cloud.File, tf.Cloud.Line))
	}
	for _, block := range tf.ProviderMeta {
		w.Addf("%s: provider_meta%s at %s is not converted",
			codeTerraformMeta, label(block.Labels), location(block.File, block.Line))
	}
}

// warnConditions reports preconditions and postconditions. Neither emitter can
// express them in generated code, and dropping them silently would lose a
// correctness check the author wrote.
func warnConditions(graph *ir.Graph, w *Warnings) {
	for _, resource := range graph.Resources {
		warnConditionFields(resource.ID, resource.File, resource.Attributes, resource.Blocks, w)
	}
	for _, output := range graph.Outputs {
		warnConditionFields("output."+output.Name, output.File, output.Attributes, output.Blocks, w)
	}
}

func warnConditionFields(owner, file string, attrs map[string]ir.Value, blocks []ir.Block, w *Warnings) {
	for _, name := range []string{"precondition", "postcondition"} {
		if _, ok := attrs[name]; ok {
			// .tf.json spells the condition as an object property, so there is no
			// per-condition line to report.
			w.Addf("%s: %s: %s is not converted", codeCondition, owner, name)
			continue
		}
		for _, block := range blocks {
			if block.Type == name {
				w.Addf("%s: %s: %s block at %s is not converted", codeCondition, owner, name, location(file, block.Line))
			}
		}
	}
}

func warnVariableValidations(variables []*ir.Variable, w *Warnings) {
	for _, variable := range variables {
		// .tf.json spells validation as an object property.
		if _, ok := variable.Attributes["validation"]; ok {
			w.Addf("%s: var.%s: validation is not converted (return an error from the stack instead)",
				codeVariableValidation, variable.Name)
		}
		for _, block := range variable.Blocks {
			if block.Type == "validation" {
				w.Addf("%s: var.%s: validation block at %s is not converted (return an error from the stack instead)",
					codeVariableValidation, variable.Name, location(variable.File, block.Line))
			}
		}
	}
}

// warnUnmodelledFields reports typed nodes carrying settings the IR has no
// field for. Those settings land in Attributes/Blocks, which no emitter reads.
func warnUnmodelledFields(graph *ir.Graph, w *Warnings) {
	if tf := graph.Terraform; tf != nil {
		warnExtras("terraform", tf.File, tf.Attributes, tf.Blocks, nil, nil, w)
	}
	for _, variable := range graph.Variables {
		warnExtras("var."+variable.Name, variable.File, variable.Attributes, variable.Blocks,
			knownBlocks("validation"), knownBlocks("validation"), w)
	}
	for _, output := range graph.Outputs {
		warnExtras("output."+output.Name, output.File, output.Attributes, output.Blocks,
			knownBlocks("precondition", "postcondition"), knownBlocks("precondition", "postcondition"), w)
	}
}

func warnExtras(owner, file string, attrs map[string]ir.Value, blocks []ir.Block, skipAttrs, skipBlocks map[string]bool, w *Warnings) {
	names := make([]string, 0, len(attrs))
	for name := range attrs {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if skipAttrs[name] {
			continue
		}
		w.Addf("%s: %s: attribute %q is not converted", codeUnsupportedField, owner, name)
	}

	seen := make(map[string]bool, len(blocks))
	for _, block := range blocks {
		if skipBlocks[block.Type] || seen[block.Type] {
			continue
		}
		seen[block.Type] = true
		w.Addf("%s: %s: block %q at %s is not converted", codeUnsupportedField, owner, block.Type, location(file, block.Line))
	}
}

func knownBlocks(names ...string) map[string]bool {
	out := make(map[string]bool, len(names))
	for _, name := range names {
		out[name] = true
	}
	return out
}

// location renders a source position as "file:line", or just the file when the
// position is unknown. Only the base name is used so warning text does not
// depend on where a checkout lives.
func location(file string, line int) string {
	name := filepath.Base(file)
	if name == "" || name == "." {
		return "an unknown location"
	}
	if line > 0 {
		return fmt.Sprintf("%s:%d", name, line)
	}
	return name
}

func label(labels []string) string {
	if len(labels) == 0 || labels[0] == "" {
		return ""
	}
	return " " + strconv.Quote(labels[0])
}

// article picks the indefinite article for a block type so warnings read
// naturally: "a moved block", "an import block".
func article(word string) string {
	if word == "" {
		return "a"
	}
	switch word[0] {
	case 'a', 'e', 'i', 'o', 'u', 'A', 'E', 'I', 'O', 'U':
		return "an"
	}
	return "a"
}
