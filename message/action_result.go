package message

// ActionResult feeds the result of an executed action back to the LLM so it can
// decide the next step (reasoning loop). (agent → LLM)
type ActionResult struct {
	BaseMessage
	Action string `json:"action"`
	Result string `json:"result"`
}

func (ActionResult) inputMessage() {}

// NewActionResult builds an action_result message.
func NewActionResult(action, result string) ActionResult {
	return ActionResult{
		BaseMessage: BaseMessage{MsgType: "action_result"},
		Action:      action,
		Result:      result,
	}
}
