package hook

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sheshisheri-hi/token-optimizer/internal/core/config"
	"github.com/sheshisheri-hi/token-optimizer/internal/core/measurement"
	"github.com/sheshisheri-hi/token-optimizer/internal/hook/compress"
)

// Handle processes a hook event. Returns JSON bytes for stdout (may be nil). Fail-open.
func Handle(target, event, dataHome string) ([]byte, error) {
	cfg, err := config.Load(dataHome)
	if err != nil {
		return nil, err
	}
	payload := readStdin()

	logPath := filepath.Join(dataHome, "logs", "hook.log")
	_ = os.MkdirAll(filepath.Dir(logPath), 0o755)
	line := fmt.Sprintf("%s target=%s event=%s\n", time.Now().UTC().Format(time.RFC3339), target, event)
	_ = appendLog(logPath, line)

	switch normalizeEvent(event) {
	case "session-start":
		if cfg.FeatureEnabled("session_continuity") {
			if out := sessionStartResponse(target, dataHome); out != nil {
				return out, nil
			}
		}
	case "pre-tool-use":
		if out := preToolUseResponse(target, payload, cfg); out != nil {
			_ = appendLog(logPath, fmt.Sprintf("%s tool rewrite\n", time.Now().UTC().Format(time.RFC3339)))
			_ = measurement.Record(dataHome, "compress", "pre-tool-use", 512)
			return out, nil
		}
	case "post-tool-use":
		if cfg.FeatureEnabled("lean_output_nudges") {
			_ = appendLog(logPath, fmt.Sprintf("%s lean_output observe\n", time.Now().UTC().Format(time.RFC3339)))
		}
		if cfg.FeatureEnabled("loop_detection") {
			detectLoop(logPath, payload)
		}
	case "stop", "session-end":
		if cfg.FeatureEnabled("session_logging") {
			_ = writeSessionSummary(dataHome)
			_ = measurement.Record(dataHome, "session", "stop", 0)
		}
	}
	return nil, nil
}

func preToolUseResponse(target string, payload map[string]any, cfg *config.File) []byte {
	if cfg.FeatureEnabled("bash_compress") {
		if cmd := extractShellCommand(payload); cmd != "" {
			if res := compress.Bash(cmd); res.Changed {
				return toolRewriteResponse(target, res.Command)
			}
		}
	}
	if cfg.FeatureEnabled("search_compress") {
		if cmd := extractShellCommand(payload); cmd != "" {
			if res := compress.Search(cmd); res.Changed {
				return toolRewriteResponse(target, res.Command)
			}
		}
		if tool, _ := payload["tool_name"].(string); strings.EqualFold(tool, "Grep") {
			// Cursor native Grep — log only; rewrite not supported without command string
			return nil
		}
	}
	return nil
}

func toolRewriteResponse(target, command string) []byte {
	switch strings.ToLower(target) {
	case "cursor":
		out := map[string]any{
			"permission":    "allow",
			"updated_input": map[string]any{"command": command},
		}
		data, _ := json.Marshal(out)
		return data
	default:
		out := map[string]any{
			"updatedInput": map[string]any{"command": command},
		}
		data, _ := json.Marshal(out)
		return data
	}
}

func sessionStartResponse(target, dataHome string) []byte {
	latest := filepath.Join(dataHome, "sessions", "latest.json")
	data, err := os.ReadFile(latest)
	if err != nil {
		return nil
	}
	var m map[string]any
	if json.Unmarshal(data, &m) != nil {
		return nil
	}
	hint, _ := m["hint"].(string)
	if hint == "" {
		hint = "Continue from last session checkpoint. Stay terse."
	}
	switch strings.ToLower(target) {
	case "cursor":
		out := map[string]any{"agent_message": hint}
		b, _ := json.Marshal(out)
		return b
	default:
		out := map[string]any{"additionalContext": hint}
		b, _ := json.Marshal(out)
		return b
	}
}

func detectLoop(logPath string, payload map[string]any) {
	cmd := extractShellCommand(payload)
	if cmd == "" {
		return
	}
	data, _ := os.ReadFile(logPath)
	lines := strings.Split(string(data), "\n")
	similar := 0
	for i := len(lines) - 1; i >= 0 && i >= len(lines)-6; i-- {
		if strings.Contains(lines[i], cmd) {
			similar++
		}
	}
	if similar >= 3 {
		_ = appendLog(logPath, fmt.Sprintf("%s loop_detected cmd=%q\n", time.Now().UTC().Format(time.RFC3339), cmd))
	}
}

func normalizeEvent(event string) string {
	switch strings.ToLower(strings.TrimSpace(event)) {
	case "sessionstart", "session-start":
		return "session-start"
	case "pretooluse", "pre-tool-use":
		return "pre-tool-use"
	case "posttooluse", "post-tool-use":
		return "post-tool-use"
	case "sessionend", "session-end":
		return "session-end"
	default:
		return strings.ToLower(event)
	}
}

func extractShellCommand(payload map[string]any) string {
	if payload == nil {
		return ""
	}
	if c, ok := payload["command"].(string); ok && c != "" {
		return c
	}
	if input, ok := payload["input"].(map[string]any); ok {
		if c, ok := input["command"].(string); ok {
			return c
		}
	}
	if input, ok := payload["tool_input"].(map[string]any); ok {
		if c, ok := input["command"].(string); ok {
			return c
		}
	}
	tool := ""
	if name, ok := payload["toolName"].(string); ok {
		tool = name
	}
	if name, ok := payload["tool_name"].(string); ok {
		tool = name
	}
	if tool != "" && !strings.EqualFold(tool, "bash") && !strings.EqualFold(tool, "Shell") {
		return ""
	}
	return ""
}

func readStdin() map[string]any {
	data, _ := io.ReadAll(io.LimitReader(os.Stdin, 4<<20))
	if len(data) == 0 {
		return nil
	}
	var m map[string]any
	_ = json.Unmarshal(data, &m)
	return m
}

func appendLog(path, line string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(line)
	return err
}

func writeSessionSummary(dataHome string) error {
	dir := filepath.Join(dataHome, "sessions")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	now := time.Now().UTC()
	payload := map[string]any{
		"endedAt": now.Format(time.RFC3339),
		"hint":    "Prior session ended " + now.Format(time.RFC3339) + ". Resume with minimal context.",
		"note":    "session stop hook",
	}
	data, _ := json.MarshalIndent(payload, "", "  ")
	_ = os.WriteFile(filepath.Join(dir, "latest.json"), append(data, '\n'), 0o644)
	name := fmt.Sprintf("%d.json", now.Unix())
	return os.WriteFile(filepath.Join(dir, name), append(data, '\n'), 0o644)
}
