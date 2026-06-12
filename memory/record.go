package memory

// Record is one stored memory returned from a search.
type Record struct {
	Content string
	Meta    map[string]any
	Score   float64
}
