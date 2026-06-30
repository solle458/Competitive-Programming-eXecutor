package merge

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseInclude(t *testing.T) {
	t.Parallel()

	tests := []struct {
		line string
		want string
		ok   bool
	}{
		{`#include <iostream>`, "iostream", true},
		{`#include "test/test.hpp"`, "test/test.hpp", true},
		{`  #include   "a/b.hpp"  `, "a/b.hpp", true},
		{`int main() {}`, "", false},
		{`#include broken`, "", false},
	}

	for _, tt := range tests {
		got, ok := parseInclude(tt.line)
		if ok != tt.ok || got != tt.want {
			t.Errorf("parseInclude(%q) = (%q, %v), want (%q, %v)", tt.line, got, ok, tt.want, tt.ok)
		}
	}
}

func TestGenerate(t *testing.T) {
	root := filepath.Join("..", "..", "test")
	libraryDir := filepath.Join(root, "library")
	mainPath := filepath.Join(root, "test_contest", "a", "main.cpp")

	source, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatalf("read main.cpp: %v", err)
	}

	got, err := Generate(string(source), []string{libraryDir})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if !strings.Contains(got, "class Util") {
		t.Fatalf("missing dependency code: Util")
	}
	if !strings.Contains(got, "class Test") {
		t.Fatalf("missing library code: Test")
	}
	if strings.Contains(got, CodeMarker) {
		t.Fatalf("library code marker leaked into submission")
	}
	if !strings.Contains(got, LibraryMarker) {
		t.Fatalf("libraries marker missing from submission")
	}
}

func TestGenerateLibrariesMarkerRequired(t *testing.T) {
	dir := t.TempDir()
	libDir := filepath.Join(dir, "lib")
	if err := os.MkdirAll(filepath.Join(libDir, "test"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	library := `#pragma once
/* -- library code --*/
int x;
/* -- library code --*/
`
	if err := os.WriteFile(filepath.Join(libDir, "test", "x.hpp"), []byte(library), 0o644); err != nil {
		t.Fatalf("write library: %v", err)
	}

	source := "#include \"test/x.hpp\"\nint main() {}\n"
	_, err := Generate(source, []string{libDir})
	if err == nil {
		t.Fatal("expected error when libraries marker is missing")
	}
	if !strings.Contains(err.Error(), LibraryMarker) {
		t.Fatalf("error should mention marker: %v", err)
	}
}
