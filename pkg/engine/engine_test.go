package engine_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/be-thomas/openiac/pkg/emitters"
	"github.com/be-thomas/openiac/pkg/engine"
)

const source = `
variable "environment" {
  type    = string
  default = "dev"
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Environment = var.environment
  }
}

resource "aws_subnet" "public" {
  vpc_id     = aws_vpc.main.id
  cidr_block = "10.0.1.0/24"
}

output "subnet_id" {
  value = aws_subnet.public.id
}
`

func convert(t *testing.T, target string) (*engine.TargetResult, []string) {
	t.Helper()
	targets, err := engine.ParseTargets(target)
	if err != nil {
		t.Fatalf("ParseTargets(%q): %v", target, err)
	}
	graph, results, err := engine.ConvertSource(source, "main.tf", targets, emitters.Options{StackName: "TestStack"})
	if err != nil {
		t.Fatalf("ConvertSource: %v", err)
	}
	if graph == nil || len(results) != 1 {
		t.Fatalf("graph=%v results=%d", graph != nil, len(results))
	}
	if len(results[0].Files) != 1 {
		t.Fatalf("files = %d, want 1", len(results[0].Files))
	}
	return &results[0], results[0].Warnings
}

func TestConvertCDK(t *testing.T) {
	result, warnings := convert(t, "cdk")
	file := result.Files[0]

	if file.Path != "index.ts" {
		t.Errorf("path = %q, want index.ts", file.Path)
	}

	checks := []string{
		"import * as ec2 from 'aws-cdk-lib/aws-ec2';",
		"export class TestStack extends cdk.Stack",
		`const environment = this.node.tryGetContext("environment") ?? "dev";`,
		`const mainVpc = new ec2.CfnVPC(this, "aws_vpc.main", {`,
		"Tags: [{ Key: \"Environment\", Value: environment }]",
		"VpcId: mainVpc.ref",
		`new cdk.CfnOutput(this, "subnet_id", { value: publicSubnet.ref });`,
	}
	for _, want := range checks {
		if !strings.Contains(file.Content, want) {
			t.Errorf("CDK output missing %q\n---\n%s", want, file.Content)
		}
	}
	if len(warnings) != 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}
}

func TestConvertPulumi(t *testing.T) {
	result, warnings := convert(t, "pulumi")
	file := result.Files[0]

	checks := []string{
		"import * as aws from '@pulumi/aws';",
		`const environment = config.get("environment") ?? "dev";`,
		`const mainVpc = new aws.ec2.Vpc("main", {`,
		"vpcId: mainVpc.id",
		"export const subnetId = publicSubnet.id;",
	}
	for _, want := range checks {
		if !strings.Contains(file.Content, want) {
			t.Errorf("Pulumi output missing %q\n---\n%s", want, file.Content)
		}
	}
	if len(warnings) != 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}
}

func TestConvertBothTargets(t *testing.T) {
	targets, err := engine.ParseTargets("both")
	if err != nil {
		t.Fatal(err)
	}
	_, results, err := engine.ConvertSource(source, "main.tf", targets, emitters.Options{StackName: "TestStack"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("results = %d, want 2", len(results))
	}
	if results[0].Target != engine.TargetCDK || results[1].Target != engine.TargetPulumi {
		t.Fatalf("targets = %s, %s", results[0].Target, results[1].Target)
	}
}

func TestParseTargets(t *testing.T) {
	cases := map[string]engine.Target{
		"":        engine.TargetCDK,
		"cdk":     engine.TargetCDK,
		"aws-cdk": engine.TargetCDK,
		"Pulumi":  engine.TargetPulumi,
	}
	for input, want := range cases {
		targets, err := engine.ParseTargets(input)
		if err != nil {
			t.Errorf("ParseTargets(%q): %v", input, err)
			continue
		}
		if len(targets) != 1 || targets[0] != want {
			t.Errorf("ParseTargets(%q) = %v, want %s", input, targets, want)
		}
	}

	both, err := engine.ParseTargets("all")
	if err != nil {
		t.Fatal(err)
	}
	if len(both) != 2 || both[0] != engine.TargetCDK || both[1] != engine.TargetPulumi {
		t.Errorf(`ParseTargets("all") = %v, want [cdk pulumi]`, both)
	}

	if _, err := engine.ParseTargets("cloudformation"); err == nil {
		t.Error("expected an error for an unknown target")
	}
}

func TestConvertInvalidHCL(t *testing.T) {
	if _, _, err := engine.ConvertSource(`resource "aws_vpc" {`, "bad.tf", []engine.Target{engine.TargetCDK}, emitters.Options{}); err == nil {
		t.Fatal("expected a parse error")
	}
}

func TestParseFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.tf")
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	graph, err := engine.Parse(path)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(graph.Resources) != 2 {
		t.Errorf("resources = %d, want 2", len(graph.Resources))
	}
}

func TestMarshalIR(t *testing.T) {
	graph, err := engine.ParseSource(source, "main.tf")
	if err != nil {
		t.Fatal(err)
	}
	data, err := engine.MarshalIR(graph)
	if err != nil {
		t.Fatal(err)
	}

	var decoded struct {
		Version   string           `json:"version"`
		Resources []map[string]any `json:"resources"`
		Edges     []map[string]any `json:"edges"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("IR is not valid JSON: %v", err)
	}
	if decoded.Version != "0.1" {
		t.Errorf("version = %q, want 0.1", decoded.Version)
	}
	if len(decoded.Resources) != 2 || len(decoded.Edges) == 0 {
		t.Errorf("resources=%d edges=%d", len(decoded.Resources), len(decoded.Edges))
	}
	if !strings.HasSuffix(string(data), "\n") {
		t.Error("IR should end with a newline")
	}
}
