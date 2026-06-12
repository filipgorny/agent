package agent

// BaseMessage carries the fields common to every protocol message. It is
// embedded into concrete message types; field promotion makes a plain
// json.Marshal emit a flat object with "msg_type". Extend it over time with
// further shared fields (timestamp, id, ...).
type BaseMessage struct {
	MsgType string `json:"msg_type"`
}

// Type returns the message's protocol type.
func (b BaseMessage) Type() string {
	return b.MsgType
}

// InputMessage is a message the agent sends TO the LLM (agent → LLM). The
// unexported marker keeps the set of input messages sealed to this package.
type InputMessage interface {
	Type() string
	inputMessage()
}

// OutputMessage is a message the LLM returns to the agent (LLM → agent).
type OutputMessage interface {
	Type() string
	outputMessage()
}

// --- Input messages (agent → LLM) ---

// UserInput is a message originating from the user.
type UserInput struct {
	BaseMessage
	Text string `json:"text"`
}

func (UserInput) inputMessage() {}

// NewUserInput builds a user_input message.
func NewUserInput(text string) UserInput {
	return UserInput{
		BaseMessage: BaseMessage{MsgType: "user_input"},
		Text:        text,
	}
}

// EventMessage wraps an EventBus event for the LLM protocol.
type EventMessage struct {
	BaseMessage
	EventType string         `json:"event_type"`
	Source    string         `json:"source"`
	Params    map[string]any `json:"params"`
}

func (EventMessage) inputMessage() {}

// NewEventMessage builds an event message.
func NewEventMessage(eventType, source string, params map[string]any) EventMessage {
	return EventMessage{
		BaseMessage: BaseMessage{MsgType: "event"},
		EventType:   eventType,
		Source:      source,
		Params:      params,
	}
}

// --- Output messages (LLM → agent) ---

// ActionCall is the LLM's decision: which action to run and with what params.
type ActionCall struct {
	BaseMessage
	Action string         `json:"action"`
	Params map[string]any `json:"params"`
}

func (ActionCall) outputMessage() {}
