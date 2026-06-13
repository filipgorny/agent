package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestFetchSendsBrowserUserAgent guards the most common real-world failure: many
// sites answer the Go default User-Agent ("Go-http-client/…") with 403/404/429.
// fetchURL must send a non-empty, non-default UA.
func TestFetchSendsBrowserUserAgent(t *testing.T) {
	var gotUA string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		_, _ = w.Write([]byte(sampleArticle))
	}))

	defer srv.Close()

	if _, err := fetchURL(context.Background(), srv.URL); err != nil {
		t.Fatalf("fetchURL: %v", err)
	}

	if gotUA == "" {
		t.Error("no User-Agent sent")
	}

	if strings.HasPrefix(gotUA, "Go-http-client") {
		t.Errorf("sent Go default User-Agent %q; real sites block it", gotUA)
	}
}

func TestWebGet404ReturnsInformativeError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))

	defer srv.Close()

	_, err := (WebGet{}).Run(context.Background(), map[string]any{"url": srv.URL})

	if err == nil {
		t.Fatal("expected error on 404")
	}

	msg := err.Error()

	if !strings.Contains(msg, "404") || !strings.Contains(msg, "Not Found") {
		t.Errorf("error should name the status and its text, got %q", msg)
	}

	if !strings.Contains(msg, srv.URL) {
		t.Errorf("error should include the URL for diagnosis, got %q", msg)
	}
}

func TestWebGet500ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))

	defer srv.Close()

	if _, err := (WebGet{}).Run(context.Background(), map[string]any{"url": srv.URL}); err == nil {
		t.Fatal("expected error on 500")
	}
}

func TestWebGetMissingURL(t *testing.T) {
	if _, err := (WebGet{}).Run(context.Background(), map[string]any{}); err == nil {
		t.Fatal("expected error when url param is missing")
	}
}

func TestWebGetBadURL(t *testing.T) {
	if _, err := (WebGet{}).Run(context.Background(), map[string]any{"url": "http://nonexistent.invalid./"}); err == nil {
		t.Fatal("expected error for an unresolvable host")
	}
}

func TestWebGetFollowsRedirect(t *testing.T) {
	final := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(sampleArticle))
	}))

	defer final.Close()

	redir := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, final.URL, http.StatusFound)
	}))

	defer redir.Close()

	out, err := (WebGet{}).Run(context.Background(), map[string]any{"url": redir.URL})

	if err != nil {
		t.Fatalf("web_get through redirect: %v", err)
	}

	if !strings.Contains(out, "Goroutines") {
		t.Errorf("redirect target content missing: %q", out)
	}
}

func TestWebGetTextFormat(t *testing.T) {
	srv := newServer(t)

	defer srv.Close()

	out, err := (WebGet{}).Run(context.Background(), map[string]any{"url": srv.URL, "format": "text"})

	if err != nil {
		t.Fatalf("web_get text: %v", err)
	}

	if !strings.Contains(out, "Goroutines make concurrent programming") {
		t.Errorf("text content missing: %q", out)
	}

	if strings.Contains(out, "<p>") {
		t.Errorf("text format should not contain HTML tags: %q", out)
	}
}

// TestWebGetNonArticleContent ensures non-article responses (here, plain JSON)
// still return their content via the fallback path rather than erroring.
func TestWebGetNonArticleContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","value":42}`))
	}))

	defer srv.Close()

	out, err := (WebGet{}).Run(context.Background(), map[string]any{"url": srv.URL})

	if err != nil {
		t.Fatalf("web_get json: %v", err)
	}

	if !strings.Contains(out, "42") {
		t.Errorf("fallback should preserve body content, got %q", out)
	}
}
