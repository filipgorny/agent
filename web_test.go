package agent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleArticle = `<!DOCTYPE html><html><head><title>Test</title></head><body>
<nav>Home About Contact navigation junk</nav>
<article>
<h1>The Go Programming Language</h1>
<p>Go is an open source programming language designed for building simple, reliable and efficient software.</p>
<p>It was created at Google and is widely used for backend services, command line tools and cloud infrastructure.</p>
<p>Goroutines make concurrent programming straightforward and the standard library is extensive.</p>
</article>
<footer>copyright junk footer</footer>
</body></html>`

func newArticleServer(t *testing.T) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(sampleArticle))
	}))
}

func TestWebDownload(t *testing.T) {
	srv := newArticleServer(t)

	defer srv.Close()

	ctx := context.Background()

	out, err := (WebDownload{}).Run(ctx, map[string]any{"url": srv.URL})

	if err != nil {
		t.Fatalf("web_download: %v", err)
	}

	if !strings.Contains(out, "Goroutines") {
		t.Errorf("body missing expected content: %q", out)
	}

	// to dest
	dest := filepath.Join(t.TempDir(), "page.html")

	if _, err := (WebDownload{}).Run(ctx, map[string]any{"url": srv.URL, "dest": dest}); err != nil {
		t.Fatalf("web_download dest: %v", err)
	}

	data, _ := os.ReadFile(dest)

	if !strings.Contains(string(data), "Goroutines") {
		t.Errorf("saved file missing content")
	}
}

func TestWebGet(t *testing.T) {
	srv := newArticleServer(t)

	defer srv.Close()

	out, err := (WebGet{}).Run(context.Background(), map[string]any{"url": srv.URL})

	if err != nil {
		t.Fatalf("web_get: %v", err)
	}

	if !strings.Contains(out, "Goroutines make concurrent programming") {
		t.Errorf("web_get missing article content: %q", out)
	}
}
