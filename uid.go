package agent

import (
	"crypto/rand"
	"encoding/hex"
)

// newUID returns a short random hex id (no external dependency).
func newUID() string {
	var b [8]byte

	if _, err := rand.Read(b[:]); err != nil {
		// rand.Read never fails on supported platforms; fall back to a constant.
		return "0000000000000000"
	}

	return hex.EncodeToString(b[:])
}
