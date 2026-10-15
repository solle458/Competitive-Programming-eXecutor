package merge

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"Competitive-Programming-eXecutor/internal/problem"

	"github.com/gammazero/deque"
)

// LibraryMarker はコンテスト解答の挿入位置。
// CodeMarker は公開ヘッダに残っている本体区切り。提出ではその行を落とす。
const (
	LibraryMarker = "/* -- libraries --*/"
	CodeMarker    = "/* -- library code --*/"
)

var ErrLibrariesMarkerMissing = errors.New("libraries marker not found in main source")

func WriteSubmission(p problem.Problem, libraryDirs []string) (string, error) {
	sourcePath := p.Source()
	content, err := os.ReadFile(sourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("source code not found: %q", sourcePath)
		}
		return "", fmt.Errorf("read source code %q: %w", sourcePath, err)
	}

	submissionCode, err := Generate(string(content), libraryDirs)
	if err != nil {
		return "", err
	}

	submissionPath := p.Submission()
	if err := os.WriteFile(submissionPath, []byte(submissionCode), 0o644); err != nil {
		return "", fmt.Errorf("write submission file %q: %w", submissionPath, err)
	}
	return submissionPath, nil
}

func Generate(sourceCode string, libraryDirs []string) (string, error) {
	lines := strings.Split(sourceCode, "\n")
	var queue deque.Deque[string]
	seen := make(map[string]struct{})
	var stdIncludes []string

	extractInclude := func(line string) bool {
		include, ok := parseInclude(line)
		if !ok {
			return false
		}
		if _, exists := seen[include]; !exists {
			if strings.HasSuffix(include, ".hpp") {
				queue.PushBack(include)
			} else {
				stdIncludes = append(stdIncludes, include)
			}
		}
		seen[include] = struct{}{}
		return true
	}

	for _, line := range lines {
		extractInclude(line)
	}

	mergedLibraries := deque.Deque[string]{}
	for queue.Len() > 0 {
		include := queue.PopFront()
		libraryLines, err := readLibraryFile(include, libraryDirs)
		if err != nil {
			return "", err
		}

		code := extractLibraryCode(libraryLines, extractInclude)
		mergedLibraries.PushFront(code)
	}

	hasLibrariesMarker := false
	var out strings.Builder

	for _, key := range stdIncludes {
		if strings.HasSuffix(key, ".hpp") {
			continue
		}
		out.WriteString("#include <")
		out.WriteString(key)
		out.WriteString(">\n")
	}

	for _, line := range lines {
		if strings.HasPrefix(line, "#include") {
			continue
		}
		out.WriteString(line)
		out.WriteString("\n")

		if strings.HasPrefix(line, LibraryMarker) {
			hasLibrariesMarker = true
			if mergedLibraries.Len() == 0 {
				continue
			}
			for mergedLibraries.Len() > 0 {
				out.WriteString(mergedLibraries.PopFront())
			}
		}
	}

	if mergedLibraries.Len() > 0 && !hasLibrariesMarker {
		return "", fmt.Errorf("%w: add %q to main source", ErrLibrariesMarkerMissing, LibraryMarker)
	}

	return strings.TrimSuffix(out.String(), "\n") + "\n", nil
}

func parseInclude(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "#include") {
		return "", false
	}

	rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "#include"))
	if len(rest) < 3 {
		return "", false
	}

	switch rest[0] {
	case '"':
		end := strings.Index(rest[1:], "\"")
		if end == -1 {
			return "", false
		}
		return rest[1 : 1+end], true
	case '<':
		end := strings.Index(rest, ">")
		if end <= 1 {
			return "", false
		}
		return rest[1:end], true
	default:
		return "", false
	}
}

func readLibraryFile(include string, libraryDirs []string) ([]string, error) {
	for _, libraryDir := range libraryDirs {
		libraryPath := filepath.Join(libraryDir, include)
		if _, err := os.Stat(libraryPath); os.IsNotExist(err) {
			continue
		}
		content, err := os.ReadFile(libraryPath)
		if err != nil {
			return nil, fmt.Errorf("read library %q: %w", libraryPath, err)
		}
		return strings.Split(string(content), "\n"), nil
	}
	return nil, fmt.Errorf("library not found: %q (searched in %v)", include, libraryDirs)
}

func extractLibraryCode(lines []string, extractInclude func(string) bool) string {
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	var out strings.Builder
	for _, line := range lines {
		if extractInclude(line) || droppedLibraryLine(line) {
			continue
		}
		out.WriteString(line)
		out.WriteString("\n")
	}
	return out.String()
}

func droppedLibraryLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	switch {
	case trimmed == "#pragma once":
		return true
	case strings.HasPrefix(trimmed, "#include"):
		return true
	case strings.HasPrefix(trimmed, "using namespace"):
		return true
	case trimmed == CodeMarker:
		return true
	default:
		return false
	}
}
