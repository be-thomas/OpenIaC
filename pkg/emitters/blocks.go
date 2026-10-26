package emitters

import (
	"sort"

	"github.com/be-thomas/openiac/pkg/ir"
	"github.com/be-thomas/openiac/pkg/mappings"
)

// UnflattenBlocks recovers nested blocks from a flat .tf.json body.
//
// Terraform's JSON syntax has no way to spell a nested block below the top
// level, so the parser leaves e.g. "versioning": {"enabled": true} as a map
// attribute, while native HCL produces a block. An emitter that knows its
// schema passes it here: every attribute whose name is a registered block type
// is converted back into the blocks HCL would have produced, and reported in
// consumed so the caller does not also render it as a property.
//
// Values that do not fit a block shape (a scalar, a list of scalars) are left
// alone. Blocks are returned in sorted attribute-name order so output stays
// deterministic.
func UnflattenBlocks(attrs map[string]ir.Value, schema map[string]mappings.BlockMap) (blocks []ir.Block, consumed map[string]bool) {
	if len(attrs) == 0 || len(schema) == 0 {
		return nil, nil
	}

	names := make([]string, 0, len(attrs))
	for name := range attrs {
		if _, ok := schema[name]; ok {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return nil, nil
	}
	sort.Strings(names)

	consumed = make(map[string]bool, len(names))
	for _, name := range names {
		converted := blocksFromValue(name, attrs[name], schema[name])
		if len(converted) == 0 {
			continue
		}
		consumed[name] = true
		blocks = append(blocks, converted...)
	}
	return blocks, consumed
}

// blocksFromValue converts one flat attribute into the blocks it represents: a
// single object is one block, an array of objects is one block per element.
func blocksFromValue(typ string, v ir.Value, blockMap mappings.BlockMap) []ir.Block {
	if v.Kind == ir.KindList {
		blocks := make([]ir.Block, 0, len(v.Items))
		for _, item := range v.Items {
			if block, ok := blockFromMap(typ, item, blockMap); ok {
				blocks = append(blocks, block)
			}
		}
		return blocks
	}
	if block, ok := blockFromMap(typ, v, blockMap); ok {
		return []ir.Block{block}
	}
	return nil
}

func blockFromMap(typ string, v ir.Value, blockMap mappings.BlockMap) (ir.Block, bool) {
	if v.Kind != ir.KindMap {
		return ir.Block{}, false
	}

	block := ir.Block{
		Type:       typ,
		Attributes: make(map[string]ir.Value, len(v.Fields)),
	}
	for name, value := range v.Fields {
		block.Attributes[name] = value
	}

	if len(blockMap.Blocks) > 0 {
		nested, consumed := UnflattenBlocks(block.Attributes, blockMap.Blocks)
		for name := range consumed {
			delete(block.Attributes, name)
		}
		block.Blocks = nested
	}
	return block, true
}
