package atcoder

import "testing"

func TestParseTasksPageUsesDisplayedIndex(t *testing.T) {
	const body = `
<table>
<tr>
<td class="text-center no-break"><a href="/contests/adt_all_20261002_2/tasks/abc387_a">A</a></td>
<td><a href="/contests/adt_all_20261002_2/tasks/abc387_a">Happy New Year 2025</a></td>
</tr>
<tr>
<td class="text-center no-break"><a href="/contests/adt_all_20261002_2/tasks/abc297_a">B</a></td>
<td><a href="/contests/adt_all_20261002_2/tasks/abc297_a">Double Click</a></td>
</tr>
<tr>
<td><a href="/contests/abc387/tasks/abc387_a">A</a></td>
</tr>
</table>`

	problems, err := parseTasksPage(body, "adt_all_20261002_2")
	if err != nil {
		t.Fatal(err)
	}
	want := []Problem{
		{ContestID: "adt_all_20261002_2", ProblemID: "abc387_a", ProblemIndex: "A"},
		{ContestID: "adt_all_20261002_2", ProblemID: "abc297_a", ProblemIndex: "B"},
	}
	if len(problems) != len(want) {
		t.Fatalf("len %d, want %d: %+v", len(problems), len(want), problems)
	}
	for i := range want {
		if problems[i] != want[i] {
			t.Fatalf("problem %d\n got %+v\nwant %+v", i, problems[i], want[i])
		}
	}
	if problemIndexesCollide(problems) {
		t.Fatalf("indexes collide: %+v", problems)
	}
	suffixes := []Problem{
		{ProblemID: "abc387_a", ProblemIndex: problemIndexFromID("abc387_a")},
		{ProblemID: "abc297_a", ProblemIndex: problemIndexFromID("abc297_a")},
	}
	if !problemIndexesCollide(suffixes) {
		t.Fatal("original problem-id suffixes should collide, otherwise the daily-contest bug is gone")
	}
}

func TestParseTasksPageEmpty(t *testing.T) {
	_, err := parseTasksPage(`<html><tr><td>none</td></tr></html>`, "adt_all_20261002_2")
	if err == nil || err.Error() != "no problems found on tasks page" {
		t.Fatalf("error %v", err)
	}
}

func TestProblemIndexesCollide(t *testing.T) {
	if problemIndexesCollide([]Problem{{ProblemIndex: "A"}, {ProblemIndex: "B"}}) {
		t.Fatal("unique indexes reported as colliding")
	}
	if !problemIndexesCollide([]Problem{{ProblemIndex: "A"}, {ProblemIndex: "a"}}) {
		t.Fatal("case-insensitive collision not reported")
	}
}
