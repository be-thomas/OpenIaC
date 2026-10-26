package tfversion_test

import (
	"testing"

	"github.com/be-thomas/openiac/pkg/tfversion"
)

func TestCanonical(t *testing.T) {
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{"1", "1.0.0", true},
		{"1.5", "1.5.0", true},
		{"1.5.2", "1.5.2", true},
		{"v1.6.0", "1.6.0", true},
		{" v1.10.3 ", "1.10.3", true},
		{"1.2.3-beta1", "", false},
		{"1.2.3+build.5", "", false},
		{"latest", "", false},
		{"", "", false},
		{"v", "", false},
		{"1..2", "", false},
	}
	for _, c := range cases {
		got, ok := tfversion.Canonical(c.in)
		if got != c.want || ok != c.ok {
			t.Errorf("Canonical(%q) = (%q, %v), want (%q, %v)", c.in, got, ok, c.want, c.ok)
		}
	}
}

func TestCompare(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"1.9", "1.10", -1},
		{"1.10", "1.9", 1},
		{"1.10.0", "1.10", 0},
		{"v1.5.0", "1.5.0", 0},
		{"1.0.0", "1.0.1", -1},
		{"nonsense", "1.0.0", -1},
		{"1.0.0", "nonsense", 1},
		{"nonsense", "also nonsense", 0},
	}
	for _, c := range cases {
		if got := tfversion.Compare(c.a, c.b); got != c.want {
			t.Errorf("Compare(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestBaselineWithinSupportedRange(t *testing.T) {
	if tfversion.Compare(tfversion.Baseline, tfversion.Latest) >= 0 {
		t.Fatalf("baseline %s must be older than latest %s", tfversion.Baseline, tfversion.Latest)
	}
}
