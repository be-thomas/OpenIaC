// Package engine wires parsers and emitters into a conversion pipeline.
package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/be-thomas/openiac/pkg/emitters"
	"github.com/be-thomas/openiac/pkg/emitters/awscdk"
	"github.com/be-thomas/openiac/pkg/emitters/pulumi"
	"github.com/be-thomas/openiac/pkg/ir"
	hclparser "github.com/be-thomas/openiac/pkg/parsers/hcl"
)

// Target names a supported output platform.
type Target string

const (
	TargetCDK    Target = "cdk"
	TargetPulumi Target = "pulumi"
	TargetBoth   Target = "both"
)

// TargetResult pairs a target with its generated files and warnings.
type TargetResult struct {
	Target   Target          `json:"target"`
	Files    []emitters.File `json:"files"`
	Warnings []string        `json:"warnings"`
}

// Parse reads a .tf file or a directory of .tf files.
func Parse(input string) (*ir.Graph, error) {
	info, err := os.Stat(input)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", input, err)
	}
	if info.IsDir() {
		return hclparser.ParseDir(input)
	}
	return hclparser.ParseFiles([]string{input})
}

// ParseSource reads HCL held in memory. The playground uses it through WASM.
func ParseSource(source, filename string) (*ir.Graph, error) {
	return hclparser.ParseString(source, filename)
}

// ParseTargets expands a target flag.
func ParseTargets(value string) ([]Target, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "cdk", "aws-cdk", "awscdk":
		return []Target{TargetCDK}, nil
	case "pulumi":
		return []Target{TargetPulumi}, nil
	case "both", "all":
		return []Target{TargetCDK, TargetPulumi}, nil
	default:
		return nil, fmt.Errorf("unknown target %q (want cdk, pulumi or both)", value)
	}
}

// NewEmitter returns the emitter for a target.
func NewEmitter(target Target) (emitters.Emitter, error) {
	switch target {
	case TargetCDK:
		return awscdk.New(), nil
	case TargetPulumi:
		return pulumi.New(), nil
	default:
		return nil, fmt.Errorf("unknown target %q", target)
	}
}

// Convert runs one or more emitters over a graph.
func Convert(graph *ir.Graph, targets []Target, options emitters.Options) ([]TargetResult, error) {
	results := make([]TargetResult, 0, len(targets))
	for _, target := range targets {
		emitter, err := NewEmitter(target)
		if err != nil {
			return nil, err
		}
		result, err := emitter.Emit(graph, options)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", target, err)
		}
		results = append(results, TargetResult{
			Target:   target,
			Files:    result.Files,
			Warnings: result.Warnings,
		})
	}
	return results, nil
}

// ConvertSource parses HCL from a string and runs the requested targets.
func ConvertSource(source, filename string, targets []Target, options emitters.Options) (*ir.Graph, []TargetResult, error) {
	graph, err := ParseSource(source, filename)
	if err != nil {
		return nil, nil, err
	}
	results, err := Convert(graph, targets, options)
	if err != nil {
		return nil, nil, err
	}
	return graph, results, nil
}

// MarshalIR renders the intermediate representation as stable JSON.
func MarshalIR(graph *ir.Graph) ([]byte, error) {
	data, err := json.MarshalIndent(graph, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
