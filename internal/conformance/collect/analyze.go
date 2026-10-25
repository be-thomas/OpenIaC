package collect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	hcljson "github.com/hashicorp/hcl/v2/json"
	"github.com/zclconf/go-cty/cty"
)

// LocalSource is a relative module source found in a Terraform file.
type LocalSource struct {
	Dir    string
	Source string
}

// Analysis is the syntax-level summary of one module root. It is built from
// parsed HCL, so comments, strings and unrelated attributes cannot invent
// module references or provider usage.
type Analysis struct {
	Features     []string
	Providers    []string
	LocalSources []LocalSource
	Violations   []PolicyViolation
}

// AnalyzeFiles parses every Terraform file and returns its structural
// features, providers, local module sources and policy violations. A parse
// error rejects the whole module root: unparseable input is not a conformance
// fixture.
func AnalyzeFiles(files map[string][]byte) (*Analysis, error) {
	a := &Analysis{}
	features := make(map[string]bool)
	providers := make(map[string]bool)

	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		content := files[name]
		lower := strings.ToLower(name)
		var err error
		if strings.HasSuffix(lower, ".tf.json") {
			err = analyzeJSONFile(name, content, a, features, providers)
		} else {
			err = analyzeHCLFile(name, content, a, features, providers)
		}
		if err != nil {
			return nil, err
		}
		a.addRawSecretViolations(name, content)
	}

	a.Features = sortedSet(features)
	a.Providers = sortedSet(providers)
	return a, nil
}

func analyzeHCLFile(name string, content []byte, a *Analysis, features, providers map[string]bool) error {
	f, diags := hclsyntax.ParseConfig(content, name, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return fmt.Errorf("%s: %s", name, diags.Error())
	}
	body, ok := f.Body.(*hclsyntax.Body)
	if !ok {
		return fmt.Errorf("%s: unsupported HCL body", name)
	}
	analyzeHCLBody(name, content, body, a, features, providers)
	return nil
}

func analyzeHCLBody(name string, content []byte, body *hclsyntax.Body, a *Analysis, features, providers map[string]bool) {
	for attrName, attr := range body.Attributes {
		switch attrName {
		case "count":
			features["count"] = true
		case "for_each":
			features["for_each"] = true
		case "depends_on":
			features["depends_on"] = true
		case "sensitive":
			if isTrueLiteral(attr.Expr) {
				features["sensitive"] = true
			}
		}
		detectExprFeatures(attr.Expr, features)
		if strings.HasPrefix(strings.TrimSpace(expressionText(content, attr.Expr)), "<<") {
			features["heredoc"] = true
		}
		a.checkSecretAttribute(name, attrName, attr.Expr)
	}

	for _, block := range body.Blocks {
		switch block.Type {
		case "terraform":
			features["terraform"] = true
		case "variable":
			features["variables"] = true
			if len(block.Labels) > 0 {
				a.checkSecretDefault(name, block.Labels[0], block)
			}
		case "output":
			features["outputs"] = true
		case "locals":
			features["locals"] = true
		case "module":
			features["modules"] = true
			if len(block.Labels) > 0 {
				if src, ok := literalStringAttribute(block.Body.Attributes["source"]); ok {
					if isLocalSource(src) {
						a.LocalSources = append(a.LocalSources, LocalSource{Dir: dirOf(name), Source: src})
					}
				}
			}
		case "resource":
			features["resources"] = true
			if len(block.Labels) > 0 {
				if p := providerFromType(block.Labels[0]); p != "" {
					providers[p] = true
				}
			}
		case "data":
			features["data"] = true
			if len(block.Labels) > 0 {
				if p := providerFromType(block.Labels[0]); p != "" {
					providers[p] = true
				}
				if block.Labels[0] == "external" {
					a.Violations = append(a.Violations, PolicyViolation{
						Kind: "executable", Path: name,
						Detail: "external data sources are not allowed",
					})
				}
			}
		case "provider":
			features["providers"] = true
			if len(block.Labels) > 0 {
				providers[block.Labels[0]] = true
			}
		case "backend":
			features["backend"] = true
		case "required_providers":
			features["required_providers"] = true
		case "dynamic":
			features["dynamic"] = true
		case "check":
			features["check"] = true
		case "import":
			features["import"] = true
		case "moved":
			features["moved"] = true
		case "provisioner":
			a.Violations = append(a.Violations, PolicyViolation{
				Kind: "executable", Path: name,
				Detail: "provisioners are not allowed",
			})
		}
		analyzeHCLBody(name, content, block.Body, a, features, providers)
	}
}

