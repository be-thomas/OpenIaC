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

func main() {
	js.Global().Set("openiacTranspile", js.FuncOf(transpile))
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

func encode(value response) string {
	data, err := json.Marshal(value)
	if err != nil {
		return `{"ok":false,"error":"failed to encode response"}`
	}
	return string(data)
}
