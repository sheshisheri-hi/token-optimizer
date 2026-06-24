package jsonmerge

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
)

// MergeFile reads path (or starts empty), deep-merges updates, writes if changed.
// Returns list of changed top-level keys and backup of previous leaf values (dotted paths).
func MergeFile(path string, updates map[string]any) (actions []string, backup map[string]string, err error) {
	backup = map[string]string{}
	root := map[string]any{}
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &root)
	}
	if root == nil {
		root = map[string]any{}
	}
	changed := deepMerge(root, updates, "", backup)
	if !changed {
		return nil, backup, nil
	}
	if err := os.MkdirAll(dirOf(path), 0o755); err != nil {
		return nil, backup, err
	}
	data, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return nil, backup, err
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return nil, backup, err
	}
	return []string{"merge " + path}, backup, nil
}

func deepMerge(dst, src map[string]any, prefix string, backup map[string]string) bool {
	changed := false
	for k, v := range src {
		path := k
		if prefix != "" {
			path = prefix + "." + k
		}
		existing, ok := dst[k]
		srcMap, srcIsMap := v.(map[string]any)
		if srcIsMap {
			var dstMap map[string]any
			if ok {
				dstMap, _ = existing.(map[string]any)
			}
			if dstMap == nil {
				dstMap = map[string]any{}
				dst[k] = dstMap
			}
			if deepMerge(dstMap, srcMap, path, backup) {
				changed = true
			}
			continue
		}
		if ok && reflect.DeepEqual(existing, v) {
			continue
		}
		if ok {
			backup[path] = fmt.Sprintf("%v", existing)
		}
		dst[k] = v
		changed = true
	}
	return changed
}

func dirOf(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[:i]
		}
	}
	return "."
}

// RestoreFile applies backup dotted paths onto file (best-effort revert).
func RestoreFromBackup(path string, backup map[string]string) error {
	if len(backup) == 0 {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return err
	}
	for dotted, val := range backup {
		setDotted(root, dotted, val)
	}
	out, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(out, '\n'), 0o644)
}

func setDotted(root map[string]any, path string, val string) {
	parts := splitPath(path)
	walk := root
	for i, p := range parts {
		if i == len(parts)-1 {
			walk[p] = val
			return
		}
		next, ok := walk[p].(map[string]any)
		if !ok {
			next = map[string]any{}
			walk[p] = next
		}
		walk = next
	}
}

func splitPath(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	out = append(out, s[start:])
	return out
}

func LoadMap(data []byte) (map[string]any, error) {
	var m map[string]any
	if len(data) == 0 {
		return map[string]any{}, nil
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	if m == nil {
		m = map[string]any{}
	}
	return m, nil
}
