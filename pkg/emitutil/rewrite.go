package emitutil

import (
	"sort"
	"strings"
)

// ResolveFunc maps an HCL reference to a target-language expression.
// attribute is empty when the reference does not select an attribute.
type ResolveFunc func(address, attribute string) (string, bool)

// RefRewriter replaces HCL node references inside a raw expression with
// target-language identifiers.
type RefRewriter struct {
	addresses []string
	resolve   ResolveFunc
}

// NewRefRewriter builds a rewriter over the given addresses.
func NewRefRewriter(addresses []string, resolve ResolveFunc) *RefRewriter {
	unique := make(map[string]bool, len(addresses))
	ordered := make([]string, 0, len(addresses))
	for _, address := range addresses {
		if address == "" || unique[address] {
			continue
		}
		unique[address] = true
		ordered = append(ordered, address)
	}
	// Longest first so data.aws_ami.ubuntu wins over data.aws_ami.
	sort.Slice(ordered, func(i, j int) bool {
		if len(ordered[i]) != len(ordered[j]) {
			return len(ordered[i]) > len(ordered[j])
		}
		return ordered[i] < ordered[j]
	})
	return &RefRewriter{addresses: ordered, resolve: resolve}
}

// Rewrite replaces every known address in raw, including inside HCL template
// interpolations. Text inside comments is not rewritten.
func (r *RefRewriter) Rewrite(raw string) string {
	var b strings.Builder
	for i := 0; i < len(raw); {
		switch {
		case raw[i] == '"':
			end := scanQuoted(raw, i)
			b.WriteString(r.rewriteQuoted(raw[i:end]))
			i = end
		case raw[i] == '#' || (raw[i] == '/' && i+1 < len(raw) && raw[i+1] == '/'):
			// Copy the rest of the line unchanged.
			end := strings.IndexByte(raw[i:], '\n')
			if end < 0 {
				b.WriteString(raw[i:])
				i = len(raw)
				continue
			}
			b.WriteString(raw[i : i+end+1])
			i += end + 1
		default:
			if replacement, consumed, ok := r.matchAt(raw, i); ok {
				b.WriteString(replacement)
				i += consumed
				continue
			}
			b.WriteByte(raw[i])
			i++
		}
	}
	return b.String()
}

func (r *RefRewriter) matchAt(s string, i int) (string, int, bool) {
	if i > 0 && isIdentOrDot(s[i-1]) {
		return "", 0, false
	}
	for _, address := range r.addresses {
		if !strings.HasPrefix(s[i:], address) {
			continue
		}
		end := i + len(address)
		if end < len(s) && isIdentChar(s[end]) {
			continue
		}

		attribute := ""
		consumed := len(address)
		if end < len(s) && s[end] == '.' {
			j := end + 1
			for j < len(s) && isIdentChar(s[j]) {
				j++
			}
			if j > end+1 {
				attribute = s[end+1 : j]
				consumed = j - i
			}
		}
		replacement, ok := r.resolve(address, attribute)
		if !ok {
			return "", 0, false
		}
		return replacement, consumed, true
	}
	return "", 0, false
}

// rewriteQuoted rewrites interpolations inside an HCL quoted string while
// leaving literal text alone.
func (r *RefRewriter) rewriteQuoted(quoted string) string {
	var b strings.Builder
	for i := 0; i < len(quoted); {
		if quoted[i] == '\\' && i+1 < len(quoted) {
			b.WriteString(quoted[i : i+2])
			i += 2
			continue
		}
		if quoted[i] == '$' && i+2 < len(quoted) && quoted[i+1] == '{' {
			end := matchingBrace(quoted, i+1)
			if end < 0 {
				b.WriteString(quoted[i:])
				break
			}
			b.WriteString("${")
			b.WriteString(r.Rewrite(quoted[i+2 : end]))
			b.WriteByte('}')
			i = end + 1
			continue
		}
		if quoted[i] == '%' && i+2 < len(quoted) && quoted[i+1] == '{' {
			end := matchingBrace(quoted, i+1)
			if end < 0 {
				b.WriteString(quoted[i:])
				break
			}
			b.WriteString("%{")
			b.WriteString(r.Rewrite(quoted[i+2 : end]))
			b.WriteByte('}')
			i = end + 1
			continue
		}
		b.WriteByte(quoted[i])
		i++
	}
	return b.String()
}

// scanQuoted returns the index just past the closing quote.
func scanQuoted(s string, start int) int {
	for i := start + 1; i < len(s); i++ {
		if s[i] == '\\' {
			i++
			continue
		}
		if s[i] == '"' {
			return i + 1
		}
	}
	return len(s)
}

func matchingBrace(s string, open int) int {
	depth := 0
	for i := open; i < len(s); i++ {
		switch s[i] {
		case '"':
			i = scanQuoted(s, i) - 1
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func isIdentChar(b byte) bool {
	return b == '_' || b == '$' || b >= '0' && b <= '9' || b >= 'A' && b <= 'Z' || b >= 'a' && b <= 'z'
}

func isIdentOrDot(b byte) bool {
	return isIdentChar(b) || b == '.' || b == '-'
}

// HasHeredoc reports whether a raw expression contains a heredoc, which the
// v0 rewriter does not transform.
func HasHeredoc(raw string) bool {
	return strings.Contains(raw, "<<")
}
