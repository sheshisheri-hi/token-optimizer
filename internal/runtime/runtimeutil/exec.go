package runtimeutil

import (
	"os/exec"
)

func execLookPath(name string) (string, error) {
	return exec.LookPath(name)
}

func execRun(shellCmd string) (string, error) {
	cmd := exec.Command("sh", "-c", shellCmd)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
