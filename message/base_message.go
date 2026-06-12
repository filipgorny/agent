package message

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
