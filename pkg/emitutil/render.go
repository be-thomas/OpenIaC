package emitutil

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/be-thomas/openiac/pkg/ir"
)

// Quote renders a TypeScript string literal.
func Quote(s string) string {
	var b strings.Builder
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(s); err != nil {
		return strconv.Quote(s)
	}
	return strings.TrimSuffix(b.String(), "\n")
}

// RenderValue renders an IR value as a TypeScript expression.
func RenderValue(value ir.Value, rewriter *RefRewriter) string {
	switch value.Kind {
	case ir.KindString:
		s, _ := value.Literal.(string)
		return Quote(s)
	case ir.KindNumber:
		return renderNumber(value.Literal)
	case ir.KindBool:
		if b, ok := value.Literal.(bool); ok && b {
			return "true"
		}
		return "false"
	case ir.KindNull:
		return "undefined"
	case ir.KindList:
		parts := make([]string, 0, len(value.Items))
		for _, item := range value.Items {
			parts = append(parts, RenderValue(item, rewriter))
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case ir.KindMap:
		keys := make([]string, 0, len(value.Fields))
		for key := range value.Fields {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, key := range keys {
			parts = append(parts, Quote(key)+": "+RenderValue(value.Fields[key], rewriter))
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case ir.KindExpr:
		raw := value.Raw
		if raw == "" {
			raw = value.Expr
		}
		return RenderExpression(raw, rewriter)
	default:
		return "undefined"
	}
}

func renderNumber(literal any) string {
	switch n := literal.(type) {
	case int64:
		return strconv.FormatInt(n, 10)
	case int:
		return strconv.Itoa(n)
	case float64:
		return strconv.FormatFloat(n, 'f', -1, 64)
	case json.Number:
		return n.String()
	default:
		return fmt.Sprint(n)
	}
}

// RenderExpression rewrites references in a raw HCL expression and converts
// quoted template strings to TypeScript template literals.
func RenderExpression(raw string, rewriter *RefRewriter) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "undefined"
	}
	if HasHeredoc(trimmed) {
		return trimmed
	}
	if len(trimmed) >= 2 && trimmed[0] == '"' && trimmed[len(trimmed)-1] == '"' {
		return templateToTS(trimmed, rewriter)
	}
	if rewriter == nil {
		return trimmed
	}
	return rewriter.Rewrite(trimmed)
}

// templateToTS converts an HCL quoted template ("prefix-${var.suffix}") into a
// TypeScript template literal (`prefix-${suffix}`).
func templateToTS(quoted string, rewriter *RefRewriter) string {
	inner := quoted[1 : len(quoted)-1]
	var b strings.Builder
	b.WriteByte('`')
	for i := 0; i < len(inner); {
		switch {
		case inner[i] == '\\' && i+1 < len(inner):
			b.WriteString(inner[i : i+2])
			i += 2
		case inner[i] == '$' && i+2 < len(inner) && inner[i+1] == '$' && inner[i+2] == '{':
			b.WriteString("\\${")
			i += 3
		case inner[i] == '%' && i+2 < len(inner) && inner[i+1] == '%' && inner[i+2] == '{':
			b.WriteString("%{")
			i += 3
		case inner[i] == '$' && i+1 < len(inner) && inner[i+1] == '{':
			end := matchingBrace(inner, i+1)
			if end < 0 {
				b.WriteString(inner[i:])
				i = len(inner)
				continue
			}
			b.WriteString("${")
			expression := inner[i+2 : end]
			if rewriter != nil {
				expression = rewriter.Rewrite(expression)
			}
			b.WriteString(expression)
			b.WriteByte('}')
			i = end + 1
		case inner[i] == '`':
			b.WriteString("\\`")
			i++
		default:
			b.WriteByte(inner[i])
			i++
		}
	}
	b.WriteByte('`')
	return b.String()
}

// TagsToCfn converts a map value into CloudFormation's list-of-{Key,Value}
// tag shape. Non-literal values are rendered as expressions where possible.
func TagsToCfn(value ir.Value, rewriter *RefRewriter) (string, bool) {
	if value.Kind != ir.KindMap {
		return "", false
	}
	keys := make([]string, 0, len(value.Fields))
	for key := range value.Fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, "{ Key: "+Quote(key)+", Value: "+RenderValue(value.Fields[key], rewriter)+" }")
	}
	return "[" + strings.Join(parts, ", ") + "]", true
}
