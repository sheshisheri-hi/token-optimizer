package compress

import "testing"

func TestSearch_rg(t *testing.T) {
	r := Search("rg foo bar")
	if !r.Changed {
		t.Fatalf("expected cap, got %+v", r)
	}
}
