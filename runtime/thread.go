package runtime

// Thread is an execution context. The initial run creates the main thread; async
// skills run in side threads whose ParentID points at the spawning thread.
type Thread struct {
	ID       string
	ParentID string // "" for the main thread
}
