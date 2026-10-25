// Command openiac converts Terraform HCL into native AWS CDK or Pulumi
// TypeScript.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/be-thomas/openiac/pkg/emitters"
	"github.com/be-thomas/openiac/pkg/engine"
)

const version = "0.1.0-dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		usage(stderr)
		return 2
	}
	switch args[0] {
	case "convert":
		return runConvert(args[1:], stdout, stderr)
	case "ir":
		return runIR(args[1:], stdout, stderr)
	case "version", "--version", "-v":
		fmt.Fprintf(stdout, "openiac %s\n", version)
		return 0
	case "help", "--help", "-h":
		usage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "openiac: unknown command %q\n\n", args[0])
		usage(stderr)
		return 2
	}
}

func runConvert(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("convert", flag.ContinueOnError)
	flags.SetOutput(stderr)
	target := flags.String("target", "cdk", "output target: cdk, pulumi or both")
	outDir := flags.String("out", "out", "directory for generated files")
	stackName := flags.String("stack", "OpenIacStack", "CDK stack class name")
	toStdout := flags.Bool("stdout", false, "print generated files instead of writing them")
	jsonOut := flags.Bool("json", false, "print a JSON result")

	input, rest, err := splitInput(args, map[string]bool{
		"-target": true, "--target": true,
		"-out": true, "--out": true,
		"-stack": true, "--stack": true,
	})
	if err != nil {
		fmt.Fprintf(stderr, "openiac: %v\n", err)
		return 2
	}
	if err := flags.Parse(rest); err != nil {
		return 2
	}
	if input == "" {
		fmt.Fprintln(stderr, "usage: openiac convert <file|dir> [--target cdk|pulumi|both] [--out DIR]")
		return 2
	}

	targets, err := engine.ParseTargets(*target)
	if err != nil {
		fmt.Fprintf(stderr, "openiac: %v\n", err)
		return 2
	}
	graph, err := engine.Parse(input)
	if err != nil {
		fmt.Fprintf(stderr, "openiac: %v\n", err)
		return 1
	}

	options := emitters.Options{StackName: *stackName, ProjectName: "openiac"}
	results, err := engine.Convert(graph, targets, options)
	if err != nil {
		fmt.Fprintf(stderr, "openiac: %v\n", err)
		return 1
	}

	if *jsonOut {
		payload := struct {
			Graph   any                   `json:"graph"`
			Results []engine.TargetResult `json:"results"`
		}{Graph: graph, Results: results}
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(payload); err != nil {
			fmt.Fprintf(stderr, "openiac: %v\n", err)
			return 1
		}
		return 0
	}

	resources, data, variables, outputs, edges := graph.Stats()
	fmt.Fprintf(stdout, "Parsed %d resources (%d data), %d variables, %d outputs, %d edges\n",
		resources, data, variables, outputs, edges)

	for _, result := range results {
		for _, file := range result.Files {
			if *toStdout {
				fmt.Fprintf(stdout, "\n// ==== %s/%s ====\n", result.Target, file.Path)
				fmt.Fprint(stdout, file.Content)
				continue
			}
			path := filepath.Join(*outDir, string(result.Target), file.Path)
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				fmt.Fprintf(stderr, "openiac: %v\n", err)
				return 1
			}
			if err := os.WriteFile(path, []byte(file.Content), 0o644); err != nil {
				fmt.Fprintf(stderr, "openiac: %v\n", err)
				return 1
			}
			fmt.Fprintf(stdout, "%s: %s\n", result.Target, path)
		}
		for _, warning := range result.Warnings {
			fmt.Fprintf(stderr, "warning: %s\n", warning)
		}
	}
	return 0
}

func runIR(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("ir", flag.ContinueOnError)
	flags.SetOutput(stderr)
	output := flags.String("o", "", "write the IR to this file instead of stdout")

	input, rest, err := splitInput(args, map[string]bool{"-o": true, "--o": true})
	if err != nil {
		fmt.Fprintf(stderr, "openiac: %v\n", err)
		return 2
	}
	if err := flags.Parse(rest); err != nil {
		return 2
	}
	if input == "" {
		fmt.Fprintln(stderr, "usage: openiac ir <file|dir> [-o FILE]")
		return 2
	}

	graph, err := engine.Parse(input)
	if err != nil {
		fmt.Fprintf(stderr, "openiac: %v\n", err)
		return 1
	}
	data, err := engine.MarshalIR(graph)
	if err != nil {
		fmt.Fprintf(stderr, "openiac: %v\n", err)
		return 1
	}
	if *output == "" {
		stdout.Write(data)
		return 0
	}
	if err := os.WriteFile(*output, data, 0o644); err != nil {
		fmt.Fprintf(stderr, "openiac: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "wrote %s\n", *output)
	return 0
}

// splitInput pulls the single positional argument out of args so flags may
// appear before or after it. Windows-style paths are not a concern here.
func splitInput(args []string, valueFlags map[string]bool) (string, []string, error) {
	var input string
	rest := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			if input != "" {
				return "", nil, fmt.Errorf("unexpected argument %q", arg)
			}
			input = arg
			continue
		}
		rest = append(rest, arg)
		if !strings.Contains(arg, "=") && valueFlags[arg] {
			if i+1 >= len(args) {
				return "", nil, fmt.Errorf("flag %s needs a value", arg)
			}
			i++
			rest = append(rest, args[i])
		}
	}
	return input, rest, nil
}

func usage(w io.Writer) {
	fmt.Fprint(w, `OpenIaC converts Terraform HCL into native AWS CDK or Pulumi TypeScript.

Usage:
  openiac convert <file|dir> [flags]   Convert a Terraform module
  openiac ir <file|dir> [-o FILE]      Dump the universal IR
  openiac version                      Print the version

Convert flags:
  --target cdk|pulumi|both   Output target (default cdk)
  --out DIR                  Output directory (default out)
  --stack NAME               CDK stack class name (default OpenIacStack)
  --stdout                   Print generated files instead of writing them
  --json                     Print a JSON result

Examples:
  openiac convert ./examples/aws-web-app --target both
  openiac convert ./main.tf --target pulumi --stdout
  openiac ir ./examples/aws-web-app | jq '.resources[].id'
`)
}
