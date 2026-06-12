package interaction

import (
	"context"
	"fmt"

	"github.com/filipgorny/agent/core"
	"github.com/filipgorny/agent/stream"
)

// AskChoiceSkillName is the registered name of the ask_choice skill.
const AskChoiceSkillName = "ask_choice"

// AskChoice asks the user to pick one of several options and returns the choice.
type AskChoice struct {
	ask func(ctx context.Context, req stream.AskRequest) (string, error)
}

func (AskChoice) Name() string {
	return AskChoiceSkillName
}

func (AskChoice) Description() string {
	return "Ask the user to choose one of several options. params: {\"question\": string, \"choices\": [string]}"
}

func (AskChoice) IsAsync() bool {
	return false
}

func (AskChoice) GetEvents() []core.EventSpec {
	return []core.EventSpec{{Name: "ask_choice.result", Description: "Emitted with the option the user chose."}}
}

func (s AskChoice) Run(ctx context.Context, params map[string]any) (string, error) {
	if s.ask == nil {
		return "", fmt.Errorf("ask_choice: agent is non-interactive")
	}

	q, ok := core.ParamString(params, "question")

	if !ok {
		return "", fmt.Errorf("ask_choice: missing string \"question\" parameter")
	}

	choices := toStrings(params["choices"])

	if len(choices) == 0 {
		return "", fmt.Errorf("ask_choice: missing non-empty \"choices\" array")
	}

	return s.ask(ctx, stream.AskRequest{Question: q, Choices: choices})
}

// toStrings converts a JSON array parameter to []string.
func toStrings(v any) []string {
	arr, ok := v.([]any)

	if !ok {
		return nil
	}

	out := make([]string, 0, len(arr))

	for _, e := range arr {
		if s, ok := e.(string); ok {
			out = append(out, s)
		}
	}

	return out
}
