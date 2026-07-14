package submit

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	errSessionRequired = errors.New("atcoder session is required")
	errOJNotFound      = errors.New("oj command not found")
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

func submitWithOJ(taskURL, submissionPath, session string) error {
	if _, err := exec.LookPath("oj"); err != nil {
		return fmt.Errorf("%w: install online-judge-tools (pip install online-judge-tools)", errOJNotFound)
	}
	if err := syncOJCookie(session); err != nil {
		return err
	}

	cmd := exec.Command("oj", "submit", taskURL, submissionPath, "--yes")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("oj submit failed: %w (submit works only for ongoing contests with a valid REVEL_SESSION; for past problems use the browser)", err)
	}
	return nil
}
