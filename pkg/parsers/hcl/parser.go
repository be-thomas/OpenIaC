// Package hclparser reads Terraform configuration into the OpenIaC
// intermediate representation.
//
// Both syntaxes Terraform accepts are supported: native HCL (.tf) and JSON
// (.tf.json). Native HCL keeps its full block structure; JSON bodies below the
// top level cannot express blocks, so an object property becomes a map value.
//
// The parser accepts the whole Terraform 1.x configuration language. Constructs
// the IR has no typed node for — moved, import, removed, check, ephemeral,
// action and anything newer — are preserved in the graph instead of dropped, so
// emitters can warn about them. See docs/TERRAFORM-VERSIONS.md.
package hclparser

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"

	"github.com/be-thomas/openiac/pkg/ir"
	"github.com/be-thomas/openiac/pkg/tfversion"
)

// ParseDir parses every Terraform configuration file directly inside dir.
// Subdirectories are not descended into: they are usually modules.
func ParseDir(dir string) (*ir.Graph, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read directory %s: %w", dir, err)
	}

	var paths []string
	for _, entry := range entries {
		if entry.IsDir() || !isTerraformFile(entry.Name()) {
			continue
		}
		paths = append(paths, filepath.Join(dir, entry.Name()))
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no Terraform configuration files (*.tf, *.tf.json) found in %s", dir)
	}
	sort.Strings(paths)
	return ParseFiles(paths)
}

// ParseFiles parses the given Terraform configuration files into one graph.
// Native HCL and JSON files may be mixed.
func ParseFiles(paths []string) (*ir.Graph, error) {
	p := &parser{
		graph: ir.New(strings.Join(paths, ",")),
		src:   make(map[string][]byte, len(paths)),
	}
	hclParser := hclparse.NewParser()

	for _, path := range paths {
		src, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}

		var (
			file  *hcl.File
			diags hcl.Diagnostics
		)
		if isJSONFile(path) {
			file, diags = hclParser.ParseJSON(src, path)
		} else {
			file, diags = hclParser.ParseHCL(src, path)
		}
		if diags.HasErrors() {
			return nil, fmt.Errorf("parse %s:\n%s", path, diags.Error())
		}

		p.src[path] = src
		if err := p.parseFile(path, file); err != nil {
			return nil, err
		}
	}
	return p.finish()
}

// ParseString parses Terraform source held in memory. It powers the WASM
// playground. The filename's suffix selects the syntax, exactly as ParseFiles
// does.
func ParseString(src, filename string) (*ir.Graph, error) {
	if filename == "" {
		filename = "input.tf"
	}

	p := &parser{
		graph: ir.New(filename),
		src:   map[string][]byte{filename: []byte(src)},
	}

	hclParser := hclparse.NewParser()
	var (
		file  *hcl.File
		diags hcl.Diagnostics
	)
	if isJSONFile(filename) {
		file, diags = hclParser.ParseJSON([]byte(src), filename)
	} else {
		file, diags = hclParser.ParseHCL([]byte(src), filename)
	}
	if diags.HasErrors() {
		return nil, fmt.Errorf("parse %s:\n%s", filename, diags.Error())
	}

	if err := p.parseFile(filename, file); err != nil {
		return nil, err
	}
	return p.finish()
}

func isTerraformFile(name string) bool {
	return strings.HasSuffix(name, ".tf") || isJSONFile(name)
}

func isJSONFile(name string) bool {
	return strings.HasSuffix(strings.ToLower(name), ".tf.json")
}

type parser struct {
	graph *ir.Graph
	src   map[string][]byte
}

// block is the syntax-independent view of a top-level block. Native HCL and
// JSON differ only in how the block is recovered; everything downstream reads
// the body through hcl.Body.
type block struct {
	file   string
	typ    string
	labels []string
	body   hcl.Body
	line   int
}

func (p *parser) parseFile(path string, file *hcl.File) error {
	if native, ok := file.Body.(*hclsyntax.Body); ok {
		return p.parseNative(path, native)
	}
	return p.parseJSON(path, file.Body)
}

func (p *parser) parseNative(filename string, body *hclsyntax.Body) error {
	for _, b := range body.Blocks {
		blk := block{
			file:   filename,
			typ:    b.Type,
			labels: b.Labels,
			body:   b.Body,
			line:   b.TypeRange.Start.Line,
		}
		if err := p.addTopLevel(blk); err != nil {
			return err
		}
	}

	// Terraform rejects arguments at the top level, but a parser that silently
	// drops them is worse than one that keeps them.
	names := make([]string, 0, len(body.Attributes))
	for name := range body.Attributes {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		attr := body.Attributes[name]
		value := parseExpr(attr.Expr, p.src[filename])
		p.graph.Preserved = append(p.graph.Preserved, ir.Block{
			Type: name,
			Body: &value,
			File: filename,
			Line: attr.SrcRange.Start.Line,
		})
	}
	return nil
}

