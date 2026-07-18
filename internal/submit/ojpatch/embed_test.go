package ojpatch

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMaterialize(t *testing.T) {
	dir := t.TempDir()
	submitScript, err := Materialize(dir)
	if err != nil {
		t.Fatalf("Materialize: %v", err)
	}
	if filepath.Base(submitScript) != submitFile {
		t.Fatalf("got %q, want basename %q", submitScript, submitFile)
	}
	for _, name := range []string{patchFile, submitFile} {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("stat %s: %v", path, err)
		}
	}
}
