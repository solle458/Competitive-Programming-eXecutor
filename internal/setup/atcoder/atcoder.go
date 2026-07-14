package atcoder

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"Competitive-Programming-eXecutor/internal/app"
	"Competitive-Programming-eXecutor/internal/config"
	"Competitive-Programming-eXecutor/internal/setup"
	"Competitive-Programming-eXecutor/internal/template"

	"github.com/gocolly/colly/v2"
	"golang.org/x/sync/errgroup"
)

type Problem struct {
	ContestID    string
	ProblemID    string // e.g. abc464_a (used in task URL)
	ProblemIndex string // e.g. A (used as directory name)
}

type contestProblemEntry struct {
	ContestID    string `json:"contest_id"`
	ProblemID    string `json:"problem_id"`
	ProblemIndex string `json:"problem_index"`
}

type Contest struct {
	ID               string `json:"id"`
	StartEpochSecond int64  `json:"start_epoch_second"`
	DurationSecond   int64  `json:"duration_second"`
	Title            string `json:"title"`
	RateChange       string `json:"rate_change"`
}

const (
	kenkooooContestProblemURL = "https://kenkoooo.com/atcoder/resources/contest-problem.json"
	atcoderUserAgent          = "Mozilla/5.0 (compatible; cpx/1.0; +https://github.com/solle458/Competitive-Programming-eXecutor)"
	requestInterval           = 500 * time.Millisecond
)

var (
	taskLinkPattern   = regexp.MustCompile(`/contests/([a-z0-9_-]+)/tasks/([a-z0-9_-]+)`)
	sampleInputLabel  = regexp.MustCompile(`^(?:入力例|Sample Input) (\d+)$`)
	sampleOutputLabel = regexp.MustCompile(`^(?:出力例|Sample Output) (\d+)$`)
)

var (
	noSampleCasesError = errors.New("no sample cases found")
)

type AtCoder struct{}

func Session(cfg *config.Config) string {
	if v := strings.TrimSpace(os.Getenv("ATCODER_SESSION")); v != "" {
		return v
	}
	if cfg != nil {
		return strings.TrimSpace(cfg.File.AtCoderSession)
	}
	return ""
}

func (AtCoder) Supports(contestID string) bool {
	exists, err := IsContestExists(contestID)
	if err != nil {
		return false
	}
	return exists
}

func (AtCoder) Setup(req setup.Request, app *app.App) error {
	contestID := strings.ToLower(req.ContestID)
	lang := req.Lang
	workingDir := req.WorkingDir

	session := Session(app.Config)

	problems, err := GetProblems(contestID, session)
	if err != nil {
		return err
	}
	if len(problems) == 0 {
		return fmt.Errorf("contest %q has no problems", contestID)
	}

	var g errgroup.Group
	g.SetLimit(3)

	for _, problem := range problems {
		problem := problem
		g.Go(func() error {
			if err := setupProblem(problem, contestID, lang, workingDir, session, app.Config.File.RootDir); err != nil {
				if errors.Is(err, noSampleCasesError) {
					fmt.Printf("no sample cases found for %q, skipping\n", problem.ProblemID)
					return nil
				}
				fmt.Printf("setup failed for %q: %v\n", problem.ProblemID, err)
				return err
			}
			return nil
		})
		time.Sleep(500 * time.Millisecond)
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("setup completed with errors: %w", err)
	}
	return nil
}

func setupProblem(problem Problem, contestID, lang, workingDir, session, rootDir string) error {
	problemDir := filepath.Join(workingDir, contestID, strings.ToLower(problem.ProblemIndex))
	if err := os.MkdirAll(problemDir, 0o755); err != nil {
		return fmt.Errorf("create problem directory %q: %w", problemDir, err)
	}
	mainPath := filepath.Join(problemDir, "main."+lang)
	if _, err := os.Stat(mainPath); os.IsNotExist(err) {
		sourceCode, err := template.GetSourceCode(lang, rootDir)
		if err != nil {
			return fmt.Errorf("get template %q: %w", mainPath, err)
		}
		if err := os.WriteFile(mainPath, []byte(sourceCode), 0o644); err != nil {
			return fmt.Errorf("write template %q: %w", mainPath, err)
		}
	}
	testDir := filepath.Join(problemDir, "test")
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		if err := downloadSamples(problemDir, contestID, problem.ProblemID, session); err != nil {
			return fmt.Errorf("download samples for %q: %w", problem.ProblemID, err)
		}
	}
	return nil
}

func newAtCoderCollector(session string) *colly.Collector {
	col := colly.NewCollector(
		colly.UserAgent(atcoderUserAgent),
	)
	col.SetRequestTimeout(30 * time.Second)
	col.Limit(&colly.LimitRule{
		DomainGlob:  "*atcoder.jp*",
		Parallelism: 1,
		Delay:       requestInterval,
	})
	if session != "" {
		col.OnRequest(func(r *colly.Request) {
			r.Headers.Set("Cookie", "REVEL_SESSION="+session)
		})
	}
	return col
}

