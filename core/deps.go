package core

import llm "github.com/filipgorny/llm-provider"

// Deps are the agent services a skill may need at construction: the LLM provider
// (e.g. summarize_text, translate, think) and Emit to publish events from the
// skill (e.g. file_watch).
type Deps struct {
	LLM  *llm.LlmProvider
	Emit func(Event)
}
