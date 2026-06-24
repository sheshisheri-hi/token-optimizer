package capabilities

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/sheshisheri-hi/token-optimizer/internal/runtime"
)

type File struct {
	Version  int                       `json:"version"`
	ProbedAt string                    `json:"probedAt"`
	Features map[string]FeatureProbe     `json:"features"`
}

type FeatureProbe struct {
	Enabled   bool   `json:"enabled"`
	LastProbe string `json:"lastProbe"`
	ProbedAt  string `json:"probedAt"`
	Note      string `json:"note,omitempty"`
}

func Path(dataHome string) string {
	return filepath.Join(dataHome, "capabilities.json")
}

func Load(dataHome string) (*File, error) {
	p := Path(dataHome)
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &File{Version: 1, Features: map[string]FeatureProbe{}}, nil
		}
		return nil, err
	}
	var f File
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	if f.Features == nil {
		f.Features = map[string]FeatureProbe{}
	}
	return &f, nil
}

func Save(dataHome string, f *File) error {
	if err := os.MkdirAll(dataHome, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(Path(dataHome), append(data, '\n'), 0o644)
}

func Seed(dataHome string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	f := &File{
		Version:  1,
		ProbedAt: now,
		Features: map[string]FeatureProbe{
			"bash_compress":      {Enabled: true, LastProbe: string(runtime.StatusOK), ProbedAt: now},
			"search_compress":    {Enabled: true, LastProbe: string(runtime.StatusOK), ProbedAt: now},
			"session_continuity": {Enabled: true, LastProbe: string(runtime.StatusOK), ProbedAt: now},
			"session_logging":    {Enabled: true, LastProbe: string(runtime.StatusOK), ProbedAt: now},
			"lean_output_nudges": {Enabled: true, LastProbe: string(runtime.StatusDegraded), ProbedAt: now, Note: "log-only until upstream supports injection"},
			"loop_detection":     {Enabled: false, LastProbe: string(runtime.StatusOff), ProbedAt: now, Note: "Phase 2"},
		},
	}
	return Save(dataHome, f)
}

func ProbeStatus(dataHome, feature string) runtime.FeatureStatus {
	f, err := Load(dataHome)
	if err != nil {
		return runtime.StatusOff
	}
	p, ok := f.Features[feature]
	if !ok {
		return runtime.StatusOK
	}
	switch p.LastProbe {
	case string(runtime.StatusOK):
		return runtime.StatusOK
	case string(runtime.StatusDegraded):
		return runtime.StatusDegraded
	case string(runtime.StatusBlocked):
		return runtime.StatusBlocked
	default:
		if p.Enabled {
			return runtime.StatusOK
		}
		return runtime.StatusOff
	}
}
