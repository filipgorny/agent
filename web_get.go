package agent

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/markusmobius/go-trafilatura"
	"golang.org/x/net/html"
)

// WebGetSkillName is the registered name of the web_get skill.
const WebGetSkillName = "web_get"

func init() {
	RegisterSkill(WebGetSkillName, func(Deps) Skill {
		return WebGet{}
	})
}

// WebGet fetches a URL and returns its main readable content. It uses
// go-trafilatura to extract the article body (with built-in fallbacks for
// non-article/list pages), converting to markdown by default. Params: url,
// optional format ("markdown" default | "text").
type WebGet struct{}

func (WebGet) Name() string {
	return WebGetSkillName
}

func (WebGet) Run(ctx context.Context, params map[string]any) (string, error) {
	url, ok := paramString(params, "url")

	if !ok {
		return "", fmt.Errorf("web_get: missing string \"url\" parameter")
	}

	format, _ := paramString(params, "format")

	body, err := fetchURL(ctx, url)

	if err != nil {
		return "", fmt.Errorf("web_get: %w", err)
	}

	result, err := trafilatura.Extract(bytes.NewReader(body), trafilatura.Options{
		IncludeLinks: true,
	})

	if err == nil && result != nil {
		if format == "text" && strings.TrimSpace(result.ContentText) != "" {
			return result.ContentText, nil
		}

		if result.ContentNode != nil {
			md, err := nodeToMarkdown(result.ContentNode)

			if err == nil && strings.TrimSpace(md) != "" {
				return md, nil
			}
		}

		if strings.TrimSpace(result.ContentText) != "" {
			return result.ContentText, nil
		}
	}

	// Fallback: convert the whole page (handles list/social pages where article
	// extraction yields little).
	md, err := htmltomarkdown.ConvertString(string(body))

	if err != nil {
		return "", fmt.Errorf("web_get: convert fallback: %w", err)
	}

	return md, nil
}

// nodeToMarkdown renders an html node to HTML, then converts it to markdown.
func nodeToMarkdown(node *html.Node) (string, error) {
	var buf bytes.Buffer

	if err := html.Render(&buf, node); err != nil {
		return "", err
	}

	return htmltomarkdown.ConvertString(buf.String())
}
