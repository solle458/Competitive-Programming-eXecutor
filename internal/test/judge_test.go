package test

import (
	"strings"
	"testing"
	"time"
)

func TestJudge(t *testing.T) {
	limit := 2 * time.Second
	cases := []struct {
		name     string
		expected string
		actual   string
		duration time.Duration
		finish   runFinish
		want     Verdict
	}{
		{name: "match", expected: "42\n", actual: "42\n", duration: time.Millisecond, finish: ran, want: AC},
		{name: "trim", expected: "42\n", actual: "  42  ", duration: time.Millisecond, finish: ran, want: AC},
		{name: "wrong", expected: "42\n", actual: "0\n", duration: time.Millisecond, finish: ran, want: WA},
		{name: "slow match", expected: "42\n", actual: "42\n", duration: 3 * time.Second, finish: ran, want: TLE},
		{name: "slow wrong", expected: "42\n", actual: "0\n", duration: 3 * time.Second, finish: ran, want: WA},
		{name: "at limit", expected: "42\n", actual: "42\n", duration: limit, finish: ran, want: AC},
		{name: "killed match", expected: "42\n", actual: "42\n", duration: time.Second, finish: timedOut, want: TLE},
		{name: "killed wrong", expected: "42\n", actual: "0\n", duration: time.Second, finish: timedOut, want: TLE},
		{name: "crash match", expected: "42\n", actual: "42\n", duration: time.Millisecond, finish: crashed, want: RE},
		{name: "crash wrong", expected: "42\n", actual: "0\n", duration: time.Millisecond, finish: crashed, want: RE},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := judge(tc.expected, tc.actual, tc.duration, limit, tc.finish)
			if got != tc.want {
				t.Fatalf("judge(%q, %q, %s, %d) = %s, want %s", tc.expected, tc.actual, tc.duration, tc.finish, got, tc.want)
			}
		})
	}
}

func TestStderrTail(t *testing.T) {
	if got := stderrTail("boom"); got != "boom" {
		t.Fatalf("stderrTail(%q) = %q, want %q", "boom", got, "boom")
	}
	body := strings.Repeat("a", 1024) + "TAIL"
	got := stderrTail(body)
	want := strings.Repeat("a", 1020) + "TAIL"
	if got != want {
		t.Fatalf("stderrTail length %d, suffix %q, want length %d suffix %q", len(got), got[len(got)-8:], len(want), want[len(want)-8:])
	}
}
