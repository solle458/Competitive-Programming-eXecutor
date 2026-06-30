package merge

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gammazero/deque"
)

// init やテンプレート生成と同じ文字列を使う（自動生成するならここが唯一の定義元）
const (
	LibraryMarker = "/* -- libraries --*/"
	CodeMarker    = "/* -- library code --*/"
)

var (
	ErrLibrariesMarkerMissing = errors.New("libraries marker not found in main source")
	ErrCodeMarkerMissing      = errors.New("library code marker not found")
)

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

		code, err := extractLibraryCode(include, libraryLines, extractInclude)
		if err != nil {
			return "", err
		}
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

func extractLibraryCode(include string, lines []string, extractInclude func(string) bool) (string, error) {
	var out strings.Builder
	inCode := false
	foundMarker := false

	for _, line := range lines {
		if extractInclude(line) {
			continue
		}
		if strings.HasPrefix(line, CodeMarker) {
			if inCode {
				return out.String(), nil
			}
			inCode = true
			foundMarker = true
			continue
		}
		if inCode {
			out.WriteString(line)
			out.WriteString("\n")
		}
	}

	if !foundMarker {
		return "", fmt.Errorf("%w in %q: add %q", ErrCodeMarkerMissing, include, CodeMarker)
	}
	return out.String(), nil
}