func (a *Analysis) checkSecretAttribute(file, name string, expr hcl.Expression) {
	if value, ok := literalString(expr); ok && credentialLiteral(name, value) {
		a.Violations = append(a.Violations, PolicyViolation{
			Kind: "secret", Path: file,
			Detail: "credential-like literal for " + name,
		})
	}
	// A secret can also hide in an object, list, function call or conditional,
	// for example `default = { api_key = "..." }` or `jsonencode({ password = "..." })`.
	a.checkSecretValue(file, expr)
}

// checkSecretValue walks the expression syntax so nested literals are seen
// whatever wraps them.
func (a *Analysis) checkSecretValue(file string, expr hcl.Expression) {
	switch e := expr.(type) {
	case *hclsyntax.ObjectConsExpr:
		for _, item := range e.Items {
			key, _ := literalString(item.KeyExpr)
			if value, ok := literalString(item.ValueExpr); ok && key != "" && credentialLiteral(key, value) {
				a.Violations = append(a.Violations, PolicyViolation{
					Kind: "secret", Path: file,
					Detail: "credential-like literal for " + key,
				})
			}
			a.checkSecretValue(file, item.KeyExpr)
			a.checkSecretValue(file, item.ValueExpr)
		}
	case *hclsyntax.TupleConsExpr:
		for _, item := range e.Exprs {
			a.checkSecretValue(file, item)
		}
	case *hclsyntax.FunctionCallExpr:
		for _, arg := range e.Args {
			a.checkSecretValue(file, arg)
		}
	case *hclsyntax.ConditionalExpr:
		a.checkSecretValue(file, e.Condition)
		a.checkSecretValue(file, e.TrueResult)
		a.checkSecretValue(file, e.FalseResult)
	case *hclsyntax.ParenthesesExpr:
		a.checkSecretValue(file, e.Expression)
	case *hclsyntax.TemplateExpr:
		for _, part := range e.Parts {
			a.checkSecretValue(file, part)
		}
	case *hclsyntax.TemplateWrapExpr:
		a.checkSecretValue(file, e.Wrapped)
	case *hclsyntax.UnaryOpExpr:
		a.checkSecretValue(file, e.Val)
	case *hclsyntax.BinaryOpExpr:
		a.checkSecretValue(file, e.LHS)
		a.checkSecretValue(file, e.RHS)
	case *hclsyntax.IndexExpr:
		a.checkSecretValue(file, e.Collection)
		a.checkSecretValue(file, e.Key)
	case *hclsyntax.SplatExpr:
		a.checkSecretValue(file, e.Source)
		a.checkSecretValue(file, e.Each)
	case *hclsyntax.RelativeTraversalExpr:
		a.checkSecretValue(file, e.Source)
	case *hclsyntax.ForExpr:
		a.checkSecretValue(file, e.CollExpr)
		a.checkSecretValue(file, e.KeyExpr)
		a.checkSecretValue(file, e.ValExpr)
		a.checkSecretValue(file, e.CondExpr)
	case *hclsyntax.ObjectConsKeyExpr:
		a.checkSecretValue(file, e.Wrapped)
	}
}

func (a *Analysis) checkSecretDefault(file, variable string, block *hclsyntax.Block) {
	attr, ok := block.Body.Attributes["default"]
	if !ok {
		return
	}
	if value, ok := literalString(attr.Expr); ok && credentialLiteral(variable, value) {
		a.Violations = append(a.Violations, PolicyViolation{
			Kind: "secret", Path: file,
			Detail: "credential-like default for variable " + variable,
		})
	}
	// Nested object keys are scanned when the block body is walked, so this
	// must not scan the default a second time.
}

func (a *Analysis) addRawSecretViolations(file string, content []byte) {
	for _, re := range rawSecretRE {
		loc := re.FindIndex(content)
		if loc == nil {
			continue
		}
		if looksLikePlaceholder(string(content[loc[0]:loc[1]])) {
			continue
		}
		a.Violations = append(a.Violations, PolicyViolation{
			Kind: "secret", Path: file, Detail: "credential material",
		})
		return
	}
}

