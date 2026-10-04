package test

import (
	"math"
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
			got := judge(tc.expected, tc.actual, tc.duration, limit, tc.finish, 0)
			if got != tc.want {
				t.Fatalf("judge(%q, %q, %s, %d) = %s, want %s", tc.expected, tc.actual, tc.duration, tc.finish, got, tc.want)
			}
		})
	}
}

func TestJudgeEps(t *testing.T) {
	limit := 2 * time.Second
	cases := []struct {
		name     string
		expected string
		actual   string
		eps      float64
		want     Verdict
	}{
		{name: "exact", expected: "1.0 2\n", actual: "1.0 2\n", eps: 1e-6, want: AC},
		{name: "1e-9 off", expected: "1.000000000\n", actual: "1.000000001\n", eps: 1e-6, want: AC},
		{name: "1e-3 off", expected: "1.000\n", actual: "1.001\n", eps: 1e-6, want: WA},
		{name: "token count", expected: "1 2\n", actual: "1\n", eps: 1e-6, want: WA},
		{name: "non numeric", expected: "foo\n", actual: "bar\n", eps: 1e-6, want: WA},
		{name: "eps 0 formatting", expected: "1.0\n", actual: "1.00\n", eps: 0, want: WA},
		{name: "relative", expected: "1000\n", actual: "1000.0005\n", eps: 1e-6, want: AC},
		{name: "whitespace", expected: "1.0 2.0\n", actual: "1.0\n2.0\n", eps: 1e-6, want: AC},
		{name: "whitespace eps 0", expected: "1.0 2.0\n", actual: "1.0\n2.0\n", eps: 0, want: WA},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := judge(tc.expected, tc.actual, time.Millisecond, limit, ran, tc.eps)
			if got != tc.want {
				t.Fatalf("judge(%q, %q, eps %g) = %s, want %s", tc.expected, tc.actual, tc.eps, got, tc.want)
			}
		})
	}
}

func TestTimeLimitDuration(t *testing.T) {
	d, err := timeLimitDuration(1.5)
	if err != nil {
		t.Fatal(err)
	}
	if d != 1500*time.Millisecond {
		t.Fatalf("duration %s, want 1.5s", d)
	}
	if _, err := timeLimitDuration(2); err != nil {
		t.Fatal(err)
	}
	for _, seconds := range []float64{0, -1, math.NaN(), math.Inf(1)} {
		if _, err := timeLimitDuration(seconds); err == nil {
			t.Fatalf("seconds %v accepted", seconds)
		}
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
