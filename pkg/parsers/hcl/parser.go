// Package hclparser reads Terraform HCL (HCL2 / Terraform 0.12+) into the
// OpenIaC intermediate representation.
package hclparser

import (
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"

	"github.com/be-thomas/openiac/pkg/ir"
)

// ParseDir parses every top-level .tf file in dir.
func ParseDir(dir string) (*ir.Graph, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read directory %s: %w", dir, err)
	}

	var paths []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tf") {
			continue
		}
		paths = append(paths, filepath.Join(dir, entry.Name()))
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no .tf files found in %s", dir)
	}
	sort.Strings(paths)
	return ParseFiles(paths)
}

// ParseFiles parses the given .tf files into one graph.
func ParseFiles(paths []string) (*ir.Graph, error) {
	p := &parser{
		graph: ir.New(strings.Join(paths, ",")),
		src:   make(map[string][]byte, len(paths)),
	}

	type parsedFile struct {
		name string
		file *hcl.File
	}
	files := make([]parsedFile, 0, len(paths))
	hclParser := hclparse.NewParser()

	for _, path := range paths {
		src, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		file, diags := hclParser.ParseHCL(src, path)
		if diags.HasErrors() {
			return nil, fmt.Errorf("parse %s:\n%s", path, diags.Error())
		}
		p.src[path] = src
		files = append(files, parsedFile{name: path, file: file})
	}

	for _, f := range files {
		body, ok := f.file.Body.(*hclsyntax.Body)
		if !ok {
			return nil, fmt.Errorf("%s: expected native HCL syntax", f.name)
		}
		if err := p.parseTopLevel(f.name, body); err != nil {
			return nil, err
		}
	}

	p.resolveEdges()
	return p.graph, nil
}

// ParseString parses HCL source held in memory. It powers the WASM playground.
func ParseString(src, filename string) (*ir.Graph, error) {
	if filename == "" {
		filename = "input.tf"
	}
	p := &parser{
		graph: ir.New(filename),
		src:   map[string][]byte{filename: []byte(src)},
	}
	file, diags := hclparse.NewParser().ParseHCL([]byte(src), filename)
	if diags.HasErrors() {
		return nil, fmt.Errorf("parse %s:\n%s", filename, diags.Error())
	}
	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return nil, fmt.Errorf("%s: expected native HCL syntax", filename)
	}
	if err := p.parseTopLevel(filename, body); err != nil {
		return nil, err
	}
	p.resolveEdges()
	return p.graph, nil
}

type parser struct {
	graph *ir.Graph
	src   map[string][]byte
}

func (p *parser) parseTopLevel(filename string, body *hclsyntax.Body) error {
	src := p.src[filename]

	for _, block := range body.Blocks {
		switch block.Type {
		case "resource":
			if len(block.Labels) != 2 {
				return fmt.Errorf("%s:%d: resource block needs type and name labels", filename, block.TypeRange.Start.Line)
			}
			p.addResource(filename, src, ir.KindResource, block)
		case "data":
			if len(block.Labels) != 2 {
				return fmt.Errorf("%s:%d: data block needs type and name labels", filename, block.TypeRange.Start.Line)
			}
			p.addResource(filename, src, ir.KindData, block)
		case "variable":
			if len(block.Labels) != 1 {
				return fmt.Errorf("%s:%d: variable block needs a name label", filename, block.TypeRange.Start.Line)
			}
			p.addVariable(src, block)
		case "locals":
			p.addLocals(src, block)
		case "output":
			if len(block.Labels) != 1 {
				return fmt.Errorf("%s:%d: output block needs a name label", filename, block.TypeRange.Start.Line)
			}
			p.addOutput(src, block)
		case "module":
			if len(block.Labels) != 1 {
				return fmt.Errorf("%s:%d: module block needs a name label", filename, block.TypeRange.Start.Line)
			}
			p.addModule(src, block)
		case "provider":
			if len(block.Labels) != 1 {
				return fmt.Errorf("%s:%d: provider block needs a name label", filename, block.TypeRange.Start.Line)
			}
			p.addProvider(src, block)
		case "terraform":
			// Backend and required_providers metadata do not affect the graph.
		default:
			// Unknown top-level blocks are ignored, matching Terraform's forward
			// compatibility behaviour for future block types.
		}
	}
	return nil
}

func (p *parser) addResource(filename string, src []byte, kind ir.Kind, block *hclsyntax.Block) {
	resourceType, name := block.Labels[0], block.Labels[1]
	attrs, blocks := parseBody(*block.Body, src)

	id := resourceType + "." + name
	if kind == ir.KindData {
		id = "data." + resourceType + "." + name
	}

	res := &ir.Resource{
		ID:         id,
		Kind:       kind,
		Type:       resourceType,
		Name:       name,
		Provider:   providerPrefix(resourceType),
		Attributes: attrs,
		Blocks:     blocks,
		File:       filename,
		Line:       block.TypeRange.Start.Line,
	}

	if meta, ok := attrs["depends_on"]; ok {
		res.DependsOn = append(res.DependsOn, meta.Refs...)
		delete(attrs, "depends_on")
	}
	if meta, ok := attrs["provider"]; ok {
		res.ProviderAlias = strings.TrimSpace(meta.Expr)
		delete(attrs, "provider")
	}

	p.graph.Resources = append(p.graph.Resources, res)
}

