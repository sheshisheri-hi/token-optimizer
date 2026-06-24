package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnablePreservesMeta(t *testing.T) {
	dir := t.TempDir()
	f := &File{Modules: map[string]ModuleState{
		"settings": {Enabled: false, Meta: map[string]string{"backup.x": "y"}},
	}}
	f.EnableModules([]string{"settings"})
	if f.Modules["settings"].Meta["backup.x"] != "y" {
		t.Fatal("meta wiped on enable")
	}
	if !f.Modules["settings"].Enabled {
		t.Fatal("not enabled")
	}
	if err := Save(dir, f); err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(Path(dir)) {
		t.Fatal("path")
	}
	_ = os.MkdirAll(dir, 0o755)
}
