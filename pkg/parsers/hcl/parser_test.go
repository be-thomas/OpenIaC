package hclparser_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/be-thomas/openiac/pkg/ir"
	hclparser "github.com/be-thomas/openiac/pkg/parsers/hcl"
)

const sample = `
terraform {
  required_version = ">= 1.0"
}

variable "region" {
  type        = string
  description = "AWS region"
  default     = "us-east-1"
}

variable "environment" {
  type    = string
  default = "dev"
}

provider "aws" {
  region = var.region
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name        = "app-${var.environment}"
    Environment = var.environment
  }
}

resource "aws_subnet" "public" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = "${var.region}a"
}

resource "aws_security_group" "web" {
  name   = "web"
  vpc_id = aws_vpc.main.id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_instance" "web" {
  ami                    = "ami-123"
  subnet_id              = aws_subnet.public.id
  vpc_security_group_ids = [aws_security_group.web.id]
  depends_on             = [aws_s3_bucket.logs]
}

resource "aws_s3_bucket" "logs" {
  bucket = "logs-${var.environment}"
}

data "aws_ami" "ubuntu" {
  most_recent = true
}

output "bucket_arn" {
  value       = aws_s3_bucket.logs.arn
  description = "Logs bucket"
}
`

func parseSample(t *testing.T) *ir.Graph {
	t.Helper()
	graph, err := hclparser.ParseString(sample, "main.tf")
	if err != nil {
		t.Fatalf("ParseString: %v", err)
	}
	return graph
}

func TestParseResources(t *testing.T) {
	graph := parseSample(t)

	if got, want := len(graph.Resources), 6; got != want {
		t.Fatalf("resources = %d, want %d", got, want)
	}
	if graph.ResourceByID("aws_vpc.main") == nil {
		t.Error("aws_vpc.main missing from graph")
	}
	data := graph.ResourceByID("data.aws_ami.ubuntu")
	if data == nil {
		t.Fatal("data.aws_ami.ubuntu missing from graph")
	}
	if data.Kind != ir.KindData {
		t.Errorf("data kind = %q, want %q", data.Kind, ir.KindData)
	}
	if data.Provider != "aws" {
		t.Errorf("provider = %q, want aws", data.Provider)
	}
	if vpc := graph.ResourceByID("aws_vpc.main"); vpc == nil || vpc.Provider != "aws" {
		t.Errorf("aws_vpc.main provider not parsed")
	}
}

func TestParseStructuredValues(t *testing.T) {
	graph := parseSample(t)
	vpc := graph.ResourceByID("aws_vpc.main")

	cidr := vpc.Attributes["cidr_block"]
	if cidr.Kind != ir.KindString || cidr.Literal != "10.0.0.0/16" {
		t.Errorf("cidr_block = %#v, want literal string", cidr)
	}

	tags := vpc.Attributes["tags"]
	if tags.Kind != ir.KindMap {
		t.Fatalf("tags kind = %q, want map", tags.Kind)
	}
	environment, ok := tags.Fields["Environment"]
	if !ok {
		t.Fatal("tags.Environment missing")
	}
	if environment.Kind != ir.KindExpr {
		t.Errorf("tags.Environment kind = %q, want expr", environment.Kind)
	}
	if !contains(environment.Refs, "var.environment") {
		t.Errorf("tags.Environment refs = %v, want var.environment", environment.Refs)
	}

	name := tags.Fields["Name"]
	if name.Kind != ir.KindExpr || name.Raw == "" {
		t.Errorf("tags.Name = %#v, want raw expression", name)
	}
}

func TestParseBlocks(t *testing.T) {
	graph := parseSample(t)
	sg := graph.ResourceByID("aws_security_group.web")
	if len(sg.Blocks) != 1 {
		t.Fatalf("security group blocks = %d, want 1", len(sg.Blocks))
	}
	ingress := sg.Blocks[0]
	if ingress.Type != "ingress" {
		t.Errorf("block type = %q, want ingress", ingress.Type)
	}
	if got := ingress.Attributes["from_port"]; got.Kind != ir.KindNumber || got.Literal != int64(80) {
		t.Errorf("from_port = %#v, want number 80", got)
	}
}

