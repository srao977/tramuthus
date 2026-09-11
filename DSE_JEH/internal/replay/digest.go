// Package replay provides deterministic comparison material for proving and replay runs.
// Input is a completed run result; output is a SHA-256 digest of canonical Go JSON encoding.
// The digest has no configurable scientific parameters and does not alter source evidence.
// It is an audit comparison aid, not a strategy score, phase metric, or profitability measure.
package replay

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"tramuthus/dse-jeh/internal/evidence"
)

// Digest returns a reproducible digest of all admitted inputs, events, and final state.
func Digest(result evidence.RunResult) (string, error) {
	encoded, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("encode replay result: %w", err)
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:]), nil
}
