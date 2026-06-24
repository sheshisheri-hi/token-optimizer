package compress

import "strings"

// Search caps grep/rg/find style commands.
func Search(command string) Result {
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
	switch {
	case strings.HasPrefix(lower, "rg "), strings.HasPrefix(lower, "grep "), strings.HasPrefix(lower, "find "):
		if strings.Contains(cmd, " head ") || strings.Contains(cmd, "|head") {
			return Result{Command: command}
		}
		return rewrite(cmd, cmd+" | head -80", "cap search output")
	case strings.HasPrefix(lower, "git grep"):
		if !strings.Contains(cmd, " head ") {
			return rewrite(cmd, cmd+" | head -80", "cap git grep")
		}
	}
	return Result{Command: command}
}
