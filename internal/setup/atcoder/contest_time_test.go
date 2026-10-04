package atcoder

import (
	"testing"
	"time"
)

func TestContestInProgress(t *testing.T) {
	start := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	c := Contest{
		ID:               "abc100",
		StartEpochSecond: start.Unix(),
		DurationSecond:   6000,
	}

	if !c.InProgress(start) {
		t.Fatal("at start, want in progress")
	}
	if !c.InProgress(start.Add(5999 * time.Second)) {
		t.Fatal("one second before the end, want in progress")
	}
	if c.InProgress(start.Add(6000 * time.Second)) {
		t.Fatal("at end, want finished")
	}
	if c.InProgress(start.Add(-time.Second)) {
		t.Fatal("before start, want not started")
	}
}
