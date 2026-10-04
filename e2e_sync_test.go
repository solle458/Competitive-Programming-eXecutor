package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestE2ESyncPastTestPushes(t *testing.T) {
	requireJJ(t)
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not installed")
	}
	ws, origin := gitWorkspace(t)
	now := time.Now().Unix()
	url := serveContests(t, fmt.Sprintf(`[{"id":"abc100","start_epoch_second":%d,"duration_second":100,"title":"Past","rate_change":"-"}]`, now-10_000))
	initGitWorkspace(t, ws, url)

	writePySample(t, ws, "abc100/a", "print(42)\n", "42\n")
	stdout, stderr, code := runAtEnv(t, ws, []string{"CPX_CONTESTS_URL=" + url}, "test", "abc100/a", "-l", "py")
	if code != 0 || stderr != "" {
		t.Fatalf("code=%d stderr=%q stdout=%q", code, stderr, stdout)
	}
	if !strings.Contains(stdout, "[INFO] synced workspace to origin/main\n") {
		t.Fatalf("stdout %q", stdout)
	}
	files := originFiles(t, origin)
	if !strings.Contains(files, "abc100/a/main.py\n") {
		t.Fatalf("origin/main missing solution:\n%s", files)
	}
}

func TestE2ESyncInProgressSkips(t *testing.T) {
	requireJJ(t)
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not installed")
	}
	ws, origin := gitWorkspace(t)
	now := time.Now().Unix()
	url := serveContests(t, fmt.Sprintf(`[{"id":"abc100","start_epoch_second":%d,"duration_second":7200,"title":"Live","rate_change":"-"}]`, now-60))
	initGitWorkspace(t, ws, url)
	before := originFiles(t, origin)

	writePySample(t, ws, "abc100/a", "print(42)\n", "42\n")
	stdout, stderr, code := runAtEnv(t, ws, []string{"CPX_CONTESTS_URL=" + url}, "test", "abc100/a", "-l", "py")
	if code != 0 || stderr != "" {
		t.Fatalf("code=%d stderr=%q stdout=%q", code, stderr, stdout)
	}
	if !strings.Contains(stdout, "[INFO] skip sync: contest abc100 is in progress\n") {
		t.Fatalf("stdout %q", stdout)
	}
	if strings.Contains(stdout, "synced workspace to origin/main") {
		t.Fatalf("pushed during contest\nstdout %q", stdout)
	}
	after := originFiles(t, origin)
	if after != before {
		t.Fatalf("origin/main changed\n before:\n%s\n after:\n%s", before, after)
	}
	if strings.Contains(after, "abc100/a/main.py") {
		t.Fatalf("in-progress solution is on origin/main:\n%s", after)
	}
}

func TestE2ESyncFailedTestDoesNotPush(t *testing.T) {
	requireJJ(t)
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not installed")
	}
	ws, origin := gitWorkspace(t)
	now := time.Now().Unix()
	url := serveContests(t, fmt.Sprintf(`[{"id":"abc100","start_epoch_second":%d,"duration_second":100,"title":"Past","rate_change":"-"}]`, now-10_000))
	initGitWorkspace(t, ws, url)
	before := originFiles(t, origin)

	writePySample(t, ws, "abc100/a", "print(0)\n", "42\n")
	stdout, stderr, code := runAtEnv(t, ws, []string{"CPX_CONTESTS_URL=" + url}, "test", "abc100/a", "-l", "py")
	if code != 1 {
		t.Fatalf("exit %d, want 1\nstdout %q\nstderr %q", code, stdout, stderr)
	}
	if strings.Contains(stdout, "synced workspace to origin/main") || strings.Contains(stdout, "skip sync") {
		t.Fatalf("stdout %q", stdout)
	}
	after := originFiles(t, origin)
	if after != before {
		t.Fatalf("origin/main changed after a failed test\n before:\n%s\n after:\n%s", before, after)
	}
}

func TestE2ESyncLibraryDuringLiveContestPushes(t *testing.T) {
	requireJJ(t)
	ws, origin := gitWorkspace(t)
	now := time.Now().Unix()
	url := serveContests(t, fmt.Sprintf(`[{"id":"abc100","start_epoch_second":%d,"duration_second":7200,"title":"Live","rate_change":"-"}]`, now-60))
	initGitWorkspace(t, ws, url)

	stdout, stderr, code := runAtEnv(t, ws, []string{"CPX_CONTESTS_URL=" + url}, "new-lib", "graph/dsu.hpp")
	if code != 0 || stderr != "" {
		t.Fatalf("code=%d stderr=%q stdout=%q", code, stderr, stdout)
	}
	if !strings.Contains(stdout, "[INFO] synced workspace to origin/main\n") {
		t.Fatalf("stdout %q", stdout)
	}
	files := originFiles(t, origin)
	if !strings.Contains(files, "library/graph/dsu.hpp\n") {
		t.Fatalf("origin/main missing library:\n%s", files)
	}
}

func requireJJ(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj not installed")
	}
}

func gitWorkspace(t *testing.T) (string, string) {
	t.Helper()
	base := resolvedTemp(t)
	origin := filepath.Join(base, "origin.git")
	ws := filepath.Join(base, "ws")
	runGit(t, base, "init", "--bare", "-b", "main", origin)
	runGit(t, base, "clone", origin, ws)
	runGit(t, ws, "config", "user.email", "cpx-test@example.com")
	runGit(t, ws, "config", "user.name", "cpx-test")
	writeFile(t, filepath.Join(ws, "README.md"), "seed\n")
	runGit(t, ws, "add", "README.md")
	runGit(t, ws, "commit", "-m", "seed")
	runGit(t, ws, "push", "origin", "main")
	return ws, origin
}

func initGitWorkspace(t *testing.T, ws, contestsURL string) {
	t.Helper()
	stdout, stderr, code := runAtEnv(t, ws, []string{"CPX_CONTESTS_URL=" + contestsURL}, "init")
	if code != 0 || stderr != "" {
		t.Fatalf("init code=%d stderr=%q stdout=%q", code, stderr, stdout)
	}
	if !strings.Contains(stdout, "[INFO] initialized cpx workspace in "+ws+"\n") {
		t.Fatalf("init stdout %q", stdout)
	}
	if !strings.Contains(stdout, "[INFO] synced workspace to origin/main\n") {
		t.Fatalf("init did not sync\nstdout %q", stdout)
	}
}

func writePySample(t *testing.T, ws, problem, source, output string) {
	t.Helper()
	writeFile(t, filepath.Join(ws, problem, "main.py"), source)
	writeFile(t, filepath.Join(ws, problem, "test", "sample-1.in"), "\n")
	writeFile(t, filepath.Join(ws, problem, "test", "sample-1.out"), output)
}

func serveContests(t *testing.T, body string) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

func originFiles(t *testing.T, origin string) string {
	t.Helper()
	cmd := exec.Command("git", "--git-dir", origin, "ls-tree", "-r", "--name-only", "main")
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

func runAtEnv(t *testing.T, dir string, env []string, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(cpxBin, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
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
