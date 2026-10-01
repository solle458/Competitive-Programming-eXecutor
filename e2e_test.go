package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"testing"
	"time"
)

var cpxBin string

func TestMain(m *testing.M) {
	bin := os.Getenv("CPX_BIN")
	var buildDir string
	if bin == "" {
		dir, err := os.MkdirTemp("", "cpx-e2e-bin")
		if err != nil {
			fmt.Fprintf(os.Stderr, "mkdir: %v\n", err)
			os.Exit(1)
		}
		buildDir = dir
		bin = filepath.Join(dir, "cpx")
		_, file, _, ok := runtime.Caller(0)
		if !ok {
			fmt.Fprintln(os.Stderr, "runtime.Caller failed")
			os.Exit(1)
		}
		cmd := exec.Command("go", "build", "-o", bin, ".")
		cmd.Dir = filepath.Dir(file)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			os.RemoveAll(buildDir)
			fmt.Fprintf(os.Stderr, "go build: %v\n", err)
			os.Exit(1)
		}
	} else if !filepath.IsAbs(bin) {
		abs, err := filepath.Abs(bin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "CPX_BIN: %v\n", err)
			os.Exit(1)
		}
		bin = abs
	}
	cpxBin = bin
	code := m.Run()
	if buildDir != "" {
		os.RemoveAll(buildDir)
	}
	os.Exit(code)
}

