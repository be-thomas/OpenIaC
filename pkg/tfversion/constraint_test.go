package tfversion_test

import (
	"testing"

	"github.com/be-thomas/openiac/pkg/tfversion"
)

func TestParseConstraintLowerBound(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"1.2.3", "1.2.3"},
		{"= 1.2.3", "1.2.3"},
		{">= 1.5.0", "1.5.0"},
		{">= 1.5", "1.5.0"},
		{"~> 1.3", "1.3.0"},
		{"~> 1.3.7", "1.3.7"},
		{"> 1.5.0", "1.5.1"},
		{"> 1.5", "1.5.1"},
		{"< 2.0", ""},
		{"<= 2.0", ""},
		{"!= 1.4.0", ""},
		{"!= 1.4.0, >= 1.6.2", "1.6.2"},
		{"< 2.0, >= 1.6.2", "1.6.2"},
		{">= 1.0, < 2.0, ~> 1.2.7", "1.2.7"},
		{"  >=1.6.0  ,  <2.0  ", "1.6.0"},
	}
	for _, c := range cases {
		got, err := tfversion.ParseConstraint(c.in)
		if err != nil {
			t.Errorf("ParseConstraint(%q): %v", c.in, err)
			continue
		}
		if got.LowerBound != c.want {
			t.Errorf("ParseConstraint(%q).LowerBound = %q, want %q", c.in, got.LowerBound, c.want)
		}
		if got.Raw != c.in {
			t.Errorf("ParseConstraint(%q).Raw = %q", c.in, got.Raw)
		}
	}
}

func TestParseConstraintClauses(t *testing.T) {
	c, err := tfversion.ParseConstraint(">= 1.5.0, != 1.5.3, < 2.0")
	if err != nil {
		t.Fatalf("ParseConstraint: %v", err)
	}
	want := []tfversion.Clause{
		{Op: ">=", Version: "1.5.0"},
		{Op: "!=", Version: "1.5.3"},
		{Op: "<", Version: "2.0.0"},
	}
	if len(c.Clauses) != len(want) {
		t.Fatalf("clauses = %#v, want %#v", c.Clauses, want)
	}
	for i := range want {
		if c.Clauses[i] != want[i] {
			t.Errorf("clause %d = %#v, want %#v", i, c.Clauses[i], want[i])
		}
	}
}

func TestParseConstraintErrors(t *testing.T) {
	for _, in := range []string{
		"",
		"   ",
		"== 1.0.0",
		">=",
		"garbage",
		">= 1.2.3-beta",
		">= 1.2.3, ,",
	} {
		if _, err := tfversion.ParseConstraint(in); err == nil {
			t.Errorf("ParseConstraint(%q) succeeded, want an error", in)
		}
	}
}
