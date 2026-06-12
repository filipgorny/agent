package text

import (
	"context"
	"strings"
	"testing"

	llm "github.com/filipgorny/llm-provider"
)

// fakeLlm records the prompt and returns a fixed reply.
type fakeLlm struct {
	prompt string
	reply  string
}

func (f *fakeLlm) Prompt(ctx context.Context, prompt string) (string, error) {
	f.prompt = prompt

	return f.reply, nil
}

func TestSummarizeText(t *testing.T) {
	f := &fakeLlm{reply: "SUMMARY"}
	skill := SummarizeText{llm: llm.NewLlmProvider(f)}

	out, err := skill.Run(context.Background(), map[string]any{"text": "long text", "max_words": float64(10)})

	if err != nil {
		t.Fatalf("summarize: %v", err)
	}

	if out != "SUMMARY" {
		t.Errorf("out = %q", out)
	}

	if !strings.Contains(f.prompt, "Summarize") || !strings.Contains(f.prompt, "10 words") {
		t.Errorf("prompt = %q", f.prompt)
	}
}

func TestTranslate(t *testing.T) {
	f := &fakeLlm{reply: "Cześć"}
	skill := Translate{llm: llm.NewLlmProvider(f)}

	out, err := skill.Run(context.Background(), map[string]any{"text": "Hello", "to": "Polish"})

	if err != nil {
		t.Fatalf("translate: %v", err)
	}

	if out != "Cześć" {
		t.Errorf("out = %q", out)
	}

	if !strings.Contains(f.prompt, "Polish") || !strings.Contains(f.prompt, "Hello") {
		t.Errorf("prompt = %q", f.prompt)
	}
}

func TestRequireLLM(t *testing.T) {
	if _, err := (SummarizeText{}).Run(context.Background(), map[string]any{"text": "x"}); err == nil {
		t.Error("summarize_text should error without LLM")
	}

	if _, err := (Translate{}).Run(context.Background(), map[string]any{"text": "x", "to": "fr"}); err == nil {
		t.Error("translate should error without LLM")
	}
}
