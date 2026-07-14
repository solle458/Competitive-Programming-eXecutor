package submit

import (
	"fmt"
	"os"
	"path/filepath"

	"Competitive-Programming-eXecutor/internal/config"
	"Competitive-Programming-eXecutor/internal/merge"
	setupatcoder "Competitive-Programming-eXecutor/internal/setup/atcoder"
	"Competitive-Programming-eXecutor/internal/test"
)

type Request struct {
	ProblemPath string
	Lang        string
	TimeLimit   int
	SkipTest    bool
}

func Run(cfg *config.Config, req Request) error {
	if err := validProblemPath(req.ProblemPath); err != nil {
		return err
	}

	lang := req.Lang
	if lang == "" {
		lang = cfg.File.DefaultLang
	}
	if lang == "" {
		lang = "cpp"
	}

	if !req.SkipTest {
		executableFilePath, err := test.Compile(req.ProblemPath, lang, cfg)
		if err != nil {
			return err
		}
		executionTimes, err := test.Run(req.ProblemPath, executableFilePath, lang)
		if err != nil {
			return err
		}
		if err := test.Compare(req.ProblemPath, executionTimes, req.TimeLimit); err != nil {
			return err
		}
	}

	submissionPath, err := generateSubmission(req.ProblemPath, lang, cfg)
	if err != nil {
		return err
	}

	url, err := taskURL(req.ProblemPath)
	if err != nil {
		return err
	}

	session := setupatcoder.Session(cfg)
	return submitWithOJ(url, submissionPath, session)
}

func validProblemPath(problemPath string) error {
	dir := filepath.Join(".", problemPath)
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("problem %q not found in current directory", problemPath)
		}
		return fmt.Errorf("stat problem directory %q: %w", dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%q is not a directory", dir)
	}
	return nil
}

func generateSubmission(problemPath, lang string, cfg *config.Config) (string, error) {
	sourcePath := filepath.Join(".", problemPath, "main."+lang)
	content, err := os.ReadFile(sourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("source code not found: %q", sourcePath)
		}
		return "", fmt.Errorf("read source code %q: %w", sourcePath, err)
	}

	submissionCode, err := merge.Generate(string(content), cfg.File.LibraryDirs)
	if err != nil {
		return "", err
	}

	submissionPath := filepath.Join(".", problemPath, "submission."+lang)
	if err := os.WriteFile(submissionPath, []byte(submissionCode), 0o644); err != nil {
		return "", fmt.Errorf("write submission file %q: %w", submissionPath, err)
	}
	return submissionPath, nil
}