func TestParseVariablesProvidersOutputs(t *testing.T) {
	graph := parseSample(t)

	if len(graph.Variables) != 2 {
		t.Fatalf("variables = %d, want 2", len(graph.Variables))
	}
	region := graph.Variables[0]
	if region.Name != "region" {
		t.Fatalf("first variable = %q, want region", region.Name)
	}
	if region.Type != "string" {
		t.Errorf("region type = %q, want string", region.Type)
	}
	if region.Default == nil || region.Default.Kind != ir.KindString || region.Default.Literal != "us-east-1" {
		t.Errorf("region default = %#v", region.Default)
	}
	if region.Description != "AWS region" {
		t.Errorf("region description = %q", region.Description)
	}

	if len(graph.Providers) != 1 || graph.Providers[0].Name != "aws" {
		t.Errorf("providers = %#v, want one aws provider", graph.Providers)
	}

	if len(graph.Outputs) != 1 {
		t.Fatalf("outputs = %d, want 1", len(graph.Outputs))
	}
	out := graph.Outputs[0]
	if out.Name != "bucket_arn" {
		t.Errorf("output name = %q", out.Name)
	}
	if !contains(out.Value.Refs, "aws_s3_bucket.logs") {
		t.Errorf("output refs = %v, want aws_s3_bucket.logs", out.Value.Refs)
	}
}

func TestParseEdges(t *testing.T) {
	graph := parseSample(t)

	cases := []struct{ from, to string }{
		{"aws_subnet.public", "aws_vpc.main"},
		{"aws_security_group.web", "aws_vpc.main"},
		{"aws_instance.web", "aws_subnet.public"},
		{"aws_instance.web", "aws_security_group.web"},
		{"aws_instance.web", "aws_s3_bucket.logs"},
	}
	for _, c := range cases {
		if !hasEdge(graph, c.from, c.to) {
			t.Errorf("missing edge %s -> %s", c.from, c.to)
		}
	}

	instance := graph.ResourceByID("aws_instance.web")
	if _, ok := instance.Attributes["depends_on"]; ok {
		t.Error("depends_on should be moved out of attributes")
	}
	if len(instance.DependsOn) != 1 || instance.DependsOn[0] != "aws_s3_bucket.logs" {
		t.Errorf("DependsOn = %v", instance.DependsOn)
	}
}

func TestParseOrder(t *testing.T) {
	graph := parseSample(t)
	ordered, err := graph.Order()
	if err != nil {
		t.Fatalf("Order: %v", err)
	}
	position := make(map[string]int, len(ordered))
	for i, resource := range ordered {
		position[resource.ID] = i
	}
	pairs := [][2]string{
		{"aws_vpc.main", "aws_subnet.public"},
		{"aws_vpc.main", "aws_security_group.web"},
		{"aws_subnet.public", "aws_instance.web"},
		{"aws_security_group.web", "aws_instance.web"},
	}
	for _, pair := range pairs {
		if position[pair[0]] > position[pair[1]] {
			t.Errorf("%s should be emitted before %s", pair[0], pair[1])
		}
	}
}

func TestParseInvalidHCL(t *testing.T) {
	if _, err := hclparser.ParseString(`resource "aws_vpc" "main" {`, "broken.tf"); err == nil {
		t.Fatal("expected a parse error")
	}
}

func TestParseDir(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"vpc.tf":        `resource "aws_vpc" "main" { cidr_block = "10.0.0.0/16" }`,
		"subnet.tf":     `resource "aws_subnet" "public" { vpc_id = aws_vpc.main.id }`,
		"notes.txt":     "ignored",
		"variables.hcl": "not terraform",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	graph, err := hclparser.ParseDir(dir)
	if err != nil {
		t.Fatalf("ParseDir: %v", err)
	}
	if len(graph.Resources) != 2 {
		t.Fatalf("resources = %d, want 2", len(graph.Resources))
	}
	if !hasEdge(graph, "aws_subnet.public", "aws_vpc.main") {
		t.Error("cross-file edge missing")
	}
	if !strings.Contains(graph.Source, "subnet.tf") || !strings.Contains(graph.Source, "vpc.tf") {
		t.Errorf("graph source = %q, want both parsed files", graph.Source)
	}
}

func TestParseProviderAlias(t *testing.T) {
	graph, err := hclparser.ParseString(`
provider "aws" {
  alias  = "west"
  region = "us-west-2"
}

resource "aws_vpc" "main" {
  provider   = aws.west
  cidr_block = "10.0.0.0/16"
}
`, "main.tf")
	if err != nil {
		t.Fatal(err)
	}
	vpc := graph.ResourceByID("aws_vpc.main")
	if vpc.ProviderAlias != "aws.west" {
		t.Errorf("provider alias = %q, want aws.west", vpc.ProviderAlias)
	}
	if graph.Providers[0].Alias != "west" {
		t.Errorf("provider block alias = %q, want west", graph.Providers[0].Alias)
	}
}

func hasEdge(graph *ir.Graph, from, to string) bool {
	for _, edge := range graph.Edges {
		if edge.From == from && edge.To == to {
			return true
		}
	}
	return false
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
