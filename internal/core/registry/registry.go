package registry

import (
	"fmt"
	"strings"

	"github.com/sheshisheri-hi/token-optimizer/templates"
	"gopkg.in/yaml.v3"
)

type FeatureDef struct {
	Label         string   `yaml:"label"`
	Phase         int      `yaml:"phase"`
	Targets       []string `yaml:"targets"`
	Implemented   []string `yaml:"implemented"`
}

type ModuleDef struct {
	Label    string                `yaml:"label"`
	Features map[string]FeatureDef `yaml:"features"`
}

type Registry struct {
	Version int                    `yaml:"version"`
	Modules map[string]ModuleDef   `yaml:"modules"`
}

func Load() (*Registry, error) {
	data, err := templates.FS.ReadFile("feature-registry.yaml")
	if err != nil {
		return nil, fmt.Errorf("read feature-registry: %w", err)
	}
	var reg Registry
	if err := yaml.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("parse feature-registry: %w", err)
	}
	return &reg, nil
}

func MustLoad() *Registry {
	reg, err := Load()
	if err != nil {
		panic(err)
	}
	return reg
}

func (r *Registry) ModuleNames() []string {
	names := make([]string, 0, len(r.Modules))
	for name := range r.Modules {
		names = append(names, name)
	}
	return names
}

func (r *Registry) DefaultModules() []string {
	return []string{"hooks", "skills", "settings", "vscode", "measurement"}
}

func (r *Registry) FeatureCount(target string) (total, implemented int) {
	for _, mod := range r.Modules {
		for _, feat := range mod.Features {
			if !hasTarget(feat.Targets, target) {
				continue
			}
			total++
			if hasTarget(feat.Implemented, target) {
				implemented++
			}
		}
	}
	return total, implemented
}

func hasTarget(list []string, target string) bool {
	for _, t := range list {
		if strings.EqualFold(t, target) {
			return true
		}
	}
	return false
}
