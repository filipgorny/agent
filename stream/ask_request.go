package stream

// AskRequest is a question the agent poses to the user. Empty Choices means a
// free-text answer; non-empty Choices means the UI should offer a selection.
type AskRequest struct {
	Question string
	Choices  []string
}
