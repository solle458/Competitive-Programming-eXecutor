package template

const cppTemplate = `
#include <iostream>

using namespace std;

/* -- libraries --*/


void solve() {

}

int main() {
	cin.tie(0);
	ios::sync_with_stdio(false);
	solve();
	return 0;
}
`

func GetSourceCode(lang string) string {
	switch lang {
	case "cpp":
		return cppTemplate
	default:
		return ""
	}
}
