package web

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/filipgorny/agent/core"
)

// browserUserAgent is sent on every fetch: many sites answer the Go default
// User-Agent with 403/404/429, so we present a common desktop-browser string.
const browserUserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0 Safari/537.36"

// fetchClient bounds every request so a hung server can't block the agent when
// the caller passes a context without a deadline.
var fetchClient = &http.Client{Timeout: 30 * time.Second}

// WebDownloadSkillName is the registered name of the web_download skill.
const WebDownloadSkillName = "web_download"

// WebDownload fetches the raw content at a URL. Params: url, optional dest.
type WebDownload struct{}

func (WebDownload) Name() string {
	return WebDownloadSkillName
}

func (WebDownload) Description() string {
	return "Download raw content from a URL. params: {\"url\": string, \"dest\": string?}"
}

func (WebDownload) IsAsync() bool {
	return false
}

func (WebDownload) GetEvents() []core.EventSpec {
	return []core.EventSpec{{Name: "web_download.result", Description: "Emitted when web_download finishes."}}
}

func (WebDownload) Run(ctx context.Context, params map[string]any) (string, error) {
	url, ok := core.ParamString(params, "url")

	if !ok {
		return "", fmt.Errorf("web_download: missing string \"url\" parameter")
	}

	body, err := fetchURL(ctx, url)

	if err != nil {
		return "", fmt.Errorf("web_download: %w", err)
	}

	if dest, ok := core.ParamString(params, "dest"); ok && dest != "" {
		if err := os.WriteFile(dest, body, 0o644); err != nil {
			return "", fmt.Errorf("web_download: %w", err)
		}

		return fmt.Sprintf("saved %d bytes to %s", len(body), dest), nil
	}

	return string(body), nil
}

// fetchURL GETs url and returns the response body.
func fetchURL(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", browserUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := fetchClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get %s: status %d %s", url, resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	return io.ReadAll(resp.Body)
}
