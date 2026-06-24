package app

import (
	"fmt"
	"strings"

	"github.com/sheshisheri-hi/token-optimizer/internal/runtime"
	"github.com/sheshisheri-hi/token-optimizer/internal/runtime/claude"
	"github.com/sheshisheri-hi/token-optimizer/internal/runtime/copilot"
	"github.com/sheshisheri-hi/token-optimizer/internal/runtime/cursor"
)

func RuntimeFor(target string) (runtime.Runtime, error) {
	switch strings.ToLower(strings.TrimSpace(target)) {
	case "", "copilot":
		return copilot.New(), nil
	case "cursor":
		return cursor.New(), nil
	case "claude":
		return claude.New(), nil
	default:
		return nil, fmt.Errorf("unknown target %q (use copilot, cursor, or claude)", target)
	}
}
