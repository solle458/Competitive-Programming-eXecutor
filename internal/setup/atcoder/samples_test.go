package atcoder

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadSamplesKeepsEmptyOutput(t *testing.T) {
	const page = `<!DOCTYPE html><html><body>
<section><h3>Sample Input 3</h3><pre>x</pre></section>
<section><h3>Sample Output 3</h3><pre>
</pre></section>
<section><h3>入力例 2</h3><pre>in</pre></section>
<section><h3>出力例 2</h3><pre>out</pre></section>
</body></html>`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte("User-agent: *\nAllow: /\n"))
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(page))
	}))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	if err := downloadSamplesFromURL(dir, srv.URL+"/contests/abc350/tasks/abc350_f", ""); err != nil {
		t.Fatal(err)
	}

	assertSampleFile(t, filepath.Join(dir, "test", "sample-3.in"), "x\n")
	assertSampleFile(t, filepath.Join(dir, "test", "sample-3.out"), "\n")
	assertSampleFile(t, filepath.Join(dir, "test", "sample-2.in"), "in\n")
	assertSampleFile(t, filepath.Join(dir, "test", "sample-2.out"), "out\n")
}

func TestDownloadSamplesRejectsOneSidedSample(t *testing.T) {
	const page = `<!DOCTYPE html><html><body>
<section><h3>Sample Input 1</h3><pre>x</pre></section>
</body></html>`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			_, _ = w.Write([]byte("User-agent: *\nAllow: /\n"))
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(page))
	}))
	t.Cleanup(srv.Close)

	err := downloadSamplesFromURL(t.TempDir(), srv.URL+"/task", "")
	if err == nil || err.Error() != "sample 1 is incomplete" {
		t.Fatalf("error %v, want sample 1 is incomplete", err)
	}
}

func assertSampleFile(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Fatalf("%s\n got %q\nwant %q", path, got, want)
	}
}
