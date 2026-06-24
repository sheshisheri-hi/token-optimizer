package compress

import (
	"strings"
	"testing"
)

func TestBash_gitLog(t *testing.T) {
	r := Bash("git log")
	if !r.Changed || !strings.Contains(r.Command, "--oneline") {
		t.Fatalf("expected rewrite, got %+v", r)
	}
}

func TestBash_rejectsSemicolon(t *testing.T) {
	r := Bash("git log; rm -rf /")
	if r.Changed {
		t.Fatalf("expected no rewrite, got %+v", r)
	}
}

func TestBash_pytest(t *testing.T) {
	r := Bash("pytest tests/")
	if !r.Changed || !strings.Contains(r.Command, "-q") {
		t.Fatalf("expected quiet pytest, got %+v", r)
	}
}
