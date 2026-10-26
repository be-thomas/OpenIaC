package hclparser

import (
	"fmt"
	"sort"
	"strings"

	"github.com/be-thomas/openiac/pkg/ir"
)

// addTerraform merges one terraform block into the graph's Terraform metadata.
//
// Terraform permits only one terraform block per module, but modules in the
// wild repeat it and the merge is harmless, so every block is folded in.
//
// Native HCL spells required_providers, backend, cloud and provider_meta as
// nested blocks. JSON spells each of them as an object property, which the flat
// JSON body reports as a map value; both shapes are accepted here.
func (p *parser) addTerraform(blk block) error {
	attrs, blocks, err := parseBody(blk.body, p.src[blk.file])
	if err != nil {
		return fmt.Errorf("%s:%d: %w", blk.file, blk.line, err)
	}

	tf := p.graph.Terraform
	if tf == nil {
		tf = &ir.Terraform{File: blk.file, Line: blk.line}
		p.graph.Terraform = tf
	}

	mergeTerraformAttrs(tf, attrs)
	mergeTerraformBlocks(tf, blocks)

	if len(attrs) > 0 {
		if tf.Attributes == nil {
			tf.Attributes = make(map[string]ir.Value, len(attrs))
		}
		for name, value := range attrs {
			tf.Attributes[name] = value
		}
	}
	return nil
}

// mergeTerraformAttrs consumes the typed settings written as arguments and
// leaves everything else in attrs.
func mergeTerraformAttrs(tf *ir.Terraform, attrs map[string]ir.Value) {
	if v, ok := attrs["required_version"]; ok {
		if s := literalString(v); s != "" {
			tf.RequiredVersion = s
			delete(attrs, "required_version")
		} else if s := strings.TrimSpace(v.Expr); s != "" {
			tf.RequiredVersion = s
			delete(attrs, "required_version")
		}
	}

	if v, ok := attrs["experiments"]; ok {
		if list, ok := stringList(v); ok {
			tf.Experiments = append(tf.Experiments, list...)
			delete(attrs, "experiments")
		}
	}

	if v, ok := attrs["required_providers"]; ok && v.Kind == ir.KindMap {
		if tf.RequiredProviders == nil {
			tf.RequiredProviders = make(map[string]ir.Value, len(v.Fields))
		}
		for name, decl := range v.Fields {
			tf.RequiredProviders[name] = decl
		}
		delete(attrs, "required_providers")
	}

	if v, ok := attrs["backend"]; ok {
		if b, ok := labelledBlock("backend", v); ok {
			if tf.Backend == nil {
				// JSON bodies are flat, so the nested block has no range of its
				// own; the enclosing terraform block supplies the file.
				b.File = tf.File
				tf.Backend = &b
			}
			delete(attrs, "backend")
		}
	}

	if v, ok := attrs["cloud"]; ok && v.Kind == ir.KindMap {
		if tf.Cloud == nil {
			tf.Cloud = &ir.Block{Type: "cloud", Attributes: v.Fields, File: tf.File}
		}
		delete(attrs, "cloud")
	}

	if v, ok := attrs["provider_meta"]; ok && v.Kind == ir.KindMap {
		names := make([]string, 0, len(v.Fields))
		for name := range v.Fields {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			body := v.Fields[name]
			if body.Kind != ir.KindMap {
				continue
			}
			tf.ProviderMeta = append(tf.ProviderMeta, ir.Block{
				Type:       "provider_meta",
				Labels:     []string{name},
				Attributes: body.Fields,
				File:       tf.File,
			})
		}
		delete(attrs, "provider_meta")
	}
}

// mergeTerraformBlocks folds nested terraform blocks into the metadata.
func mergeTerraformBlocks(tf *ir.Terraform, blocks []ir.Block) {
	for _, b := range blocks {
		switch b.Type {
		case "required_providers":
			if tf.RequiredProviders == nil {
				tf.RequiredProviders = make(map[string]ir.Value, len(b.Attributes))
			}
			for name, decl := range b.Attributes {
				tf.RequiredProviders[name] = decl
			}
		case "backend":
			if tf.Backend == nil {
				backend := b
				tf.Backend = &backend
			} else {
				tf.Blocks = append(tf.Blocks, b)
			}
		case "cloud":
			if tf.Cloud == nil {
				cloud := b
				tf.Cloud = &cloud
			} else {
				tf.Blocks = append(tf.Blocks, b)
			}
		case "provider_meta":
			tf.ProviderMeta = append(tf.ProviderMeta, b)
		default:
			tf.Blocks = append(tf.Blocks, b)
		}
	}
}

// labelledBlock turns JSON's {"label": {…}} spelling of a single-label block
// into the block it represents.
func labelledBlock(typ string, v ir.Value) (ir.Block, bool) {
	if v.Kind != ir.KindMap || len(v.Fields) != 1 {
		return ir.Block{}, false
	}
	for label, body := range v.Fields {
		if body.Kind != ir.KindMap {
			return ir.Block{}, false
		}
		return ir.Block{
			Type:       typ,
			Labels:     []string{label},
			Attributes: body.Fields,
		}, true
	}
	return ir.Block{}, false
}

// stringList reads a list of literal strings. Native HCL reports a bare
// identifier such as module_variable_optional_attrs as an expression, so the
// source text is accepted too.
func stringList(v ir.Value) ([]string, bool) {
	if v.Kind != ir.KindList {
		return nil, false
	}
	out := make([]string, 0, len(v.Items))
	for _, item := range v.Items {
		switch item.Kind {
		case ir.KindString:
			out = append(out, literalString(item))
		case ir.KindExpr:
			text := strings.TrimSpace(item.Expr)
			if text == "" {
				return nil, false
			}
			out = append(out, text)
		default:
			return nil, false
		}
	}
	return out, true
}
