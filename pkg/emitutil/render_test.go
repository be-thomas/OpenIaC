package emitutil_test

import (
	"strings"
	"testing"

	"github.com/be-thomas/openiac/pkg/emitutil"
	"github.com/be-thomas/openiac/pkg/ir"
)

func testRewriter() *emitutil.RefRewriter {
	return emitutil.NewRefRewriter(
		[]string{"aws_vpc.main", "data.aws_ami.ubuntu", "var.environment"},
		func(address, attribute string) (string, bool) {
			base := ""
			switch address {
			case "aws_vpc.main":
				base = "mainVpc"
			case "data.aws_ami.ubuntu":
				base = "ubuntuImage"
			case "var.environment":
				if attribute != "" {
					return "", false
				}
				return "environment", true
			default:
				return "", false
			}
			if attribute == "" {
				return base, true
			}
			return base + "." + attribute, true
		},
	)
}

func TestQuote(t *testing.T) {
	cases := map[string]string{
		"plain":     `"plain"`,
		`quote"x`:   `"quote\"x"`,
		"<a&b>":     `"<a&b>"`,
		"tab\there": `"tab\there"`,
	}
	for in, want := range cases {
		if got := emitutil.Quote(in); got != want {
			t.Errorf("Quote(%q) = %s, want %s", in, got, want)
		}
	}
}

func TestRenderValue(t *testing.T) {
	rewriter := testRewriter()

	cases := []struct {
		name  string
		value ir.Value
		want  string
	}{
		{"string", ir.Value{Kind: ir.KindString, Literal: "hello"}, `"hello"`},
		{"number", ir.Value{Kind: ir.KindNumber, Literal: int64(42)}, "42"},
		{"float", ir.Value{Kind: ir.KindNumber, Literal: 0.5}, "0.5"},
		{"bool", ir.Value{Kind: ir.KindBool, Literal: true}, "true"},
		{"null", ir.Value{Kind: ir.KindNull}, "undefined"},
		{
			"list",
			ir.Value{Kind: ir.KindList, Items: []ir.Value{
				{Kind: ir.KindNumber, Literal: int64(1)},
				{Kind: ir.KindString, Literal: "two"},
			}},
			`[1, "two"]`,
		},
		{
			"map is sorted",
			ir.Value{Kind: ir.KindMap, Fields: map[string]ir.Value{
				"b": {Kind: ir.KindNumber, Literal: int64(2)},
				"a": {Kind: ir.KindNumber, Literal: int64(1)},
			}},
			`{"a": 1, "b": 2}`,
		},
		{
			"expression",
			ir.Value{Kind: ir.KindExpr, Raw: "aws_vpc.main.id"},
			"mainVpc.id",
		},
	}
	for _, c := range cases {
		if got := emitutil.RenderValue(c.value, rewriter); got != c.want {
			t.Errorf("%s: RenderValue = %s, want %s", c.name, got, c.want)
		}
	}
}

func TestRenderExpressionTemplates(t *testing.T) {
	rewriter := testRewriter()

	cases := map[string]string{
		`"prefix-${aws_vpc.main.id}"`: "`prefix-${mainVpc.id}`",
		`"${data.aws_ami.ubuntu.id}"`: "`${ubuntuImage.id}`",
		`"$${literal}"`:               "`\\${literal}`",
		`aws_vpc.main.id`:             "mainVpc.id",
		``:                            "undefined",
		`"plain string"`:              "`plain string`",
	}
	for in, want := range cases {
		if got := emitutil.RenderExpression(in, rewriter); got != want {
			t.Errorf("RenderExpression(%q) = %s, want %s", in, got, want)
		}
	}
}

func TestRewriteBoundaries(t *testing.T) {
	rewriter := testRewriter()
	cases := map[string]string{
		"aws_vpc.main":           "mainVpc",
		"aws_vpc.main.id":        "mainVpc.id",
		"aws_vpc.mainx":          "aws_vpc.mainx",
		"xaws_vpc.main":          "xaws_vpc.main",
		"data.aws_ami.ubuntu.id": "ubuntuImage.id",
	}
	for in, want := range cases {
		if got := rewriter.Rewrite(in); got != want {
			t.Errorf("Rewrite(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRewriteSkipsComments(t *testing.T) {
	rewriter := testRewriter()
	in := "aws_vpc.main.id # aws_vpc.main.id\naws_vpc.main.id"
	want := "mainVpc.id # aws_vpc.main.id\nmainVpc.id"
	if got := rewriter.Rewrite(in); got != want {
		t.Errorf("Rewrite = %q, want %q", got, want)
	}
}

func TestRewriteInsideInterpolation(t *testing.T) {
	rewriter := testRewriter()
	in := `"name-${aws_vpc.main.id}"`
	want := `"name-${mainVpc.id}"`
	if got := rewriter.Rewrite(in); got != want {
		t.Errorf("Rewrite = %q, want %q", got, want)
	}
}

func TestTagsToCfn(t *testing.T) {
	rewriter := testRewriter()
	value := ir.Value{Kind: ir.KindMap, Fields: map[string]ir.Value{
		"Name": {Kind: ir.KindString, Literal: "web"},
		"Env":  {Kind: ir.KindExpr, Raw: "var.environment"},
	}}
	got, ok := emitutil.TagsToCfn(value, rewriter)
	if !ok {
		t.Fatal("TagsToCfn returned false")
	}
	want := `[{ Key: "Env", Value: environment }, { Key: "Name", Value: "web" }]`
	if got != want {
		t.Errorf("TagsToCfn = %s, want %s", got, want)
	}

	if _, ok := emitutil.TagsToCfn(ir.Value{Kind: ir.KindString, Literal: "x"}, rewriter); ok {
		t.Error("TagsToCfn should reject non-map values")
	}
}

func TestHasHeredoc(t *testing.T) {
	if !emitutil.HasHeredoc(`<<-EOT
hello
EOT`) {
		t.Error("HasHeredoc = false, want true")
	}
	if emitutil.HasHeredoc(`"plain"`) {
		t.Error("HasHeredoc = true, want false")
	}
}

func TestRewriteUnknownAddressLeftAlone(t *testing.T) {
	rewriter := testRewriter()
	in := "aws_s3_bucket.logs.arn"
	if got := rewriter.Rewrite(in); strings.TrimSpace(got) != in {
		t.Errorf("Rewrite = %q, want unchanged", got)
	}
}
