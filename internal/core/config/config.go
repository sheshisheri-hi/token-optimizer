package config

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type File struct {
	Features map[string]bool `yaml:"features"`
	Modules  map[string]bool `yaml:"modules"`
}

func Default() *File {
	return &File{
		Features: map[string]bool{
			"bash_compress":       true,
			"search_compress":     true,
			"session_continuity":  true,
			"session_logging":     true,
			"lean_output_nudges":  true,
			"loop_detection":      false,
			"output-control":      true,
			"token-economy":       true,
			"mcp-hygiene":         true,
			"workflow-habits":     true,
			"mcp_hygiene":         true,
			"copilot_defaults":    true,
			"debug_logs":          true,
			"rtk_proxy":           false,
		},
		Modules: map[string]bool{
			"hooks":       true,
			"skills":      true,
			"settings":    true,
			"vscode":      true,
			"measurement": true,
			"hooks-rtk":   false,
			"hooks_rtk":   false,
		},
	}
}

func Path(dataHome string) string {
	return filepath.Join(dataHome, "config.yaml")
}

func Load(dataHome string) (*File, error) {
	path := Path(dataHome)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			f := Default()
			f.ApplyEnvOverrides()
			return f, nil
		}
		return nil, err
	}
	var f File
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	if f.Features == nil {
		f.Features = map[string]bool{}
	}
	if f.Modules == nil {
		f.Modules = map[string]bool{}
	}
	f.ApplyEnvOverrides()
	return &f, nil
}

func Save(dataHome string, f *File) error {
	if err := os.MkdirAll(dataHome, 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(f)
	if err != nil {
		return err
	}
	return os.WriteFile(Path(dataHome), data, 0o644)
}

func (f *File) ApplyEnvOverrides() {
	envMap := map[string]string{
		"ORG_SLIM_BASH_COMPRESS":      "bash_compress",
		"ORG_SLIM_SEARCH_COMPRESS":    "search_compress",
		"ORG_SLIM_SESSION_CONTINUITY": "session_continuity",
		"ORG_SLIM_SESSION_LOGGING":    "session_logging",
		"ORG_SLIM_LEAN_OUTPUT_NUDGES": "lean_output_nudges",
		"ORG_SLIM_LOOP_DETECTION":     "loop_detection",
	}
	for env, feat := range envMap {
		if v := os.Getenv(env); v != "" {
			f.SetFeature(feat, envBool(v))
		}
	}
	if v := os.Getenv("ORG_SLIM_MEASUREMENT"); v != "" {
		if f.Modules == nil {
			f.Modules = map[string]bool{}
		}
		f.Modules["measurement"] = envBool(v)
	}
}

func envBool(s string) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	return s == "1" || s == "true" || s == "on" || s == "yes"
}

func (f *File) FeatureEnabled(name string) bool {
	if v, ok := f.Features[name]; ok {
		return v
	}
	if v, ok := Default().Features[name]; ok {
		return v
	}
	return true
}

func (f *File) SetFeature(name string, on bool) {
	if f.Features == nil {
		f.Features = map[string]bool{}
	}
	f.Features[name] = on
}
