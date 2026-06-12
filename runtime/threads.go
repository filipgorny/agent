package runtime

import "sync"

// Threads manages the agent's thread tree.
type Threads struct {
	mu   sync.Mutex
	all  map[string]*Thread
	main string
}

func NewThreads() *Threads {
	return &Threads{all: map[string]*Thread{}}
}

// EnsureMain creates the main thread once and returns its id.
func (t *Threads) EnsureMain() string {
	t.mu.Lock()

	defer t.mu.Unlock()

	if t.main == "" {
		id := NewUID()
		t.all[id] = &Thread{ID: id}
		t.main = id
	}

	return t.main
}

// Spawn creates a side thread with the given parent and returns its id.
func (t *Threads) Spawn(parentID string) string {
	t.mu.Lock()

	defer t.mu.Unlock()

	id := NewUID()
	t.all[id] = &Thread{ID: id, ParentID: parentID}

	return id
}

// Get returns a thread by id.
func (t *Threads) Get(id string) (*Thread, bool) {
	t.mu.Lock()

	defer t.mu.Unlock()

	th, ok := t.all[id]

	return th, ok
}
