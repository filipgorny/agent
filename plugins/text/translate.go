package text

import (
	"context"
	"fmt"

	"github.com/filipgorny/agent/core"
	llm "github.com/filipgorny/llm-provider"
)

// TranslateSkillName is the registered name of the translate skill.
const TranslateSkillName = "translate"

// Translate translates text using the agent's LLM. Params: text, to, optional from.
type Translate struct {
	llm *llm.LlmProvider
}

func (Translate) Name() string {
	return TranslateSkillName
}

func (Translate) Description() string {
	return "Translate text using the LLM. params: {\"text\": string, \"to\": string, \"from\": string?}"
}

func (Translate) IsAsync() bool {
	return false
}

func (Translate) GetEvents() []core.EventSpec {
	return []core.EventSpec{{Name: "translate.result", Description: "Emitted with the translation when translate finishes."}}
}

func (t Translate) Run(ctx context.Context, params map[string]any) (string, error) {
	if t.llm == nil {
		return "", fmt.Errorf("translate: no LLM available")
	}

	textInput, ok := core.ParamString(params, "text")

	if !ok {
		return "", fmt.Errorf("translate: missing string \"text\" parameter")
	}

	to, ok := core.ParamString(params, "to")

	if !ok {
		return "", fmt.Errorf("translate: missing string \"to\" parameter")
	}

	from := ""

	if f, ok := core.ParamString(params, "from"); ok && f != "" {
		from = " from " + f
	}

	prompt := fmt.Sprintf("Translate the following text%s to %s. Output only the translation.\n\n%s", from, to, textInput)

	return t.llm.Prompt(ctx, prompt)
}
