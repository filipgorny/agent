package agent

import (
	"context"
	"strings"
	"testing"

	llm "github.com/filipgorny/llm-provider"
)

func TestSummarizeText(t *testing.T) {
	strat := &scriptedLlm{reply: func(int, string) string { return "SUMMARY" }}
	skill := SummarizeText{llm: llm.NewLlmProvider(strat)}

	out, err := skill.Run(context.Background(), map[string]any{
		"text":      "A very long piece of text about many things.",
		"max_words": float64(10),
	})

	if err != nil {
		t.Fatalf("summarize: %v", err)
	}

	if out != "SUMMARY" {
		t.Errorf("out = %q", out)
	}

	if !strings.Contains(strat.calls[0], "Summarize") || !strings.Contains(strat.calls[0], "10 words") {
		t.Errorf("prompt missing expectations: %q", strat.calls[0])
	}
}

func TestTranslate(t *testing.T) {
	strat := &scriptedLlm{reply: func(int, string) string { return "Cześć" }}
	skill := Translate{llm: llm.NewLlmProvider(strat)}

	out, err := skill.Run(context.Background(), map[string]any{
		"text": "Hello",
		"to":   "Polish",
	})

	if err != nil {
		t.Fatalf("translate: %v", err)
	}

	if out != "Cześć" {
		t.Errorf("out = %q", out)
	}

	if !strings.Contains(strat.calls[0], "Polish") || !strings.Contains(strat.calls[0], "Hello") {
		t.Errorf("prompt missing expectations: %q", strat.calls[0])
	}
}

func TestLLMSkillsRequireLLM(t *testing.T) {
	if _, err := (SummarizeText{}).Run(context.Background(), map[string]any{"text": "x"}); err == nil {
		t.Error("summarize_text should error without LLM")
	}

	if _, err := (Translate{}).Run(context.Background(), map[string]any{"text": "x", "to": "fr"}); err == nil {
		t.Error("translate should error without LLM")
	}
}
