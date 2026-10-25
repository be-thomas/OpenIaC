package emitutil_test

import (
	"testing"

	"github.com/be-thomas/openiac/pkg/emitutil"
)

func TestWordsAndCase(t *testing.T) {
	cases := []struct{ in, pascal, camel string }{
		{"s3_bucket", "S3Bucket", "s3Bucket"},
		{"cloudwatch-log-group", "CloudwatchLogGroup", "cloudwatchLogGroup"},
		{"main", "Main", "main"},
	}
	for _, c := range cases {
		if got := emitutil.PascalCase(c.in); got != c.pascal {
			t.Errorf("PascalCase(%q) = %q, want %q", c.in, got, c.pascal)
		}
		if got := emitutil.CamelCase(c.in); got != c.camel {
			t.Errorf("CamelCase(%q) = %q, want %q", c.in, got, c.camel)
		}
	}
}

func TestSanitize(t *testing.T) {
	cases := map[string]string{
		"class":  "classVar",
		"my-var": "myVar",
		"a b c":  "aBC",
		"$weird": "$weird",
		"9lives": "lives",
	}
	for in, want := range cases {
		if got := emitutil.Sanitize(in); got != want {
			t.Errorf("Sanitize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestResourceIdent(t *testing.T) {
	cases := map[string]string{
		"logs": "logsS3Bucket",
		"main": "mainVpc",
	}
	if got := emitutil.ResourceIdent("aws", "aws_s3_bucket", "logs"); got != cases["logs"] {
		t.Errorf("ResourceIdent = %q, want %q", got, cases["logs"])
	}
	if got := emitutil.ResourceIdent("aws", "aws_vpc", "main"); got != cases["main"] {
		t.Errorf("ResourceIdent = %q, want %q", got, cases["main"])
	}
	if got := emitutil.ResourceIdent("aws", "aws_eip", "nat"); got != "natEip" {
		t.Errorf("ResourceIdent = %q, want natEip", got)
	}
}

func TestNamer(t *testing.T) {
	namer := emitutil.NewNamer()
	if got := namer.Take("logs"); got != "logs" {
		t.Fatalf("Take = %q", got)
	}
	if got := namer.Take("logs"); got != "logs2" {
		t.Fatalf("second Take = %q, want logs2", got)
	}
	if got := namer.Take("logs"); got != "logs3" {
		t.Fatalf("third Take = %q, want logs3", got)
	}

	namer.Reserve("config")
	if got := namer.Take("config"); got != "config2" {
		t.Fatalf("reserved Take = %q, want config2", got)
	}

	if got := emitutil.NewNamer().Take(""); got != "node" {
		t.Fatalf("empty Take = %q, want node", got)
	}
}
