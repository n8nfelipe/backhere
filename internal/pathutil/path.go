package pathutil

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// NormalizePath returns a stable, absolute representation of a directory.
// Existing symlinks are resolved so a symlink and its target share reminders.
func NormalizePath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("o caminho não pode estar vazio")
	}

	if path == "~" || strings.HasPrefix(path, "~/") ||
		(runtime.GOOS == "windows" && strings.HasPrefix(path, `~\`)) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("localizar diretório pessoal: %w", err)
		}
		if path == "~" {
			path = home
		} else {
			path = filepath.Join(home, filepath.FromSlash(strings.ReplaceAll(path[2:], `\`, "/")))
		}
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("normalizar caminho: %w", err)
	}
	abs = filepath.Clean(abs)

	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", fmt.Errorf("resolver caminho %q: %w", abs, err)
	}
	resolved = filepath.Clean(resolved)
	if runtime.GOOS == "windows" {
		resolved = strings.ToLower(resolved)
	}
	return resolved, nil
}
