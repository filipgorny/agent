// Package stream defines the agent's outbound message protocol: the structured
// stream of what the agent does while working (answers, logs, questions, root
// changes). It is distinct from skill events (core.Event).
package stream

import "time"

// Message is one item on the agent's outbound stream.
type Message struct {
	Type      string    `json:"type"`
	Subtype   string    `json:"subtype,omitempty"`
	Payload   any       `json:"payload,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Message types.
const (
	TypeAnswerUser = "ANSWER_USER"
	TypeLog        = "LOG"
	TypeAskUser    = "ASK_USER"
	TypeChangeRoot = "CHANGE_ROOT_FOLDER"
	TypeSession    = "SESSION"
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
