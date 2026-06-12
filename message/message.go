// Package message defines the typed JSON protocol exchanged with the LLM.
package message

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
