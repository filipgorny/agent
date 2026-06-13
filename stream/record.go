// Package stream defines the agent's outbound message protocol: the structured
// stream of what the agent does while working (answers, logs, questions, root
// changes). It is distinct from skill events (core.Event).
package stream

import "time"

// Record is one item on the agent's outbound stream.
type Record struct {
	Type      string    `json:"type"`
	Subtype   string    `json:"subtype,omitempty"`
	Payload   any       `json:"payload,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Record types.
const (
	TypeAnswerUser = "ANSWER_USER"
	TypeLog        = "LOG"
	TypeAskUser    = "ASK_USER"
	TypeChangeRoot = "CHANGE_ROOT_FOLDER"
	TypeSession    = "SESSION"
	TypeStatus     = "STATUS"
)

// STATUS subtypes mark the lifecycle of a reasoning turn. They let a UI show
// whether the agent is currently waiting on the LLM: a request is in flight from
// StatusLLMRequest until the matching StatusLLMResponse.
const (
	StatusInput       = "INPUT"        // user input received; a turn begins
	StatusLLMRequest  = "LLM_REQUEST"  // prompt sent to the LLM; awaiting a reply
	StatusLLMResponse = "LLM_RESPONSE" // the LLM replied
)

// Default LOG subtypes.
const (
	LogToolCall   = "TOOL_CALL"
	LogToolResult = "TOOL_RESULT"
	LogAction     = "ACTION"
	LogReasoning  = "REASONING"
	LogSkillEvent = "SKILL_EVENT"
	LogMemory     = "MEMORY"
	LogError      = "ERROR"
)

// SubtypeChoice marks an ASK_USER message that carries selectable choices.
const SubtypeChoice = "CHOICE"
