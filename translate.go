package agent

import (
	"context"
	"fmt"

	llm "github.com/filipgorny/llm-provider"
)

// TranslateSkillName is the registered name of the translate skill.
const TranslateSkillName = "translate"

func init() {
	RegisterSkill(TranslateSkillName, func(d Deps) Skill {
		return Translate{llm: d.LLM}
	})
}

// Translate translates text using the agent's LLM. Params: text, to, optional from.
type Translate struct {
	llm *llm.LlmProvider
}

func (Translate) Name() string {
	return TranslateSkillName
}

func (t Translate) Run(ctx context.Context, params map[string]any) (string, error) {
	if t.llm == nil {
		return "", fmt.Errorf("translate: no LLM available")
	}

	text, ok := paramString(params, "text")

	if !ok {
		return "", fmt.Errorf("translate: missing string \"text\" parameter")
	}

	to, ok := paramString(params, "to")

	if !ok {
		return "", fmt.Errorf("translate: missing string \"to\" parameter")
	}

	from := ""

	if f, ok := paramString(params, "from"); ok && f != "" {
		from = " from " + f
	}

	prompt := fmt.Sprintf("Translate the following text%s to %s. Output only the translation.\n\n%s", from, to, text)

	return t.llm.Prompt(ctx, prompt)
}
