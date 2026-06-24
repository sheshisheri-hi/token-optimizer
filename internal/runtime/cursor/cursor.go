package cursor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sheshisheri-hi/token-optimizer/internal/core/capabilities"
	"github.com/sheshisheri-hi/token-optimizer/internal/core/config"
	"github.com/sheshisheri-hi/token-optimizer/internal/core/manifest"
	"github.com/sheshisheri-hi/token-optimizer/internal/core/registry"
	"github.com/sheshisheri-hi/token-optimizer/internal/runtime"
	"github.com/sheshisheri-hi/token-optimizer/internal/runtime/runtimeutil"
	"github.com/sheshisheri-hi/token-optimizer/internal/runtime/shared"
	"github.com/sheshisheri-hi/token-optimizer/internal/version"
)

const hookFileName = "hooks.json"
const slimMarker = "slim internal hook"

type Runtime struct {
	slimBin string
}

func New() *Runtime {
	bin, _ := os.Executable()
	return &Runtime{slimBin: bin}
}

func (r *Runtime) ID() string          { return "cursor" }
func (r *Runtime) DisplayName() string { return "Cursor" }

func (r *Runtime) DataHome() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cursor", "slim"), nil
}

func (r *Runtime) cursorHome() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cursor"), nil
}

func (r *Runtime) skillsDir() (string, error) {
	home, err := r.cursorHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "skills"), nil
}

func (r *Runtime) hooksPath() (string, error) {
	home, err := r.cursorHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, hookFileName), nil
}

func (r *Runtime) defaultModules() []string {
	return registry.MustLoad().DefaultModules()
}

func (r *Runtime) Setup(ctx context.Context, opts runtime.SetupOpts) (*runtime.SetupResult, error) {
	reg, err := registry.Load()
	if err != nil {
		return nil, err
	}
	modules := opts.Modules
	if opts.All || len(modules) == 0 {
		modules = r.defaultModules()
	}

	result := &runtime.SetupResult{}
	dataHome, err := r.DataHome()
	if err != nil {
		return nil, err
	}

	for _, mod := range modules {
		switch mod {
		case "hooks":
			actions, err := r.setupHooks(dataHome, opts.DryRun)
			if err != nil {
				return nil, err
			}
			result.Actions = append(result.Actions, actions...)
		case "skills":
			actions, err := r.setupSkills(opts.DryRun)
			if err != nil {
				return nil, err
			}
			result.Actions = append(result.Actions, actions...)
		case "settings":
			actions, backup, err := r.setupSettings(opts.DryRun)
			if err != nil {
				return nil, err
			}
			result.Actions = append(result.Actions, actions...)
			if !opts.DryRun && len(backup) > 0 {
				man, _ := manifest.Load(dataHome)
				runtimeutil.SaveModuleBackup(man, "settings", backup)
				_ = manifest.Save(dataHome, man)
			}
		case "vscode":
			actions, backup, err := r.setupVSCode(opts.DryRun)
			if err != nil {
				return nil, err
			}
			result.Actions = append(result.Actions, actions...)
			if !opts.DryRun && len(backup) > 0 {
				man, _ := manifest.Load(dataHome)
				runtimeutil.SaveModuleBackup(man, "vscode", backup)
				_ = manifest.Save(dataHome, man)
			}
		case "measurement":
			actions, err := runtimeutil.SetupMeasurement(dataHome, opts.DryRun)
			if err != nil {
				return nil, err
			}
			result.Actions = append(result.Actions, actions...)
		case "hooks-rtk":
			actions, err := runtimeutil.SetupRTK(r.ID(), opts.DryRun)
			if err != nil {
				return nil, err
			}
			result.Actions = append(result.Actions, actions...)
		default:
			if contains(reg.ModuleNames(), mod) {
				result.Actions = append(result.Actions, fmt.Sprintf("skip unimplemented module %q for cursor", mod))
			} else {
				result.Actions = append(result.Actions, fmt.Sprintf("skip unknown module %q", mod))
			}
		}
	}

	if opts.DryRun {
		result.Actions = append(result.Actions, "dry-run: manifest and config not written")
		return result, nil
	}

	if err := os.MkdirAll(filepath.Join(dataHome, "logs"), 0o755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(dataHome, "sessions"), 0o755); err != nil {
		return nil, err
	}

	cfg := config.Default()
	if err := config.Save(dataHome, cfg); err != nil {
		return nil, err
	}
	result.Actions = append(result.Actions, "write "+config.Path(dataHome))

	man, err := manifest.Load(dataHome)
	if err != nil {
		return nil, err
	}
	man.Touch(r.ID(), version.Version)
	man.EnableModules(modules)
	if err := manifest.Save(dataHome, man); err != nil {
		return nil, err
	}
	result.Actions = append(result.Actions, "write "+manifest.Path(dataHome))
	_ = capabilities.Seed(dataHome)

	return result, nil
}