func TestE2EInit(t *testing.T) {
	dir := resolvedTemp(t)
	stdout, stderr, code := runAt(t, dir, "init")
	if code != 0 {
		t.Fatalf("exit %d\nstderr: %s", code, stderr)
	}
	if stderr != "" {
		t.Fatalf("stderr %q", stderr)
	}
	wantOut := fmt.Sprintf("[INFO] initialized cpx workspace in %s\n", dir)
	if stdout != wantOut {
		t.Fatalf("stdout\n got %q\nwant %q", stdout, wantOut)
	}

	cfg, err := os.ReadFile(filepath.Join(dir, ".cpx", "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	wantCfg := fmt.Sprintf("file:\n    root_dir: %s\n    library_dirs:\n        - %s/library\n    default_lang: cpp\n    atcoder_session: \"\"\n", dir, dir)
	if string(cfg) != wantCfg {
		t.Fatalf("config\n got %q\nwant %q", cfg, wantCfg)
	}

	src, err := os.ReadFile(filepath.Join(dir, ".cpx", "templates", "source", "source_template.cpp"))
	if err != nil {
		t.Fatal(err)
	}
	if string(src) != sourceTemplate {
		t.Fatalf("source template\n got %q\nwant %q", src, sourceTemplate)
	}
	lib, err := os.ReadFile(filepath.Join(dir, ".cpx", "templates", "library", "library_template.hpp"))
	if err != nil {
		t.Fatal(err)
	}
	if string(lib) != libraryTemplate {
		t.Fatalf("library template\n got %q\nwant %q", lib, libraryTemplate)
	}
	if _, err := os.Stat(filepath.Join(dir, "library")); !os.IsNotExist(err) {
		t.Fatalf("library directory should not exist, stat err=%v", err)
	}
}

func TestE2ENewLib(t *testing.T) {
	dir := workspace(t)

	stdout, stderr, code := runAt(t, dir, "new-lib", "graph/dsu.hpp")
	if code != 0 || stderr != "" {
		t.Fatalf("create code=%d stderr=%q stdout=%q", code, stderr, stdout)
	}
	created := filepath.Join(dir, "library", "graph", "dsu.hpp")
	wantOut := fmt.Sprintf("[INFO] created %s\n[INFO] include: #include \"graph/dsu.hpp\"\n", created)
	if stdout != wantOut {
		t.Fatalf("stdout\n got %q\nwant %q", stdout, wantOut)
	}
	body, err := os.ReadFile(created)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != libraryTemplate {
		t.Fatalf("created file\n got %q\nwant %q", body, libraryTemplate)
	}

	stdout, stderr, code = runAt(t, dir, "new-lib", "graph/dsu.hpp")
	if code != 1 {
		t.Fatalf("overwrite exit %d, want 1\nstdout %q\nstderr %q", code, stdout, stderr)
	}
	if stdout != "" {
		t.Fatalf("overwrite stdout %q", stdout)
	}
	wantErr := fmt.Sprintf("Error: library already exists: %s\n%s", created, newLibUsage)
	if stderr != wantErr {
		t.Fatalf("overwrite stderr\n got %q\nwant %q", stderr, wantErr)
	}

	stdout, stderr, code = runAt(t, dir, "new-lib", "../x")
	if code != 1 || stdout != "" {
		t.Fatalf("reject code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	wantErr = "Error: path must be relative to the library directory: \"../x\"\n" + newLibUsage
	if stderr != wantErr {
		t.Fatalf("reject stderr\n got %q\nwant %q", stderr, wantErr)
	}
}

func TestE2EMerge(t *testing.T) {
	dir := workspace(t)
	writeFile(t, filepath.Join(dir, "library", "util.hpp"), utilLibrary)
	writeFile(t, filepath.Join(dir, "library", "widget.hpp"), widgetLibrary)
	contest := filepath.Join(dir, "contest")
	writeFile(t, filepath.Join(contest, "a", "main.cpp"), mergeMain)

	stdout, stderr, code := runAt(t, contest, "merge", "a")
	if code != 0 || stderr != "" {
		t.Fatalf("merge code=%d stderr=%q stdout=%q", code, stderr, stdout)
	}
	if stdout != "[INFO] wrote a/submission.cpp\n" {
		t.Fatalf("stdout %q", stdout)
	}
	got, err := os.ReadFile(filepath.Join(contest, "a", "submission.cpp"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != mergedSubmission {
		t.Fatalf("submission.cpp\n got %q\nwant %q", got, mergedSubmission)
	}

	writeFile(t, filepath.Join(dir, "library", "only.hpp"), onlyLibrary)
	writeFile(t, filepath.Join(contest, "b", "main.cpp"), "#include \"only.hpp\"\nint main() { return 0; }\n")
	stdout, stderr, code = runAt(t, contest, "merge", "b")
	if code != 1 || stdout != "" {
		t.Fatalf("missing marker code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	wantErr := "Error: libraries marker not found in main source: add \"/* -- libraries --*/\" to main source\n" + mergeUsage
	if stderr != wantErr {
		t.Fatalf("missing marker stderr\n got %q\nwant %q", stderr, wantErr)
	}
	if _, err := os.Stat(filepath.Join(contest, "b", "submission.cpp")); !os.IsNotExist(err) {
		t.Fatalf("submission.cpp should not exist, stat err=%v", err)
	}

	stdout, stderr, code = runAt(t, contest, "merge", "nope")
	if code != 1 || stdout != "" {
		t.Fatalf("missing dir code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	wantErr = "Error: problem \"nope\" not found in current directory\n" + mergeUsage
	if stderr != wantErr {
		t.Fatalf("missing dir stderr\n got %q\nwant %q", stderr, wantErr)
	}
}

func TestE2ETestPython(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not installed")
	}
	dir := workspace(t)

	writeFile(t, filepath.Join(dir, "a", "main.py"), "print(42)\n")
	writeFile(t, filepath.Join(dir, "a", "test", "sample-1.in"), "\n")
	writeFile(t, filepath.Join(dir, "a", "test", "sample-1.out"), "42\n")
	stdout, stderr, code := runAt(t, dir, "test", "a", "-l", "py")
	if code != 0 || stderr != "" {
		t.Fatalf("ac code=%d stderr=%q stdout=%q", code, stderr, stdout)
	}
	if normalizeDurations(stdout) != sampleAC {
		t.Fatalf("ac stdout\n got %q\nwant %q", normalizeDurations(stdout), sampleAC)
	}
	assertFile(t, filepath.Join(dir, "a", "test", "sample-1.test"), "42\n")

	writeFile(t, filepath.Join(dir, "b", "main.py"), "print(0)\n")
	writeFile(t, filepath.Join(dir, "b", "test", "sample-1.in"), "1\n")
	writeFile(t, filepath.Join(dir, "b", "test", "sample-1.out"), "42\n")
	stdout, stderr, code = runAt(t, dir, "test", "b", "-l", "py")
	if code != 1 {
		t.Fatalf("wa exit %d, want 1\nstdout %q\nstderr %q", code, stdout, stderr)
	}
	if stderr != "Error: samples did not pass: WA\n" {
		t.Fatalf("wa stderr\n got %q\nwant %q", stderr, "Error: samples did not pass: WA\n")
	}
	if normalizeDurations(stdout) != sampleWA {
		t.Fatalf("wa stdout\n got %q\nwant %q", normalizeDurations(stdout), sampleWA)
	}
	assertFile(t, filepath.Join(dir, "b", "test", "sample-1.test"), "0\n")

	writeFile(t, filepath.Join(dir, "empty", "main.py"), "print(1)\n")
	if err := os.MkdirAll(filepath.Join(dir, "empty", "test"), 0o755); err != nil {
		t.Fatal(err)
	}
	stdout, stderr, code = runAt(t, dir, "test", "empty", "-l", "py")
	if code != 1 || stdout != "" {
		t.Fatalf("no samples code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	wantErr := "Error: no input files found in empty/test\n" + testUsage
	if stderr != wantErr {
		t.Fatalf("no samples stderr\n got %q\nwant %q", stderr, wantErr)
	}

	stdout, stderr, code = runAt(t, dir, "test", "nomain", "-l", "py")
	if code != 1 || stdout != "" {
		t.Fatalf("missing dir code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	wantErr = "Error: problem \"nomain\" not found in current directory\n" + testUsage
	if stderr != wantErr {
		t.Fatalf("missing dir stderr\n got %q\nwant %q", stderr, wantErr)
	}

	if err := os.MkdirAll(filepath.Join(dir, "nosrc", "test"), 0o755); err != nil {
		t.Fatal(err)
	}
	stdout, stderr, code = runAt(t, dir, "test", "nosrc", "-l", "py")
	if code != 1 || stdout != "" {
		t.Fatalf("missing main code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	wantErr = "Error: main.py not found: stat nosrc/main.py: no such file or directory\n" + testUsage
	if stderr != wantErr {
		t.Fatalf("missing main stderr\n got %q\nwant %q", stderr, wantErr)
	}
}

func TestE2ETestHonorsDefaultLang(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not installed")
	}
	dir := workspace(t)
	cfgPath := filepath.Join(dir, ".cpx", "config.yaml")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	updated := strings.Replace(string(data), "default_lang: cpp", "default_lang: py", 1)
	if updated == string(data) {
		t.Fatalf("default_lang not rewritten:\n%s", data)
	}
	if err := os.WriteFile(cfgPath, []byte(updated), 0o644); err != nil {
		t.Fatal(err)
	}

	writeFile(t, filepath.Join(dir, "a", "main.py"), "print(42)\n")
	writeFile(t, filepath.Join(dir, "a", "test", "sample-1.in"), "\n")
	writeFile(t, filepath.Join(dir, "a", "test", "sample-1.out"), "42\n")
	stdout, stderr, code := runAt(t, dir, "test", "a")
	if code != 0 || stderr != "" {
		t.Fatalf("code=%d stderr=%q stdout=%q", code, stderr, stdout)
	}
	if normalizeDurations(stdout) != sampleAC {
		t.Fatalf("stdout\n got %q\nwant %q", normalizeDurations(stdout), sampleAC)
	}
}

func TestE2ETestKillsAtTimeLimit(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not installed")
	}
	dir := workspace(t)
	writeFile(t, filepath.Join(dir, "a", "main.py"), "while True:\n    pass\n")
	writeFile(t, filepath.Join(dir, "a", "test", "sample-1.in"), "\n")
	writeFile(t, filepath.Join(dir, "a", "test", "sample-1.out"), "42\n")

	stdout, stderr, code := runAtDeadline(t, 5*time.Second, dir, "test", "a", "-l", "py", "-t", "1")
	if code != 1 {
		t.Fatalf("exit %d, want 1\nstdout %q\nstderr %q", code, stdout, stderr)
	}
	if stderr != "Error: samples did not pass: TLE\n" {
		t.Fatalf("stderr\n got %q\nwant %q", stderr, "Error: samples did not pass: TLE\n")
	}
	if normalizeDurations(stdout) != sampleTLE {
		t.Fatalf("stdout\n got %q\nwant %q", normalizeDurations(stdout), sampleTLE)
	}
	assertFile(t, filepath.Join(dir, "a", "test", "sample-1.test"), "")
}

func TestE2ETestRuntimeErrorContinues(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not installed")
	}
	dir := workspace(t)
	writeFile(t, filepath.Join(dir, "a", "main.py"), "import sys\nn = sys.stdin.read().strip()\nif n == \"bad\":\n    sys.stderr.write(\"boom\")\n    raise SystemExit(3)\nprint(42)\n")
	writeFile(t, filepath.Join(dir, "a", "test", "sample-1.in"), "bad\n")
	writeFile(t, filepath.Join(dir, "a", "test", "sample-1.out"), "0\n")
	writeFile(t, filepath.Join(dir, "a", "test", "sample-2.in"), "ok\n")
	writeFile(t, filepath.Join(dir, "a", "test", "sample-2.out"), "7\n")

	stdout, stderr, code := runAt(t, dir, "test", "a", "-l", "py", "-t", "1")
	if code != 1 {
		t.Fatalf("exit %d, want 1\nstdout %q\nstderr %q", code, stdout, stderr)
	}
	if stderr != "Error: samples did not pass: RE\n" {
		t.Fatalf("stderr\n got %q\nwant %q", stderr, "Error: samples did not pass: RE\n")
	}
	if normalizeDurations(stdout) != sampleRE {
		t.Fatalf("stdout\n got %q\nwant %q", normalizeDurations(stdout), sampleRE)
	}
	assertFile(t, filepath.Join(dir, "a", "test", "sample-1.test"), "")
	assertFile(t, filepath.Join(dir, "a", "test", "sample-2.test"), "42\n")
}

func TestE2ETestCpp(t *testing.T) {
	if _, err := exec.LookPath("g++"); err != nil {
		t.Skip("g++ not installed")
	}
	dir := workspace(t)

	writeFile(t, filepath.Join(dir, "a", "main.cpp"), "#include <iostream>\nint main() {\n  std::cout << 42 << std::endl;\n}\n")
	writeFile(t, filepath.Join(dir, "a", "test", "sample-1.in"), "\n")
	writeFile(t, filepath.Join(dir, "a", "test", "sample-1.out"), "42\n")
	stdout, stderr, code := runAt(t, dir, "test", "a")
	if code != 0 || stderr != "" {
		t.Fatalf("ac code=%d stderr=%q stdout=%q", code, stderr, stdout)
	}
	if normalizeDurations(stdout) != sampleAC {
		t.Fatalf("ac stdout\n got %q\nwant %q", normalizeDurations(stdout), sampleAC)
	}
	assertFile(t, filepath.Join(dir, "a", "test", "sample-1.test"), "42\n")
	if _, err := os.Stat(filepath.Join(dir, "a", "a.out")); err != nil {
		t.Fatalf("a.out: %v", err)
	}

	writeFile(t, filepath.Join(dir, "bad", "main.cpp"), "int main( { }\n")
	writeFile(t, filepath.Join(dir, "bad", "test", "sample-1.in"), "1\n")
	writeFile(t, filepath.Join(dir, "bad", "test", "sample-1.out"), "1\n")
	stdout, stderr, code = runAt(t, dir, "test", "bad")
	if code != 1 || stdout != "" {
		t.Fatalf("compile error code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	if !strings.Contains(stderr, "error:") {
		t.Fatalf("compiler diagnostic missing:\n%s", stderr)
	}
	wantTail := "Error: exit status 1\n" + testUsage
	if !strings.HasSuffix(stderr, wantTail) {
		t.Fatalf("compile error stderr\n got %q\nwant suffix %q", stderr, wantTail)
	}
	if _, err := os.Stat(filepath.Join(dir, "bad", "a.out")); !os.IsNotExist(err) {
		t.Fatalf("a.out should not exist, stat err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "bad", "test", "sample-1.test")); !os.IsNotExist(err) {
		t.Fatalf("sample output should not exist, stat err=%v", err)
	}
}

func TestE2ESubmitStopsOnWA(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not installed")
	}
	dir := workspace(t)
	writeFile(t, filepath.Join(dir, "a", "main.py"), "print(0)\n")
	writeFile(t, filepath.Join(dir, "a", "test", "sample-1.in"), "1\n")
	writeFile(t, filepath.Join(dir, "a", "test", "sample-1.out"), "42\n")

	stdout, stderr, code := runAt(t, dir, "submit", "a", "-l", "py", "--copy")
	if code != 1 {
		t.Fatalf("exit %d, want 1\nstdout %q\nstderr %q", code, stdout, stderr)
	}
	const wantErr = "Error: samples did not pass: WA (use --skip-test to submit anyway)\n"
	if stderr != wantErr {
		t.Fatalf("stderr\n got %q\nwant %q", stderr, wantErr)
	}
	if normalizeDurations(stdout) != sampleWA {
		t.Fatalf("stdout\n got %q\nwant %q", normalizeDurations(stdout), sampleWA)
	}
	if _, err := os.Stat(filepath.Join(dir, "a", "submission.py")); !os.IsNotExist(err) {
		t.Fatalf("submission.py should not exist, stat err=%v", err)
	}
}

func TestE2ESubmitMissingProblem(t *testing.T) {
	dir := workspace(t)
	before := relTree(t, dir)
	stdout, stderr, code := runAt(t, dir, "submit", "nope")
	if code != 1 || stdout != "" {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	wantErr := "Error: problem \"nope\" not found in current directory\n" + submitUsage
	if stderr != wantErr {
		t.Fatalf("stderr\n got %q\nwant %q", stderr, wantErr)
	}
	after := relTree(t, dir)
	if before != after {
		t.Fatalf("workspace changed\n before:\n%s\n after:\n%s", before, after)
	}
}

func resolvedTemp(t *testing.T) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

func workspace(t *testing.T) string {
	t.Helper()
	dir := resolvedTemp(t)
	stdout, stderr, code := runAt(t, dir, "init")
	if code != 0 || stderr != "" {
		t.Fatalf("init code=%d stderr=%q stdout=%q", code, stderr, stdout)
	}
	want := fmt.Sprintf("[INFO] initialized cpx workspace in %s\n", dir)
	if stdout != want {
		t.Fatalf("init stdout\n got %q\nwant %q", stdout, want)
	}
	return dir
}

func runAtDeadline(t *testing.T, limit time.Duration, dir string, args ...string) (string, string, int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), limit)
	defer cancel()
	cmd := exec.CommandContext(ctx, cpxBin, args...)
	cmd.Dir = dir
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		if cmd.Process != nil {
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
		t.Fatalf("cpx %s did not finish within %s\nstdout %q\nstderr %q", strings.Join(args, " "), limit, stdout.String(), stderr.String())
	}
	if err == nil {
		return stdout.String(), stderr.String(), 0
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("cpx %s: %v\nstderr: %s", strings.Join(args, " "), err, stderr.String())
	}
	return stdout.String(), stderr.String(), exitErr.ExitCode()
}

func runAt(t *testing.T, dir string, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(cpxBin, args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		return stdout.String(), stderr.String(), 0
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("cpx %s: %v\nstderr: %s", strings.Join(args, " "), err, stderr.String())
	}
	return stdout.String(), stderr.String(), exitErr.ExitCode()
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertFile(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Fatalf("%s\n got %q\nwant %q", path, got, want)
	}
}

func relTree(t *testing.T, root string) string {
	t.Helper()
	var rels []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rels = append(rels, rel)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(rels)
	return strings.Join(rels, "\n")
}

var durationLine = regexp.MustCompile(`(?m)^(\[INFO\] (?:Execution time|slowest execution time): ).*$`)

func normalizeDurations(s string) string {
	return durationLine.ReplaceAllString(s, "${1}<duration>")
}

const newLibUsage = `Usage:
  cpx new-lib <path> [flags]

Flags:
  -d, --dir int   index into config.library_dirs (default 0)
  -h, --help      help for new-lib

`

const mergeUsage = `Usage:
  cpx merge <problem> [flags]

Flags:
  -h, --help          help for merge
  -l, --lang string   language of the source code (default: default_lang in config, else cpp)

`

const testUsage = `Usage:
  cpx test <problem> [flags]

Flags:
  -h, --help             help for test
  -l, --lang string      language of the source code (default: default_lang in config, else cpp)
  -t, --time-limit int   time limit in seconds for sample tests (default 2)

`

const submitUsage = `Usage:
  cpx submit <problem> [flags]

Flags:
  -c, --copy             copy merged source to clipboard instead of submitting
  -h, --help             help for submit
  -l, --lang string      language of the source code (default: default_lang in config, else cpp)
      --skip-test        skip sample tests before submit or copy
  -t, --time-limit int   time limit in seconds for sample tests (default 2)

`

const sampleAC = `========================================
[INFO] sample-1: AC
[INFO] Execution time: <duration>
[INFO] Expected: 42

[INFO] Actual: 42

========================================
========================================
[INFO] slowest execution time: <duration>
[STATUS] AC
========================================
`

const sampleTLE = `========================================
[INFO] sample-1: TLE
[INFO] Execution time: <duration>
[INFO] Expected: 42

[INFO] Actual: 
========================================
========================================
[INFO] slowest execution time: <duration>
[STATUS] TLE
========================================
`

const sampleRE = `========================================
[INFO] sample-1: RE
[INFO] Execution time: <duration>
[INFO] Expected: 0

[INFO] Actual: 
[INFO] Stderr: boom
========================================
========================================
[INFO] sample-2: WA
[INFO] Execution time: <duration>
[INFO] Expected: 7

[INFO] Actual: 42

========================================
========================================
[INFO] slowest execution time: <duration>
[STATUS] RE
========================================
`

const sampleWA = `========================================
[INFO] sample-1: WA
[INFO] Execution time: <duration>
[INFO] Expected: 42

[INFO] Actual: 0

========================================
========================================
[INFO] slowest execution time: <duration>
[STATUS] WA
========================================
`

const sourceTemplate = `#include <iostream>

using namespace std;

/* -- libraries --*/


void solve() {

}

int main() {
	cin.tie(0);
	ios::sync_with_stdio(false);
	solve();
	return 0;
}
`

const libraryTemplate = `#pragma once

/* -- library code --*/


`

const utilLibrary = `#pragma once
#include <vector>
/* -- library code --*/
class Util {};
/* -- library code --*/
`

const widgetLibrary = `#pragma once
#include <iostream>
#include "util.hpp"
/* -- library code --*/
class Test {};
/* -- library code --*/
`

const onlyLibrary = `#pragma once
/* -- library code --*/
class Only {};
/* -- library code --*/
`

const mergeMain = `#include <iostream>
#include <vector>
#include "widget.hpp"

/* -- libraries --*/

int main() {
	return 0;
}
`

const mergedSubmission = `#include <iostream>
#include <vector>

/* -- libraries --*/
class Util {};
class Test {};

int main() {
	return 0;
}

`
