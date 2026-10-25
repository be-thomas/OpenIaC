// Package emitters defines the interface every output target implements.
package emitters

import (
	"fmt"
	"sort"

	"github.com/be-thomas/openiac/pkg/ir"
)

// File is one generated source file.
type File struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// Result is the output of one emitter run.
type Result struct {
	Files    []File   `json:"files"`
	Warnings []string `json:"warnings"`
}

// Options are emitter-wide settings.
type Options struct {
	StackName   string
	ProjectName string
}

// Emitter renders an IR graph as target-language source.
type Emitter interface {
	Name() string
	Emit(graph *ir.Graph, options Options) (*Result, error)
}

// Warnings collects non-fatal conversion problems.
type Warnings struct {
	items []string
	seen  map[string]bool
}

func NewWarnings() *Warnings {
	return &Warnings{seen: make(map[string]bool)}
}

// Add records a warning once.
func (w *Warnings) Add(message string) {
	if w.seen[message] {
		return
	}
	w.seen[message] = true
	w.items = append(w.items, message)
}

func (w *Warnings) Addf(format string, args ...any) {
	w.Add(fmt.Sprintf(format, args...))
}

func (w *Warnings) List() []string {
	return append([]string(nil), w.items...)
}

// GroupBlocks groups nested blocks by type while preserving source order.
func GroupBlocks(blocks []ir.Block) ([]string, map[string][]ir.Block) {
	byType := make(map[string][]ir.Block)
	var order []string
	for _, block := range blocks {
		if _, ok := byType[block.Type]; !ok {
			order = append(order, block.Type)
		}
		byType[block.Type] = append(byType[block.Type], block)
	}
	sort.Strings(order)
	return order, byType
}