func (r *Runtime) setupSettings(dryRun bool) ([]string, map[string]string, error) {
	home, err := r.cursorHome()
	if err != nil {
		return nil, nil, err
	}
	path := filepath.Join(home, "settings.json")
	return runtimeutil.MergeTemplateJSON("cursor/settings.json", path, dryRun)
}

func (r *Runtime) setupVSCode(dryRun bool) ([]string, map[string]string, error) {
	path, err := shared.IDEUserSettings("Cursor")
	if err != nil {
		return nil, nil, err
	}
	return runtimeutil.MergeTemplateJSON("cursor/vscode-settings.json", path, dryRun)
}

func (r *Runtime) setupHooks(dataHome string, dryRun bool) ([]string, error) {
	binDir := filepath.Join(dataHome, "bin")
	binPath := filepath.Join(binDir, "slim")
	hookPath, err := r.hooksPath()
	if err != nil {
		return nil, err
	}

	if dryRun {
		return []string{
			fmt.Sprintf("would copy binary → %s", binPath),
			fmt.Sprintf("would merge hooks → %s", hookPath),
		}, nil
	}

	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return nil, err
	}
	if err := shared.CopyFile(r.slimBin, binPath); err != nil {
		return nil, err
	}

	slimHooks := slimHookEntries(binPath)
	merged, err := mergeHooksFile(hookPath, slimHooks)
	if err != nil {
		return nil, err
	}
	data, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(hookPath), 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(hookPath, append(data, '\n'), 0o644); err != nil {
		return nil, err
	}

	return []string{"install " + binPath, "merge " + hookPath}, nil
}

func slimHookEntries(binPath string) map[string][]any {
	cmd := func(event string) map[string]any {
		return map[string]any{
			"command": fmt.Sprintf("%s internal hook %s --target cursor", quotePath(binPath), event),
		}
	}
	return map[string][]any{
		"sessionStart": {cmd("session-start")},
		"preToolUse": {
			map[string]any{
				"matcher": "Shell",
				"command": fmt.Sprintf("%s internal hook pre-tool-use --target cursor", quotePath(binPath)),
			},
		},
		"postToolUse": {cmd("post-tool-use")},
		"stop":        {cmd("stop")},
	}
}

func mergeHooksFile(path string, slimEntries map[string][]any) (map[string]any, error) {
	var root map[string]any
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &root)
	}
	if root == nil {
		root = map[string]any{}
	}
	if _, ok := root["version"]; !ok {
		root["version"] = 1
	}
	hooks, _ := root["hooks"].(map[string]any)
	if hooks == nil {
		hooks = map[string]any{}
	}
	for event, entries := range slimEntries {
		existing, _ := hooks[event].([]any)
		var kept []any
		for _, e := range existing {
			m, ok := e.(map[string]any)
			if !ok {
				kept = append(kept, e)
				continue
			}
			c, _ := m["command"].(string)
			if strings.Contains(c, slimMarker) {
				continue
			}
			kept = append(kept, e)
		}
		hooks[event] = append(kept, entries...)
	}
	root["hooks"] = hooks
	return root, nil
}

func (r *Runtime) setupSkills(dryRun bool) ([]string, error) {
	skillsDir, err := r.skillsDir()
	if err != nil {
		return nil, err
	}
	return runtimeutil.InstallSkills("cursor/skills", skillsDir, dryRun)
}