func (p *parser) parseJSON(filename string, body hcl.Body) error {
	content, remain, diags := body.PartialContent(topLevelSchema)
	if diags.HasErrors() {
		return fmt.Errorf("parse %s:\n%s", filename, diags.Error())
	}

	for _, b := range content.Blocks {
		blk := block{
			file:   filename,
			typ:    b.Type,
			labels: b.Labels,
			body:   b.Body,
			line:   b.TypeRange.Start.Line,
		}
		if err := p.addTopLevel(blk); err != nil {
			return err
		}
	}

	// The schema hides what it consumed, so everything left is an unknown
	// top-level property: either a block type this parser does not know yet, or
	// an argument Terraform would reject. Preserve it.
	attributes, attrDiags := remain.JustAttributes()
	if attrDiags.HasErrors() {
		return fmt.Errorf("parse %s:\n%s", filename, attrDiags.Error())
	}
	names := make([]string, 0, len(attributes))
	for name := range attributes {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		p.preserveJSONProperty(filename, name, attributes[name].Expr, attributes[name].NameRange.Start.Line)
	}
	return nil
}

func (p *parser) addTopLevel(blk block) error {
	if want, known := topLevelLabelCount[blk.typ]; known && len(blk.labels) != want {
		return labelError(blk)
	}

	switch blk.typ {
	case "terraform":
		return p.addTerraform(blk)
	case "locals":
		return p.addLocals(blk)
	case "resource":
		return p.addResource(blk, ir.KindResource)
	case "data":
		return p.addResource(blk, ir.KindData)
	case "variable":
		return p.addVariable(blk)
	case "output":
		return p.addOutput(blk)
	case "module":
		return p.addModule(blk)
	case "provider":
		return p.addProvider(blk)
	default:
		return p.preserveBlock(blk)
	}
}

func labelError(blk block) error {
	switch topLevelLabelCount[blk.typ] {
	case 0:
		return fmt.Errorf("%s:%d: %s block takes no labels", blk.file, blk.line, blk.typ)
	case 1:
		return fmt.Errorf("%s:%d: %s block needs a name label", blk.file, blk.line, blk.typ)
	default:
		return fmt.Errorf("%s:%d: %s block needs type and name labels", blk.file, blk.line, blk.typ)
	}
}

func (p *parser) addResource(blk block, kind ir.Kind) error {
	src := p.src[blk.file]
	attrs, blocks, err := parseBody(blk.body, src)
	if err != nil {
		return fmt.Errorf("%s:%d: %w", blk.file, blk.line, err)
	}

	resourceType, name := blk.labels[0], blk.labels[1]
	id := resourceType + "." + name
	if kind == ir.KindData {
		id = "data." + id
	}

	res := &ir.Resource{
		ID:         id,
		Kind:       kind,
		Type:       resourceType,
		Name:       name,
		Provider:   providerPrefix(resourceType),
		Attributes: attrs,
		Blocks:     blocks,
		File:       blk.file,
		Line:       blk.line,
	}

	if meta, ok := attrs["depends_on"]; ok {
		res.DependsOn = append(res.DependsOn, valueAddresses(meta)...)
		delete(attrs, "depends_on")
	}
	if meta, ok := attrs["provider"]; ok {
		res.ProviderAlias = providerAddress(meta)
		delete(attrs, "provider")
	}
	if len(attrs) == 0 {
		res.Attributes = nil
	}

	p.graph.Resources = append(p.graph.Resources, res)
	return nil
}

func (p *parser) addVariable(blk block) error {
	attrs, blocks, err := parseBody(blk.body, p.src[blk.file])
	if err != nil {
		return fmt.Errorf("%s:%d: %w", blk.file, blk.line, err)
	}

	v := &ir.Variable{
		Name:   blk.labels[0],
		Blocks: blocks,
		File:   blk.file,
		Line:   blk.line,
	}

	if typeExpr, ok := attrs["type"]; ok {
		v.Type = typeText(typeExpr)
		delete(attrs, "type")
	}
	if def, ok := attrs["default"]; ok {
		v.Default = &def
		delete(attrs, "default")
	}
	if desc, ok := attrs["description"]; ok {
		v.Description = literalString(desc)
		delete(attrs, "description")
	}
	if sensitive, ok := attrs["sensitive"]; ok && sensitive.Kind == ir.KindBool {
		v.Sensitive = sensitive.Literal == true
		delete(attrs, "sensitive")
	}
	if nullable, ok := attrs["nullable"]; ok && nullable.Kind == ir.KindBool {
		value := nullable.Literal == true
		v.Nullable = &value
		delete(attrs, "nullable")
	}
	if ephemeral, ok := attrs["ephemeral"]; ok && ephemeral.Kind == ir.KindBool {
		v.Ephemeral = ephemeral.Literal == true
		delete(attrs, "ephemeral")
	}
	if deprecated, ok := attrs["deprecated"]; ok {
		v.Deprecated = literalString(deprecated)
		delete(attrs, "deprecated")
	}
	if len(attrs) > 0 {
		v.Attributes = attrs
	}

	p.graph.Variables = append(p.graph.Variables, v)
	return nil
}

