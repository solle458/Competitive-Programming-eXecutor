package workspacesync

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"Competitive-Programming-eXecutor/internal/setup/atcoder"
)

func TestContestIDs(t *testing.T) {
	got := contestIDs([]string{
		"abc100/a/main.py",
		"abc100/b/a.out",
		"abc100/ex/main.cpp",
		"library/graph/dsu.hpp",
		"library/io/fast.hpp",
		"README.md",
		"arc200/a/test/sample-1.test",
	})
	want := []string{"abc100", "arc200"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestBlockReason(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	byID := map[string]atcoder.Contest{
		"abc100": {
			ID:               "abc100",
			StartEpochSecond: now.Unix() - 10,
			DurationSecond:   1000,
		},
		"abc099": {
			ID:               "abc099",
			StartEpochSecond: now.Unix() - 10_000,
			DurationSecond:   100,
		},
		"arc200": {
			ID:               "arc200",
			StartEpochSecond: now.Unix() - 5,
			DurationSecond:   1000,
		},
	}

	if got := blockReason(nil, byID, now); got != "" {
		t.Fatalf("no contests: %q", got)
	}
	if got := blockReason([]string{"abc099"}, byID, now); got != "" {
		t.Fatalf("past contest: %q", got)
	}
	if got := blockReason([]string{"abc100"}, byID, now); got != "contest abc100 is in progress" {
		t.Fatalf("one live contest: %q", got)
	}
	if got := blockReason([]string{"abc100", "arc200"}, byID, now); got != "contests abc100, arc200 are in progress" {
		t.Fatalf("two live contests: %q", got)
	}
	if got := blockReason([]string{"abc404"}, byID, now); got != "contest abc404 is missing from the contest schedule" {
		t.Fatalf("missing contest: %q", got)
	}
}

func TestLoadContests(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"id":"ABC100","start_epoch_second":10,"duration_second":5,"title":"X","rate_change":"-"}]`))
	}))
	t.Cleanup(srv.Close)
	t.Setenv("CPX_CONTESTS_URL", srv.URL)

	got, err := loadContests()
	if err != nil {
		t.Fatal(err)
	}
	contest, ok := got["abc100"]
	if !ok {
		t.Fatalf("missing abc100 in %#v", got)
	}
	if contest.StartEpochSecond != 10 || contest.DurationSecond != 5 {
		t.Fatalf("contest %+v", contest)
	}
}
