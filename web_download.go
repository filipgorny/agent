package agent

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
)

// WebDownloadSkillName is the registered name of the web_download skill.
const WebDownloadSkillName = "web_download"

func init() {
	RegisterSkill(WebDownloadSkillName, func(Deps) Skill {
		return WebDownload{}
	})
}

// WebDownload fetches the raw content at a URL. Params: url, optional dest.
// With dest the content is saved and the path returned; otherwise the body is
// returned.
type WebDownload struct{}

func (WebDownload) Name() string {
	return WebDownloadSkillName
}

func (WebDownload) Run(ctx context.Context, params map[string]any) (string, error) {
	url, ok := paramString(params, "url")

	if !ok {
		return "", fmt.Errorf("web_download: missing string \"url\" parameter")
	}

	body, err := fetchURL(ctx, url)

	if err != nil {
		return "", fmt.Errorf("web_download: %w", err)
	}

	if dest, ok := paramString(params, "dest"); ok && dest != "" {
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
