package problem

import (
	"fmt"
	"os"
	"path/filepath"
)

type Problem struct {
	Dir  string
	Lang string
}

func Open(dir, lang string) (Problem, error) {
	problemDir := filepath.Join(".", dir)
	info, err := os.Stat(problemDir)
	if err != nil {
		if os.IsNotExist(err) {
			return Problem{}, fmt.Errorf("problem %q not found in current directory", dir)
		}
		return Problem{}, fmt.Errorf("stat problem directory %q: %w", problemDir, err)
	}
	if !info.IsDir() {
		return Problem{}, fmt.Errorf("%q is not a directory", problemDir)
	}
	return Problem{Dir: dir, Lang: lang}, nil
}

func (p Problem) Source() string {
	return filepath.Join(p.Dir, "main."+p.Lang)
}

func (p Problem) Submission() string {
	return filepath.Join(p.Dir, "submission."+p.Lang)
}

func (p Problem) SampleDir() string {
	return filepath.Join(p.Dir, "test")
}

func (p Problem) Binary() string {
	return filepath.Join(p.Dir, "a.out")
}
