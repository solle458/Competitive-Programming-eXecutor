package test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"Competitive-Programming-eXecutor/internal/config"
	"Competitive-Programming-eXecutor/internal/problem"
)

type Verdict int

const (
	AC Verdict = iota
	TLE
	WA
	RE
)

func (v Verdict) String() string {
	switch v {
	case AC:
		return "AC"
	case TLE:
		return "TLE"
	case WA:
		return "WA"
	case RE:
		return "RE"
	default:
		return fmt.Sprintf("Verdict(%d)", int(v))
	}
}

type CaseResult struct {
	Name     string
	Verdict  Verdict
	Duration time.Duration
	Expected string
	Actual   string
	Stderr   string
}

var ErrNotAccepted = errors.New("samples did not pass")

var languages = map[string]func(problem.Problem, *config.Config) ([]string, error){
	"cpp": buildCpp,
	"py":  buildPy,
}

func RunSamples(p problem.Problem, timeLimit int, cfg *config.Config) error {
	argv, err := build(p, cfg)
	if err != nil {
		return err
	}
	results, err := runCases(p, argv, timeLimit)
	if err != nil {
		return err
	}
	worst := printResults(results)
	if worst != AC {
		return fmt.Errorf("%w: %s", ErrNotAccepted, worst)
	}
	return nil
}

func build(p problem.Problem, cfg *config.Config) ([]string, error) {
	fn, ok := languages[p.Lang]
	if !ok {
		return nil, fmt.Errorf("unsupported language %q (supported: cpp, py)", p.Lang)
	}
	return fn(p, cfg)
}

func buildPy(p problem.Problem, _ *config.Config) ([]string, error) {
	if _, err := os.Stat(p.Source()); err != nil {
		return nil, fmt.Errorf("main.py not found: %w", err)
	}
	return []string{"python3", p.Source()}, nil
}

func buildCpp(p problem.Problem, cfg *config.Config) ([]string, error) {
	outPath := p.Binary()
	args := []string{"-std=c++20", "-O3"}
	for _, dir := range cfg.File.LibraryDirs {
		args = append(args, "-I", dir)
	}
	args = append(args, "-o", outPath, p.Source())
	cmd := exec.Command("g++", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return []string{outPath}, nil
}

type runFinish int

const (
	ran runFinish = iota
	timedOut
	crashed
)

type sampleRun struct {
	name     string
	actual   string
	stderr   string
	duration time.Duration
	finish   runFinish
}

func runCases(p problem.Problem, argv []string, timeLimit int) ([]CaseResult, error) {
	testDir := p.SampleDir()
	inputs, err := inputFiles(testDir)
	if err != nil {
		return nil, err
	}
	if len(inputs) == 0 {
		return nil, fmt.Errorf("no input files found in %s", testDir)
	}

	limit := time.Duration(timeLimit) * time.Second
	runs := make([]sampleRun, 0, len(inputs))
	for _, inputFile := range inputs {
		stem := strings.TrimSuffix(filepath.Base(inputFile), ".in")
		actual, stderr, duration, finish, err := runSample(argv, inputFile, limit)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", stem, err)
		}
		testPath := filepath.Join(testDir, stem+".test")
		if err := os.WriteFile(testPath, []byte(actual), 0o644); err != nil {
			return nil, err
		}
		runs = append(runs, sampleRun{
			name:     stem,
			actual:   actual,
			stderr:   stderr,
			duration: duration,
			finish:   finish,
		})
	}

	results := make([]CaseResult, 0, len(runs))
	for _, run := range runs {
		expected, err := os.ReadFile(filepath.Join(testDir, run.name+".out"))
		if err != nil {
			return nil, err
		}
		results = append(results, CaseResult{
			Name:     run.name,
			Verdict:  judge(string(expected), run.actual, run.duration, limit, run.finish),
			Duration: run.duration,
			Expected: string(expected),
			Actual:   run.actual,
			Stderr:   run.stderr,
		})
	}
	return results, nil
}

func runSample(argv []string, inputPath string, limit time.Duration) (string, string, time.Duration, runFinish, error) {
	in, err := os.Open(inputPath)
	if err != nil {
		return "", "", 0, ran, err
	}
	defer in.Close()

	ctx, cancel := context.WithTimeout(context.Background(), limit)
	defer cancel()
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Stdin = in
	cmd.WaitDelay = time.Second
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	runErr := cmd.Run()
	duration := time.Since(start)
	actual := stdout.String()
	errText := stderr.String()
	if runErr == nil {
		return actual, errText, duration, ran, nil
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return actual, errText, duration, timedOut, nil
	}
	var exitErr *exec.ExitError
	if errors.As(runErr, &exitErr) {
		return actual, errText, duration, crashed, nil
	}
	return actual, errText, duration, ran, runErr
}

func judge(expected, actual string, duration, limit time.Duration, finish runFinish) Verdict {
	switch finish {
	case timedOut:
		return TLE
	case crashed:
		return RE
	default:
		if strings.TrimSpace(actual) != strings.TrimSpace(expected) {
			return WA
		}
		if duration > limit {
			return TLE
		}
		return AC
	}
}

func stderrTail(s string) string {
	const max = 1024
	if len(s) <= max {
		return s
	}
	return s[len(s)-max:]
}

func printResults(results []CaseResult) Verdict {
	slowest := time.Duration(0)
	worst := AC
	for _, r := range results {
		slowest = max(slowest, r.Duration)
		if r.Verdict > worst {
			worst = r.Verdict
		}
		fmt.Println("========================================")
		fmt.Printf("[INFO] %s: %s\n", r.Name, r.Verdict)
		fmt.Printf("[INFO] Execution time: %s\n", r.Duration)
		fmt.Printf("[INFO] Expected: %s\n", r.Expected)
		fmt.Printf("[INFO] Actual: %s\n", r.Actual)
		if r.Verdict == RE {
			fmt.Printf("[INFO] Stderr: %s\n", stderrTail(r.Stderr))
		}
		fmt.Println("========================================")
	}
	fmt.Println("========================================")
	fmt.Printf("[INFO] slowest execution time: %s\n", slowest.String())
	fmt.Printf("[STATUS] %s\n", worst)
	fmt.Println("========================================")
	return worst
}

func inputFiles(testDir string) ([]string, error) {
	files, err := os.ReadDir(testDir)
	if err != nil {
		return nil, err
	}
	var inputs []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasSuffix(file.Name(), ".in") {
			inputs = append(inputs, filepath.Join(testDir, file.Name()))
		}
	}
	sort.Slice(inputs, func(i, j int) bool {
		return naturalLess(
			strings.TrimSuffix(filepath.Base(inputs[i]), ".in"),
			strings.TrimSuffix(filepath.Base(inputs[j]), ".in"),
		)
	})
	return inputs, nil
}

func naturalLess(a, b string) bool {
	aPrefix, aNum, aHasNum := splitNumericSuffix(a)
	bPrefix, bNum, bHasNum := splitNumericSuffix(b)
	if aPrefix != bPrefix {
		return aPrefix < bPrefix
	}
	if aHasNum && bHasNum {
		return aNum < bNum
	}
	return a < b
}

func splitNumericSuffix(s string) (string, int, bool) {
	i := len(s)
	for i > 0 && s[i-1] >= '0' && s[i-1] <= '9' {
		i--
	}
	if i == len(s) {
		return s, 0, false
	}
	n, _ := strconv.Atoi(s[i:])
	return s[:i], n, true
}
