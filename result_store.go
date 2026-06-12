package agent

import "sync"

// resultStore keeps full skill/action results out of the LLM context but
// retrievable by id (via the get_result action), so large outputs do not blow a
// limited context window while remaining losslessly accessible.
type resultStore struct {
	mu sync.Mutex
	m  map[string]string
}

func newResultStore() *resultStore {
	return &resultStore{m: map[string]string{}}
}

// put stores content and returns its id.
func (s *resultStore) put(content string) string {
	id := "r_" + newUID()

	s.mu.Lock()

	defer s.mu.Unlock()

	s.m[id] = content

	return id
}

// get returns the stored content for id.
func (s *resultStore) get(id string) (string, bool) {
	s.mu.Lock()

	defer s.mu.Unlock()

	v, ok := s.m[id]

	return v, ok
}
