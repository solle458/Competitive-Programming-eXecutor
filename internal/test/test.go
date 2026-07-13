package test

import (
	"Competitive-Programming-eXecutor/internal/config"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

var errTestFailed = errors.New("test failed")

func Compile(problemID string, lang string, config *config.Config) (string, error) {
	switch lang {
	case "py":
		mainPath := filepath.Join(problemID, "main.py")
		if _, err := os.Stat(mainPath); err != nil {
			return "", fmt.Errorf("main.py not found: %w", err)
		}
		return "", nil
	default:
		outPath := filepath.Join(problemID, "a.out")
		args := []string{"-std=c++20", "-O3"}
		for _, dir := range config.File.LibraryDirs {
			args = append(args, "-I", dir)
		}
		args = append(args, "-o", outPath, filepath.Join(problemID, "main.cpp"))
		cmd := exec.Command("g++", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return "", err
		}
		return outPath, nil
	}
}

func Run(problemID string, executableFilePath string, lang string) (map[string]time.Duration, error) {
	testDir := filepath.Join(problemID, "test")
	inputFiles, err := GetInputFiles(testDir)
	if err != nil {
		return nil, err
	}
	if len(inputFiles) == 0 {
		return nil, fmt.Errorf("no input files found in %s", testDir)
	}

	executionTimes := make(map[string]time.Duration, len(inputFiles))
	for _, inputFile := range inputFiles {
		stem := strings.TrimSuffix(filepath.Base(inputFile), ".in")

		var cmd *exec.Cmd
		switch lang {
		case "py":
			cmd = exec.Command("python3", filepath.Join(problemID, "main.py"))
		default:
			cmd = exec.Command(executableFilePath)
		}

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

func GetInputFiles(testDir string) ([]string, error) {
	files, err := os.ReadDir(testDir)
	if err != nil {
		return nil, err
	}
	var inputFiles []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasSuffix(file.Name(), ".in") {
			inputFiles = append(inputFiles, filepath.Join(testDir, file.Name()))
		}
	}
	sort.Slice(inputFiles, func(i, j int) bool {
		return naturalLess(
			strings.TrimSuffix(filepath.Base(inputFiles[i]), ".in"),
			strings.TrimSuffix(filepath.Base(inputFiles[j]), ".in"),
		)
	})
	return inputFiles, nil
}

func Compare(problemID string, executionTimes map[string]time.Duration, timeLimit int) error {
	testDir := filepath.Join(problemID, "test")
	inputFiles, err := GetInputFiles(testDir)
	if err != nil {
		return err
	}
	if len(inputFiles) == 0 {
		return fmt.Errorf("no input files found in %s", testDir)
	}

	status := "AC"
	slowestExecutionTime := time.Duration(0)
	timeLimitDuration := time.Duration(timeLimit) * time.Second

	for _, inputFile := range inputFiles {
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
