package submit

import (
	"fmt"
	"os"

	"github.com/atotto/clipboard"

	"Competitive-Programming-eXecutor/internal/config"
	"Competitive-Programming-eXecutor/internal/merge"
	"Competitive-Programming-eXecutor/internal/problem"
	setupatcoder "Competitive-Programming-eXecutor/internal/setup/atcoder"
	"Competitive-Programming-eXecutor/internal/test"
)

type Request struct {
	ProblemPath string
	Lang        string
	TimeLimit   int
	SkipTest    bool
	Copy        bool
}

func Run(cfg *config.Config, req Request) error {
	p, err := problem.Open(req.ProblemPath, req.Lang)
	if err != nil {
		return err
	}
	if p.Lang == "" {
		p.Lang = cfg.File.DefaultLang
	}
	if p.Lang == "" {
		p.Lang = "cpp"
	}

	if !req.SkipTest {
		executableFilePath, err := test.Compile(p.Dir, p.Lang, cfg)
		if err != nil {
			return err
		}
		executionTimes, err := test.Run(p.Dir, executableFilePath, p.Lang)
		if err != nil {
			return err
		}
		if err := test.Compare(p.Dir, executionTimes, req.TimeLimit); err != nil {
			return err
		}
	}

	submissionPath, err := merge.WriteSubmission(p, cfg.File.LibraryDirs)
	if err != nil {
		return err
	}
	fmt.Printf("[INFO] wrote %s\n", submissionPath)

	if req.Copy {
		return copySourceCode(submissionPath)
	}

	url, err := taskURL(req.ProblemPath)
	if err != nil {
		return err
	}
	fmt.Printf("[INFO] submitting to %s\n", url)

	session := setupatcoder.Session(cfg)
	return submitWithOJ(url, submissionPath, session)
}

func copySourceCode(submissionPath string) error {
	content, err := os.ReadFile(submissionPath)
	if err != nil {
		return fmt.Errorf("read submission file %q: %w", submissionPath, err)
	}
	if err := clipboard.WriteAll(string(content)); err != nil {
		return fmt.Errorf("write to clipboard: %w", err)
	}
	fmt.Printf("[INFO] copied %s to clipboard\n", submissionPath)
	return nil
}