func (r *Runtime) Off(ctx context.Context, opts runtime.OffOpts) (*runtime.OffResult, error) {
	dataHome, err := r.DataHome()
	if err != nil {
		return nil, err
	}
	man, err := manifest.Load(dataHome)
	if err != nil {
		return nil, err
	}

	modules := opts.Modules
	if opts.All || len(modules) == 0 {
		modules = man.InstalledModuleNames()
		if len(modules) == 0 {
			modules = r.defaultModules()
		}
	}

	result := &runtime.OffResult{}
	archiveRoot := shared.ArchiveRoot(dataHome)

	for _, mod := range modules {
		switch mod {
		case "hooks":
			hookPath, _ := r.hooksPath()
			if err := stripSlimHooks(hookPath); err == nil {
				result.Actions = append(result.Actions, "removed slim hooks from "+hookPath)
			}
		case "skills":
			skillsDir, _ := r.skillsDir()
			for _, name := range []string{"output-control", "token-economy", "mcp-hygiene", "workflow-habits"} {
				p := filepath.Join(skillsDir, name)
				if err := shared.ArchiveDir(p, archiveRoot); err == nil {
					result.Actions = append(result.Actions, "archived "+p)
				}
			}
		}
		man.DisableModules([]string{mod})
	}
	if err := manifest.Save(dataHome, man); err != nil {
		return nil, err
	}
	result.Actions = append(result.Actions, "updated manifest")
	return result, nil
}

func stripSlimHooks(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return err
	}
	hooks, ok := root["hooks"].(map[string]any)
	if !ok {
		return nil
	}
	for event, raw := range hooks {
		list, ok := raw.([]any)
		if !ok {
			continue
		}
		var kept []any
		for _, e := range list {
			m, ok := e.(map[string]any)
			if !ok {
				kept = append(kept, e)
				continue
			}
			c, _ := m["command"].(string)
			if strings.Contains(c, slimMarker) {
				continue
			}
			kept = append(kept, e)
		}
		hooks[event] = kept
	}
	root["hooks"] = hooks
	out, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(out, '\n'), 0o644)
}

func (r *Runtime) Status(ctx context.Context) (*runtime.StatusReport, error) {
	return r.buildReport(false)
}

func (r *Runtime) Doctor(ctx context.Context, opts runtime.DoctorOpts) (*runtime.DoctorReport, error) {
	rep, err := r.buildReport(true)
	if err != nil {
		return nil, err
	}
	doc := &runtime.DoctorReport{StatusReport: *rep}
	dataHome, _ := r.DataHome()
	if !shared.FileExists(filepath.Join(dataHome, "logs", "hook.log")) {
		doc.Warnings = append(doc.Warnings, "hook.log not found yet — hooks fire after Cursor agent use")
	}
	return doc, nil
}

func (r *Runtime) Probe(ctx context.Context) (*runtime.ProbeReport, error) {
	return &runtime.ProbeReport{
		Features: map[string]runtime.FeatureStatus{
			"bash_compress":      runtime.StatusOK,
			"session_continuity": runtime.StatusOK,
			"session_logging":    runtime.StatusOK,
		},
	}, nil
}

func (r *Runtime) HookCommand(event string) []string {
	return []string{r.slimBin, "internal", "hook", event, "--target", "cursor"}
}

func (r *Runtime) buildReport(doctor bool) (*runtime.StatusReport, error) {
	skillsDir, _ := r.skillsDir()
	paths := map[string]string{
		"output-control":  filepath.Join(skillsDir, "output-control", "SKILL.md"),
		"token-economy":   filepath.Join(skillsDir, "token-economy", "SKILL.md"),
		"mcp-hygiene":     filepath.Join(skillsDir, "mcp-hygiene", "SKILL.md"),
		"workflow-habits": filepath.Join(skillsDir, "workflow-habits", "SKILL.md"),
	}
	dataHome, err := r.DataHome()
	if err != nil {
		return nil, err
	}
	return runtimeutil.BuildReport(r.ID(), dataHome, paths)
}

func quotePath(s string) string {
	if strings.ContainsAny(s, " \t") {
		return fmt.Sprintf("%q", s)
	}
	return s
}

func contains(list []string, v string) bool {
	return shared.Contains(list, v)
}
