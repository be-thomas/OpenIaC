//go:build js && wasm

// Command openiac-wasm exposes the conversion pipeline to browsers.
package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/be-thomas/openiac/pkg/emitters"
	"github.com/be-thomas/openiac/pkg/engine"
)

type response struct {
	OK        bool     `json:"ok"`
	Error     string   `json:"error,omitempty"`
	Files     []file   `json:"files,omitempty"`
	Warnings  []string `json:"warnings,omitempty"`
	Resources int      `json:"resources,omitempty"`
	Data      int      `json:"data,omitempty"`
	Variables int      `json:"variables,omitempty"`
	Outputs   int      `json:"outputs,omitempty"`
	Edges     int      `json:"edges,omitempty"`
}

type file struct {
	Target  string `json:"target"`
	Path    string `json:"path"`
	Content string `json:"content"`
}

// irResponse carries the parsed graph as JSON text, so the playground can show
// the IR without importing a TypeScript emitter.
type irResponse struct {
	OK        bool   `json:"ok"`
	Error     string `json:"error,omitempty"`
	IR        string `json:"ir,omitempty"`
	Resources int    `json:"resources,omitempty"`
	Variables int    `json:"variables,omitempty"`
	Outputs   int    `json:"outputs,omitempty"`
	Edges     int    `json:"edges,omitempty"`
}

func main() {
	js.Global().Set("openiacTranspile", js.FuncOf(transpile))
	js.Global().Set("openiacIR", js.FuncOf(parseIR))
	<-make(chan struct{})
}

// transpile(source, target) returns a JSON string so JavaScript never has to
// marshal Go values.
func transpile(_ js.Value, args []js.Value) any {
	if len(args) == 0 || args[0].Type() != js.TypeString {
		return encode(response{Error: "missing HCL source"})
	}
	source := args[0].String()
	targetName := "cdk"
	if len(args) > 1 && args[1].Type() == js.TypeString {
		targetName = args[1].String()
	}

	targets, err := engine.ParseTargets(targetName)
	if err != nil {
		return encode(response{Error: err.Error()})
	}
	graph, results, err := engine.ConvertSource(source, "playground.tf", targets, emitters.Options{StackName: "OpenIacStack"})
	if err != nil {
		return encode(response{Error: err.Error()})
	}
	if graph == nil {
		return encode(response{Error: "no graph produced"})
	}

	resources, data, variables, outputs, edges := graph.Stats()
	out := response{
		OK:        true,
		Resources: resources,
		Data:      data,
		Variables: variables,
		Outputs:   outputs,
		Edges:     edges,
	}
	for _, result := range results {
		for _, generated := range result.Files {
			out.Files = append(out.Files, file{
				Target:  string(result.Target),
				Path:    generated.Path,
				Content: generated.Content,
			})
		}
		out.Warnings = append(out.Warnings, result.Warnings...)
	}
	return encode(out)
}

// parseIR(source) returns the JSON IR for source as a string field, using the
// same parser and schema version the CLI prints with `openiac ir`.
func parseIR(_ js.Value, args []js.Value) any {
	if len(args) == 0 || args[0].Type() != js.TypeString {
		return encode(irResponse{Error: "missing HCL source"})
	}
	graph, err := engine.ParseSource(args[0].String(), "playground.tf")
	if err != nil {
		return encode(irResponse{Error: err.Error()})
	}
	if graph == nil {
		return encode(irResponse{Error: "no graph produced"})
	}
	data, err := engine.MarshalIR(graph)
	if err != nil {
		return encode(irResponse{Error: err.Error()})
	}

	resources, _, variables, outputs, edges := graph.Stats()
	return encode(irResponse{
		OK:        true,
		IR:        string(data),
		Resources: resources,
		Variables: variables,
		Outputs:   outputs,
		Edges:     edges,
	})
}

func encode(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return `{"ok":false,"error":"failed to encode response"}`
	}
	return string(data)
}
