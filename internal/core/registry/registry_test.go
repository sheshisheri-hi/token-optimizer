package registry

import "testing"

func TestLoadDefaultModules(t *testing.T) {
	reg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	mods := reg.DefaultModules()
	if len(mods) < 4 {
		t.Fatalf("expected default modules, got %v", mods)
	}
}
