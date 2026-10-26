package main

import (
	"bytes"
	"testing"

	"github.com/be-thomas/openiac/internal/buildinfo"
)

func TestVersionCommand(t *testing.T) {
	for _, arg := range []string{"version", "--version", "-v"} {
		var stdout, stderr bytes.Buffer
		if code := run([]string{arg}, &stdout, &stderr); code != 0 {
			t.Fatalf("%s: exit code %d, want 0 (stderr %q)", arg, code, stderr.String())
		}
		want := "openiac " + buildinfo.Version + "\n"
		if stdout.String() != want {
			t.Errorf("%s: stdout = %q, want %q", arg, stdout.String(), want)
		}
		if stderr.Len() != 0 {
			t.Errorf("%s: stderr = %q, want empty", arg, stderr.String())
		}
	}
}
