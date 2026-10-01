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
	dir := t.TempDir()
	libraryDir := filepath.Join(dir, "library")
	if err := os.MkdirAll(libraryDir, 0o755); err != nil {
		t.Fatalf("mkdir library: %v", err)
	}
	util := "#pragma once\n#include <vector>\n/* -- library code --*/\nclass Util {};\n/* -- library code --*/\n"
	widget := "#pragma once\n#include \"util.hpp\"\n/* -- library code --*/\nclass Test {};\n/* -- library code --*/\n"
	if err := os.WriteFile(filepath.Join(libraryDir, "util.hpp"), []byte(util), 0o644); err != nil {
		t.Fatalf("write util.hpp: %v", err)
	}
	if err := os.WriteFile(filepath.Join(libraryDir, "widget.hpp"), []byte(widget), 0o644); err != nil {
		t.Fatalf("write widget.hpp: %v", err)
	}
	mainPath := filepath.Join(dir, "main.cpp")
	main := "#include <iostream>\n#include \"widget.hpp\"\n" + LibraryMarker + "\nint main() {}\n"
	if err := os.WriteFile(mainPath, []byte(main), 0o644); err != nil {
		t.Fatalf("write main.cpp: %v", err)
	}

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
