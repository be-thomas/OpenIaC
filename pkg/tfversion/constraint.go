package tfversion

import (
	"fmt"
	"strings"
)

// Clause is one operator-and-version term of a required_version constraint.
type Clause struct {
	Op      string // "=", "!=", ">", ">=", "<", "<=", "~>"
	Version string // canonical "MAJOR.MINOR.PATCH"
}

// Constraint is a parsed required_version value.
//
// Only a conservative lower bound is modelled: "~> 1.5" is treated as
// ">= 1.5.0" and upper bounds are ignored. Raw always holds the original text,
// which is what the IR preserves.
type Constraint struct {
	Raw        string
	Clauses    []Clause
	LowerBound string // canonical; empty when the constraint implies no lower bound
}

// ParseConstraint parses a Terraform required_version constraint.
//
// It accepts comma-separated clauses with the operators =, !=, >, >=, <, <=
// and ~>, and bare versions (which Terraform reads as =). It is deliberately
// best-effort: callers treat an error as "no usable bound", never as a parse
// failure of the module.
func ParseConstraint(s string) (Constraint, error) {
	raw := s
	c := Constraint{Raw: raw}

	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return c, fmt.Errorf("empty version constraint")
	}

	for _, part := range strings.Split(trimmed, ",") {
		clause, err := parseClause(part)
		if err != nil {
			return c, err
		}
		c.Clauses = append(c.Clauses, clause)

		lower, has := clause.lowerBound()
		if has && (c.LowerBound == "" || Compare(lower, c.LowerBound) > 0) {
			c.LowerBound = lower
		}
	}
	return c, nil
}

func parseClause(part string) (Clause, error) {
	part = strings.TrimSpace(part)
	if part == "" {
		return Clause{}, fmt.Errorf("empty version constraint clause")
	}

	op := "="
	rest := part
	for _, candidate := range []string{">=", "<=", "~>", "!=", "==", ">", "<", "="} {
		if strings.HasPrefix(part, candidate) {
			if candidate == "==" {
				return Clause{}, fmt.Errorf("unsupported version operator %q", candidate)
			}
			op = candidate
			rest = strings.TrimSpace(part[len(candidate):])
			break
		}
	}
	if rest == "" {
		return Clause{}, fmt.Errorf("version constraint %q has no version", part)
	}

	version, ok := Canonical(rest)
	if !ok {
		return Clause{}, fmt.Errorf("invalid version %q", rest)
	}
	return Clause{Op: op, Version: version}, nil
}

// lowerBound reports the smallest version a clause guarantees.
func (c Clause) lowerBound() (string, bool) {
	switch c.Op {
	case "=", ">=", "~>":
		return c.Version, true
	case ">":
		v, ok := bumpPatch(c.Version)
		return v, ok
	default:
		// "!=", "<" and "<=" say nothing about the minimum.
		return "", false
	}
}
