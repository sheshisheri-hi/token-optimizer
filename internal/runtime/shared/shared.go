package shared

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	goruntime "runtime"
	"time"
)

func ArchiveDirName() string {
	return time.Now().UTC().Format("20060102T150405Z")
}

func IDEUserSettings(app string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	switch goruntime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", app, "User", "settings.json"), nil
	case "windows":
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, app, "User", "settings.json"), nil
		}
		return "", fmt.Errorf("APPDATA not set")
	default:
		return filepath.Join(home, ".config", app, "User", "settings.json"), nil
	}
}

func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func ArchiveFile(path, archiveRoot string) error {
	if !FileExists(path) {
		return os.ErrNotExist
	}
	dest := filepath.Join(archiveRoot, filepath.Base(path))
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	return os.Rename(path, dest)
}

func ArchiveDir(path, archiveRoot string) error {
	if !FileExists(path) {
		return os.ErrNotExist
	}
	dest := filepath.Join(archiveRoot, filepath.Base(path))
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	return os.Rename(path, dest)
}

func ArchiveRoot(dataHome string) string {
	return filepath.Join(filepath.Dir(dataHome), "slim-archive", ArchiveDirName())
}

func Contains(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}
