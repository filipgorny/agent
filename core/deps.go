package core

import (
	"context"

	"github.com/filipgorny/agent/stream"
	llm "github.com/filipgorny/llm-provider"
)

// Deps are the agent services a skill may need at construction: the LLM provider
// (e.g. summarize_text, translate, think), Emit to publish events from the skill
// (e.g. file_watch), and Ask to put a question to the user (nil = non-interactive).
type Deps struct {
	LLM  *llm.LlmProvider
	Emit func(Event)
	Ask  func(ctx context.Context, req stream.AskRequest) (string, error)
}
