package message

// ActionCall is the LLM's decision: which action to run and with what params.
// (LLM → agent)
type ActionCall struct {
	BaseMessage
	Action string         `json:"action"`
	Params map[string]any `json:"params"`
}

func (ActionCall) outputMessage() {}
