package web

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/filipgorny/agent/core"
)

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

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
