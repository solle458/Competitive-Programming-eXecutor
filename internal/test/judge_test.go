package test

import (
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
		want     Verdict
	}{
		{name: "match", expected: "42\n", actual: "42\n", duration: time.Millisecond, want: AC},
		{name: "trim", expected: "42\n", actual: "  42  ", duration: time.Millisecond, want: AC},
		{name: "wrong", expected: "42\n", actual: "0\n", duration: time.Millisecond, want: WA},
		{name: "slow match", expected: "42\n", actual: "42\n", duration: 3 * time.Second, want: TLE},
		{name: "slow wrong", expected: "42\n", actual: "0\n", duration: 3 * time.Second, want: WA},
		{name: "at limit", expected: "42\n", actual: "42\n", duration: limit, want: AC},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := judge(tc.expected, tc.actual, tc.duration, limit)
			if got != tc.want {
				t.Fatalf("judge(%q, %q, %s) = %s, want %s", tc.expected, tc.actual, tc.duration, got, tc.want)
			}
		})
	}
}
