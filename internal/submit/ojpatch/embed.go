package ojpatch

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed oj_patch.py oj_submit.py
var scripts embed.FS

const (
	patchFile  = "oj_patch.py"
	submitFile = "oj_submit.py"
)

// Materialize writes the embedded oj patch scripts into dir and returns the
// absolute path to oj_submit.py.
func Materialize(dir string) (submitScript string, err error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create ojpatch dir: %w", err)
	}
	for _, name := range []string{patchFile, submitFile} {
		data, err := scripts.ReadFile(name)
		if err != nil {
			return "", fmt.Errorf("read embedded %s: %w", name, err)
		}
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return "", fmt.Errorf("write %s: %w", path, err)
		}
	}
	return filepath.Join(dir, submitFile), nil
}

// CacheDir returns a stable cache directory for the embedded oj patch scripts.
func CacheDir() (string, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cache, "cpx", "ojpatch"), nil
}
