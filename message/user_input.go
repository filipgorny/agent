package message

// UserInput is a message originating from the user (agent → LLM).
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