func (p *parser) addVariable(src []byte, block *hclsyntax.Block) {
	attrs, _ := parseBody(*block.Body, src)
	v := &ir.Variable{Name: block.Labels[0]}
	if typeExpr, ok := attrs["type"]; ok {
		v.Type = strings.TrimSpace(typeExpr.Expr)
		if v.Type == "" {
			v.Type = renderValue(typeExpr)
		}
	}
	if def, ok := attrs["default"]; ok {
		v.Default = &def
	}
	if desc, ok := attrs["description"]; ok {
		v.Description = literalString(desc)
	}
	if sensitive, ok := attrs["sensitive"]; ok {
		v.Sensitive = sensitive.Kind == ir.KindBool && sensitive.Literal == true
	}
	p.graph.Variables = append(p.graph.Variables, v)
}

func (p *parser) addLocals(src []byte, block *hclsyntax.Block) {
	body := block.Body
	names := make([]string, 0, len(body.Attributes))
	for name := range body.Attributes {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		p.graph.Locals = append(p.graph.Locals, &ir.Local{
			Name:  name,
			Value: parseExpr(body.Attributes[name].Expr, src),
		})
	}
}

func (p *parser) addOutput(src []byte, block *hclsyntax.Block) {
	attrs, _ := parseBody(*block.Body, src)
	out := &ir.Output{Name: block.Labels[0]}
	if value, ok := attrs["value"]; ok {
		out.Value = value
	}
	if desc, ok := attrs["description"]; ok {
		out.Description = literalString(desc)
	}
	if sensitive, ok := attrs["sensitive"]; ok {
		out.Sensitive = sensitive.Kind == ir.KindBool && sensitive.Literal == true
	}
	p.graph.Outputs = append(p.graph.Outputs, out)
}

func (p *parser) addModule(src []byte, block *hclsyntax.Block) {
	attrs, blocks := parseBody(*block.Body, src)
	m := &ir.Module{
		Name:       block.Labels[0],
		Attributes: attrs,
		Blocks:     blocks,
	}
	if source, ok := attrs["source"]; ok {
		m.Source = literalString(source)
		delete(attrs, "source")
	}
	if version, ok := attrs["version"]; ok {
		m.Version = literalString(version)
		delete(attrs, "version")
	}
	p.graph.Modules = append(p.graph.Modules, m)
}

func (p *parser) addProvider(src []byte, block *hclsyntax.Block) {
	attrs, _ := parseBody(*block.Body, src)
	provider := &ir.Provider{Name: block.Labels[0]}
	if alias, ok := attrs["alias"]; ok {
		provider.Alias = literalString(alias)
		delete(attrs, "alias")
	}
	provider.Attributes = attrs
	p.graph.Providers = append(p.graph.Providers, provider)
}

// resolveEdges turns value references into graph edges now that every node
// address is known.
func (p *parser) resolveEdges() {
	for _, res := range p.graph.Resources {
		for _, ref := range valueRefs(res.Attributes) {
			if p.graph.HasNode(ref) {
				p.graph.AddEdge(res.ID, ref, ir.EdgeReference)
			}
		}
		for _, block := range res.Blocks {
			for _, ref := range blockRefs(block) {
				if p.graph.HasNode(ref) {
					p.graph.AddEdge(res.ID, ref, ir.EdgeReference)
				}
			}
		}
		for _, dep := range res.DependsOn {
			if p.graph.HasNode(dep) {
				p.graph.AddEdge(res.ID, dep, ir.EdgeDependsOn)
			}
		}
	}
}

func valueRefs(values map[string]ir.Value) []string {
	seen := map[string]bool{}
	var out []string
	var walk func(ir.Value)
	walk = func(v ir.Value) {
		for _, ref := range v.Refs {
			if !seen[ref] {
				seen[ref] = true
				out = append(out, ref)
			}
		}
		for _, item := range v.Items {
			walk(item)
		}
		for _, field := range v.Fields {
			walk(field)
		}
	}
	for _, v := range values {
		walk(v)
	}
	sort.Strings(out)
	return out
}

func blockRefs(block ir.Block) []string {
	refs := valueRefs(block.Attributes)
	for _, nested := range block.Blocks {
		refs = append(refs, blockRefs(nested)...)
	}
	sort.Strings(refs)
	return refs
}

func parseBody(body hclsyntax.Body, src []byte) (map[string]ir.Value, []ir.Block) {
	attrs := make(map[string]ir.Value, len(body.Attributes))
	names := make([]string, 0, len(body.Attributes))
	for name := range body.Attributes {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		attrs[name] = parseExpr(body.Attributes[name].Expr, src)
	}

	blocks := make([]ir.Block, 0, len(body.Blocks))
	for _, block := range body.Blocks {
		blockAttrs, nested := parseBody(*block.Body, src)
		blocks = append(blocks, ir.Block{
			Type:       block.Type,
			Labels:     block.Labels,
			Attributes: blockAttrs,
			Blocks:     nested,
			Line:       block.TypeRange.Start.Line,
		})
	}
	return attrs, blocks
}

func parseExpr(expr hclsyntax.Expression, src []byte) ir.Value {
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
func structuredValue(expr hclsyntax.Expression, src []byte) (ir.Value, bool) {
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
func literalValue(expr hclsyntax.Expression) (value ir.Value, ok bool) {
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

func collectRefs(expr hclsyntax.Expression) []string {
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
