package tally

import (
	"fmt"
	"strings"

	"github.com/sheshisheri-hi/token-optimizer/internal/core/config"
	"github.com/sheshisheri-hi/token-optimizer/internal/core/manifest"
	"github.com/sheshisheri-hi/token-optimizer/internal/core/registry"
)

type Summary struct {
	ModulesInstalled int
	ModulesTotal     int
	FeaturesActive   int
	FeaturesTotal    int
	FeaturesOff      int
	FeaturesNotShipped int
}

func Compute(reg *registry.Registry, man *manifest.File, cfg *config.File, target string) Summary {
	moduleNames := reg.ModuleNames()
	installed := map[string]bool{}
	for _, n := range man.InstalledModuleNames() {
		installed[n] = true
	}

	s := Summary{ModulesTotal: len(moduleNames)}
	for _, name := range moduleNames {
		if installed[name] {
			s.ModulesInstalled++
		}
	}

	for modName, mod := range reg.Modules {
		modOn := installed[modName]
		if cfg != nil {
			if v, ok := cfg.Modules[modName]; ok && !v {
				modOn = false
			}
		}
		for featName, feat := range mod.Features {
			if !hasTarget(feat.Targets, target) {
				continue
			}
			s.FeaturesTotal++
			shipped := hasTarget(feat.Implemented, target)
			if !shipped {
				s.FeaturesNotShipped++
				continue
			}
			on := modOn
			if cfg != nil && cfg.FeatureEnabled(featName) == false {
				on = false
			}
			if on {
				s.FeaturesActive++
			} else {
				s.FeaturesOff++
			}
		}
	}
	return s
}

func (s Summary) Footer() string {
	return fmt.Sprintf(
		"Modules: %d/%d installed | Features: %d/%d active (%d OFF, %d not shipped)",
		s.ModulesInstalled, s.ModulesTotal,
		s.FeaturesActive, s.FeaturesTotal,
		s.FeaturesOff, s.FeaturesNotShipped,
	)
}

func hasTarget(list []string, target string) bool {
	for _, t := range list {
		if strings.EqualFold(t, target) {
			return true
		}
	}
	return false
}
