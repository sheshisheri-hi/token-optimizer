package compress

import (
	"strings"
)

type Result struct {
	Command string
	Changed bool
	Reason  string
}

var dangerous = []string{";", "|", "&", "$", "`", "\n", "\r", ">", "<"}

// Bash rewrites safe read-only commands to shorter output. Fail-open: return original on any doubt.
func Bash(command string) Result {
	cmd := strings.TrimSpace(command)
	if cmd == "" {
		return Result{Command: command}
	}
	for _, ch := range dangerous {
		if strings.Contains(cmd, ch) {
			return Result{Command: command, Reason: "dangerous char"}
		}
	}

	lower := strings.ToLower(cmd)
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return Result{Command: command}
	}

	switch {
	case strings.HasPrefix(lower, "git log"):
		return rewrite(cmd, "git log --oneline -30", "limit git log lines")
	case strings.HasPrefix(lower, "git status"):
		return rewrite(cmd, "git status -sb", "short git status")
	case strings.HasPrefix(lower, "git diff"):
		if !strings.Contains(cmd, " -") {
			return rewrite(cmd, cmd+" --stat", "git diff stat summary")
		}
	case strings.HasPrefix(lower, "pytest"):
		if !strings.Contains(lower, "-q") {
			return rewrite(cmd, cmd+" -q --tb=no", "quiet pytest")
		}
	case strings.HasPrefix(lower, "python -m pytest"):
		if !strings.Contains(lower, "-q") {
			return rewrite(cmd, cmd+" -q --tb=no", "quiet pytest")
		}
	case strings.HasPrefix(lower, "cargo test"):
		if !strings.Contains(lower, "--") {
			return rewrite(cmd, cmd+" -- --quiet", "quiet cargo test")
		}
	case strings.HasPrefix(lower, "go test"):
		if !strings.Contains(cmd, " -") {
			return rewrite(cmd, cmd+" -count=1", "go test count=1")
		}
	case strings.HasPrefix(lower, "npm test"), strings.HasPrefix(lower, "pnpm test"), strings.HasPrefix(lower, "yarn test"):
		// leave as-is; flags vary
	case strings.HasPrefix(lower, "rg "), strings.HasPrefix(lower, "grep "), strings.HasPrefix(lower, "find "):
		if !strings.Contains(cmd, " head ") && !strings.Contains(cmd, "| head") {
			return rewrite(cmd, cmd+" | head -50", "cap search output")
		}
	}

	return Result{Command: command}
}

func rewrite(original, next, reason string) Result {
	if original == next {
		return Result{Command: original}
	}
	return Result{Command: next, Changed: true, Reason: reason}
}
