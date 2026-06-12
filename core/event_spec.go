package core

// EventSpec describes an event a skill or plugin can emit. The description is in
// English so the LLM understands what the event means.
type EventSpec struct {
	Name        string
	Description string
}
