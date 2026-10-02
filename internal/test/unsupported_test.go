package test

import (
	"testing"

	"Competitive-Programming-eXecutor/internal/config"
	"Competitive-Programming-eXecutor/internal/problem"
)

func TestRunSamplesUnsupportedLanguage(t *testing.T) {
	err := RunSamples(problem.Problem{Dir: "a", Lang: "rs"}, Options{TimeLimit: 2}, config.NewConfig())
	if err == nil {
		t.Fatal("expected error")
	}
	const want = "unsupported language \"rs\" (supported: cpp, py)"
	if err.Error() != want {
		t.Fatalf("error %q, want %q", err.Error(), want)
	}
}
