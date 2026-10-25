//go:build !(js && wasm)

// Package main is the WebAssembly entry point. This stub exists so native
// `go build ./...` and `go vet ./...` do not fail on a directory whose real
// build constraints exclude every file.
package main

func main() {}
