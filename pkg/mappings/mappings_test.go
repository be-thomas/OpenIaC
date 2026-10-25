package mappings_test

import (
	"encoding/json"
	"testing"

	"github.com/be-thomas/openiac/pkg/mappings"
)

func TestLoad(t *testing.T) {
	registry, err := mappings.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if registry.Provider != "aws" {
		t.Errorf("provider = %q, want aws", registry.Provider)
	}
	if len(registry.Resources) < 14 {
		t.Errorf("mapped resources = %d, want at least 14", len(registry.Resources))
	}
}

func TestLookupS3Bucket(t *testing.T) {
	mapping, ok := mappings.Lookup("aws_s3_bucket")
	if !ok {
		t.Fatal("aws_s3_bucket is not mapped")
	}
	if mapping.CDK == nil || mapping.Pulumi == nil {
		t.Fatal("aws_s3_bucket needs both CDK and Pulumi mappings")
	}
	if mapping.CDK.Kind != "l2" || mapping.CDK.Class != "Bucket" || mapping.CDK.Module != "aws-s3" {
		t.Errorf("CDK mapping = %#v", mapping.CDK)
	}
	if got := mapping.CDK.Properties["bucket"].Name; got != "bucketName" {
		t.Errorf("bucket property = %q, want bucketName", got)
	}
	if got := mapping.CDK.Attributes["arn"]; got != "bucketArn" {
		t.Errorf("arn attribute = %q, want bucketArn", got)
	}
	versioning, ok := mapping.CDK.Blocks["versioning"]
	if !ok || versioning.Shape != "presence" {
		t.Errorf("versioning block = %#v", versioning)
	}
}

func TestLookupSecurityGroupBlocks(t *testing.T) {
	mapping, ok := mappings.Lookup("aws_security_group")
	if !ok {
		t.Fatal("aws_security_group is not mapped")
	}
	ingress, ok := mapping.CDK.Blocks["ingress"]
	if !ok {
		t.Fatal("ingress block missing")
	}
	if ingress.Property != "SecurityGroupIngress" || ingress.Shape != "array" {
		t.Errorf("ingress block = %#v", ingress)
	}
	if got := ingress.Properties["cidr_blocks"]; got.Name != "CidrIp" || got.Transform != "first" {
		t.Errorf("cidr_blocks property = %#v", got)
	}
}

func TestLookupVPC(t *testing.T) {
	mapping, _ := mappings.Lookup("aws_vpc")
	if mapping.CDK.Kind != "l1" || mapping.CDK.Class != "CfnVPC" {
		t.Errorf("CDK mapping = %#v", mapping.CDK)
	}
	if got := mapping.CDK.Properties["tags"]; got.Transform != "cfnTags" || got.Name != "Tags" {
		t.Errorf("tags property = %#v", got)
	}
	if mapping.Pulumi.Module != "ec2" || mapping.Pulumi.Class != "Vpc" {
		t.Errorf("Pulumi mapping = %#v", mapping.Pulumi)
	}
}

func TestLookupUnknown(t *testing.T) {
	if _, ok := mappings.Lookup("aws_does_not_exist"); ok {
		t.Fatal("unknown resource should not be mapped")
	}
}

func TestPropertyMapForms(t *testing.T) {
	var short mappings.PropertyMap
	if err := json.Unmarshal([]byte(`"bucketName"`), &short); err != nil {
		t.Fatal(err)
	}
	if short.Name != "bucketName" || short.Transform != "" {
		t.Errorf("short form = %#v", short)
	}

	var full mappings.PropertyMap
	if err := json.Unmarshal([]byte(`{"name":"Tags","transform":"cfnTags"}`), &full); err != nil {
		t.Fatal(err)
	}
	if full.Name != "Tags" || full.Transform != "cfnTags" {
		t.Errorf("full form = %#v", full)
	}
}
