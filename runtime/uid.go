// Package runtime holds the agent's internal execution primitives: threads,
// event listeners and the result store.
package runtime

import (
	"crypto/rand"
	"encoding/hex"
)

// NewUID returns a short random hex id (no external dependency).
func NewUID() string {
	var b [8]byte

	if _, err := rand.Read(b[:]); err != nil {
		// rand.Read never fails on supported platforms; fall back to a constant.
		return "0000000000000000"
	}

	return hex.EncodeToString(b[:])
}
