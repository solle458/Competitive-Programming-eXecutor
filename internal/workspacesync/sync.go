package workspacesync

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"Competitive-Programming-eXecutor/internal/setup/atcoder"
)

const defaultContestsURL = "https://kenkoooo.com/atcoder/resources/contests.json"

var (
	contestIDPattern    = regexp.MustCompile(`^[a-z]+[0-9]+$`)
	problemIndexPattern = regexp.MustCompile(`^[a-z]{1,2}$`)
)

func Sync(workspace string) error {
	root, ok := findGitRoot(workspace)
	if !ok {
		return nil
	}
	if _, err := exec.LookPath("jj"); err != nil {
		return errors.New("jj not found on PATH")
	}
	if _, err := os.Stat(filepath.Join(root, ".jj")); err != nil {
		if _, err := jj(root, nil, "git", "init", "--colocate"); err != nil {
			return err
		}
	}
	if _, err := jj(root, nil, "git", "fetch"); err != nil {
		return err
	}

	from := "main@origin"
	if !revExists(root, from) {
		from = "root()"
	}
	diffOut, err := jj(root, nil, "diff", "--from", from, "--name-only")
	if err != nil {
		return err
	}
	paths := nonEmptyLines(diffOut)
	ids := contestIDs(paths)
	if len(ids) > 0 {
		byID, err := loadContests()
		if err != nil {
			fmt.Printf("[INFO] skip sync: contest schedule: %v\n", err)
			return nil
		}
		if reason := blockReason(ids, byID, time.Now()); reason != "" {
			fmt.Printf("[INFO] skip sync: %s\n", reason)
			return nil
		}
	}
	if len(paths) == 0 {
		return nil
	}

	if from == "main@origin" {
		if _, err := jj(root, nil, "bookmark", "track", "main@origin"); err != nil {
			return err
		}
		if _, err := jj(root, nil, "rebase", "-d", "main@origin"); err != nil {
			return err
		}
		conflict, err := jj(root, nil, "log", "-r", "@", "--no-graph", "-T", "conflict")
		if err != nil {
			return err
		}
		if strings.TrimSpace(conflict) == "true" {
			return errors.New("jj rebase onto main@origin left a conflict")
		}
	}

	name, email, err := gitIdentity(root)
	if err != nil {
		return err
	}
	env := []string{"JJ_USER=" + name, "JJ_EMAIL=" + email}
	desc, err := jj(root, env, "log", "-r", "@", "--no-graph", "-T", "description")
	if err != nil {
		return err
	}
	if strings.TrimSpace(desc) == "" {
		if _, err := jj(root, env, "describe", "-m", "cpx sync"); err != nil {
			return err
		}
	}
	if _, err := jj(root, env, "metaedit", "--update-author", "-r", "@"); err != nil {
		return err
	}
	if _, err := jj(root, env, "bookmark", "set", "main", "-r", "@"); err != nil {
		return err
	}
	if _, err := jj(root, env, "git", "push", "--bookmark", "main"); err != nil {
		return err
	}
	fmt.Printf("[INFO] synced workspace to origin/main\n")
	return nil
}

func contestIDs(paths []string) []string {
	seen := map[string]struct{}{}
	var ids []string
	for _, path := range paths {
		parts := strings.Split(filepath.ToSlash(path), "/")
		if len(parts) < 2 {
			continue
		}
		id := strings.ToLower(parts[0])
		index := strings.ToLower(parts[1])
		if !contestIDPattern.MatchString(id) || !problemIndexPattern.MatchString(index) {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func blockReason(ids []string, byID map[string]atcoder.Contest, now time.Time) string {
	var live []string
	var missing []string
	for _, id := range ids {
		contest, ok := byID[id]
		if !ok {
			missing = append(missing, id)
			continue
		}
		if contest.InProgress(now) {
			live = append(live, id)
		}
	}
	sort.Strings(live)
	sort.Strings(missing)
	if len(live) == 1 {
		return "contest " + live[0] + " is in progress"
	}
	if len(live) > 1 {
		return "contests " + strings.Join(live, ", ") + " are in progress"
	}
	if len(missing) == 1 {
		return "contest " + missing[0] + " is missing from the contest schedule"
	}
	if len(missing) > 1 {
		return "contests " + strings.Join(missing, ", ") + " are missing from the contest schedule"
	}
	return ""
}

func loadContests() (map[string]atcoder.Contest, error) {
	url := os.Getenv("CPX_CONTESTS_URL")
	if url == "" {
		url = defaultContestsURL
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "cpx")
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	var contests []atcoder.Contest
	if err := json.Unmarshal(body, &contests); err != nil {
		return nil, err
	}
	byID := make(map[string]atcoder.Contest, len(contests))
	for _, contest := range contests {
		byID[strings.ToLower(contest.ID)] = contest
	}
	return byID, nil
}

func findGitRoot(start string) (string, bool) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", false
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

func revExists(dir, rev string) bool {
	_, err := jj(dir, nil, "log", "-r", rev, "--no-graph", "-T", "commit_id")
	return err == nil
}

func gitIdentity(dir string) (string, string, error) {
	name, err := gitConfig(dir, "user.name")
	if err != nil {
		return "", "", err
	}
	email, err := gitConfig(dir, "user.email")
	if err != nil {
		return "", "", err
	}
	if name == "" || email == "" {
		return "", "", errors.New("git user.name and user.email are required to sync to origin/main")
	}
	return name, email, nil
}

func gitConfig(dir, key string) (string, error) {
	cmd := exec.Command("git", "config", "--get", key)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func jj(dir string, extraEnv []string, args ...string) (string, error) {
	cmd := exec.Command("jj", append([]string{"--color=never"}, args...)...)
	cmd.Dir = dir
	if len(extraEnv) > 0 {
		cmd.Env = append(os.Environ(), extraEnv...)
	}
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	if err != nil {
		return buf.String(), fmt.Errorf("jj %s: %w\n%s", strings.Join(args, " "), err, buf.String())
	}
	return buf.String(), nil
}

func nonEmptyLines(s string) []string {
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}
