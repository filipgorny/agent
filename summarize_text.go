package agent

import (
	"context"
	"fmt"

	llm "github.com/filipgorny/llm-provider"
)

// SummarizeTextSkillName is the registered name of the summarize_text skill.
const SummarizeTextSkillName = "summarize_text"

func init() {
	RegisterSkill(SummarizeTextSkillName, func(d Deps) Skill {
		return SummarizeText{llm: d.LLM}
	})
}

// SummarizeText summarizes text using the agent's LLM. Params: text, optional
// max_words.
type SummarizeText struct {
	llm *llm.LlmProvider
}

func (SummarizeText) Name() string {
	return SummarizeTextSkillName
}

func (s SummarizeText) Run(ctx context.Context, params map[string]any) (string, error) {
	if s.llm == nil {
		return "", fmt.Errorf("summarize_text: no LLM available")
	}

	text, ok := paramString(params, "text")

	if !ok {
		return "", fmt.Errorf("summarize_text: missing string \"text\" parameter")
	}

	constraint := ""

	if max, ok := paramInt(params, "max_words"); ok && max > 0 {
		constraint = fmt.Sprintf(" in at most %d words", max)
	}

	prompt := fmt.Sprintf("Summarize the following text%s. Output only the summary.\n\n%s", constraint, text)

	return s.llm.Prompt(ctx, prompt)
}