func analyzeJSONFile(name string, content []byte, a *Analysis, features, providers map[string]bool) error {
	if _, diags := hcljson.Parse(content, name); diags.HasErrors() {
		return fmt.Errorf("%s: %s", name, diags.Error())
	}
	var root map[string]any
	dec := json.NewDecoder(bytes.NewReader(content))
	dec.UseNumber()
	if err := dec.Decode(&root); err != nil {
		return fmt.Errorf("%s: %s", name, err)
	}
	dir := dirOf(name)
	for key, value := range root {
		switch key {
		case "terraform":
			features["terraform"] = true
			analyzeJSONBody(name, object(value), a, features, providers)
		case "variable":
			features["variables"] = true
			for variable, body := range object(value) {
				bodyObj := object(body)
				a.checkSecretJSONDefault(name, variable, bodyObj)
				analyzeJSONBody(name, bodyObj, a, features, providers)
			}
		case "output":
			features["outputs"] = true
			for _, body := range object(value) {
				analyzeJSONBody(name, object(body), a, features, providers)
			}
		case "locals":
			features["locals"] = true
			analyzeJSONBody(name, object(value), a, features, providers)
		case "module":
			features["modules"] = true
			for _, body := range object(value) {
				bodyObj := object(body)
				if src, ok := bodyObj["source"].(string); ok && isLocalSource(src) {
					a.LocalSources = append(a.LocalSources, LocalSource{Dir: dir, Source: src})
				}
				analyzeJSONBody(name, bodyObj, a, features, providers)
			}
		case "resource", "data":
			if key == "resource" {
				features["resources"] = true
			} else {
				features["data"] = true
			}
			for resourceType, byName := range object(value) {
				if p := providerFromType(resourceType); p != "" {
					providers[p] = true
				}
				if key == "data" && resourceType == "external" {
					a.Violations = append(a.Violations, PolicyViolation{
						Kind: "executable", Path: name,
						Detail: "external data sources are not allowed",
					})
				}
				for _, body := range object(byName) {
					analyzeJSONBody(name, object(body), a, features, providers)
				}
			}
		case "provider":
			features["providers"] = true
			for providerName, body := range object(value) {
				providers[providerName] = true
				analyzeJSONBody(name, object(body), a, features, providers)
			}
		case "check":
			features["check"] = true
		case "import":
			features["import"] = true
		case "moved":
			features["moved"] = true
		}
	}
	return nil
}

func analyzeJSONBody(file string, body map[string]any, a *Analysis, features, providers map[string]bool) {
	for name, value := range body {
		switch name {
		case "count":
			features["count"] = true
		case "for_each":
			features["for_each"] = true
		case "depends_on":
			features["depends_on"] = true
		case "sensitive":
			if b, ok := value.(bool); ok && b {
				features["sensitive"] = true
			}
		case "backend":
			features["backend"] = true
		case "required_providers":
			features["required_providers"] = true
		case "provisioner":
			a.Violations = append(a.Violations, PolicyViolation{
				Kind: "executable", Path: file, Detail: "provisioners are not allowed",
			})
		}
		if s, ok := value.(string); ok && credentialLiteral(name, s) {
			a.Violations = append(a.Violations, PolicyViolation{
				Kind: "secret", Path: file, Detail: "credential-like literal for " + name,
			})
		}
		if s, ok := value.(string); ok {
			detectJSONStringFeatures(s, features)
		}
		switch v := value.(type) {
		case map[string]any:
			analyzeJSONBody(file, v, a, features, providers)
		case []any:
			for _, item := range v {
				if m, ok := item.(map[string]any); ok {
					analyzeJSONBody(file, m, a, features, providers)
				}
			}
		}
	}
}

func (a *Analysis) checkSecretJSONDefault(file, variable string, body map[string]any) {
	s, ok := body["default"].(string)
	if !ok || !credentialLiteral(variable, s) {
		return
	}
	a.Violations = append(a.Violations, PolicyViolation{
		Kind: "secret", Path: file,
		Detail: "credential-like default for variable " + variable,
	})
}

