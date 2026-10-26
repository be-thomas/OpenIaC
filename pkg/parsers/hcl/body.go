package hclparser

import (
	"fmt"
	"math/big"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	hcljson "github.com/hashicorp/hcl/v2/json"
	"github.com/zclconf/go-cty/cty"

	"github.com/be-thomas/openiac/pkg/ir"
)

// parseBody converts the content of a block into IR attributes and nested
// blocks.
//
// Native HCL keeps its full structure. JSON bodies are flat by necessity: HCL
// exposes no block structure below the top level, so every object property
// becomes a map value.
func parseBody(body hcl.Body, src []byte) (map[string]ir.Value, []ir.Block, error) {
	if hcljson.IsJSONBody(body) {
		return parseJSONBody(body, src)
	}

	native, ok := body.(*hclsyntax.Body)
	if !ok {
		return nil, nil, fmt.Errorf("unsupported HCL body type %T", body)
	}

	attrs := make(map[string]ir.Value, len(native.Attributes))
	names := make([]string, 0, len(native.Attributes))
	for name := range native.Attributes {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		attrs[name] = parseExpr(native.Attributes[name].Expr, src)
	}

	blocks := make([]ir.Block, 0, len(native.Blocks))
	for _, block := range native.Blocks {
		blockAttrs, nested, err := parseBody(block.Body, src)
		if err != nil {
			return nil, nil, err
		}
		blocks = append(blocks, ir.Block{
			Type:       block.Type,
			Labels:     append([]string(nil), block.Labels...),
			Attributes: blockAttrs,
			Blocks:     nested,
			File:       block.TypeRange.Filename,
			Line:       block.TypeRange.Start.Line,
		})
	}
	return attrs, blocks, nil
}

func parseJSONBody(body hcl.Body, src []byte) (map[string]ir.Value, []ir.Block, error) {
	attributes, diags := body.JustAttributes()
	if diags.HasErrors() {
		return nil, nil, diags
	}

	attrs := make(map[string]ir.Value, len(attributes))
	names := make([]string, 0, len(attributes))
	for name := range attributes {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		attrs[name] = parseExpr(attributes[name].Expr, src)
	}
	// JSON cannot distinguish a nested block from an object property, so nested
	// blocks are never produced here. Callers that know their block schema (the
	// terraform block) request it explicitly.
	return attrs, nil, nil
}

// parseExpr converts an HCL expression to an IR value.
func parseExpr(expr hcl.Expression, src []byte) ir.Value {
	if hcljson.IsJSONExpression(expr) {
		return parseJSONExpr(expr, src)
	}
	return parseHCLExpr(expr, src)
}

func parseHCLExpr(expr hcl.Expression, src []byte) ir.Value {
	raw := sliceRange(src, expr.Range())

	if value, ok := literalValue(expr); ok {
		return value
	}
	if value, ok := structuredValue(expr, src); ok {
		return value
	}
	return ir.Value{
		Kind: ir.KindExpr,
		Expr: raw,
		Raw:  raw,
		Refs: collectRefs(expr),
	}
}

// structuredValue handles collections that mix literals and expressions, such
// as tags = { Name = "x", Env = var.environment }. Evaluating such a value
// fails, but its shape is still static and can be represented in the IR.
func structuredValue(expr hcl.Expression, src []byte) (ir.Value, bool) {
	switch e := expr.(type) {
	case *hclsyntax.ParenthesesExpr:
		return structuredValue(e.Expression, src)
	case *hclsyntax.ObjectConsExpr:
		fields := make(map[string]ir.Value, len(e.Items))
		for _, item := range e.Items {
			key, diags := item.KeyExpr.Value(nil)
			if diags.HasErrors() || !key.IsKnown() || key.Type() != cty.String {
				return ir.Value{}, false
			}
			fields[key.AsString()] = parseExpr(item.ValueExpr, src)
		}
		return ir.Value{
			Kind:   ir.KindMap,
			Fields: fields,
			Refs:   collectRefs(expr),
		}, true
	case *hclsyntax.TupleConsExpr:
		items := make([]ir.Value, 0, len(e.Exprs))
		for _, item := range e.Exprs {
			items = append(items, parseExpr(item, src))
		}
		return ir.Value{
			Kind:  ir.KindList,
			Items: items,
			Refs:  collectRefs(expr),
		}, true
	default:
		return ir.Value{}, false
	}
}

// literalValue evaluates an expression only if it is self-contained. Anything
// that reads a variable, resource or function is left as an expression.
func literalValue(expr hcl.Expression) (value ir.Value, ok bool) {
	defer func() {
		if recover() != nil {
			value, ok = ir.Value{}, false
		}
	}()

	val, diags := expr.Value(nil)
	if diags.HasErrors() || val == cty.NilVal || !val.IsWhollyKnown() {
		return ir.Value{}, false
	}
	return fromCty(val)
}

