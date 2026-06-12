package reasoning

import (
	"context"
	"fmt"

	"github.com/filipgorny/agent/core"
	llm "github.com/filipgorny/llm-provider"
)

// ThinkSkillName is the registered name of the think skill.
const ThinkSkillName = "think"

// Think is an asynchronous reasoning skill: it runs an LLM reasoning step in a
// side thread. Because it is async, the agent dispatches it and continues; the
// conclusion arrives later as a think.result event. Params: topic (or prompt).
type Think struct {
	llm *llm.LlmProvider
}

func (Think) Name() string {
	return ThinkSkillName
}

func (Think) Description() string {
	return "Reason step-by-step about a topic using the LLM. Runs asynchronously; the conclusion arrives as a think.result event. params: {\"topic\": string}"
}

func (Think) IsAsync() bool {
	return true
}

func (Think) GetEvents() []core.EventSpec {
	return []core.EventSpec{{Name: "think.result", Description: "Emitted with the reasoning conclusion when think finishes."}}
}

func (t Think) Run(ctx context.Context, params map[string]any) (string, error) {
	if t.llm == nil {
		return "", fmt.Errorf("think: no LLM available")
	}

	topic, ok := core.ParamString(params, "topic")

	if !ok {
		topic, ok = core.ParamString(params, "prompt")

		if !ok {
			return "", fmt.Errorf("think: missing string \"topic\" parameter")
		}
	}

	prompt := fmt.Sprintf("Reason step-by-step about the following and state a clear conclusion.\n\n%s", topic)

	return t.llm.Prompt(ctx, prompt)
}