var (
	// credentialNameRE matches a credential word anywhere in a snake_case name,
	// so `secret_string`, `password_value` and `db_password` are all considered.
	// A name that matches only this is judged by its value, because qualifiers
	// such as `github_token_type` or `secret_path_format` land here too.
	credentialNameRE = regexp.MustCompile(`(?i)(?:^|_)(password|passwd|secret|token|api_?key|access_?key|private_?key|credentials?|connection_?string)(?:$|_)`)

	// strongCredentialNameRE ends on the credential word itself, optionally
	// followed by a value-shaped suffix. These names are treated as credentials
	// whatever the value length, so `password = "hunter2"` cannot slip through.
	strongCredentialNameRE = regexp.MustCompile(`(?i)(?:^|_)(password|passwd|secret_?key|secret|api_?key|access_?key|private_?key|credentials?|connection_?string)(?:_id|_value|_string|_pem|_text)?$`)

	rawSecretRE = []*regexp.Regexp{
		regexp.MustCompile(`\b(?:AKIA|ASIA)[0-9A-Z]{16}\b`),
		regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----`),
	}
)

// credentialLiteral reports whether a name/value pair is a credential literal.
func credentialLiteral(name, value string) bool {
	snake := snakeName(name)
	if !credentialNameRE.MatchString(snake) || credentialQualified(snake) || looksLikePlaceholder(value) {
		return false
	}
	if strongCredentialNameRE.MatchString(snake) {
		return true
	}
	return looksLikeSecretValue(value)
}

// credentialQualified reports names that describe a credential rather than
// hold one, such as `aws_secret_manager_git_private_ssh_key_name`.
func credentialQualified(snake string) bool {
	i := strings.LastIndexByte(snake, '_')
	if i < 0 {
		return false
	}
	switch snake[i+1:] {
	case "name", "type", "path", "format", "algorithm", "unit", "units",
		"strategy", "group", "arn", "description", "label":
		return true
	}
	return false
}

// snakeName converts camelCase identifiers to snake_case so one set of name
// rules covers `dbPassword` and `db_password`.
func snakeName(name string) string {
	var b strings.Builder
	var prev rune
	for i, r := range name {
		if unicode.IsUpper(r) && i > 0 && (unicode.IsLower(prev) || unicode.IsDigit(prev)) {
			b.WriteByte('_')
		}
		b.WriteRune(unicode.ToLower(r))
		prev = r
	}
	return b.String()
}

// detectJSONStringFeatures classifies expressions embedded in a .tf.json
// string, so a JSON module is not reported with fewer features than its HCL
// twin.
func detectJSONStringFeatures(s string, features map[string]bool) {
	if !strings.Contains(s, "${") && !strings.Contains(s, "%{") {
		return
	}
	features["interpolation"] = true
	for _, body := range interpolationBodies(s) {
		expr, diags := hclsyntax.ParseExpression([]byte(body), "json", hcl.Pos{Line: 1, Column: 1})
		if diags.HasErrors() {
			continue
		}
		detectExprFeatures(expr, features)
	}
}

// interpolationBodies returns the expression inside each `${...}` occurrence.
// Brace counting is textual, so a brace inside a quoted string can end the scan
// early; the caller falls back to the interpolation feature alone when the
// fragment does not parse.
func interpolationBodies(s string) []string {
	var out []string
	for i := 0; ; {
		start := strings.Index(s[i:], "${")
		if start < 0 {
			return out
		}
		start += i
		depth := 0
		end := -1
		for j := start + 1; j < len(s); j++ {
			switch s[j] {
			case '{':
				depth++
			case '}':
				depth--
				if depth == 0 {
					end = j
				}
			}
			if end >= 0 {
				break
			}
		}
		if end < 0 {
			return out
		}
		out = append(out, s[start+2:end])
		i = end + 1
	}
}

// detectExprFeatures walks the expression syntax so a string literal cannot
// invent a construct the configuration does not use.
func detectExprFeatures(expr hcl.Expression, features map[string]bool) {
	switch e := expr.(type) {
	case *hclsyntax.ConditionalExpr:
		features["conditionals"] = true
		detectExprFeatures(e.Condition, features)
		detectExprFeatures(e.TrueResult, features)
		detectExprFeatures(e.FalseResult, features)
	case *hclsyntax.ForExpr:
		features["for_expressions"] = true
		detectExprFeatures(e.CollExpr, features)
		detectExprFeatures(e.KeyExpr, features)
		detectExprFeatures(e.ValExpr, features)
		detectExprFeatures(e.CondExpr, features)
	case *hclsyntax.FunctionCallExpr:
		features["functions"] = true
		for _, arg := range e.Args {
			detectExprFeatures(arg, features)
		}
	case *hclsyntax.TemplateExpr:
		for _, part := range e.Parts {
			if _, literal := part.(*hclsyntax.LiteralValueExpr); !literal {
				features["interpolation"] = true
			}
			detectExprFeatures(part, features)
		}
	case *hclsyntax.TemplateWrapExpr:
		features["interpolation"] = true
		detectExprFeatures(e.Wrapped, features)
	case *hclsyntax.ParenthesesExpr:
		detectExprFeatures(e.Expression, features)
	case *hclsyntax.UnaryOpExpr:
		detectExprFeatures(e.Val, features)
	case *hclsyntax.BinaryOpExpr:
		detectExprFeatures(e.LHS, features)
		detectExprFeatures(e.RHS, features)
	case *hclsyntax.IndexExpr:
		detectExprFeatures(e.Collection, features)
		detectExprFeatures(e.Key, features)
	case *hclsyntax.SplatExpr:
		detectExprFeatures(e.Source, features)
		detectExprFeatures(e.Each, features)
	case *hclsyntax.RelativeTraversalExpr:
		detectExprFeatures(e.Source, features)
	case *hclsyntax.ObjectConsExpr:
		for _, item := range e.Items {
			detectExprFeatures(item.KeyExpr, features)
			detectExprFeatures(item.ValueExpr, features)
		}
	case *hclsyntax.ObjectConsKeyExpr:
		detectExprFeatures(e.Wrapped, features)
	case *hclsyntax.TupleConsExpr:
		for _, item := range e.Exprs {
			detectExprFeatures(item, features)
		}
	}
}

func expressionText(content []byte, expr hcl.Expression) string {
	r := expr.Range()
	if r.Start.Byte < 0 || r.End.Byte > len(content) || r.Start.Byte >= r.End.Byte {
		return ""
	}
	return string(content[r.Start.Byte:r.End.Byte])
}

func literalString(expr hcl.Expression) (string, bool) {
	value, diags := expr.Value(nil)
	if diags.HasErrors() || value.IsNull() || !value.Type().Equals(cty.String) {
		return "", false
	}
	return value.AsString(), true
}

func literalStringAttribute(attr *hclsyntax.Attribute) (string, bool) {
	if attr == nil {
		return "", false
	}
	return literalString(attr.Expr)
}

func isTrueLiteral(expr hcl.Expression) bool {
	value, diags := expr.Value(nil)
	if diags.HasErrors() || value.IsNull() || !value.Type().Equals(cty.Bool) {
		return false
	}
	return value.True()
}

func isLocalSource(src string) bool {
	return strings.HasPrefix(src, "./") || strings.HasPrefix(src, "../")
}

func providerFromType(resourceType string) string {
	if i := strings.IndexByte(resourceType, '_'); i > 0 {
		return resourceType[:i]
	}
	return ""
}

func dirOf(file string) string {
	dir := path.Dir(file)
	if dir == "." {
		return ""
	}
	return dir
}

func object(value any) map[string]any {
	if m, ok := value.(map[string]any); ok {
		return m
	}
	return map[string]any{}
}

func sortedSet(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// looksLikeSecretValue filters out literals that sit under a credential-shaped
// name but read as configuration: units and enum names ("minutes", "required")
// carry no digit and are too short to be a credential, while a leaked password
// or key either contains a digit or is long enough not to be a fixed word.
func looksLikeSecretValue(s string) bool {
	if looksLikePlaceholder(s) {
		return false
	}
	if len(s) < 6 {
		return false
	}
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return len(s) >= 24
}

func looksLikePlaceholder(s string) bool {
	t := strings.ToLower(strings.TrimSpace(s))
	if t == "" || t == "none" || t == "null" || t == "changeme" || t == "change-me" {
		return true
	}
	// A value wrapped in angle brackets is a template slot, not a literal.
	if strings.HasPrefix(t, "<") && strings.HasSuffix(t, ">") {
		return true
	}
	if isPlaceholderSequence(t) {
		return true
	}
	// A single format or file-extension word describes the secret's encoding,
	// not a credential: `secret_extensions = { private_key = "key" }`.
	switch t {
	case "key", "pem", "crt", "cer", "pfx", "jks", "der", "json", "yaml", "yml", "txt":
		return true
	}
	for _, marker := range []string{
		"example", "fake", "dummy", "placeholder", "redacted", "xxxx", "****",
		"your-", "your_", "${", "%{", "var.", "local.", "data.",
		"module.", "aws_", "arn:", "alias/", "foobar",
	} {
		if strings.Contains(t, marker) {
			return true
		}
	}
	return false
}

// isPlaceholderSequence reports the filler values examples use in place of a
// credential: a repeated character ("000000") or a run of ascending bytes
// ("1234567890", "abcdef").
func isPlaceholderSequence(s string) bool {
	if len(s) < 6 {
		return false
	}
	repeated, ascending := true, true
	for i := 1; i < len(s); i++ {
		if s[i] != s[0] {
			repeated = false
		}
		if s[i] != s[i-1]+1 && !(s[i-1] == '9' && s[i] == '0') {
			ascending = false
		}
	}
	return repeated || ascending
}