type sampleCase struct {
	input  string
	output string
}

func downloadSamples(problemDir, contestID, problemID, session string) error {
	url := fmt.Sprintf("https://atcoder.jp/contests/%s/tasks/%s", contestID, problemID)
	col := newAtCoderCollector(session)

	samples := make(map[int]*sampleCase)
	var visitErr error

	col.OnError(func(_ *colly.Response, err error) {
		visitErr = err
	})
	col.OnHTML("section", func(e *colly.HTMLElement) {
		h3 := strings.TrimSpace(e.ChildText("h3"))
		if h3 == "" {
			return
		}
		pre := strings.TrimRight(e.ChildText("pre"), "\n")
		if pre == "" {
			return
		}

		if m := sampleInputLabel.FindStringSubmatch(h3); m != nil {
			n, _ := strconv.Atoi(m[1])
			if samples[n] == nil {
				samples[n] = &sampleCase{}
			}
			if samples[n].input == "" {
				samples[n].input = pre
			}
			return
		}
		if m := sampleOutputLabel.FindStringSubmatch(h3); m != nil {
			n, _ := strconv.Atoi(m[1])
			if samples[n] == nil {
				samples[n] = &sampleCase{}
			}
			if samples[n].output == "" {
				samples[n].output = pre
			}
		}
	})
	if err := col.Visit(url); err != nil {
		return err
	}
	if visitErr != nil {
		return visitErr
	}
	if len(samples) == 0 {
		return noSampleCasesError
	}

	indices := make([]int, 0, len(samples))
	for i := range samples {
		indices = append(indices, i)
	}
	sort.Ints(indices)

	testDir := filepath.Join(problemDir, "test")
	if err := os.MkdirAll(testDir, 0o755); err != nil {
		return err
	}

	for _, i := range indices {
		s := samples[i]
		if s.input == "" || s.output == "" {
			return fmt.Errorf("sample %d is incomplete", i)
		}
		inPath := filepath.Join(testDir, fmt.Sprintf("sample-%d.in", i))
		outPath := filepath.Join(testDir, fmt.Sprintf("sample-%d.out", i))
		if err := os.WriteFile(inPath, []byte(s.input+"\n"), 0o644); err != nil {
			return err
		}
		if err := os.WriteFile(outPath, []byte(s.output+"\n"), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func GetProblems(contestID, session string) ([]Problem, error) {
	contestID = strings.ToLower(contestID)
	problems, err := getProblemsFromKenkoooo(contestID)
	if err == nil && len(problems) > 0 {
		return problems, nil
	}
	return getProblemsFromTasksPage(contestID, session)
}

func getProblemsFromKenkoooo(contestID string) ([]Problem, error) {
	resp, err := http.Get(kenkooooContestProblemURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("kenkoooo API: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var entries []contestProblemEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, err
	}

	result := make([]Problem, 0)
	for _, entry := range entries {
		if entry.ContestID != contestID {
			continue
		}
		result = append(result, Problem{
			ContestID:    entry.ContestID,
			ProblemID:    entry.ProblemID,
			ProblemIndex: entry.ProblemIndex,
		})
	}
	return result, nil
}

func getProblemsFromTasksPage(contestID, session string) ([]Problem, error) {
	url := fmt.Sprintf("https://atcoder.jp/contests/%s/tasks", contestID)
	resp, err := httpGet(url, session)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound && session == "" {
			return nil, errors.New("tasks page: 404 Not Found (AtCoder login required for ongoing contests; set atcoder_session in .cpx/config.yaml or ATCODER_SESSION env var)")
		}
		return nil, fmt.Errorf("tasks page: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	var problems []Problem
	for _, match := range taskLinkPattern.FindAllStringSubmatch(string(body), -1) {
		if len(match) < 3 || match[1] != contestID {
			continue
		}
		problemID := match[2]
		if _, ok := seen[problemID]; ok {
			continue
		}
		seen[problemID] = struct{}{}
		problems = append(problems, Problem{
			ContestID:    contestID,
			ProblemID:    problemID,
			ProblemIndex: problemIndexFromID(problemID),
		})
	}
	if len(problems) == 0 {
		return nil, errors.New("no problems found on tasks page")
	}
	return problems, nil
}

func problemIndexFromID(problemID string) string {
	parts := strings.Split(problemID, "_")
	if len(parts) == 0 {
		return problemID
	}
	return strings.ToUpper(parts[len(parts)-1])
}

func IsContestExists(contestID string) (bool, error) {
	url := fmt.Sprintf("https://atcoder.jp/contests/%s", strings.ToLower(contestID))
	resp, err := httpGet(url, "")
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK, nil
}

func httpGet(url, session string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", atcoderUserAgent)
	if session != "" {
		req.Header.Set("Cookie", "REVEL_SESSION="+session)
	}
	return http.DefaultClient.Do(req)
}
