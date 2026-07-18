package submit

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"Competitive-Programming-eXecutor/internal/submit/ojpatch"
)

var (
	errSessionRequired = errors.New("atcoder session is required")
	errOJNotFound      = errors.New("oj command not found")
	errOJPython        = errors.New("python with onlinejudge not found")
)

func taskURL(problemPath string) (string, error) {
	contestID, index, err := resolveContestAndIndex(problemPath)
	if err != nil {
		return "", err
	}
	problemID := fmt.Sprintf("%s_%s", contestID, index)
	return fmt.Sprintf("https://atcoder.jp/contests/%s/tasks/%s", contestID, problemID), nil
}

func resolveContestAndIndex(problemPath string) (string, string, error) {
	normalized := filepath.ToSlash(problemPath)
	parts := strings.Split(normalized, "/")
	parts = removeEmpty(parts)

	switch len(parts) {
	case 0:
		return "", "", fmt.Errorf("invalid problem path %q", problemPath)
	case 1:
		cwd, err := os.Getwd()
		if err != nil {
			return "", "", err
		}
		return strings.ToLower(filepath.Base(cwd)), strings.ToLower(parts[0]), nil
	default:
		return strings.ToLower(parts[len(parts)-2]), strings.ToLower(parts[len(parts)-1]), nil
	}
}

func removeEmpty(parts []string) []string {
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" && part != "." {
			result = append(result, part)
		}
	}
	return result
}

func ojCookiePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "online-judge-tools", "cookie.jar"), nil
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", errors.New("APPDATA is not set")
		}
		return filepath.Join(appData, "online-judge-tools", "cookie.jar"), nil
	default:
		return filepath.Join(home, ".local", "share", "online-judge-tools", "cookie.jar"), nil
	}
}

func syncOJCookie(session string) error {
	session = strings.TrimSpace(session)
	if session == "" {
		return fmt.Errorf("%w: set atcoder_session in .cpx/config.yaml or ATCODER_SESSION env var (use aclogin or browser DevTools)", errSessionRequired)
	}

	cookiePath, err := ojCookiePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(cookiePath), 0o700); err != nil {
		return fmt.Errorf("create oj cookie directory: %w", err)
	}

	content := fmt.Sprintf(`#LWP-Cookies-2.0
Set-Cookie3: REVEL_FLASH=""; path="/"; domain="atcoder.jp"; path_spec; secure; discard; HttpOnly=None; version=0
Set-Cookie3: REVEL_SESSION=%q; path="/"; domain="atcoder.jp"; path_spec; secure; discard; HttpOnly=None; version=0
`, session)
	if err := os.WriteFile(cookiePath, []byte(content), 0o600); err != nil {
		return fmt.Errorf("write oj cookie %q: %w", cookiePath, err)
	}
	return nil
}

// resolveOJPython finds a Python interpreter that can import onlinejudge.
// Prefer the interpreter next to (or shebang of) the oj executable so uv-tool
// installs work; fall back to python3 on PATH.
func resolveOJPython() (string, error) {
	ojPath, err := exec.LookPath("oj")
	if err != nil {
		return "", fmt.Errorf("%w: install online-judge-tools (pip install online-judge-tools)", errOJNotFound)
	}

	candidates := make([]string, 0, 4)
	if resolved, err := filepath.EvalSymlinks(ojPath); err == nil {
		ojPath = resolved
	}
	if shebang := pythonFromShebang(ojPath); shebang != "" {
		candidates = append(candidates, shebang)
	}
	ojDir := filepath.Dir(ojPath)
	candidates = append(candidates,
		filepath.Join(ojDir, "python"),
		filepath.Join(ojDir, "python3"),
	)
	if p, err := exec.LookPath("python3"); err == nil {
		candidates = append(candidates, p)
	}

	seen := make(map[string]struct{})
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}
		if canImportOnlineJudge(candidate) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("%w: ensure online-judge-tools is installed for the oj Python env", errOJPython)
}

func pythonFromShebang(scriptPath string) string {
	data, err := os.ReadFile(scriptPath)
	if err != nil || len(data) < 2 || data[0] != '#' || data[1] != '!' {
		return ""
	}
	line := string(data)
	if i := strings.IndexByte(line, '\n'); i >= 0 {
		line = line[:i]
	}
	line = strings.TrimSpace(strings.TrimPrefix(line, "#!"))
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return ""
	}
	// #!/usr/bin/env python3 → second field
	if filepath.Base(fields[0]) == "env" && len(fields) >= 2 {
		if p, err := exec.LookPath(fields[1]); err == nil {
			return p
		}
		return fields[1]
	}
	return fields[0]
}

func canImportOnlineJudge(python string) bool {
	cmd := exec.Command(python, "-c", "import onlinejudge, onlinejudge_command")
	return cmd.Run() == nil
}

func submitWithOJ(taskURL, submissionPath, session string) error {
	python, err := resolveOJPython()
	if err != nil {
		return err
	}
	if err := syncOJCookie(session); err != nil {
		return err
	}

	cacheDir, err := ojpatch.CacheDir()
	if err != nil {
		return fmt.Errorf("ojpatch cache dir: %w", err)
	}
	submitScript, err := ojpatch.Materialize(cacheDir)
	if err != nil {
		return err
	}

	cmd := exec.Command(python, submitScript, "--yes", taskURL, submissionPath)
	var buf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &buf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &buf)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w", wrapOJSubmitError(err, buf.String()))
	}
	return nil
}

func wrapOJSubmitError(err error, output string) error {
	if strings.Contains(output, "parsed_memory_limit") || strings.Contains(output, "Memory limit regex did not match") {
		return fmt.Errorf("oj submit failed: %w (AtCoder memory-limit parse failed even with MiB patch; try cpx submit --copy)", err)
	}
	return fmt.Errorf("oj submit failed: %w (need a valid REVEL_SESSION and an ongoing contest; for past problems use cpx submit --copy)", err)
}
