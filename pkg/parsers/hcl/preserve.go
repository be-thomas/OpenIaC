package hclparser

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"

	"github.com/be-thomas/openiac/pkg/ir"
)

// preserveBlock keeps a top-level block the IR has no typed node for: moved,
// import, removed, check, ephemeral, action, and any block type a later
// Terraform release adds.
//
// A preserved block never becomes a graph node, so it cannot affect
// dependencies, ordering or resource statistics. Emitters report it with a
// warning instead.
func (p *parser) preserveBlock(blk block) error {
	attrs, blocks, err := parseBody(blk.body, p.src[blk.file])
	if err != nil {
		return fmt.Errorf("%s:%d: %w", blk.file, blk.line, err)
	}

	p.graph.Preserved = append(p.graph.Preserved, ir.Block{
		Type:       blk.typ,
		Labels:     append([]string(nil), blk.labels...),
		Attributes: attrs,
		Blocks:     blocks,
		File:       blk.file,
		Line:       blk.line,
	})
	return nil
}

// preserveJSONProperty keeps a top-level property of a .tf.json document that
// the schema did not claim.
//
// JSON spells blocks in three shapes, and only one of them can be told apart
// from an ordinary value without a schema:
//
//	"moved": {"from": "a", "to": "b"}       one block, properties are arguments
//	"moved": [{...}, {...}]                 one block per element
//	"future": "scalar"                      cannot be a block
//
// The first two are recovered; anything else is kept whole in Body rather than
// guessed at.
func (p *parser) preserveJSONProperty(filename, name string, expr hcl.Expression, line int) {
	value := parseExpr(expr, p.src[filename])

	switch value.Kind {
	case ir.KindMap:
		p.graph.Preserved = append(p.graph.Preserved, ir.Block{
			Type:       name,
			Attributes: value.Fields,
			File:       filename,
			Line:       line,
		})
	case ir.KindList:
		for _, item := range value.Items {
			if item.Kind != ir.KindMap {
				p.preserveJSONBody(filename, name, value, line)
				return
			}
		}
		for _, item := range value.Items {
			p.graph.Preserved = append(p.graph.Preserved, ir.Block{
				Type:       name,
				Attributes: item.Fields,
				File:       filename,
				Line:       line,
			})
		}
	default:
		p.preserveJSONBody(filename, name, value, line)
	}
}

func (p *parser) preserveJSONBody(filename, name string, value ir.Value, line int) {
	body := value
	p.graph.Preserved = append(p.graph.Preserved, ir.Block{
		Type: name,
		Body: &body,
		File: filename,
		Line: line,
	})
}
