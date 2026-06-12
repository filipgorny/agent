package message

import "github.com/filipgorny/agent/core"

// EventMessage wraps a core.Event for the LLM protocol (agent → LLM).
type EventMessage struct {
	BaseMessage
	EventType string         `json:"event_type"`
	Source    string         `json:"source"`
	ActionUID string         `json:"action_uid,omitempty"`
	ThreadID  string         `json:"thread_id,omitempty"`
	Params    map[string]any `json:"params"`
}

func (EventMessage) inputMessage() {}

// NewEventMessage builds an event message from a core.Event.
func NewEventMessage(ev core.Event) EventMessage {
	return EventMessage{
		BaseMessage: BaseMessage{MsgType: "event"},
		EventType:   ev.Type,
		Source:      ev.Source,
		ActionUID:   ev.ActionUID,
		ThreadID:    ev.ThreadID,
		Params:      ev.Data,
	}
}
