package templates

import "embed"

// FS holds embedded templates (registry, skills).
//
//go:embed feature-registry.yaml copilot/skills copilot/settings.json copilot/vscode-settings.json cursor/skills cursor/settings.json cursor/vscode-settings.json
var FS embed.FS
