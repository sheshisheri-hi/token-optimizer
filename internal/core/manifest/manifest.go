package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const Version = 1

type ModuleState struct {
	Enabled bool              `json:"enabled"`
	Meta    map[string]string `json:"meta,omitempty"`
}

type File struct {
	Version     int                      `json:"version"`
	Target      string                   `json:"target"`
	InstalledAt string                   `json:"installedAt"`
	SlimVersion string                   `json:"slimVersion"`
	Modules     map[string]ModuleState   `json:"modules"`
}

func Path(dataHome string) string {
	return filepath.Join(dataHome, "manifest.json")
}

func Load(dataHome string) (*File, error) {
	path := Path(dataHome)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &File{
				Version: Version,
				Modules: map[string]ModuleState{},
			}, nil
		}
		return nil, err
	}
	var f File
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	if f.Modules == nil {
		f.Modules = map[string]ModuleState{}
	}
	return &f, nil
}

func Save(dataHome string, f *File) error {
	if err := os.MkdirAll(dataHome, 0o755); err != nil {
		return err
	}
	f.Version = Version
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(Path(dataHome), append(data, '\n'), 0o644)
}

func (f *File) EnableModules(names []string) {
	for _, name := range names {
		st := f.Modules[name]
		st.Enabled = true
		f.Modules[name] = st
	}
}

func (f *File) DisableModules(names []string) {
	for _, name := range names {
		st := f.Modules[name]
		st.Enabled = false
		f.Modules[name] = st
	}
}

func (f *File) InstalledModuleNames() []string {
	var out []string
	for name, st := range f.Modules {
		if st.Enabled {
			out = append(out, name)
		}
	}
	return out
}

func (f *File) Touch(target, slimVersion string) {
	f.Target = target
	f.InstalledAt = time.Now().UTC().Format(time.RFC3339)
	f.SlimVersion = slimVersion
}
