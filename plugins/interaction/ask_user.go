package interaction

import (
	"context"
	"fmt"

	"github.com/filipgorny/agent/core"
	"github.com/filipgorny/agent/stream"
)

// AskUserSkillName is the registered name of the ask_user skill.
const AskUserSkillName = "ask_user"

// AskUser asks the user a free-text question and returns their answer. It is the
// LLM-facing interface; the actual prompt travels to the UI as an ASK_USER
// message via the injected ask callback.
type AskUser struct {
	ask func(ctx context.Context, req stream.AskRequest) (string, error)
}

func (AskUser) Name() string {
	return AskUserSkillName
}

func (AskUser) Description() string {
	return "Ask the user a free-text question and return their answer. params: {\"question\": string}"
}

func (AskUser) IsAsync() bool {
	return false
}

func (AskUser) GetEvents() []core.EventSpec {
	return []core.EventSpec{{Name: "ask_user.result", Description: "Emitted with the user's answer."}}
}

func (s AskUser) Run(ctx context.Context, params map[string]any) (string, error) {
	if s.ask == nil {
		return "", fmt.Errorf("ask_user: agent is non-interactive")
	}

	q, ok := core.ParamString(params, "question")

	if !ok {
		return "", fmt.Errorf("ask_user: missing string \"question\" parameter")
	}

	return s.ask(ctx, stream.AskRequest{Question: q})
}
