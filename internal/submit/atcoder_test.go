package submit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveContestAndIndexWithPath(t *testing.T) {
	contestID, index, err := resolveContestAndIndex("abc464/a")
	if err != nil {
		t.Fatalf("resolveContestAndIndex: %v", err)
	}
	if contestID != "abc464" || index != "a" {
		t.Fatalf("got (%q, %q), want (abc464, a)", contestID, index)
	}
}

func TestResolveContestAndIndexFromCWD(t *testing.T) {
	contestDir := filepath.Join(t.TempDir(), "abc464")
	if err := os.MkdirAll(contestDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.Chdir(contestDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	contestID, index, err := resolveContestAndIndex("a")
	if err != nil {
		t.Fatalf("resolveContestAndIndex: %v", err)
	}
	if contestID != "abc464" || index != "a" {
		t.Fatalf("got (%q, %q), want (abc464, a)", contestID, index)
	}
}

func TestTaskURL(t *testing.T) {
	got, err := taskURL("abc464/a")
	if err != nil {
		t.Fatalf("taskURL: %v", err)
	}
	want := "https://atcoder.jp/contests/abc464/tasks/abc464_a"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestOJCookiePath(t *testing.T) {
	path, err := ojCookiePath()
	if err != nil {
		t.Fatalf("ojCookiePath: %v", err)
	}
	if path == "" {
		t.Fatal("expected non-empty cookie path")
	}
}
