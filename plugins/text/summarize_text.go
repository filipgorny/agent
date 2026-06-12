package text

import (
	"context"
	"fmt"

	"github.com/filipgorny/agent/core"
	llm "github.com/filipgorny/llm-provider"
)

// SummarizeTextSkillName is the registered name of the summarize_text skill.
const SummarizeTextSkillName = "summarize_text"

// SummarizeText summarizes text using the agent's LLM. Params: text, optional
// max_words.
type SummarizeText struct {
	llm *llm.LlmProvider
}

func (SummarizeText) Name() string {
	return SummarizeTextSkillName
}

func (SummarizeText) Description() string {
	return "Summarize text using the LLM. params: {\"text\": string, \"max_words\": int?}"
}

func (SummarizeText) IsAsync() bool {
	return false
}

func (SummarizeText) GetEvents() []core.EventSpec {
	return []core.EventSpec{{Name: "summarize_text.result", Description: "Emitted with the summary when summarize_text finishes."}}
}

func (s SummarizeText) Run(ctx context.Context, params map[string]any) (string, error) {
	if s.llm == nil {
		return "", fmt.Errorf("summarize_text: no LLM available")
	}

	textInput, ok := core.ParamString(params, "text")

	if !ok {
		return "", fmt.Errorf("summarize_text: missing string \"text\" parameter")
	}

	constraint := ""

	if max, ok := core.ParamInt(params, "max_words"); ok && max > 0 {
		constraint = fmt.Sprintf(" in at most %d words", max)
	}

	prompt := fmt.Sprintf("Summarize the following text%s. Output only the summary.\n\n%s", constraint, textInput)

	return s.llm.Prompt(ctx, prompt)
}
