// Package emitutil holds naming, reference rewriting and TypeScript rendering
// helpers shared by all emitters.
package emitutil

import (
	"strconv"
	"strings"
	"unicode"
)

var reserved = map[string]bool{
	"abstract": true, "any": true, "as": true, "async": true, "await": true,
	"boolean": true, "break": true, "case": true, "catch": true, "class": true,
	"const": true, "constructor": true, "continue": true, "debugger": true,
	"declare": true, "default": true, "delete": true, "do": true, "else": true,
	"enum": true, "export": true, "extends": true, "false": true, "finally": true,
	"for": true, "from": true, "function": true, "get": true, "if": true,
	"implements": true, "import": true, "in": true, "instanceof": true,
	"interface": true, "let": true, "module": true, "namespace": true,
	"new": true, "null": true, "number": true, "of": true, "package": true,
	"private": true, "protected": true, "public": true, "readonly": true,
	"return": true, "set": true, "static": true, "string": true, "super": true,
	"switch": true, "symbol": true, "this": true, "throw": true, "true": true,
	"try": true, "type": true, "typeof": true, "undefined": true, "var": true,
	"void": true, "while": true, "yield": true,
}

// Words splits an identifier into words on underscores, dashes and spaces.
func Words(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' ' || r == '.' || r == '/'
	})
}

// PascalCase converts snake_case or kebab-case to PascalCase. Short tokens
// containing digits keep their digits ("s3_bucket" -> "S3Bucket").
func PascalCase(s string) string {
	words := Words(s)
	var b strings.Builder
	for _, word := range words {
		runes := []rune(word)
		if len(runes) == 0 {
			continue
		}
		b.WriteRune(unicode.ToUpper(runes[0]))
		b.WriteString(string(runes[1:]))
	}
	return b.String()
}

// CamelCase converts to lowerCamelCase.
func CamelCase(s string) string {
	pascal := PascalCase(s)
	if pascal == "" {
		return ""
	}
	runes := []rune(pascal)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// Sanitize turns arbitrary text into a valid TypeScript identifier.
func Sanitize(s string) string {
	s = CamelCase(s)
	if s == "" {
		return ""
	}
	var b strings.Builder
	for i, r := range s {
		if unicode.IsLetter(r) || r == '_' || r == '$' || (i > 0 && unicode.IsDigit(r)) {
			b.WriteRune(r)
		}
	}
	out := b.String()
	if out == "" {
		return ""
	}
	if reserved[out] {
		out += "Var"
	}
	return out
}

// Namer hands out unique TypeScript identifiers.
type Namer struct {
	used map[string]bool
}

func NewNamer() *Namer {
	return &Namer{used: make(map[string]bool)}
}

// Take returns base, or base2/base3/... when base is already taken.
func (n *Namer) Take(base string) string {
	base = Sanitize(base)
	if base == "" {
		base = "node"
	}
	name := base
	for i := 2; n.used[name]; i++ {
		name = base + strconv.Itoa(i)
	}
	n.used[name] = true
	return name
}

// Reserve marks a name as taken without returning a new one.
func (n *Namer) Reserve(name string) {
	n.used[name] = true
}

// ResourceIdent builds a readable identifier such as logsS3Bucket for
// aws_s3_bucket.logs.
func ResourceIdent(provider, resourceType, name string) string {
	suffix := strings.TrimPrefix(resourceType, provider+"_")
	if suffix == "" {
		suffix = resourceType
	}
	return CamelCase(name) + PascalCase(suffix)
}
