package test

import (
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

type language struct {
	compile func(problem.Problem, *config.Config) (string, error)
	argv    func(problem.Problem, string) []string
}

var languages = map[string]language{
	"cpp": {compile: compileCpp, argv: cppArgv},
	"py":  {compile: compilePy, argv: pyArgv},
}

func RunSamples(p problem.Problem, timeLimit int, cfg *config.Config) error {
	lang, err := languageByName(p.Lang)
	if err != nil {
		return err
	}
	executable, err := lang.compile(p, cfg)
	if err != nil {
		return err
	}
	executionTimes, err := runCases(p, lang.argv(p, executable))
	if err != nil {
		return err
	}
	return compare(p, executionTimes, timeLimit)
}

func languageByName(name string) (language, error) {
	lang, ok := languages[name]
	if !ok {
		return language{}, fmt.Errorf("unsupported language %q (supported: cpp, py)", name)
	}
	return lang, nil
}

func compilePy(p problem.Problem, _ *config.Config) (string, error) {
	if _, err := os.Stat(p.Source()); err != nil {
		return "", fmt.Errorf("main.py not found: %w", err)
	}
	return "", nil
}

func pyArgv(p problem.Problem, _ string) []string {
	return []string{"python3", p.Source()}
}

func compileCpp(p problem.Problem, cfg *config.Config) (string, error) {
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
		return "", err
	}
	return outPath, nil
}

func cppArgv(_ problem.Problem, executable string) []string {
	return []string{executable}
}

func runCases(p problem.Problem, argv []string) (map[string]time.Duration, error) {
	testDir := p.SampleDir()
	inputFiles, err := inputFiles(testDir)
	if err != nil {
		return nil, err
	}
	if len(inputFiles) == 0 {
		return nil, fmt.Errorf("no input files found in %s", testDir)
	}

	executionTimes := make(map[string]time.Duration, len(inputFiles))
	for _, inputFile := range inputFiles {
		stem := strings.TrimSuffix(filepath.Base(inputFile), ".in")
		cmd := exec.Command(argv[0], argv[1:]...)

		in, err := os.Open(inputFile)
		if err != nil {
			return nil, err
		}
		cmd.Stdin = in

		start := time.Now()
		output, err := cmd.Output()
		in.Close()
		executionTimes[stem] = time.Since(start)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", stem, err)
		}

		testPath := filepath.Join(testDir, stem+".test")
		if err := os.WriteFile(testPath, output, 0o644); err != nil {
			return nil, err
		}
	}
	return executionTimes, nil
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

func compare(p problem.Problem, executionTimes map[string]time.Duration, timeLimit int) error {
	testDir := p.SampleDir()
	inputs, err := inputFiles(testDir)
	if err != nil {
		return err
	}
	if len(inputs) == 0 {
		return fmt.Errorf("no input files found in %s", testDir)
	}

	status := "AC"
	slowestExecutionTime := time.Duration(0)
	timeLimitDuration := time.Duration(timeLimit) * time.Second

	for _, inputFile := range inputs {
		stem := strings.TrimSuffix(filepath.Base(inputFile), ".in")
		testPath := filepath.Join(testDir, stem+".test")
		outPath := filepath.Join(testDir, stem+".out")

		actual, err := os.ReadFile(testPath)
		if err != nil {
			return err
		}
		expected, err := os.ReadFile(outPath)
		if err != nil {
			return err
		}

		executionTime, ok := executionTimes[stem]
		if !ok {
			return fmt.Errorf("execution time not found for %s", stem)
		}
		slowestExecutionTime = max(slowestExecutionTime, executionTime)

		caseStatus := "AC"
		if strings.TrimSpace(string(actual)) != strings.TrimSpace(string(expected)) {
			caseStatus = "WA"
		} else if executionTime > timeLimitDuration {
			caseStatus = "TLE"
		}

		fmt.Println("========================================")
		fmt.Printf("[INFO] %s: %s\n", stem, caseStatus)
		fmt.Printf("[INFO] Execution time: %s\n", executionTime)
		fmt.Printf("[INFO] Expected: %s\n", string(expected))
		fmt.Printf("[INFO] Actual: %s\n", string(actual))
		fmt.Println("========================================")

		status = worseStatus(status, caseStatus)
	}

	fmt.Println("========================================")
	fmt.Printf("[INFO] slowest execution time: %s\n", slowestExecutionTime.String())
	fmt.Printf("[STATUS] %s\n", status)
	fmt.Println("========================================")
	return nil
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

func worseStatus(current, newStatus string) string {
	priority := map[string]int{
		"AC":  0,
		"TLE": 1,
		"WA":  2,
	}
	if priority[newStatus] > priority[current] {
		return newStatus
	}
	return current
}
