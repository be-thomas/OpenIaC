package hclparser

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"

	"github.com/be-thomas/openiac/pkg/ir"
)

// parseJSONExpr converts an expression that came from a .tf.json document.
//
// It differs from the native path in two ways that make the shared logic
// unusable on its own:
//
//   - A JSON expression's Value(nil) treats every string as a literal, so
//     "${var.region}" would be stored verbatim with no refs.
//   - Value(nil) also evaluates objects and arrays, which would lose the
//     refs inside them.
//
// The source shape therefore decides the path: '{' and '[' are decomposed with
// hcl.ExprMap and hcl.ExprList, and only scalars are handed to Value(nil).
func parseJSONExpr(expr hcl.Expression, src []byte) ir.Value {
	trimmed := strings.TrimSpace(sliceRange(src, expr.Range()))

	switch {
	case strings.HasPrefix(trimmed, "{"):
		if value, ok := jsonObject(expr, src); ok {
			return value
		}
	case strings.HasPrefix(trimmed, "["):
		if value, ok := jsonArray(expr, src); ok {
			return value
		}
	}

	if value, ok := literalValue(expr); ok {
		text, isString := value.Literal.(string)
		if value.Kind == ir.KindString && isString && isTemplateText(text) {
			if inner, ok := singleInterpolation(text); ok {
				// Terraform evaluates "${expr}" to the expression's own value
				// and type, so keep the bare expression: this renders exactly
				// like the native HCL form instead of wrapping it in a string
				// template.
				return ir.Value{
					Kind: ir.KindExpr,
					Expr: inner,
					Raw:  inner,
					Refs: collectRefs(expr),
				}
			}
			// Rendering the decoded text back as an HCL template keeps the
			// emitters' existing template handling working unchanged.
			quoted := hclQuoteTemplate(text)
			return ir.Value{
				Kind: ir.KindExpr,
				Expr: quoted,
				Raw:  quoted,
				Refs: collectRefs(expr),
			}
		}
		return value
	}

	// Anything left has no literal value. Keep the source text so nothing is
	// silently dropped.
	raw := strings.TrimSpace(sliceRange(src, expr.Range()))
	return ir.Value{
		Kind: ir.KindExpr,
		Expr: raw,
		Raw:  raw,
		Refs: collectRefs(expr),
	}
}

func jsonObject(expr hcl.Expression, src []byte) (ir.Value, bool) {
	pairs, diags := hcl.ExprMap(expr)
	if diags.HasErrors() {
		return ir.Value{}, false
	}

	fields := make(map[string]ir.Value, len(pairs))
	for _, pair := range pairs {
		key, keyDiags := pair.Key.Value(nil)
		if keyDiags.HasErrors() || !key.IsKnown() || key.Type() != cty.String {
			return ir.Value{}, false
		}
		fields[key.AsString()] = parseExpr(pair.Value, src)
	}
	return ir.Value{
		Kind:   ir.KindMap,
		Fields: fields,
		Refs:   collectRefs(expr),
	}, true
}

func jsonArray(expr hcl.Expression, src []byte) (ir.Value, bool) {
	exprs, diags := hcl.ExprList(expr)
	if diags.HasErrors() {
		return ir.Value{}, false
	}

	items := make([]ir.Value, 0, len(exprs))
	for _, item := range exprs {
		items = append(items, parseExpr(item, src))
	}
	return ir.Value{
		Kind:  ir.KindList,
		Items: items,
		Refs:  collectRefs(expr),
	}, true
}

// isTemplateText reports whether decoded JSON text contains HCL template
// markers, meaning it is an expression rather than a plain string.
func isTemplateText(s string) bool {
	return strings.Contains(s, "${") || strings.Contains(s, "%{")
}

// singleInterpolation returns the HCL source of the expression inside a JSON
// string that is nothing but one interpolation, for example "${var.port}".
//
// Terraform evaluates such a string to the wrapped expression's value and
// type, not to a string, so it must be kept as the bare expression. Only a
// string that cannot be recognised as a single interpolation falls back to the
// general template handling.
func singleInterpolation(s string) (string, bool) {
	if !strings.HasPrefix(s, "${") || !strings.HasSuffix(s, "}") {
		return "", false
	}
	parsed, diags := hclsyntax.ParseTemplate([]byte(s), "json", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return "", false
	}
	if _, ok := parsed.(*hclsyntax.TemplateWrapExpr); !ok {
		return "", false
	}
	inner := strings.TrimSpace(s[2 : len(s)-1])
	if inner == "" {
		return "", false
	}
	return inner, true
}

// hclQuoteTemplate renders text as an HCL quoted template.
//
// strconv.Quote cannot be used: HCL does not accept every JSON escape, and the
// template markers themselves must survive. Only the characters that would end
// the string or change its meaning on the way back are escaped.
func hclQuoteTemplate(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 2)
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString(`\\`)
		case '"':
			b.WriteString(`\"`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if r < 0x20 || r == 0x7f {
				fmt.Fprintf(&b, `\u%04x`, r)
				continue
			}
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}
