package agent

import "sync"

// threads manages the agent's thread tree.
type threads struct {
	mu   sync.Mutex
	all  map[string]*Thread
	main string
}

func newThreads() *threads {
	return &threads{all: map[string]*Thread{}}
}

// ensureMain creates the main thread once and returns its id.
func (t *threads) ensureMain() string {
	t.mu.Lock()

	defer t.mu.Unlock()

	if t.main == "" {
		id := newUID()
		t.all[id] = &Thread{ID: id}
		t.main = id
	}

	return t.main
}

// spawn creates a side thread with the given parent and returns its id.
func (t *threads) spawn(parentID string) string {
	t.mu.Lock()

	defer t.mu.Unlock()

	id := newUID()
	t.all[id] = &Thread{ID: id, ParentID: parentID}

	return id
}

// get returns a thread by id.
func (t *threads) get(id string) (*Thread, bool) {
	t.mu.Lock()

	defer t.mu.Unlock()

	th, ok := t.all[id]

	return th, ok
}
