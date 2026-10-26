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
// existing lists the blocks the caller already has. A type that is present
// there is left alone, so an attribute that duplicates a native block is still
// reported as an unsupported property rather than rendered a second time.
//
// Values that do not fit a block shape (a scalar, a list holding a scalar) are
// left alone. An empty array counts as a block shape: it means the block is
// absent, which is what native HCL with no such block means. Blocks are
// returned in sorted attribute-name order so output stays deterministic.
func UnflattenBlocks(attrs map[string]ir.Value, schema map[string]mappings.BlockMap, existing []ir.Block) (blocks []ir.Block, consumed map[string]bool) {
	if len(attrs) == 0 || len(schema) == 0 {
		return nil, nil
	}

	native := make(map[string]bool, len(existing))
	for _, block := range existing {
		native[block.Type] = true
	}

	names := make([]string, 0, len(attrs))
	for name, value := range attrs {
		if _, ok := schema[name]; ok && !native[name] && blockShape(value) {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return nil, nil
	}
	sort.Strings(names)

	consumed = make(map[string]bool, len(names))
	for _, name := range names {
		consumed[name] = true
		blocks = append(blocks, blocksFromValue(name, attrs[name], schema[name])...)
	}
	return blocks, consumed
}

// blockShape reports whether a flat JSON value has the shape of nested blocks:
// one object, or an array of objects. An empty array is included, because it
// says the block is absent rather than that it is an attribute.
func blockShape(v ir.Value) bool {
	switch v.Kind {
	case ir.KindMap:
		return true
	case ir.KindList:
		for _, item := range v.Items {
			if item.Kind != ir.KindMap {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// blocksFromValue converts a value that blockShape accepted into the blocks it
// represents: a single object is one block, an array of objects is one block
// per element.
func blocksFromValue(typ string, v ir.Value, blockMap mappings.BlockMap) []ir.Block {
	if v.Kind == ir.KindList {
		blocks := make([]ir.Block, 0, len(v.Items))
		for _, item := range v.Items {
			block, _ := blockFromMap(typ, item, blockMap)
			blocks = append(blocks, block)
		}
		return blocks
	}
	block, _ := blockFromMap(typ, v, blockMap)
	return []ir.Block{block}
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
		nested, consumed := UnflattenBlocks(block.Attributes, blockMap.Blocks, nil)
		for name := range consumed {
			delete(block.Attributes, name)
		}
		block.Blocks = nested
	}
	return block, true
}
