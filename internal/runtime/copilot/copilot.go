package copilot

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sheshisheri-hi/token-optimizer/internal/version"
	"github.com/sheshisheri-hi/token-optimizer/internal/core/capabilities"
	"github.com/sheshisheri-hi/token-optimizer/internal/core/config"
	"github.com/sheshisheri-hi/token-optimizer/internal/core/manifest"
	"github.com/sheshisheri-hi/token-optimizer/internal/core/registry"
	"github.com/sheshisheri-hi/token-optimizer/internal/runtime"
	"github.com/sheshisheri-hi/token-optimizer/internal/runtime/runtimeutil"
	"github.com/sheshisheri-hi/token-optimizer/internal/runtime/shared"
)

const hookFileName = "copilot-slim.json"

type Runtime struct {
	slimBin string
}

func New() *Runtime {
	bin, _ := os.Executable()
	return &Runtime{slimBin: bin}
}

func (r *Runtime) ID() string          { return "copilot" }
func (r *Runtime) DisplayName() string { return "GitHub Copilot (VS Code)" }

func (r *Runtime) DataHome() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".copilot", "slim"), nil
}

func (r *Runtime) copilotHome() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".copilot"), nil
}

func (r *Runtime) hooksDir() (string, error) {
	home, err := r.copilotHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "hooks"), nil
}

func (r *Runtime) skillsDir() (string, error) {
	home, err := r.copilotHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "skills"), nil
}

func (r *Runtime) Setup(ctx context.Context, opts runtime.SetupOpts) (*runtime.SetupResult, error) {
	reg, err := registry.Load()
	if err != nil {
		return nil, err
	}
	modules := opts.Modules
	if opts.All || len(modules) == 0 {
		modules = reg.DefaultModules()
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
			result.Actions = append(result.Actions, fmt.Sprintf("skip unknown module %q", mod))
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
	home, err := r.copilotHome()
	if err != nil {
		return nil, nil, err
	}
	path := filepath.Join(home, "settings.json")
	return runtimeutil.MergeTemplateJSON("copilot/settings.json", path, dryRun)
}

func (r *Runtime) setupVSCode(dryRun bool) ([]string, map[string]string, error) {
	path, err := shared.IDEUserSettings("Code")
	if err != nil {
		return nil, nil, err
	}
	return runtimeutil.MergeTemplateJSON("copilot/vscode-settings.json", path, dryRun)
}

func (r *Runtime) setupHooks(dataHome string, dryRun bool) ([]string, error) {
	var actions []string
	binDir := filepath.Join(dataHome, "bin")
	binPath := filepath.Join(binDir, "slim")
	hooksDir, err := r.hooksDir()
	if err != nil {
		return nil, err
	}
	hookPath := filepath.Join(hooksDir, hookFileName)

	if dryRun {
		return []string{
			fmt.Sprintf("would copy binary → %s", binPath),
			fmt.Sprintf("would write hook → %s", hookPath),
		}, nil
	}

	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return nil, err
	}
	if err := shared.CopyFile(r.slimBin, binPath); err != nil {
		return nil, err
	}
	actions = append(actions, "install "+binPath)

	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		return nil, err
	}
	cfg := hookConfig(binPath)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(hookPath, append(data, '\n'), 0o644); err != nil {
		return nil, err
	}
	actions = append(actions, "write "+hookPath)
	return actions, nil
}

func hookConfig(binPath string) map[string]any {
	quote := func(s string) string {
		if strings.ContainsAny(s, " \t") {
			return fmt.Sprintf("%q", s)
		}
		return s
	}
	bin := quote(binPath)
	cmd := func(event string) map[string]any {
		return map[string]any{
			"type":       "command",
			"bash":       fmt.Sprintf("%s internal hook %s --target copilot", bin, event),
			"timeoutSec": 10,
		}
	}
	pre := cmd("pre-tool-use")
	pre["matcher"] = map[string]any{"toolName": "bash"}
	return map[string]any{
		"version": 1,
		"hooks": map[string]any{
			"sessionStart": []any{cmd("session-start")},
			"preToolUse":   []any{pre},
			"postToolUse":  []any{cmd("post-tool-use")},
			"stop":         []any{cmd("stop")},
		},
	}
}

func (r *Runtime) setupSkills(dryRun bool) ([]string, error) {
	skillsDir, err := r.skillsDir()
	if err != nil {
		return nil, err
	}
	return runtimeutil.InstallSkills("copilot/skills", skillsDir, dryRun)
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
			modules = registry.MustLoad().DefaultModules()
		}
	}

	result := &runtime.OffResult{}
	archiveRoot := shared.ArchiveRoot(dataHome)

	for _, mod := range modules {
		switch mod {
		case "hooks":
			hooksDir, _ := r.hooksDir()
			hookPath := filepath.Join(hooksDir, hookFileName)
			if err := shared.ArchiveFile(hookPath, archiveRoot); err == nil {
				result.Actions = append(result.Actions, "archived "+hookPath)
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

func (r *Runtime) Status(ctx context.Context) (*runtime.StatusReport, error) {
	return r.buildReport(false)
}

func (r *Runtime) Doctor(ctx context.Context, opts runtime.DoctorOpts) (*runtime.DoctorReport, error) {
	rep, err := r.buildReport(true)
	if err != nil {
		return nil, err
	}
	doc := &runtime.DoctorReport{StatusReport: *rep}
	if !shared.FileExists(filepath.Join(mustDataHome(r), "logs", "hook.log")) {
		doc.Warnings = append(doc.Warnings, "hook.log not found yet — hooks fire after Copilot agent use")
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
	return []string{r.slimBin, "internal", "hook", event}
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

func mustDataHome(r *Runtime) string {
	p, err := r.DataHome()
	if err != nil {
		return ""
	}
	return p
}
