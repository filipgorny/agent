package runtime

import "sync"

// ResultStore keeps full skill/action results out of the LLM context but
// retrievable by id (via the get_result action), so large outputs do not blow a
// limited context window while remaining losslessly accessible.
type ResultStore struct {
	mu sync.Mutex
	m  map[string]string
}

func NewResultStore() *ResultStore {
	return &ResultStore{m: map[string]string{}}
}

// Put stores content and returns its id.
func (s *ResultStore) Put(content string) string {
	id := "r_" + NewUID()

	s.mu.Lock()

	defer s.mu.Unlock()

	s.m[id] = content

	return id
}

// Get returns the stored content for id.
func (s *ResultStore) Get(id string) (string, bool) {
	s.mu.Lock()

	defer s.mu.Unlock()

	v, ok := s.m[id]

	return v, ok
}