func (p *parser) addLocals(blk block) error {
	attrs, blocks, err := parseBody(blk.body, p.src[blk.file])
	if err != nil {
		return fmt.Errorf("%s:%d: %w", blk.file, blk.line, err)
	}

	names := make([]string, 0, len(attrs))
	for name := range attrs {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		p.graph.Locals = append(p.graph.Locals, &ir.Local{
			Name:  name,
			Value: attrs[name],
			File:  blk.file,
			Line:  blk.line,
		})
	}

	// A locals block cannot contain nested blocks. If one appears anyway, keep
	// it rather than drop it.
	if len(blocks) > 0 {
		p.graph.Preserved = append(p.graph.Preserved, ir.Block{
			Type:   blk.typ,
			Blocks: blocks,
			File:   blk.file,
			Line:   blk.line,
		})
	}
	return nil
}

func (p *parser) addOutput(blk block) error {
	attrs, blocks, err := parseBody(blk.body, p.src[blk.file])
	if err != nil {
		return fmt.Errorf("%s:%d: %w", blk.file, blk.line, err)
	}

	out := &ir.Output{
		Name:   blk.labels[0],
		Blocks: blocks,
		File:   blk.file,
		Line:   blk.line,
	}

	if value, ok := attrs["value"]; ok {
		out.Value = value
		delete(attrs, "value")
	}
	if desc, ok := attrs["description"]; ok {
		out.Description = literalString(desc)
		delete(attrs, "description")
	}
	if sensitive, ok := attrs["sensitive"]; ok && sensitive.Kind == ir.KindBool {
		out.Sensitive = sensitive.Literal == true
		delete(attrs, "sensitive")
	}
	if typeExpr, ok := attrs["type"]; ok {
		out.Type = typeText(typeExpr)
		delete(attrs, "type")
	}
	if ephemeral, ok := attrs["ephemeral"]; ok && ephemeral.Kind == ir.KindBool {
		out.Ephemeral = ephemeral.Literal == true
		delete(attrs, "ephemeral")
	}
	if deprecated, ok := attrs["deprecated"]; ok {
		out.Deprecated = literalString(deprecated)
		delete(attrs, "deprecated")
	}
	if len(attrs) > 0 {
		out.Attributes = attrs
	}

	p.graph.Outputs = append(p.graph.Outputs, out)
	return nil
}

func (p *parser) addModule(blk block) error {
	attrs, blocks, err := parseBody(blk.body, p.src[blk.file])
	if err != nil {
		return fmt.Errorf("%s:%d: %w", blk.file, blk.line, err)
	}

	m := &ir.Module{
		Name:       blk.labels[0],
		Attributes: attrs,
		Blocks:     blocks,
		File:       blk.file,
		Line:       blk.line,
	}
	if source, ok := attrs["source"]; ok {
		m.Source = literalString(source)
		delete(attrs, "source")
	}
	if version, ok := attrs["version"]; ok {
		m.Version = literalString(version)
		delete(attrs, "version")
	}
	if len(attrs) == 0 {
		m.Attributes = nil
	}

	p.graph.Modules = append(p.graph.Modules, m)
	return nil
}

func (p *parser) addProvider(blk block) error {
	attrs, blocks, err := parseBody(blk.body, p.src[blk.file])
	if err != nil {
		return fmt.Errorf("%s:%d: %w", blk.file, blk.line, err)
	}

	provider := &ir.Provider{
		Name:   blk.labels[0],
		Blocks: blocks,
		File:   blk.file,
		Line:   blk.line,
	}
	if alias, ok := attrs["alias"]; ok {
		provider.Alias = literalString(alias)
		delete(attrs, "alias")
	}
	if len(attrs) > 0 {
		provider.Attributes = attrs
	}

	p.graph.Providers = append(p.graph.Providers, provider)
	return nil
}

// finish resolves edges and records the language features the graph needs.
func (p *parser) finish() (*ir.Graph, error) {
	p.resolveEdges()
	p.graph.Features, p.graph.MinVersion = tfversion.Analyze(p.graph)
	return p.graph, nil
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

// providerAddress renders a provider meta-argument. Native HCL gives an
// expression (aws.west); JSON gives a plain string when it is written as one.
func providerAddress(v ir.Value) string {
	if s := strings.TrimSpace(v.Expr); s != "" {
		return s
	}
	return strings.TrimSpace(literalString(v))
}

// valueAddresses flattens a depends_on argument into node addresses. Both
// depends_on = [aws_vpc.main] and JSON's ["aws_vpc.main"] are accepted.
func valueAddresses(v ir.Value) []string {
	switch v.Kind {
	case ir.KindList:
		var out []string
		for _, item := range v.Items {
			out = append(out, valueAddresses(item)...)
		}
		return out
	case ir.KindString:
		if s := strings.TrimSpace(literalString(v)); s != "" {
			return []string{s}
		}
		return nil
	case ir.KindExpr:
		if len(v.Refs) > 0 {
			return v.Refs
		}
		if s := strings.TrimSpace(v.Expr); s != "" {
			return []string{s}
		}
	}
	return nil
}
