package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveLang(t *testing.T) {
	cfg := NewConfig()
	cfg.File.DefaultLang = "py"
	if got := cfg.ResolveLang("cpp"); got != "cpp" {
		t.Fatalf("ResolveLang(%q) = %q, want %q", "cpp", got, "cpp")
	}
	if got := cfg.ResolveLang(""); got != "py" {
		t.Fatalf("ResolveLang(%q) = %q, want %q", "", got, "py")
	}
	cfg.File.DefaultLang = ""
	if got := cfg.ResolveLang(""); got != "cpp" {
		t.Fatalf("ResolveLang(%q) = %q, want %q", "", got, "cpp")
	}
}

func TestUpdateWritesEmptyAtCoderSession(t *testing.T) {
	root := t.TempDir()
	cfg := NewConfig()
	cfg.File.RootDir = root
	cfg.File.LibraryDirs = []string{filepath.Join(root, "library")}
	cfg.File.DefaultLang = "cpp"
	cfg.File.AtCoderSession = ""

	if err := Update(cfg); err != nil {
		t.Fatalf("Update: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, ".cpx", "config.yaml"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "atcoder_session:") {
		t.Fatalf("expected atcoder_session in config, got:\n%s", content)
	}
}
