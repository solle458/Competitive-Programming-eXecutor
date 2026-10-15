package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"Competitive-Programming-eXecutor/internal/config"
	"Competitive-Programming-eXecutor/internal/merge"
	"Competitive-Programming-eXecutor/internal/problem"
)

func TestSubmissionTextMatchesMerge(t *testing.T) {
	dir := t.TempDir()
	lib := filepath.Join(dir, "library")
	if err := os.MkdirAll(lib, 0o755); err != nil {
		t.Fatal(err)
	}
	header := "" +
		"#pragma once\n" +
		"#include <vector>\n" +
		"using namespace header_only;\n" +
		"using Value = int;\n" +
		"using Base::name;\n" +
		"/* -- library code --*/\n" +
		"Value answer() { return 1; }\n"
	if err := os.WriteFile(filepath.Join(lib, "util.hpp"), []byte(header), 0o644); err != nil {
		t.Fatal(err)
	}
	main := "" +
		"#include <iostream>\n" +
		"#include \"util.hpp\"\n" +
		"using namespace std;\n" +
		"\n" +
		merge.LibraryMarker + "\n" +
		"int main() { std::cout << answer(); }\n"
	problemDir := filepath.Join(dir, "a")
	mainPath := filepath.Join(problemDir, "main.cpp")
	if err := os.MkdirAll(problemDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(mainPath, []byte(main), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := config.NewConfig()
	cfg.File.LibraryDirs = []string{lib}
	p := problem.Problem{Dir: problemDir, Lang: "cpp"}
	got, err := submissionText(p, cfg)
	if err != nil {
		t.Fatal(err)
	}
	want, err := merge.Generate(main, cfg.File.LibraryDirs)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("submissionText()\n got %q\nwant %q", got, want)
	}
	if strings.Contains(got, merge.CodeMarker) || strings.Contains(got, "#pragma once") || strings.Contains(got, "using namespace header_only") {
		t.Fatalf("header boilerplate leaked: %q", got)
	}
	if !strings.Contains(got, "using Value = int;") || !strings.Contains(got, "using Base::name;") {
		t.Fatalf("using declarations dropped: %q", got)
	}
	if !strings.Contains(got, "using namespace std;") {
		t.Fatalf("solution using namespace dropped: %q", got)
	}
	if strings.Contains(got, "#include \"util.hpp\"") {
		t.Fatalf("hpp include was not expanded: %q", got)
	}
	iostream := strings.Index(got, "#include <iostream>")
	vector := strings.Index(got, "#include <vector>")
	body := strings.Index(got, "Value answer()")
	if iostream < 0 || vector < 0 || body < 0 || !(iostream < vector && vector < body) {
		t.Fatalf("includes were not gathered before the body: %q", got)
	}
	disk, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(disk) != main {
		t.Fatalf("main.cpp was rewritten\n got %q", disk)
	}
	if _, err := os.Stat(filepath.Join(problemDir, "submission.cpp")); !os.IsNotExist(err) {
		t.Fatalf("submission.cpp should not exist, stat err=%v", err)
	}
}

func TestRunSamplesCompilesShapedSubmission(t *testing.T) {
	if _, err := exec.LookPath("g++"); err != nil {
		t.Skip("g++ not installed")
	}
	dir := t.TempDir()
	lib := filepath.Join(dir, "library")
	if err := os.MkdirAll(lib, 0o755); err != nil {
		t.Fatal(err)
	}
	header := "" +
		"#pragma once\n" +
		"#include <iostream>\n" +
		"using namespace missing_ns;\n" +
		"using Value = int;\n" +
		merge.CodeMarker + "\n" +
		"Value answer() { return 42; }\n"
	if err := os.WriteFile(filepath.Join(lib, "util.hpp"), []byte(header), 0o644); err != nil {
		t.Fatal(err)
	}
	main := "" +
		"#include \"util.hpp\"\n" +
		merge.LibraryMarker + "\n" +
		"int main() { std::cout << answer() << std::endl; }\n"
	problemDir := filepath.Join(dir, "a")
	if err := os.MkdirAll(filepath.Join(problemDir, "test"), 0o755); err != nil {
		t.Fatal(err)
	}
	mainPath := filepath.Join(problemDir, "main.cpp")
	if err := os.WriteFile(mainPath, []byte(main), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(problemDir, "test", "sample-1.in"), []byte("\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(problemDir, "test", "sample-1.out"), []byte("42\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := config.NewConfig()
	cfg.File.LibraryDirs = []string{lib}
	err := RunSamples(problem.Problem{Dir: problemDir, Lang: "cpp"}, Options{TimeLimit: 2}, cfg)
	if err != nil {
		t.Fatal(err)
	}
	disk, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(disk) != main {
		t.Fatalf("main.cpp was rewritten\n got %q", disk)
	}
	if _, err := os.Stat(filepath.Join(problemDir, "submission.cpp")); !os.IsNotExist(err) {
		t.Fatalf("submission.cpp should not exist, stat err=%v", err)
	}
}
