package runtimeutil

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sheshisheri-hi/token-optimizer/internal/core/capabilities"
	"github.com/sheshisheri-hi/token-optimizer/internal/core/config"
	"github.com/sheshisheri-hi/token-optimizer/internal/core/jsonmerge"
	"github.com/sheshisheri-hi/token-optimizer/internal/core/manifest"
	"github.com/sheshisheri-hi/token-optimizer/internal/core/measurement"
	"github.com/sheshisheri-hi/token-optimizer/internal/core/registry"
	"github.com/sheshisheri-hi/token-optimizer/internal/core/tally"
	"github.com/sheshisheri-hi/token-optimizer/internal/runtime"
	"github.com/sheshisheri-hi/token-optimizer/internal/runtime/shared"
	"github.com/sheshisheri-hi/token-optimizer/templates"
)

func MergeTemplateJSON(templatePath, destPath string, dryRun bool) ([]string, map[string]string, error) {
	data, err := templates.FS.ReadFile(templatePath)
	if err != nil {
		return nil, nil, err
	}
	var updates map[string]any
	if err := json.Unmarshal(data, &updates); err != nil {
		return nil, nil, err
	}
	if dryRun {
		return []string{fmt.Sprintf("would merge %s → %s", templatePath, destPath)}, nil, nil
	}
	actions, backup, err := jsonmerge.MergeFile(destPath, updates)
	return actions, backup, err
}

func InstallSkills(prefix, skillsDir string, dryRun bool) ([]string, error) {
	var actions []string
	err := fs.WalkDir(templates.FS, prefix, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if filepath.Base(path) != "SKILL.md" {
			return nil
		}
		rel, _ := filepath.Rel(prefix, filepath.Dir(path))
		dest := filepath.Join(skillsDir, rel, "SKILL.md")
		if dryRun {
			actions = append(actions, "would write "+dest)
			return nil
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		src, err := templates.FS.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.WriteFile(dest, src, 0o644); err != nil {
			return err
		}
		actions = append(actions, "write "+dest)
		return nil
	})
	return actions, err
}

func SetupMeasurement(dataHome string, dryRun bool) ([]string, error) {
	if dryRun {
		return []string{"would init " + measurement.DBPath(dataHome)}, nil
	}
	return []string{"init " + measurement.DBPath(dataHome)}, measurement.Init(dataHome)
}

func SetupRTK(target string, dryRun bool) ([]string, error) {
	rtk, err := execLookPath("rtk")
	if err != nil {
		return []string{"hooks-rtk: rtk not in PATH — install from https://github.com/rtk-ai/rtk"}, nil
	}
	flag := "--copilot"
	if target == "cursor" {
		flag = "--agent cursor"
	}
	cmd := fmt.Sprintf("%s init -g %s", rtk, flag)
	if dryRun {
		return []string{"would run: " + cmd}, nil
	}
	if out, err := execRun(cmd); err != nil {
		return []string{fmt.Sprintf("hooks-rtk: %v (%s)", err, strings.TrimSpace(out))}, nil
	}
	return []string{"hooks-rtk: " + cmd}, nil
}

func BuildReport(targetID, dataHome string, skillPaths map[string]string) (*runtime.StatusReport, error) {
	reg := registry.MustLoad()
	man, _ := manifest.Load(dataHome)
	cfg, _ := config.Load(dataHome)
	sum := tally.Compute(reg, man, cfg, targetID)

	rep := &runtime.StatusReport{
		Target:   targetID,
		DataHome: dataHome,
		Tally:    sum.Footer(),
	}

	for _, name := range sortedModuleKeys(reg.Modules) {
		st := runtime.StatusOff
		detail := "not installed"
		if shared.Contains(man.InstalledModuleNames(), name) {
			st = runtime.StatusOK
			detail = "enabled"
		}
		rep.Modules = append(rep.Modules, runtime.FeatureReport{Name: name, Status: st, Detail: detail})
	}

	for modName, mod := range reg.Modules {
		for featName, feat := range mod.Features {
			if !hasTarget(feat.Implemented, targetID) {
				continue
			}
			st := runtime.StatusOff
			detail := "module off"
			if shared.Contains(man.InstalledModuleNames(), modName) {
				st = runtime.StatusOK
				detail = "active"
				if modName == "skills" {
					if p, ok := skillPaths[featName]; ok && !shared.FileExists(p) {
						st = runtime.StatusDegraded
						detail = "skill file missing"
					}
				} else if cfg != nil && !cfg.FeatureEnabled(featName) {
					st = runtime.StatusOff
					detail = "disabled in config"
				} else if probe := capabilities.ProbeStatus(dataHome, featName); probe == runtime.StatusDegraded {
					st = runtime.StatusDegraded
					detail = "degraded capability"
				}
			}
			rep.Features = append(rep.Features, runtime.FeatureReport{
				Name:   featName,
				Status: st,
				Detail: detail,
			})
		}
	}
	sort.Slice(rep.Features, func(i, j int) bool { return rep.Features[i].Name < rep.Features[j].Name })
	return rep, nil
}

func sortedModuleKeys(m map[string]registry.ModuleDef) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func hasTarget(list []string, target string) bool {
	for _, t := range list {
		if strings.EqualFold(t, target) {
			return true
		}
	}
	return false
}

func SaveModuleBackup(man *manifest.File, module string, backup map[string]string) {
	if len(backup) == 0 {
		return
	}
	st := man.Modules[module]
	if st.Meta == nil {
		st.Meta = map[string]string{}
	}
	for k, v := range backup {
		st.Meta["backup."+k] = v
	}
	man.Modules[module] = st
}