func fromCty(val cty.Value) (ir.Value, bool) {
	if val.IsNull() {
		return ir.Value{Kind: ir.KindNull}, true
	}
	if !val.IsKnown() {
		return ir.Value{}, false
	}

	typ := val.Type()
	switch {
	case typ == cty.String:
		return ir.Value{Kind: ir.KindString, Literal: val.AsString()}, true
	case typ == cty.Number:
		return numberValue(val), true
	case typ == cty.Bool:
		return ir.Value{Kind: ir.KindBool, Literal: val.True()}, true
	case typ.IsListType(), typ.IsSetType(), typ.IsTupleType():
		items := make([]ir.Value, 0, val.LengthInt())
		for it := val.ElementIterator(); it.Next(); {
			_, item := it.Element()
			converted, ok := fromCty(item)
			if !ok {
				return ir.Value{}, false
			}
			items = append(items, converted)
		}
		return ir.Value{Kind: ir.KindList, Items: items}, true
	case typ.IsMapType(), typ.IsObjectType():
		fields := make(map[string]ir.Value, val.LengthInt())
		for it := val.ElementIterator(); it.Next(); {
			key, item := it.Element()
			converted, ok := fromCty(item)
			if !ok {
				return ir.Value{}, false
			}
			fields[key.AsString()] = converted
		}
		return ir.Value{Kind: ir.KindMap, Fields: fields}, true
	default:
		return ir.Value{}, false
	}
}

func numberValue(val cty.Value) ir.Value {
	f := val.AsBigFloat()
	if f.IsInt() {
		if i, acc := f.Int64(); acc == big.Exact {
			return ir.Value{Kind: ir.KindNumber, Literal: i}
		}
	}
	fl, _ := f.Float64()
	return ir.Value{Kind: ir.KindNumber, Literal: fl}
}

func collectRefs(expr hcl.Expression) []string {
	seen := map[string]bool{}
	var refs []string
	for _, traversal := range expr.Variables() {
		addr, ok := traversalAddress(traversal)
		if !ok || seen[addr] {
			continue
		}
		seen[addr] = true
		refs = append(refs, addr)
	}
	sort.Strings(refs)
	return refs
}

// traversalAddress maps an HCL traversal to the canonical address of the node
// it starts at, dropping the attribute path.
//
//	aws_vpc.main.id        -> aws_vpc.main
//	var.region             -> var.region
//	data.aws_ami.ubuntu.id -> data.aws_ami.ubuntu
func traversalAddress(traversal hcl.Traversal) (string, bool) {
	parts := make([]string, 0, len(traversal))
	for _, step := range traversal {
		switch s := step.(type) {
		case hcl.TraverseRoot:
			parts = append(parts, s.Name)
		case hcl.TraverseAttr:
			parts = append(parts, s.Name)
		case hcl.TraverseIndex:
			// Indexed roots (e.g. each.value) are not static addresses.
			return "", false
		}
	}
	if len(parts) == 0 {
		return "", false
	}

	switch parts[0] {
	case "var", "local", "module":
		if len(parts) >= 2 {
			return parts[0] + "." + parts[1], true
		}
	case "data":
		if len(parts) >= 3 {
			return "data." + parts[1] + "." + parts[2], true
		}
	case "path", "terraform", "count", "each", "self":
		return "", false
	default:
		if len(parts) >= 2 {
			return parts[0] + "." + parts[1], true
		}
	}
	return "", false
}

func sliceRange(src []byte, r hcl.Range) string {
	if r.Start.Byte < 0 || r.End.Byte > len(src) || r.Start.Byte > r.End.Byte {
		return ""
	}
	return string(src[r.Start.Byte:r.End.Byte])
}

func providerPrefix(resourceType string) string {
	if i := strings.IndexByte(resourceType, '_'); i > 0 {
		return resourceType[:i]
	}
	return resourceType
}

func literalString(v ir.Value) string {
	if v.Kind == ir.KindString {
		s, _ := v.Literal.(string)
		return s
	}
	return ""
}

func renderValue(v ir.Value) string {
	switch v.Kind {
	case ir.KindString:
		s, _ := v.Literal.(string)
		return fmt.Sprintf("%q", s)
	case ir.KindNumber, ir.KindBool, ir.KindNull:
		return fmt.Sprint(v.Literal)
	case ir.KindExpr:
		return v.Expr
	default:
		return ""
	}
}

// typeText renders a type constraint the way Terraform writes it: the source
// text of the expression, or a plain string for JSON's "type": "string".
func typeText(v ir.Value) string {
	if s := strings.TrimSpace(v.Expr); s != "" {
		return s
	}
	if s := literalString(v); s != "" {
		return s
	}
	return strings.TrimSpace(renderValue(v))
}
